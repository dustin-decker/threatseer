package pipeline

import (
	"runtime"

	"github.com/elastic/beats/libbeat/beat"
	log "github.com/sirupsen/logrus"

	"github.com/dustin-decker/threatseer/server/engines/dynamic"
	"github.com/dustin-decker/threatseer/server/engines/profile"
	"github.com/dustin-decker/threatseer/server/engines/shipper"
	"github.com/dustin-decker/threatseer/server/engines/static"
	"github.com/dustin-decker/threatseer/server/event"
)

// NewPipelineFlow wires up the engine pipeline network
func NewPipelineFlow(b *beat.Beat, numPipelines int, in chan event.Event) {

	staticRulesEngine := static.NewStaticRulesEngine()
	log.WithFields(log.Fields{"engine": "static"}).Info("started engine")

	dynamicRulesEngine := dynamic.NewDynamicRulesEngine()
	log.WithFields(log.Fields{"engine": "dynamic"}).Info("started engine")

	profileEngine := profile.NewProfileEngine()
	log.WithFields(log.Fields{"engine": "profile"}).Info("started engine")

	shipperEngine := shipper.NewShipperEngine(b)
	log.WithFields(log.Fields{"engine": "shipper"}).Info("started engine")

	if numPipelines == 0 {
		numPipelines = runtime.NumCPU()
	}

	// start multiple pipelines in parallel
	for w := 0; w <= numPipelines; w++ {
		// add engines to the pipeline network
		// each one feeds the next through a channel
		go staticRulesEngine.AnalyzeFromPipeline(in)

		go dynamicRulesEngine.AnalyzeFromPipeline(staticRulesEngine.Out)

		go profileEngine.AnalyzeFromPipeline(dynamicRulesEngine.Out)

		// Final output without an output channel terminates the pipeline network
		go shipperEngine.PublishFromPipeline(profileEngine.Out)
	}
}
