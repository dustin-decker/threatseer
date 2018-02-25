package agent

import (
	"runtime"
	"time"

	"github.com/coreos/go-systemd/dbus"
	"github.com/golang/glog"
	"github.com/jpillora/backoff"
	log "github.com/sirupsen/logrus"
)

const (
	interval = 5
)

func (srv *Server) Systemd() {

	b := &backoff.Backoff{
		Max: 10 * time.Minute,
	}

	for {
		conn, err := dbus.New()
		if err != nil {
			d := b.Duration()
			log.Info("error %s connecting to systemd, reconnecting in %s", err, d)
			time.Sleep(d)
			continue
		}
		b.Reset()
		// defer conn.Close()

		evntChan, errChan := conn.SubscribeUnits(interval * time.Second)

		go func() {
			for {
				err := <-errChan
				if err != nil {
					glog.Error("systemd connection error: ", err)
				}
				time.Sleep(1 * time.Second)
			}
		}()

		go func() {
			for {
				evnt := <-evntChan
				if evnt != nil {
					for _, v := range evnt {
						log.WithFields(log.Fields{
							"type":       "systemd unit",
							"name":       v.Name,
							"load_state": v.LoadState,
							"sub_state":  v.SubState,
							"path":       v.Path,
						}).Info(v.Description)
					}
				}
			}
		}()

		runtime.Goexit()
	}

}
