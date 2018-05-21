package dynamic

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/dustin-decker/threatseer/server/event"
	yaml "gopkg.in/yaml.v2"
)

type dynamicRule struct {
	eventType   string
	description string
	actions     []string
	query       string
	score       int
}

type DynamicRulesEngine struct {
	Out          chan event.Event
	dynamicRules []dynamicRule
}

// Run initiates the engine on the pipeline
func (engine *DynamicRulesEngine) Run(in chan event.Event) {
	for {
		e := <-in
		// log.Print(e)
		engine.Out <- e
	}
}

// NewDynamicRulesEngine returns engine with configs loaded
func NewDynamicRulesEngine() DynamicRulesEngine {
	var e DynamicRulesEngine

	// load risky_process.yaml information
	filename := "config/dynamic_rules.yaml"
	var dr []dynamicRule
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("%s does not exist, not loading any data for that check", filename)
	} else {
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("could not read %s, got %s", filename, err.Error())
		}
		err = yaml.Unmarshal(bytes, &dr)
		if err != nil {
			log.Fatalf("could not parse %s, got %s", filename, err.Error())
		}
	}
	e.dynamicRules = dr
	e.Out = make(chan event.Event, 0)

	return e
}
