package pipeline

import (
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
	// add engines to the network
	n.Add(new(static.StaticRulesEngine), "StaticRulesEngine")
	n.Add(new(dynamic.DynamicRulesEngine), "DynamicRulesEngine")
	// connect them with a channel
	n.Connect("StaticRulesEngine", "Out", "DynamicRulesEngine", "In")
	// our net has 1 inport mapped to StaticRulesEngine.In
	n.MapInPort("In", "StaticRulesEngine", "In")
	return n
}
