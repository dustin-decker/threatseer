package dynamic

import (
	"github.com/dustin-decker/threatseer/server/event"
	flow "github.com/trustmaster/goflow"
)

func (engine *DynamicRulesEngine) OnIn(e event.Event) {
	// log.Print(e.Event)
	engine.Out <- e
}

type DynamicRulesEngine struct {
	flow.Component
	In  <-chan event.Event
	Out chan<- event.Event
}
