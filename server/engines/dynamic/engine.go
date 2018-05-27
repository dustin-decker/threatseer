package dynamic

import (
	"context"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/caibirdme/yql"
	"github.com/dustin-decker/threatseer/server/event"
	"github.com/fatih/structs"
	yaml "gopkg.in/yaml.v2"
)

// Rules are user defined rules loaded at run time from a yaml file
type Rules []struct {
	Name          string    `yaml:"name"`
	Description   string    `yaml:"description"`
	ExtraInfo     string    `yaml:"extra_info"`
	EventType     string    `yaml:"event_type"`
	Query         string    `yaml:"query"`
	Actions       []string  `yaml:"actions"`
	IndicatorType string    `yaml:"indicator_type"`
	Score         int       `yaml:"score"`
	yql           yql.Ruler // compiled AST of query from ANTLR
}

// RulesEngine stores engine state
type RulesEngine struct {
	Out   chan event.Event
	Rules Rules
	ctx   context.Context
}

// AnalyzeFromPipeline initiates the engine on the pipeline
func (engine *RulesEngine) AnalyzeFromPipeline(in chan event.Event) {
	defer close(engine.Out)
	for e := range in {
		// convert struct to map[string]interface{}
		evnt := structs.Map(e)

		for _, rule := range engine.Rules {
			if len(rule.Query) > 0 {
				result, err := rule.yql.Match(evnt)
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
							ExtraInfo:     rule.ExtraInfo,
							Score:         rule.Score,
							RuleName:      rule.Name,
						},
					)
				}
			}

		}

		engine.Out <- e
	}
}

// NewDynamicRulesEngine returns engine with configs loaded
func NewDynamicRulesEngine(ctx context.Context) RulesEngine {
	var e RulesEngine

	e.ctx = ctx

	// load risky_process.yaml information
	filename := "config/dynamic_rules.yaml"
	var dr Rules
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.WithFields(log.Fields{"engine": "dynamic", "err": err, "filename": filename}).Warn("config not found, not using engine")
		return e
	}
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.WithFields(log.Fields{"engine": "dynamic", "err": err, "filename": filename}).Fatal("could not read")
	}
	err = yaml.Unmarshal(bytes, &dr)
	if err != nil {
		log.WithFields(log.Fields{"engine": "dynamic", "err": err, "filename": filename}).Fatal("could not parse")
	}
	e.Rules = dr

	// compile the rules
	for i, rule := range dr {
		compiledRule, err := yql.Rule(rule.Query)
		if err != nil || compiledRule == nil {
			log.WithFields(log.Fields{"engine": "dynamic", "err": err, "filename": filename, "rule": rule}).Fatal("could not compile rule")
		}
		dr[i].yql = compiledRule
	}
	e.Out = make(chan event.Event, 10)

	return e
}
