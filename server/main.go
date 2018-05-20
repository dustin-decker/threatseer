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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"time"

	api "github.com/capsule8/capsule8/api/v0"
	"github.com/capsule8/capsule8/pkg/expression"
	"github.com/dustin-decker/threatseer/server/event"
	"github.com/dustin-decker/threatseer/server/pipeline"
	"github.com/golang/protobuf/ptypes/wrappers"
	flow "github.com/trustmaster/goflow"

	"google.golang.org/grpc"
)

// TCP server and GRPC client

func createSubscription() *api.Subscription {
	processEvents := []*api.ProcessEventFilter{
		//
		// Get all process lifecycle events
		//
		// &api.ProcessEventFilter{
		// 	Type: api.ProcessEventType_PROCESS_EVENT_TYPE_FORK,
		// },
		&api.ProcessEventFilter{
			Type: api.ProcessEventType_PROCESS_EVENT_TYPE_EXEC,
		},
		// &api.ProcessEventFilter{
		// 	Type: api.ProcessEventType_PROCESS_EVENT_TYPE_EXIT,
		// },
	}

	syscallEvents := []*api.SyscallEventFilter{
		// Get all open(2) syscalls that return an error
		// &api.SyscallEventFilter{
		// 	Type: api.SyscallEventType_SYSCALL_EVENT_TYPE_EXIT,

		// 	Id: &wrappers.Int64Value{
		// 		Value: 2, // SYS_OPEN
		// 	},
		// },
	}

	fileEvents := []*api.FileEventFilter{
		//
		// Get all attempts to open files matching glob *foo*
		//
		&api.FileEventFilter{
			Type: api.FileEventType_FILE_EVENT_TYPE_OPEN,

			//
			// The glob accepts a wild card character
			// (*,?) and character classes ([).
			//
			FilenamePattern: &wrappers.StringValue{
				Value: "*foo*",
			},
		},
	}

	sinFamilyFilter := expression.Equal(
		expression.Identifier("sin_family"),
		expression.Value(uint16(2)))
	kernelCallEvents := []*api.KernelFunctionCallFilter{
		//
		// Install a kprobe on connect(2)
		//
		&api.KernelFunctionCallFilter{
			Type:   api.KernelFunctionCallEventType_KERNEL_FUNCTION_CALL_EVENT_TYPE_ENTER,
			Symbol: "SyS_connect",
			Arguments: map[string]string{
				"sin_family": "+0(%si):u16",
				"sin_port":   "+2(%si):u16",
				"sin_addr":   "+4(%si):u32",
			},
			FilterExpression: sinFamilyFilter,
		},
	}

	containerEvents := []*api.ContainerEventFilter{
		//
		// Get all container lifecycle events
		//
		&api.ContainerEventFilter{
			Type: api.ContainerEventType_CONTAINER_EVENT_TYPE_CREATED,
		},
		&api.ContainerEventFilter{
			Type: api.ContainerEventType_CONTAINER_EVENT_TYPE_RUNNING,
		},
		&api.ContainerEventFilter{
			Type: api.ContainerEventType_CONTAINER_EVENT_TYPE_EXITED,
		},
		&api.ContainerEventFilter{
			Type: api.ContainerEventType_CONTAINER_EVENT_TYPE_DESTROYED,
		},
	}

	// Ticker events are used for debugging and performance testing
	tickerEvents := []*api.TickerEventFilter{
		// &api.TickerEventFilter{
		// 	Interval: int64(1 * time.Second),
		// },
	}

	chargenEvents := []*api.ChargenEventFilter{
		/*
			&api.ChargenEventFilter{
				Length: 16,
			},
		*/
	}

	eventFilter := &api.EventFilter{
		ProcessEvents:   processEvents,
		SyscallEvents:   syscallEvents,
		KernelEvents:    kernelCallEvents,
		FileEvents:      fileEvents,
		ContainerEvents: containerEvents,
		TickerEvents:    tickerEvents,
		ChargenEvents:   chargenEvents,
	}

	sub := &api.Subscription{
		EventFilter: eventFilter,
	}

	return sub
}

func main() {
	flag.Parse()

	log.Println("launching tcp server...")

	// start tcp listener on all interfaces
	// note that each connection consumes a file descriptor
	// you may need to increase your fd limits if you have many concurrent clients
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("could not listen: %s", err)
	}
	defer ln.Close()

	log.Println("starting engine pipeline...")
	// create the network
	n := pipeline.NewPipelineFlow()
	// we need a channel to talk to it
	eventChan := make(chan event.Event)
	n.SetInPort("In", eventChan)
	// run the pipeline network
	flow.RunNet(n)
	// close the input to shut the network down
	defer close(eventChan)
	// // wait until the app has done its job
	// <-net.Wait()

	log.Println("waiting for incoming TCP connections...")

	go func() {
		for {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			log.Print(map[string]string{
				"alloc":              fmt.Sprintf("%v", m.Alloc),
				"total-alloc":        fmt.Sprintf("%v", m.TotalAlloc/1024),
				"sys":                fmt.Sprintf("%v", m.Sys/1024),
				"num-gc":             fmt.Sprintf("%v", m.NumGC),
				"goroutines":         fmt.Sprintf("%v", runtime.NumGoroutine()),
				"stop-pause-nanosec": fmt.Sprintf("%v", m.PauseTotalNs),
			})
			time.Sleep(10 * time.Second)
		}
	}()

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
		go handleConn(conn, eventChan, incomingConn.RemoteAddr())
	}
}

func handleConn(conn *grpc.ClientConn, eventChan chan event.Event, clientAddr net.Addr) {
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
				Score:      map[string]int{},
				ClientAddr: clientAddr,
			}
		}
	}
}
