package static

import (
	"fmt"
	"strings"

	"github.com/dustin-decker/threatseer/server/event"

	api "github.com/capsule8/capsule8/api/v0"
)

type riskyProcess struct {
	Name   string
	Reason string
	Score  int
}

func (rp *riskyProcess) Indicator() event.Indicator {
	return event.Indicator{
		Engine:        "static",
		IndicatorType: "risky_process",
		Description:   fmt.Sprintf("%s is a risky process often used for %s", rp.Name, rp.Reason),
		Score:         rp.Score,
	}
}

func (engine *RulesEngine) checkRiskyProcess(processInfo *api.ProcessEvent) *riskyProcess {
	if processInfo != nil {
		cl := processInfo.GetExecFilename()
		for _, rp := range engine.riskyProcesses {
			if strings.Contains(cl, "/"+rp.Name) {
				return &rp
			}
		}
	}
	return nil
}
