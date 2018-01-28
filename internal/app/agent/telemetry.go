package agent

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"

	api "github.com/capsule8/capsule8/api/v0"
	"github.com/capsule8/capsule8/pkg/expression"
	"github.com/golang/protobuf/ptypes/wrappers"
)

var config struct {
	server string
	image  string
}

func init() {
	flag.StringVar(&config.server, "server",
		"unix:/var/run/capsule8/sensor.sock",
		"Capsule8 gRPC API server address")

	flag.StringVar(&config.image, "image", "",
		"container image wildcard pattern to monitor")
}

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

func createSubscription() *api.Subscription {
	processEvents := []*api.ProcessEventFilter{
		//
		// Get all process lifecycle events
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
	// // Get all open(2) syscalls that return an error
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
		&api.TickerEventFilter{
			Interval: int64(1 * time.Second),
		},
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

	if config.image != "" {
		fmt.Fprintf(os.Stderr,
			"Watching for container images matching %s\n",
			config.image)

		containerFilter := &api.ContainerFilter{}

		containerFilter.ImageNames =
			append(containerFilter.ImageNames, config.image)

		sub.ContainerFilter = containerFilter
	}

	return sub
}

// Telemetry collects telemetry from Capsule8 API
func (srv *Server) Telemetry() {
	log.Info("starting telemetry")

	// Create telemetry service client
	conn, err := grpc.Dial(config.server,
		grpc.WithDialer(dialer),
		grpc.WithInsecure())

	c := api.NewTelemetryServiceClient(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "grpc.Dial: %s\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	stream, err := c.GetEvents(ctx, &api.GetEventsRequest{
		Subscription: createSubscription(),
	})

	go func() {
		<-srv.Signals
		cancel()
	}()

	if err != nil {
		fmt.Fprintf(os.Stderr, "GetEvents: %s\n", err)
		os.Exit(1)
	}

	log.Info("monitoring telemetry")
	for {
		ev, err := stream.Recv()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Recv: %s\n", err)
			os.Exit(1)
		}

		for _, e := range ev.Events {
			// e.GetEvent()
			// log.Info(e.String())
			log.WithFields(log.Fields{
				"event": e.GetEvent(),
			}).Info()
			// log.WithFields(log.Fields{
			// 	"PID":           e.Event.GetProcessPid(),
			// 	"program":       e.Event.GetProcessLineage(),
			// 	"containerName": e.Event.GetContainerName(),
			// 	"process":       e.Event.GetProcess().,
			// }).Info()
			// log.Info(e)
		}
	}
}
