package shipper

import (
	"fmt"
	"log"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/cmd/instance"
	"github.com/elastic/beats/libbeat/common"

	flow "github.com/trustmaster/goflow"

	api "github.com/capsule8/capsule8/api/v0"
)

type Shipper struct {
	flow.Component
	ThreatseerBeat *ThreatseerBeat
}

// ThreatseerBeat tracks stuff needed for event shipping
type ThreatseerBeat struct {
	done   chan struct{}
	config Config
	client beat.Client
	In     <-chan event.Event

func Start() *Shipper {
	bt, err := instance.NewBeat("threatseer", "", "")
	if err != nil {
		log.Fatal("could not start shipper beat, got: ", err)
	}
	if err := bt.Init(); err != nil {
		return err
	}
	threatseerBeat, err := NewShipper(&bt, DefaultConfig)
	if err != nil {
		log.Fatal("could not start shipper, got: ", err)
	}

	return &Shipper{ThreatseerBeat: threatseerBeat}
}

// NewShipper creates beater
func NewShipper(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &ThreatseerBeat{
		done:   make(chan struct{}),
		config: config,
	}
	return bt, nil
}

// Run starts the beater daemon
func (bt *ThreatseerBeat) Run(b *beat.Beat) error {

	client, err := b.Publisher.Connect()
	if err != nil {
		log.Fatal("error connecting to shipper output, got: ", err)
	}

	bt.client = client
	ticker := time.NewTicker(bt.config.Interval)
	for {
		select {
		case <-bt.done:
			log.Print("recieved done signal, shutting down event shipper")
			return nil
		case <-ticker.C:
		}

		e <- bt.In

		// goes to output
		client.Publish(e)
		log.Print("api", "event sent: %v", e)
	}
}

// Stop gets called when libbeat gets a SIGTERM. It sends a message in a channel to
// stop the worker.
func (bt *ThreatseerBeat) Stop() {
	err := bt.client.Close()
	if err != nil {
		log.Print("stopping the beat client failed because of: ", err)
	}
	close(bt.done)
}
