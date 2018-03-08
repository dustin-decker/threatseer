// Copyright 2017 Capsule8, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// modified for JSON logging and to run as a component of threatseer

package agent

import (
	"os"
	"os/signal"
	"runtime"

	"github.com/capsule8/capsule8/pkg/sensor"
	"github.com/capsule8/capsule8/pkg/sys/perf"
	log "github.com/sirupsen/logrus"
)

// most of this is from https://github.com/capsule8/capsule8/blob/master/examples/cache-side-channel

const (
	// LLCLoadSampleSize define number of cache loads to sample on.
	// After each sample period of this many cache loads, the cache
	// miss rate is calculated and examined. This value tunes the
	// trade-off between CPU load and detection accuracy.
	LLCLoadSampleSize = 10000

	// Alarm thresholds as cache miss rates (between 0 and 1).
	// These values tune the trade-off between false negatives and
	// false positives.
	alarmThresholdWarning = 0.98
	alarmThresholdError   = 0.99

	// perf_event_attr config value for LL cache loads
	perfConfigLLCLoads = perf.PERF_COUNT_HW_CACHE_LL |
		(perf.PERF_COUNT_HW_CACHE_OP_READ << 8) |
		(perf.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16)

	// perf_event_attr config value for LL cache misses
	perfConfigLLCLoadMisses = perf.PERF_COUNT_HW_CACHE_LL |
		(perf.PERF_COUNT_HW_CACHE_OP_READ << 8) |
		(perf.PERF_COUNT_HW_CACHE_RESULT_MISS << 16)
)

type eventCounters struct {
	LLCLoads      uint64
	LLCLoadMisses uint64
}

var cpuCounters []eventCounters

type counterTracker struct {
	sensor   *sensor.Sensor
	counters []eventCounters
}

// L3missDetector detects large ammounts of L3 cache misses,
// which occur during cache timing attacks. Cache timing
// attacks are utilized in Meldown, Spectre, and Rowhammer type exploits.
func (srv *Server) L3missDetector() {

	log.Info("starting cache side channel detector")

	tracker := counterTracker{
		sensor:   srv.Sensor,
		counters: make([]eventCounters, runtime.NumCPU()),
	}

	// Create our event group to read LL cache accesses and misses
	//
	// We ask the kernel to sample every llcLoadSampleSize LLC
	// loads. During each sample, the LLC load misses are also
	// recorded, as well as CPU number, PID/TID, and sample time.
	attr := perf.EventAttr{
		SamplePeriod: LLCLoadSampleSize,
		SampleType:   perf.PERF_SAMPLE_TID | perf.PERF_SAMPLE_CPU,
	}
	groupID, err := tracker.sensor.Monitor.RegisterHardwareCacheEventGroup(
		[]uint64{
			perfConfigLLCLoads,
			perfConfigLLCLoadMisses,
		},
		tracker.decodeConfigLLCLoads,
		perf.WithEventAttr(&attr))
	if err != nil {
		log.Fatalf("could not register hardware cache event: %s", err)
	}

	log.Info("monitoring cache side channel misses")

	tracker.sensor.Monitor.EnableGroup(groupID)

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)
	<-signals
	close(signals)

	log.Info("shutting down cache miss sensor")
	tracker.sensor.Monitor.Close()
}

func (t *counterTracker) decodeConfigLLCLoads(
	sample *perf.SampleRecord,
	counters map[uint64]uint64,
	totalTimeElapsed uint64,
	totalTimeRunning uint64,
) (interface{}, error) {
	cpu := sample.CPU
	prevCounters := t.counters[cpu]
	t.counters[cpu] = eventCounters{
		LLCLoads:      counters[perfConfigLLCLoads],
		LLCLoadMisses: counters[perfConfigLLCLoadMisses],
	}

	counterDeltas := eventCounters{
		LLCLoads:      t.counters[cpu].LLCLoads - prevCounters.LLCLoads,
		LLCLoadMisses: t.counters[cpu].LLCLoadMisses - prevCounters.LLCLoadMisses,
	}

	t.alarm(sample, counterDeltas)
	return nil, nil
}

func (t *counterTracker) alarm(sr *perf.SampleRecord, counters eventCounters) {
	LLCLoadMissRate := float32(counters.LLCLoadMisses) / float32(counters.LLCLoads)

	if LLCLoadMissRate > alarmThresholdWarning {
		hostname, _ := os.Hostname()
		evnt := log.Fields{
			"hostname":        hostname,
			"attack":          "L3 cache miss timing",
			"pid":             sr.Pid,
			"LLCLoadMissRate": LLCLoadMissRate,
		}

		if sr.Pid > 0 {
			task := t.sensor.ProcessCache.LookupTask(int(sr.Pid))
			containerInfo := t.sensor.ProcessCache.LookupTaskContainerInfo(task)
			if containerInfo != nil {
				evnt["container_name"] = containerInfo.Name
				evnt["container_id"] = containerInfo.ID
				evnt["container_image"] = containerInfo.ImageName
				evnt["container_id"] = containerInfo.ImageID
			}

			log.WithFields(evnt).Warn("possible Meltdown | Spectre | Rowhammer | other attack utilizing L3 cache miss timing detected")
		}
	}
}
