package static

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/dustin-decker/threatseer/server/event"
	flow "github.com/trustmaster/goflow"
)

type StaticRulesEngine struct {
	flow.Component
	In             <-chan event.Event
	Out            chan<- event.Event
	riskyProcesses []riskyProcess
}

func (engine *StaticRulesEngine) OnIn(e event.Event) {
	processInfo := e.Event.GetProcess()

	if processInfo != nil {
		// check for risky processes
		rp := engine.checkRiskyProcess(processInfo)
		if rp != nil {
			e.Indicators = append(e.Indicators, rp.Indicator())
		}
	}

	log.Print(e.Event)
	engine.Out <- e
}

// NewStaticRulesEngine returns engine with configs loaded
func NewStaticRulesEngine() StaticRulesEngine {
	var e StaticRulesEngine

	// load risky_process.yaml information
	filename := "config/risky_processes.yaml"
	var rp []riskyProcess
	if _, err := os.Stat("config/risky_processes.yaml"); os.IsNotExist(err) {
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

	return e
}
