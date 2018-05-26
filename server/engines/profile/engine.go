package profile

import (
	"sync"
	"time"

	"github.com/dustin-decker/threatseer/server/config"
	"github.com/dustin-decker/threatseer/server/event"
	cf "github.com/seiflotfy/cuckoofilter"
)

// Engine stores engine state
type Engine struct {
	// pipeline output
	Out chan event.Event

	// cuckoo filter for if the application/image/whatever has been profiled
	IsProfiledFilter *cf.CuckooFilter
	// cuckoo filter for if the event is present in the profile
	EventFilter *cf.CuckooFilter
	// tracks when profiling started so the application can be added to the IsProfiledFilter
	IsProfiling map[string]time.Time

	ProfileBuildingDuration time.Duration

	Mutex *sync.Mutex
}

// AnalyzeFromPipeline initiates the engine on the pipeline
func (engine *Engine) AnalyzeFromPipeline(in chan event.Event) {
	for {
		// incoming event from the pipeline
		e := <-in

		// process profiling
		processInfo := e.Event.GetProcess()
		if processInfo != nil {
			// exec profiling
			cmd := processInfo.GetExecCommandLine()
			if len(cmd) > 0 {
				score := engine.profileExecEvent(e, cmd)
				if score < 0 {
					e.Indicators = append(e.Indicators, event.Indicator{
						Engine:        "profile",
						RuleName:      "",
						IndicatorType: "normal_behavior",
						Description:   "subject is behaving according to its profile",
						ExtraInfo:     "",
						Score:         score,
					})
				} else if score > 0 {
					e.Indicators = append(e.Indicators, event.Indicator{
						Engine:        "profile",
						RuleName:      "",
						IndicatorType: "abnormal_behavior",
						Description:   "subject is behaving outside of its profile",
						ExtraInfo:     "",
						Score:         score,
					})
				}
			}
		}
		// make event available to the next pipeline engine
		engine.Out <- e
	}
}

// NewProfileEngine returns engine with configs loaded
func NewProfileEngine(c config.Config) Engine {
	e := Engine{
		Out: make(chan event.Event, 10),
		// 10000 subject capacity
		IsProfiledFilter: cf.NewCuckooFilter(100000),
		// 4000 nodes * 2000 eventProfiles per node = 8000000
		EventFilter:             cf.NewCuckooFilter(c.ProfileEventFilterCacheSize),
		IsProfiling:             map[string]time.Time{},
		ProfileBuildingDuration: c.ProfileBuildingDuration,
		Mutex: &sync.Mutex{},
	}

	return e
}
