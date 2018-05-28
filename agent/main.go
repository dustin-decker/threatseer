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
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io"
	"io/ioutil"
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

func establishUplink(c cfg) {
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
		establishUplink(c)
	}
	defer sensorConn.Close()
	log.Info("connecting to remote")

	var serverConn net.Conn
	var cert tls.Certificate
	if c.TLSEnabled {
		certPool := x509.NewCertPool()
		var bs []byte
		bs, err = ioutil.ReadFile(c.TLSRootCAPath)
		if err != nil {
			log.WithFields(log.Fields{"err": err, "filepath": c.TLSRootCAPath}).Fatal("failed read CA certs")
		}
		ok := certPool.AppendCertsFromPEM(bs)
		if !ok {
			log.WithFields(log.Fields{"err": err}).Fatal("failed to add CA certs")
		}

		if len(c.TLSServerCertPath) > 0 && len(c.TLSServerKeyPath) > 0 {
			cert, err = tls.LoadX509KeyPair(c.TLSServerCertPath, c.TLSServerKeyPath)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Fatal("error loading server keys")
			}
		}

		randReader := rand.Reader
		serverConn, err = tls.Dial("tcp", c.Server,
			&tls.Config{
				Rand:         randReader,
				RootCAs:      certPool,
				Certificates: []tls.Certificate{cert},
				ServerName:   c.TLSOverrideCommonName,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				},
			},
		)
		if err != nil {
			if err != nil {
				log.Println(err)
				sensorConn.Close()
				log.Warn("reconnecting in 5 seconds")
				time.Sleep(5 * time.Second)
				establishUplink(c)
			}
		}
	} else {
		serverConn, err = net.DialTimeout("tcp", c.Server, time.Second*5)
		if err != nil {
			log.Println(err)
			sensorConn.Close()
			log.Warn("reconnecting in 5 seconds")
			time.Sleep(5 * time.Second)
			establishUplink(c)
		}
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
	establishUplink(c)
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

type cfg struct {
	Server                string `yaml:"server"`
	TLSEnabled            bool   `yaml:"tls_enabled"`
	TLSRootCAPath         string `yaml:"tls_root_ca_path"`
	TLSServerKeyPath      string `yaml:"tls_server_key_path"`
	TLSServerCertPath     string `yaml:"tls_server_cert_path"`
	TLSOverrideCommonName string `yaml:"tls_override_common_name"`
}

func main() {
	var c cfg
	flag.StringVar(&c.Server, "server", "127.0.0.1:8081", "remote server to send telemetry to")
	flag.BoolVar(&c.TLSEnabled, "tls", false, "enable tls")
	flag.StringVar(&c.TLSRootCAPath, "ca", "", "custom certificate authority for the remote server to send telemetry to")
	flag.StringVar(&c.TLSServerKeyPath, "key", "", "key for agent")
	flag.StringVar(&c.TLSServerCertPath, "cert", "", "certificate for agent")
	flag.StringVar(&c.TLSOverrideCommonName, "cn", "", "override the expected common name of the remote server")

	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true") // disable logging to file
	log.SetFormatter(&log.JSONFormatter{})

	log.Info("starting threatseer agent")

	go startAgent()

	waitForSensor()

	establishUplink(c)

	log.Warn("goodbye")
}
