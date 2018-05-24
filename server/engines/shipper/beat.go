package shipper

import (
	"time"

	log "github.com/sirupsen/logrus"

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

// PublishFromPipeline is the entrypoint from the flow pipeline
func (s *Shipper) PublishFromPipeline(in chan event.Event) {
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
		log.WithFields(log.Fields{"engine": "shipper", "err": err}).Fatal("could not instantiate shipper")
	}

	err = bt.Setup(newShipper, false, false, false)
	if err != nil {
		log.WithFields(log.Fields{"engine": "shipper", "err": err}).Fatal("error setting up shipper")
	}

	client, err := bt.Publisher.Connect()
	if err != nil {
		log.WithFields(log.Fields{"engine": "shipper", "err": err}).Fatal("could not connect to shipper publisher")
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
		log.WithFields(log.Fields{"engine": "shipper", "err": err}).Fatal("could not read config")
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
			log.WithFields(log.Fields{"engine": "shipper"}).Info("received done signal, shutting down")
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
		log.WithFields(log.Fields{"engine": "shipper", "err": err}).Error("stopping shipper failed")
	}
	close(s.done)
}
