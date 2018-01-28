package main

import (
	"flag"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/capsule8/capsule8/pkg/sensor"
	"github.com/dustin-decker/threatseer/internal/app/agent"
)

func main() {
	flag.Set("alsologtostderr", "true")
	flag.Parse()
	log.SetFormatter(&log.JSONFormatter{})

	srv := agent.NewAgentServer()

	go sensor.Main()

	time.Sleep(10 * time.Second)

	go srv.L3missDetector()

	go srv.Telemetry()

	// keep running
	runtime.Goexit()

}
