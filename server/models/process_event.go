package models

import (
	"strings"
	"time"
)

// ProcessEvent is that
type ProcessEvent struct {
	ID            string    `orm:"column(id);index;pk" json:"id"`
	CreatedAt     time.Time `orm:"column(created_at);auto_now_add;type(datetime)"`
	SensorID      string    `orm:"column(sensor_id);index"`
	ProcessID     string    `orm:"column(process_id);index"`
	ContainerID   string    `orm:"column(container_id)"`
	ContainerName string    `orm:"column(container_name)"`
	ProcessPID    int32     `orm:"column(process_pid);index"`
	Credentials   string    `orm:"column(credentials)"`

	ProcessEventType string `orm:"column(process_event_type)"`
	ExecFilename     string `orm:"column(exec_filename)"`
	ExecCmdLine      string `orm:"column(exec_cmd_line)"`
	ForkChildPID     int32  `orm:"column(fork_child_pid);index"`
	ForkChildID      string `orm:"column(fork_child_id)"`
}

// GetProcessEventContext returns a ProcessEvent from an Event
// returns nil if it is not a ProcessEvent
func GetProcessEventContext(e Event) *ProcessEvent {
	p := e.Event.GetProcess()
	if p != nil {
		return &ProcessEvent{
			CreatedAt: time.Now().UTC(),

			SensorID:      e.Event.SensorId,
			ProcessID:     e.Event.ProcessId,
			ContainerID:   e.Event.ContainerId,
			ContainerName: e.Event.ContainerName,
			ProcessPID:    e.Event.ProcessPid,
			Credentials:   e.Event.Credentials.String(),

			ProcessEventType: p.GetType().String(),
			ExecFilename:     p.ExecFilename,
			ExecCmdLine:      strings.Join(p.ExecCommandLine, " "),
			ForkChildPID:     p.GetForkChildPid(),
			ForkChildID:      p.GetForkChildId(),
		}
	}
	return nil
}
