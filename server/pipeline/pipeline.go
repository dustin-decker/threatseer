package pipeline

import (
	"runtime"

	"github.com/dustin-decker/threatseer/server/engines/dynamic"
	"github.com/dustin-decker/threatseer/server/engines/shipper"
	"github.com/dustin-decker/threatseer/server/engines/static"
	"github.com/dustin-decker/threatseer/server/event"
)

// NewPipelineFlow wires up the engine pipeline network
func NewPipelineFlow(numPipelines int, in chan event.Event) {

	se := static.NewStaticRulesEngine()
	de := dynamic.NewDynamicRulesEngine()
	bt := shipper.NewShipperEngine()

	if numPipelines == 0 {
		numPipelines = runtime.NumCPU()
	}

	// start multiple pipelines in parallel
	for w := 0; w <= numPipelines; w++ {
		// add engines to the pipeline network
		// each one feeds the next through a channel
		go se.AnalyzeFromPipeline(in)

		go de.AnalyzeFromPipeline(se.Out)

		// Final output without an output channel terminates the pipeline network
		go bt.PublishFromPipeline(de.Out)
	}
}
