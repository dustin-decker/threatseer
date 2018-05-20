package static

import (
	"fmt"
	"log"

	"github.com/dustin-decker/threatseer/server/event"
	flow "github.com/trustmaster/goflow"
)

func (engine *StaticRulesEngine) OnIn(e event.Event) {
	msg := fmt.Sprintf("Got event on static engine!")
	log.Print(msg)
	engine.Out <- e
}

type StaticRulesEngine struct {
	flow.Component
	In  <-chan event.Event
	Out chan<- event.Event
}
