package profile

import (
	"strings"
	"time"

	"github.com/dustin-decker/threatseer/server/models"
	log "github.com/sirupsen/logrus"
)

func (e *Engine) getBestIdentifier(evnt models.Event) string {
	imageID := evnt.Event.GetImageId()
	if len(imageID) > 0 {
		return imageID
	}

	return evnt.Event.GetProcessId()
}

func (e *Engine) profileExecEvent(evnt models.Event, cmd []string) int {
	bestIdentifier := e.getBestIdentifier(evnt)
	eventProfile := []byte(bestIdentifier + strings.Join(cmd, " "))

	// if subject has been profiled
	if _, subjectHasBeenProfiled := e.IsProfiled.Get(bestIdentifier); subjectHasBeenProfiled {
		// if the event profile has been seen before
		if e.EventFilter.Lookup(eventProfile) {
			// return a negative risk indicator
			return -50
		}
		// if the event profile has not been seen
		// return a positive risk indicator
		return 50
	}

	// if subject has not been profiled
	_, subjectIsProfiling := e.IsProfiling.Get(bestIdentifier)
	if !subjectIsProfiling {
		// mark it as currently profiling
		e.IsProfiling.Add(bestIdentifier, time.Now())
		// insert this event profile
		e.EventFilter.Insert(eventProfile)
		// return a neutral risk indicator
		return 0
	}

	// if subject has been profiled over the req'd time period
	startTime, ok := e.IsProfiling.Get(bestIdentifier)
	if ok && time.Since(startTime.(time.Time)) > e.ProfileBuildingDuration {
		log.WithFields(log.Fields{"engine": "profile",
			"identifier": bestIdentifier,
			"duration":   e.ProfileBuildingDuration}).Error("done profiling subject")
		// add the last eventProfile
		e.EventFilter.Insert(eventProfile)
		// add it to the IsProfiled LRU cache
		e.IsProfiled.Add(bestIdentifier, time.Now())
		// remove from IsProfiling LRU cache
		e.IsProfiling.Remove(bestIdentifier)
		return 0
	}

	// if subject is still being profiled, insert the eventProfile
	e.EventFilter.Insert(eventProfile)
	return 0

}
