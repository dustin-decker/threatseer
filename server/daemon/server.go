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
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	api "github.com/capsule8/capsule8/api/v0"
	"github.com/dustin-decker/threatseer/server/event"
	"github.com/dustin-decker/threatseer/server/pipeline"

	"google.golang.org/grpc"
)

// Server state
type Server struct {
	Config Config
}

// Start is the entrypoint for starting the TCP server and GRPC client
func Start() {
	flag.Parse()

	config := LoadConfigFromFile()
	server := Server{Config: config}

	log.Println("launching tcp server")

	// start tcp listener on all interfaces
	// note that each connection consumes a file descriptor
	// you may need to increase your fd limits if you have many concurrent clients
	ln, err := net.Listen("tcp", server.Config.ListenAddress)
	if err != nil {
		log.Fatalf("could not listen: %s", err)
	}
	defer ln.Close()

	log.Println("starting engine pipeline")
	// create the network
	eventChan := make(chan event.Event)
	pipeline.NewPipelineFlow(server.Config.NumberOfPipelines, eventChan)

	log.Println("waiting for incoming TCP connections")

	go ProcessStats()

	for {
		// Accept blocks until there is an incoming TCP connection
		incomingConn, connErr := ln.Accept()
		log.Println("starting a gRPC client over incoming TCP connection")
		var conn *grpc.ClientConn
		// gRPC dial over incoming net.Conn
		conn, err := grpc.Dial(":7777",
			grpc.WithInsecure(),
			grpc.WithDialer(func(target string, timeout time.Duration) (net.Conn, error) {
				return incomingConn, connErr
			}),
		)
		if err != nil {
			log.Fatalf("could not connect: %s", err)
		}

		// handle connection in goroutine so we can accept new TCP connections
		go server.handleConn(conn, eventChan, incomingConn.RemoteAddr().String())
	}
}

func (s *Server) handleConn(conn *grpc.ClientConn, eventChan chan event.Event, clientAddr string) {
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// cancel context with ctrl-c interrupt
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)

	go func() {
		<-signals
		cancel()
	}()

	c := api.NewTelemetryServiceClient(conn)

	stream, err := c.GetEvents(ctx, &api.GetEventsRequest{
		Subscription: createSubscription(),
	})

	if err != nil {
		log.Println("error subscribing to events: ", err)
		return
	}

	for {
		ev, err := stream.Recv()
		if err != nil {
			log.Println("error receiving events: ", err)
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
