package models

import (
	"encoding/binary"
	"net"
	"time"
)

// KernelEvent is that
type KernelEvent struct {
	ID            string    `orm:"column(id);index;pk"`
	CreatedAt     time.Time `orm:"column(created_at);auto_now_add;type(datetime)"`
	SensorID      string    `orm:"column(sensor_id);index"`
	ProcessID     string    `orm:"column(process_id);index"`
	ContainerID   string    `orm:"column(container_id)"`
	ContainerName string    `orm:"column(container_name)"`
	ProcessPID    int32     `orm:"column(process_pid);index"`
	Credentials   string    `orm:"column(credentials)"`

	Dest string `orm:"column(dest);index"`
}

// GetKernelEventContext returns a KernelEvent from an Event
// returns nil if it is not a KernelEvent
func GetKernelEventContext(e Event) *KernelEvent {
	k := e.Event.GetKernelCall()
	if k != nil {
		ke := KernelEvent{
			CreatedAt: time.Now().UTC(),

			SensorID:      e.Event.SensorId,
			ProcessID:     e.Event.ProcessId,
			ContainerID:   e.Event.ContainerId,
			ContainerName: e.Event.ContainerName,
			ProcessPID:    e.Event.ProcessPid,
			Credentials:   e.Event.Credentials.String(),
		}
		if dest := k.GetArguments()["sin_addr"]; dest != nil {
			ke.Dest = uint2ip(dest.GetUnsignedValue()).String()
		}
		return &ke
	}
	return nil
}

func uint2ip(n uint64) net.IP {
	ip := make(net.IP, 4)
	binary.LittleEndian.PutUint32(ip, uint32(n))
	return ip
}
