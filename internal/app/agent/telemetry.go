package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"

	api "github.com/capsule8/capsule8/api/v0"
	"github.com/capsule8/capsule8/pkg/expression"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/wrappers"
)

// Custom gRPC Dialer that understands "unix:/path/to/sock" as well as TCP addrs
func dialer(addr string, timeout time.Duration) (net.Conn, error) {
	var network, address string

	parts := strings.Split(addr, ":")
	if len(parts) > 1 && parts[0] == "unix" {
		network = "unix"
		address = parts[1]
	} else {
		network = "tcp"
		address = addr
	}

	return net.DialTimeout(network, address, timeout)
}

func createSubscription(srv *Server) *api.Subscription {
	processEvents := []*api.ProcessEventFilter{
		//
		// Get process lifecycle events
		//
		&api.ProcessEventFilter{
			Type: api.ProcessEventType_PROCESS_EVENT_TYPE_FORK,
		},
		&api.ProcessEventFilter{
			Type: api.ProcessEventType_PROCESS_EVENT_TYPE_EXEC,
		},
		&api.ProcessEventFilter{
			Type: api.ProcessEventType_PROCESS_EVENT_TYPE_EXIT,
		},
	}

	syscallEvents := []*api.SyscallEventFilter{
		// Get all open(2) syscalls that return an error
		&api.SyscallEventFilter{
			Type: api.SyscallEventType_SYSCALL_EVENT_TYPE_EXIT,

			Id: &wrappers.Int64Value{
				Value: 2, // SYS_OPEN
			},
		},
	}

	fileEvents := []*api.FileEventFilter{
		//
		// Get all attempts to open files matching pattern
		//

		// user password hashes
		&api.FileEventFilter{
			Type: api.FileEventType_FILE_EVENT_TYPE_OPEN,

			//
			// The glob accepts a wild card character
			// (*,?) and character classes ([).
			//
			FilenamePattern: &wrappers.StringValue{
				Value: "/etc/shadow",
			},
		},

		// mysql data dir
		&api.FileEventFilter{
			Type: api.FileEventType_FILE_EVENT_TYPE_OPEN,

			//
			// The glob accepts a wild card character
			// (*,?) and character classes ([).
			//
			FilenamePattern: &wrappers.StringValue{
				Value: "/var/lib/mysql/*",
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
		// Get container lifecycle events
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
		&api.ContainerEventFilter{
			Type: api.ContainerEventType_CONTAINER_EVENT_TYPE_UPDATED,
		},
	}

	networkEvents := []*api.NetworkEventFilter{
		// get interesting network events
		&api.NetworkEventFilter{
			Type: api.NetworkEventType_NETWORK_EVENT_TYPE_LISTEN_RESULT,
		},
		&api.NetworkEventFilter{
			Type: api.NetworkEventType_NETWORK_EVENT_TYPE_BIND_RESULT,
		},
		&api.NetworkEventFilter{
			Type: api.NetworkEventType_NETWORK_EVENT_TYPE_LISTEN_ATTEMPT,
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
		TickerEvents:  tickerEvents,
		ChargenEvents: chargenEvents,
	}

	if srv.Config.ProcessEvents {
		eventFilter.ProcessEvents = processEvents
	}

	if srv.Config.SyscallEvents {
		eventFilter.SyscallEvents = syscallEvents
	}

	if srv.Config.KernelCallEvents {
		eventFilter.KernelEvents = kernelCallEvents
	}

	if srv.Config.FileEvents {
		eventFilter.FileEvents = fileEvents
	}

	if srv.Config.ContainerEvents {
		eventFilter.ContainerEvents = containerEvents
	}

	if srv.Config.NetworkEvents {
		eventFilter.NetworkEvents = networkEvents
	}

	sub := &api.Subscription{
		EventFilter: eventFilter,
	}

	return sub
}

// Telemetry collects telemetry from Capsule8 API
func (srv *Server) Telemetry() {
	log.Info("starting telemetry")

	ctx, cancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)

	go func() {
		<-signals
		cancel()
	}()

	// Create telemetry service client
	conn, err := grpc.DialContext(ctx, "unix:/var/run/capsule8/sensor.sock",
		grpc.WithDialer(dialer),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(1*time.Second))
	if err != nil {
		log.Fatal("could not start telemetry service client: ", err)
	}

	c := api.NewTelemetryServiceClient(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "grpc.Dial: %s\n", err)
		os.Exit(1)
	}

	stream, err := c.GetEvents(ctx, &api.GetEventsRequest{
		Subscription: createSubscription(srv),
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating event stream: %s\n", err)
		os.Exit(1)
	}

	marshaler := &jsonpb.Marshaler{EmitDefaults: true}

	log.Info("monitoring telemetry")
	for {
		ev, err := stream.Recv()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error recieving event: %s\n", err)
			os.Exit(1)
		}

		for _, e := range ev.Events {
			var b bytes.Buffer
			err := marshaler.Marshal(&b, e)
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"unable to decode event: %v", err)
				continue
			}

			var evnt map[string]interface{}
			json.Unmarshal(b.Bytes(), &evnt)
			log.WithFields(evnt).Info()
		}
	}
}
