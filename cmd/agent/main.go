package main

import (
	"flag"
	"runtime"
	"time"

	"github.com/dustin-decker/threatseer/configs"
	"github.com/dustin-decker/threatseer/internal/app/agent"
	log "github.com/sirupsen/logrus"
)

func main() {
	config := configs.Config{
		ContainerEvents:  true,
		SystemdEvents:    true,
		CacheMissEvents:  true,
		ProcessEvents:    true,
		NetworkEvents:    true,
		SyscallEvents:    false,
		KernelCallEvents: true,
		FileEvents:       true,
		FilePatterns: []string{
			"/etc/shadow",
			"/var/lib/mysql/*",
		},
	}

	flag.Set("alsologtostderr", "true")
	flag.Parse()
	log.SetFormatter(&log.JSONFormatter{})

	srv := agent.NewAgentServer(config)

	time.Sleep(5 * time.Second)

	if srv.Config.CacheMissEvents {
		go srv.L3missDetector()
	}

	go srv.Telemetry()

	if srv.Config.SystemdEvents {
		go srv.Systemd()
	}

	// keep running
	runtime.Goexit()

}
