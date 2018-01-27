package main

import (
	"flag"

	"github.com/dustin-decker/threatseer/internal/app/agent"
)

func main() {
	flag.Parse()

	srv := agent.NewAgentServer()

	go srv.L3missDetector()

	// keep running
	for {
	}
}
