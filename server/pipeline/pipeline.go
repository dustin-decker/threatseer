package pipeline

import (
	"runtime"

	"github.com/dustin-decker/threatseer/server/engines/dynamic"
	"github.com/dustin-decker/threatseer/server/engines/static"
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
	se := new(static.StaticRulesEngine)
	se.Component.Mode = flow.ComponentModePool
	se.Component.PoolSize = goroutinesPerEngine
	n.Add(se, "StaticRulesEngine")
	de := new(dynamic.DynamicRulesEngine)
	de.Component.Mode = flow.ComponentModePool
	de.Component.PoolSize = goroutinesPerEngine
	n.Add(de, "DynamicRulesEngine")
	// connect them with a channel
	n.Connect("StaticRulesEngine", "Out", "DynamicRulesEngine", "In")
	// our net has 1 inport mapped to StaticRulesEngine.In
	n.MapInPort("In", "StaticRulesEngine", "In")
	return n
}
