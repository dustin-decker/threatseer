// Copyright 2018 Dustin Decker

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"io"
	"net"
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/capsule8/capsule8/pkg/config"
	"github.com/capsule8/capsule8/pkg/sensor"
	"github.com/capsule8/capsule8/pkg/services"
)

func startAgent() {
	manager := services.NewServiceManager()
	if len(config.Global.ProfilingListenAddr) > 0 {
		service := services.NewProfilingService(
			config.Global.ProfilingListenAddr)
		manager.RegisterService(service)
	}

	if len(config.Sensor.ListenAddr) > 0 {
		s, err := sensor.NewSensor()
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal("could not create sensor")
		}
		if err := s.Start(); err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal("could not start sensor")
		}
		defer s.Stop()
		service := sensor.NewTelemetryService(s, "unix:/var/run/threatseer.sock")
		manager.RegisterService(service)
	}

	manager.Run()
}

func joinConn(conn1, conn2 net.Conn) chan error {
	connErrChan := make(chan error)
	go func() {
		_, err := io.Copy(conn1, conn2)
		connErrChan <- err
	}()
	go func() {
		_, err := io.Copy(conn2, conn1)
		connErrChan <- err
	}()
	return connErrChan
}

func establishUplink() {
	// exit with ctrl-c
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)

	var exit bool
	go func() {
		<-signals
		exit = true
	}()

	log.Info("connecting to agent")

	sensorConn, err := net.DialTimeout("unix", "/var/run/threatseer.sock", time.Second*5)
	if err != nil {
		log.Error(err)
		log.Warn("reconnecting in 5 seconds")
		time.Sleep(5 * time.Second)
		establishUplink()
	}
	defer sensorConn.Close()
	log.Info("connecting to remote")

	serverConn, err := net.DialTimeout("tcp", "127.0.0.1:8081", time.Second*5)
	if err != nil {
		log.Println(err)
		sensorConn.Close()
		log.Warn("reconnecting in 5 seconds")
		time.Sleep(5 * time.Second)
		establishUplink()
	}
	defer serverConn.Close()

	log.Print("persisting telemetry uplink")
	err = <-joinConn(sensorConn, serverConn)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("connection error")
	}
	sensorConn.Close()
	serverConn.Close()
	if exit {
		log.Warn("shutting down")
		os.Exit(0)
	} else {
		log.Warn("connection interrupted")
	}
	log.Warn("reconnecting in 5 seconds")
	time.Sleep(5 * time.Second)
	establishUplink()
}

func waitForSensor() {
	for {
		if _, err := os.Stat("/var/run/threatseer.sock"); os.IsNotExist(err) {
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
}

func main() {
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true") // disable logging to file
	log.SetFormatter(&log.JSONFormatter{})

	log.Info("starting threatseer agent")

	go startAgent()

	waitForSensor()

	establishUplink()

	log.Warn("goodbye")
}
