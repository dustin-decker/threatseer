package pipeline

import (
	"runtime"

	"github.com/dustin-decker/threatseer/server/engines/dynamic"
	"github.com/dustin-decker/threatseer/server/engines/static"
	"github.com/dustin-decker/threatseer/server/shipper"
	flow "github.com/trustmaster/goflow"
)

type pipelineFlow struct {
	flow.Graph
}

func NewPipelineFlow() *pipelineFlow {
	n := new(pipelineFlow)
	n.InitGraphState()

	goroutinesPerEngine := uint8(runtime.NumCPU())

	// add engines to the network
	se := static.NewStaticRulesEngine()
	se.Component.Mode = flow.ComponentModePool
	se.Component.PoolSize = goroutinesPerEngine
	n.Add(&se, "StaticRulesEngine")
	de := new(dynamic.DynamicRulesEngine)
	de.Component.Mode = flow.ComponentModePool
	de.Component.PoolSize = goroutinesPerEngine
	n.Add(de, "DynamicRulesEngine")
	// connect them with a channel
	n.Connect("StaticRulesEngine", "Out", "DynamicRulesEngine", "In")

	bt := shipper.Start()
	bt.Component.Mode = flow.ComponentModePool
	bt.Component.PoolSize = goroutinesPerEngine
	n.Add(bt, "ThreatseerBeat")
	n.Connect("DynamicRulesEngine", "Out", "ThreatseerBeat", "In")

	// err := beat.Run("threatseer", "", shipper.NewShipper)
	// if err != nil {
	// 	log.Fatal("could not start shipper, got: ", err)
	// }

	// our net has 1 inport mapped to StaticRulesEngine.In
	n.MapInPort("In", "StaticRulesEngine", "In")
	return n
}
