package agent

import (
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
)

// Server tracks state
type Server struct {
	Hostname string
	MacAddr  string
	IP       string
	Signals  chan os.Signal
}

// NewAgentServer populates initial state
func NewAgentServer() *Server {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("couldn't get hostname: %v", err)
	}

	// Exit cleanly on Control-C
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)

	srv := Server{
		Hostname: hostname,
		Signals:  signals,
	}
	return &srv
}
