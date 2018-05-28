package daemon

import (
	api "github.com/capsule8/capsule8/api/v0"

	"github.com/golang/protobuf/ptypes/wrappers"
)

// this is the capsule8 sensor telemetry subscription
func createSubscription() *api.Subscription {
	processEvents := []*api.ProcessEventFilter{
		&api.ProcessEventFilter{
			Type: api.ProcessEventType_PROCESS_EVENT_TYPE_EXEC,
		},
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

	// sinFamilyFilter := expression.Equal(
	// 	expression.Identifier("sin_family"),
	// 	expression.Value(uint16(2)))
	kernelCallEvents := []*api.KernelFunctionCallFilter{
		//
		// Install a kprobe on connect(2)
		//
		// &api.KernelFunctionCallFilter{
		// 	Type:   api.KernelFunctionCallEventType_KERNEL_FUNCTION_CALL_EVENT_TYPE_ENTER,
		// 	Symbol: "SyS_connect",
		// 	Arguments: map[string]string{
		// 		"sin_family": "+0(%si):u16",
		// 		"sin_port":   "+2(%si):u16",
		// 		"sin_addr":   "+4(%si):u32",
		// 	},
		// 	FilterExpression: sinFamilyFilter,
		// },
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
