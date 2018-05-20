package event

import (
	api "github.com/capsule8/capsule8/api/v0"
)

type Event struct {
	Event *api.ReceivedTelemetryEvent
	Score map[string]int
}
