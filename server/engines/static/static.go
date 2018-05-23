package static

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/dustin-decker/threatseer/server/event"
)

// RulesEngine stores engine state
type RulesEngine struct {
	Out            chan event.Event
	riskyProcesses []riskyProcess
}

// Run initiates the engine on the pipeline
func (engine *RulesEngine) Run(in chan event.Event) {
	for {
		// incoming event from the pipeline
		e := <-in

		// process checks
		processInfo := e.Event.GetProcess()
		if processInfo != nil {
			// check for risky processes
			rp := engine.checkRiskyProcess(processInfo)
			if rp != nil {
				e.Indicators = append(e.Indicators, rp.Indicator())
			}
		}

		// make event available to the next pipeline engine
		engine.Out <- e
	}
}

// NewStaticRulesEngine returns engine with configs loaded
func NewStaticRulesEngine() RulesEngine {
	var e RulesEngine

	// load risky_process.yaml information
	filename := "config/risky_processes.yaml"
	var rp []riskyProcess
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("%s does not exist, not loading any data for that check", filename)
	} else {
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("could not read %s, got %s", filename, err.Error())
		}
		err = yaml.Unmarshal(bytes, &rp)
		if err != nil {
			log.Fatalf("could not parse %s, got %s", filename, err.Error())
		}
	}
	e.riskyProcesses = rp
	e.Out = make(chan event.Event, 0)

	return e
}
