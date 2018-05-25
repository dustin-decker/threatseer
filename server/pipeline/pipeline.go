package pipeline

import (
	"runtime"

	"github.com/dustin-decker/threatseer/server/engines/dynamic"
	"github.com/dustin-decker/threatseer/server/engines/profile"
	"github.com/dustin-decker/threatseer/server/engines/shipper"
	"github.com/dustin-decker/threatseer/server/engines/static"
	"github.com/dustin-decker/threatseer/server/event"
)

// NewPipelineFlow wires up the engine pipeline network
func NewPipelineFlow(numPipelines int, in chan event.Event) {

	staticRulesEngine := static.NewStaticRulesEngine()
	dynamicRulesEngine := dynamic.NewDynamicRulesEngine()
	profileEngine := profile.NewProfileEngine()
	shipperEngine := shipper.NewShipperEngine()

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
