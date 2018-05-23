package daemon

import (
	"fmt"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
)

// ProcessStats just logs process stats at intervals
// Eventually will emit processing stats too.
func ProcessStats() {
	for {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		log.WithFields(log.Fields{
			"alloc":              fmt.Sprintf("%v", m.Alloc),
			"total-alloc":        fmt.Sprintf("%v", m.TotalAlloc/1024),
			"sys":                fmt.Sprintf("%v", m.Sys/1024),
			"num-gc":             fmt.Sprintf("%v", m.NumGC),
			"goroutines":         fmt.Sprintf("%v", runtime.NumGoroutine()),
			"stop-pause-nanosec": fmt.Sprintf("%v", m.PauseTotalNs),
		}).Info("process stats")

		time.Sleep(30 * time.Second)
	}
}
