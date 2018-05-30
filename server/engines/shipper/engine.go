package shipper

import (
	"time"

	"github.com/dustin-decker/threatseer/server/event"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	log "github.com/sirupsen/logrus"
)

// Shipper makes it compatible flow pipeline
type Shipper struct {
	b      *beat.Beat
	client beat.Client
}

// PublishFromPipeline is the entrypoint from the flow pipeline
func (s *Shipper) PublishFromPipeline(in chan event.Event) {
	for e := range in {
		var riskScore int
		for _, indicator := range e.Indicators {
			riskScore = riskScore + indicator.Score
		}
		evnt := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"event":      e.Event,
				"indicators": e.Indicators,
				"risk_score": riskScore,
				"src_ip":     e.ClientAddr,
			},
		}

		// goes to output
		s.client.Publish(evnt)
	}
}

// NewShipperEngine is the entrypoint for the datashipper
func NewShipperEngine(b *beat.Beat) Shipper {
	client, err := b.Publisher.Connect()
	if err != nil {
		log.WithFields(log.Fields{"engine": "shipper", "err": err}).Fatal("could not connect to shipper publisher")
	}

	return Shipper{
		client: client,
	}
}
