package static

import (
	"log"

	"github.com/dustin-decker/threatseer/server/event"
	flow "github.com/trustmaster/goflow"
)

func (engine *StaticRulesEngine) OnIn(e event.Event) {
	log.Print(e.Event)
	engine.Out <- e
}

type StaticRulesEngine struct {
	flow.Component
	In  <-chan event.Event
	Out chan<- event.Event
}
