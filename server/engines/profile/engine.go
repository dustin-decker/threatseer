package profile

import (
	"context"
	"time"

	lru "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"

	"github.com/dustin-decker/threatseer/server/config"
	"github.com/dustin-decker/threatseer/server/models"
	cf "github.com/seiflotfy/cuckoofilter"
)

// Engine stores engine state
type Engine struct {
	// pipeline output
	Out chan models.Event

	// cuckoo filter for if the event is present in the profile
	EventFilter *cf.CuckooFilter
	// tracks when profiling started so the application can be added to the IsProfiledFilter
	IsProfiling *lru.Cache
	// tracks if the application/image/whatever has been profiled
	IsProfiled *lru.Cache

	ProfileBuildingDuration time.Duration

	ctx context.Context
}

// AnalyzeFromPipeline initiates the engine on the pipeline
func (engine *Engine) AnalyzeFromPipeline(in chan models.Event) {
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
					e.Indicators = append(e.Indicators, models.Indicator{
						Engine:        "profile",
						RuleName:      "",
						IndicatorType: "normal_behavior",
						Description:   "subject is behaving according to its profile",
						ExtraInfo:     "",
						Score:         score,
					})
				} else if score > 0 {
					e.Indicators = append(e.Indicators, models.Indicator{
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
	profilingLRUCache, err := lru.New(50000)
	if err != nil {
		log.WithFields(log.Fields{"engine": "profile", "err": err}).Fatal("could not make LRU cache")
	}
	profiledLRUCache, err := lru.New(100000)
	if err != nil {
		log.WithFields(log.Fields{"engine": "profile", "err": err}).Fatal("could not make LRU cache")
	}

	e := Engine{
		Out:                     make(chan models.Event, 10),
		IsProfiled:              profiledLRUCache,
		EventFilter:             cf.NewCuckooFilter(c.ProfileEventFilterCacheSize),
		IsProfiling:             profilingLRUCache,
		ProfileBuildingDuration: c.ProfileBuildingDuration,
		ctx: ctx,
	}

	return e
}
