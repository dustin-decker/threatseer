package agent

import (
	"os"
	"os/signal"

	"github.com/capsule8/capsule8/pkg/config"
	"github.com/capsule8/capsule8/pkg/sensor"
	"github.com/capsule8/capsule8/pkg/services"
	"github.com/dustin-decker/threatseer/configs"
	log "github.com/sirupsen/logrus"
)

// Server tracks state
type Server struct {
	Hostname string
	MacAddr  string
	IP       string
	Signals  chan os.Signal
	Sensor   *sensor.Sensor
	Config   configs.Config
}

// NewAgentServer populates initial state
func NewAgentServer(conf configs.Config) *Server {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("couldn't get hostname: ", err)
	}

	// increase tracing buffer size
	// ref https://lwn.net/Articles/322731/
	tracer, err := os.OpenFile("/sys/kernel/debug/tracing/current_tracer", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err == nil {
		_, err = tracer.WriteString("nop")
		if err != nil {
			log.Fatal(err)
		}
	}
	// unit is KB, and is per CPU core
	bufSize, err := os.OpenFile("/sys/kernel/debug/tracing/buffer_size_kb", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err == nil {
		_, err = bufSize.WriteString("5120")
		if err != nil {
			log.Fatal(err)
		}
	}

	// Exit cleanly on Control-C
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)

	manager := services.NewServiceManager()
	if len(config.Global.ProfilingListenAddr) > 0 {
		service := services.NewProfilingService(
			config.Global.ProfilingListenAddr)
		manager.RegisterService(service)
	}

	s, err := sensor.NewSensor()
	if err != nil {
		log.Fatal("could not create sensor: ", err.Error())
	}
	if err := s.Start(); err != nil {
		log.Fatal("could not start sensor: ", err.Error())
	}
	service := sensor.NewTelemetryService(s, config.Sensor.ListenAddr)
	manager.RegisterService(service)

	go manager.Run()

	srv := Server{
		Hostname: hostname,
		Signals:  signals,
		Sensor:   s,
		Config:   conf,
	}
	return &srv
}
