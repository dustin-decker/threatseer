package pipeline

import (
	"runtime"

	"github.com/dustin-decker/threatseer/server/engines/dynamic"
	"github.com/dustin-decker/threatseer/server/engines/shipper"
	"github.com/dustin-decker/threatseer/server/engines/static"
	"github.com/dustin-decker/threatseer/server/event"
)

// NewPipelineFlow wires up the engine pipeline network
func NewPipelineFlow(in chan event.Event) {

	se := static.NewStaticRulesEngine()
	de := dynamic.NewDynamicRulesEngine()
	bt := shipper.NewShipperEngine()

	numPipelines := runtime.NumCPU()

	for w := 0; w <= numPipelines; w++ {
		// add engines to the network
		go se.Run(in)

		go de.Run(se.Out)

		go bt.Start(de.Out)
	}
}
