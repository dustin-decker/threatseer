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
	"net"
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"

	api "github.com/capsule8/capsule8/api/v0"
	"github.com/dustin-decker/threatseer/server/event"
	"github.com/dustin-decker/threatseer/server/pipeline"

	"google.golang.org/grpc"
)

// Server state
type Server struct {
	Config Config
	Ctx    context.Context
}

// Start is the entrypoint for starting the TCP server and GRPC client
func Start() {
	flag.Parse()
	log.SetFormatter(&log.JSONFormatter{})

	ctx, cancel := context.WithCancel(context.Background())

	// cancel context with ctrl-c interrupt
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)

	go func() {
		<-signals
		cancel()
	}()

	config := LoadConfigFromFile()
	server := Server{Config: config, Ctx: ctx}

	log.Info("launching tcp server")

	// start tcp listener on all interfaces
	// note that each connection consumes a file descriptor
	// you may need to increase your fd limits if you have many concurrent clients
	ln, err := net.Listen("tcp", server.Config.ListenAddress)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("could not listen")
	}
	defer ln.Close()

	log.Info("starting engine pipeline")
	// create the network
	eventChan := make(chan event.Event)
	pipeline.NewPipelineFlow(server.Config.NumberOfPipelines, eventChan)

	log.Info("waiting for incoming TCP connections")

	go ProcessStats()

	for {
		// Accept blocks until there is an incoming TCP connection
		incomingConn, connErr := ln.Accept()
		log.Info("starting a gRPC client over incoming TCP connection")
		var conn *grpc.ClientConn
		// gRPC dial over incoming net.Conn
		conn, err := grpc.Dial(":7777",
			grpc.WithInsecure(),
			grpc.WithDialer(func(target string, timeout time.Duration) (net.Conn, error) {
				return incomingConn, connErr
			}),
		)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("could not connect")
		}

		// handle connection in goroutine so we can accept new TCP connections
		go server.handleConn(conn, eventChan, incomingConn.RemoteAddr().(*net.TCPAddr).IP.String())
	}
}

func (s *Server) handleConn(conn *grpc.ClientConn, eventChan chan event.Event, clientAddr string) {
	defer conn.Close()

	c := api.NewTelemetryServiceClient(conn)

	stream, err := c.GetEvents(s.Ctx, &api.GetEventsRequest{
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
