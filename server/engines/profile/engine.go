package profile

import (
	"context"
	"time"

	lru "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"

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
	IsProfiling *lru.Cache

	ProfileBuildingDuration time.Duration

	ctx context.Context
}

// AnalyzeFromPipeline initiates the engine on the pipeline
func (engine *Engine) AnalyzeFromPipeline(in chan event.Event) {
	defer close(engine.Out)
	for e := range in {
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

		engine.Out <- e
	}
}

// NewProfileEngine returns engine with configs loaded
func NewProfileEngine(ctx context.Context, c config.Config) Engine {
	lruCache, err := lru.New(50000)
	if err != nil {
		log.WithFields(log.Fields{"engine": "profile", "err": err}).Fatal("could not make LRU cache")
	}

	e := Engine{
		Out: make(chan event.Event, 10),
		// 10000 subject capacity
		IsProfiledFilter: cf.NewCuckooFilter(100000),
		// 4000 nodes * 2000 eventProfiles per node = 8000000
		EventFilter:             cf.NewCuckooFilter(c.ProfileEventFilterCacheSize),
		IsProfiling:             lruCache,
		ProfileBuildingDuration: c.ProfileBuildingDuration,
		ctx: ctx,
	}

	return e
}
