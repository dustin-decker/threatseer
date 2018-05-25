package profile

import (
	"strings"
	"time"

	"github.com/dustin-decker/threatseer/server/event"
	log "github.com/sirupsen/logrus"
)

func (e *Engine) getBestIdentifier(evnt event.Event) string {
	containerName := evnt.Event.GetContainerName()
	if len(containerName) > 0 {
		return containerName
	}

	return evnt.Event.GetProcessId()
}

func (e *Engine) profileExecEvent(evnt event.Event, cmd []string) int {
	bestIdentifier := e.getBestIdentifier(evnt)
	eventProfile := []byte(bestIdentifier + strings.Join(cmd, " "))

	// if subject has been profiled
	if e.IsProfiledFilter.Lookup([]byte(bestIdentifier)) {
		// if event has not been seen before, return a positive risk indicator
		if !e.HasBeenProfiledFilter.Lookup(eventProfile) {
			return 50
		}
		// it has been seen in the profile, so return a negative risk indicator
		return -50
	}

	// if subject has not been profiled, mark it as profiling now, and insert this eventProfile
	e.Mutex.Lock()
	startTime, ok := e.IsProfiling[bestIdentifier]
	e.Mutex.Unlock()
	if !ok {
		e.Mutex.Lock()
		e.IsProfiling[bestIdentifier] = time.Now()
		e.Mutex.Unlock()
		e.HasBeenProfiledFilter.Insert(eventProfile)
		return 0
	}

	// if subject has been profiled over 3 hours,
	// add the last eventProfile,
	// add it to the IsProfiledFilter,
	// and remove from IsProfiling map
	if time.Since(startTime) > time.Minute*3 {
		log.WithFields(log.Fields{"engine": "profile", "identifier": bestIdentifier}).Error("done profiling subject")
		e.HasBeenProfiledFilter.Insert(eventProfile)
		e.IsProfiledFilter.Insert([]byte(bestIdentifier))
		e.Mutex.Lock()
		delete(e.IsProfiling, bestIdentifier)
		e.Mutex.Unlock()
		return 0
	}
	// if subject is still being profiled, insert the eventProfile
	e.HasBeenProfiledFilter.Insert(eventProfile)
	return 0

}
