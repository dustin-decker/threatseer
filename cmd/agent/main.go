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
		ProcessEvents:    false,
		NetworkEvents:    false,
		SyscallEvents:    false,
		KernelCallEvents: false,
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

	go srv.L3missDetector()

	go srv.Telemetry()

	go srv.Systemd()

	// keep running
	runtime.Goexit()

}
