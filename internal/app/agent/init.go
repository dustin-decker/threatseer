package agent

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// Server tracks state
type Server struct {
	Hostname string
	MacAddr  string
	IP       string
}

// NewAgentServer populates initial state
func NewAgentServer() *Server {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Couldn't get hostname: %v", err)
	}

	srv := Server{
		Hostname: hostname,
	}
	return &srv
}
