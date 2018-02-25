package main

import (
	"flag"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dustin-decker/threatseer/internal/app/agent"
)

func main() {
	flag.Set("alsologtostderr", "true")
	flag.Parse()
	log.SetFormatter(&log.JSONFormatter{})

	srv := agent.NewAgentServer()

	time.Sleep(5 * time.Second)

	go srv.L3missDetector()

	go srv.Telemetry()

	// keep running
	runtime.Goexit()

}
