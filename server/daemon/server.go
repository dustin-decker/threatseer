// Copyright 2018 Dustin Decker

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package daemon

import (
	"context"
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	log "github.com/sirupsen/logrus"

	api "github.com/capsule8/capsule8/api/v0"
	"github.com/dustin-decker/threatseer/server/config"
	"github.com/dustin-decker/threatseer/server/event"

	"google.golang.org/grpc"
)

// Server state
type Server struct {
	// threatsser stuff
	Beat         *beat.Beat
	done         chan struct{}
	Config       config.Config
	stopPipeline chan struct{}

	// contexts for clean shut down
	grpcCtx           context.Context
	grpcCtxCancel     context.CancelFunc
	pipelineCtx       context.Context
	pipelineCtxCancel context.CancelFunc
}

// New creates new Server object
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	grpcCtx, grpcCtxCancel := context.WithCancel(context.Background())
	pipelineCtx, pipelineCtxCancel := context.WithCancel(context.Background())

	bt := &Server{
		done:              make(chan struct{}),
		stopPipeline:      make(chan struct{}),
		Config:            config,
		grpcCtx:           grpcCtx,
		grpcCtxCancel:     grpcCtxCancel,
		pipelineCtx:       pipelineCtx,
		pipelineCtxCancel: pipelineCtxCancel,
	}
	return bt, nil
}

// Stop cleanly shuts down threatseer
func (s *Server) Stop() {
	// cancel current agent connections
	s.grpcCtxCancel()
	// stop the flow in the pipeline
	s.stopPipeline <- struct{}{}
	// shut down the pipeline
	s.pipelineCtxCancel()
}

// Run is the entrypoint for starting the TCP server and GRPC client
// The main event loop that should block until signalled to stop by an
// invocation of the Stop() method.
func (s *Server) Run(b *beat.Beat) error {
	flag.Parse()
	log.SetFormatter(&log.JSONFormatter{})
	logp.Info("launching tcp server")

	// start tcp listener on all interfaces
	// note that each connection consumes a file descriptor
	// you may need to increase your fd limits if you have many concurrent clients
	ln, err := net.Listen("tcp", s.Config.ListenAddress)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("could not listen")
		return err
	}
	defer ln.Close()
	log.WithFields(log.Fields{"listen_address": s.Config.ListenAddress}).Info("threatseer server listening for connections")

	logp.Info("starting engine pipeline")
	eventChan := s.newPipelineFlow(b, s.Config.NumberOfPipelines)

	log.Info("waiting for incoming TCP connections")

	go ProcessStats()

	for {
		// Accept blocks until there is an incoming TCP connection
		incomingConn, connErr := ln.Accept()
		clientAddr := incomingConn.RemoteAddr().(*net.TCPAddr).IP.String()

		log.WithFields(log.Fields{"client_addr": clientAddr}).Info("connecting to gRPC sensor over incoming TCP connection")
		var conn *grpc.ClientConn
		// gRPC dial over incoming net.Conn
		conn, err := grpc.DialContext(s.grpcCtx, ":7777",
			grpc.WithInsecure(),
			grpc.WithDialer(func(target string, timeout time.Duration) (net.Conn, error) {
				return incomingConn, connErr
			}),
		)
		if err != nil {
			log.WithFields(log.Fields{"err": err, "client_addr": clientAddr}).Error("could not connect to sensor")
		}

		// handle connection in goroutine so we can accept new TCP connections
		go s.handleConn(conn, eventChan, clientAddr)
	}
}

func (s *Server) handleConn(conn *grpc.ClientConn, eventChan chan event.Event, clientAddr string) {
	defer conn.Close()

	c := api.NewTelemetryServiceClient(conn)

	stream, err := c.GetEvents(s.grpcCtx, &api.GetEventsRequest{
		Subscription: createSubscription(),
	})

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("error subscribing to events")
		return
	}

	for {
		ev, err := stream.Recv()
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("error receiving events")
			return
		}

		for _, e := range ev.Events {
			// send the event down the pipeline
			eventChan <- event.Event{
				Event:      e.GetEvent(),
				Indicators: make([]event.Indicator, 0),
				ClientAddr: clientAddr,
			}
		}
	}
}
