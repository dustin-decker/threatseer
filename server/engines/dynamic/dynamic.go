package dynamic

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/caibirdme/yql"
	"github.com/dustin-decker/threatseer/server/event"
	"github.com/fatih/structs"
	yaml "gopkg.in/yaml.v2"
)

// DynamicRules are user defined rules loaded at run time from a yaml file
type DynamicRules []struct {
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description"`
	EventType     string   `yaml:"event_type"`
	Query         string   `yaml:"query"`
	Actions       []string `yaml:"actions"`
	IndicatorType string   `yaml:"indicator_type"`
	Score         int      `yaml:"score"`
}

// RulesEngine stores engine state
type RulesEngine struct {
	Out          chan event.Event
	DynamicRules DynamicRules
}

// Run initiates the engine on the pipeline
func (engine *RulesEngine) Run(in chan event.Event) {
	for {
		// incoming event from the pipeline
		e := <-in

		// convert struct to map[string]interface{}
		evnt := structs.Map(e)

		for _, rule := range engine.DynamicRules {
			if len(rule.Query) > 0 {
				result, err := yql.Match(rule.Query, evnt)
				if err != nil {
					if err.Error() == "interface conversion: interface is nil, not antlr.ParserRuleContext" {
						log.WithFields(log.Fields{"rule": rule.Name, "query": rule.Query}).Error("incorrect syntax for dynamic engine rule")

					} else {
						log.WithFields(log.Fields{"err": err, "rule": rule.Name}).Error("dynamic engine got error while testing rule")
					}
				}
				if result {
					e.Indicators = append(
						e.Indicators,
						event.Indicator{
							Engine:        "dynamic",
							IndicatorType: rule.IndicatorType,
							Description:   rule.Description,
							Score:         rule.Score,
							RuleName:      rule.Name,
						},
					)
				}
			}
		}

		// make event available to the next pipeline engine
		engine.Out <- e
	}
}

// NewDynamicRulesEngine returns engine with configs loaded
func NewDynamicRulesEngine() RulesEngine {
	var e RulesEngine

	// load risky_process.yaml information
	filename := "config/dynamic_rules.yaml"
	var dr DynamicRules
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.WithFields(log.Fields{"err": err}).Warnf("%s does not exist, not loading any data for that check", filename)
	} else {
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Errorf("could not read %s", filename)
		}
		err = yaml.Unmarshal(bytes, &dr)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Errorf("could not parse %s", filename)
		}
	}
	e.DynamicRules = dr
	e.Out = make(chan event.Event, 0)

	return e
}
