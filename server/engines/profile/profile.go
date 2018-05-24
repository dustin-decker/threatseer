package profile

import (
	"github.com/dustin-decker/threatseer/server/event"
)

// Engine stores engine state
type Engine struct {
	Out chan event.Event
}

// Run initiates the engine on the pipeline
func (engine *Engine) Run(in chan event.Event) {
	for {
		// incoming event from the pipeline
		e := <-in

		//// does nothing right now

		// make event available to the next pipeline engine
		engine.Out <- e
	}
}

// NewProfileEngine returns engine with configs loaded
func NewProfileEngine() Engine {
	var e Engine
	e.Out = make(chan event.Event, 0)

	return e
}
