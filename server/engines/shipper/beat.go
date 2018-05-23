package shipper

import (
	"fmt"
	"log"
	"time"

	"github.com/dustin-decker/threatseer/server/event"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/cmd/instance"
	"github.com/elastic/beats/libbeat/common"
)

// Shipper makes it compatible flow pipeline
type Shipper struct {
	done   chan struct{}
	config Config
	client beat.Client
}

// Start is the entrypoint from the flow pipeline
func (s *Shipper) Start(in chan event.Event) {
	for {
		// incoming event from the pipeline
		e := <-in

		evnt := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"event":      e.Event,
				"indicators": e.Indicators,
				"src_ip":     e.ClientAddr,
			},
		}

		// goes to output
		s.client.Publish(evnt)
	}
}

// NewShipperEngine is the entrypoint for the datashipper
func NewShipperEngine() Shipper {
	bt, err := instance.NewBeat("threatseer", "", "")
	if err != nil {
		log.Fatal("could not instantiate beat, got: ", err)
	}

	err = bt.Setup(newShipper, false, false, false)
	if err != nil {
		log.Fatal("error setting up the shipper, got: ", err)
	}

	client, err := bt.Publisher.Connect()
	if err != nil {
		log.Fatal("error connecting to shipper output, got: ", err)
	}

	return Shipper{
		done:   make(chan struct{}),
		config: DefaultConfig,
		client: client,
	}
}

// just here to satisfty instance.Beat.Setup
// you can load in custom configs here as usual
func newShipper(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Shipper{}
	return bt, nil
}

// Run starts the beater daemon
// This is only here to satisfy the interface
func (s *Shipper) Run(b *beat.Beat) error {
	ticker := time.NewTicker(s.config.Interval)
	for {
		select {
		case <-s.done:
			log.Print("recieved done signal, shutting down event shipper")
			return nil
		case <-ticker.C:
		}
	}
}

// Stop gets called when libbeat gets a SIGTERM. It sends a message in a channel to
// stop the worker.
func (s *Shipper) Stop() {
	err := s.client.Close()
	if err != nil {
		log.Print("stopping the beat client failed because of: ", err)
	}
	close(s.done)
}
