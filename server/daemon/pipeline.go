package daemon

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

// newPipelineFlow wires up the engine pipeline network
func (s *Server) newPipelineFlow(b *beat.Beat, numPipelines uint) (eventChan chan event.Event) {

	eventChan = make(chan event.Event, 1000)
	go func() {
		<-s.stopPipeline
		close(eventChan)
	}()

	staticRulesEngine := static.NewStaticRulesEngine(s.pipelineCtx)
	log.WithFields(log.Fields{"engine": "static"}).Info("started engine")

	dynamicRulesEngine := dynamic.NewDynamicRulesEngine(s.pipelineCtx)
	log.WithFields(log.Fields{"engine": "dynamic"}).Info("started engine")

	profileEngine := profile.NewProfileEngine(s.pipelineCtx, s.Config)
	log.WithFields(log.Fields{"engine": "profile"}).Info("started engine")

	shipperEngine := shipper.NewShipperEngine(b)
	log.WithFields(log.Fields{"engine": "shipper"}).Info("started engine")

	if numPipelines == 0 {
		numPipelines = uint(runtime.NumCPU())
	}

	// start multiple pipelines in parallel
	var w uint
	for w = 0; w <= numPipelines; w++ {
		// add engines to the pipeline network
		// each one feeds the next through a channel
		go staticRulesEngine.AnalyzeFromPipeline(eventChan)

		go dynamicRulesEngine.AnalyzeFromPipeline(staticRulesEngine.Out)

		go profileEngine.AnalyzeFromPipeline(dynamicRulesEngine.Out)

		// Final output without an output channel terminates the pipeline network
		go shipperEngine.PublishFromPipeline(profileEngine.Out)
	}

	return
}
