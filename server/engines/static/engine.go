package static

import (
	"context"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"

	"github.com/dustin-decker/threatseer/server/event"
)

// RulesEngine stores engine state
type RulesEngine struct {
	Out            chan event.Event
	riskyProcesses []riskyProcess
	ctx            context.Context
}

// AnalyzeFromPipeline initiates the engine on the pipeline
func (engine *RulesEngine) AnalyzeFromPipeline(in chan event.Event) {
	defer close(engine.Out)
	for e := range in {
		// process checks
		processInfo := e.Event.GetProcess()
		if processInfo != nil {
			// check for risky processes
			rp := engine.checkRiskyProcess(processInfo)
			if rp != nil {
				e.Indicators = append(e.Indicators, rp.Indicator())
			}
		}

		engine.Out <- e
	}
}

// NewStaticRulesEngine returns engine with configs loaded
func NewStaticRulesEngine(ctx context.Context) RulesEngine {
	var e RulesEngine

	// load risky_process.yaml information
	filename := "config/risky_processes.yaml"
	var rp []riskyProcess
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.WithFields(log.Fields{"engine": "static", "err": err, "filename": filename}).Warn("config not found, not using engine")
	} else {
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			log.WithFields(log.Fields{"engine": "static", "err": err, "filename": filename}).Fatal("could not read")
		}
		err = yaml.Unmarshal(bytes, &rp)
		if err != nil {
			log.WithFields(log.Fields{"engine": "static", "err": err, "filename": filename}).Fatal("could not parse")
		}
	}
	e.riskyProcesses = rp
	e.Out = make(chan event.Event, 10)
	e.ctx = ctx

	return e
}
