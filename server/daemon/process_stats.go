package daemon

import (
	"fmt"
	"log"
	"runtime"
	"time"
)

// ProcessStats just logs process stats at intervals
// Eventually will emit processing stats too.
func ProcessStats() {
	for {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		log.Print(map[string]string{
			"alloc":              fmt.Sprintf("%v", m.Alloc),
			"total-alloc":        fmt.Sprintf("%v", m.TotalAlloc/1024),
			"sys":                fmt.Sprintf("%v", m.Sys/1024),
			"num-gc":             fmt.Sprintf("%v", m.NumGC),
			"goroutines":         fmt.Sprintf("%v", runtime.NumGoroutine()),
			"stop-pause-nanosec": fmt.Sprintf("%v", m.PauseTotalNs),
		})
		time.Sleep(30 * time.Second)
	}
}
