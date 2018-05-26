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
	"github.com/dustin-decker/threatseer/server/daemon"
	cmd "github.com/elastic/beats/libbeat/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	var Name = "threatseer"
	var Version = "0.1.1"

	RootCmd := cmd.GenRootCmd(Name, Version, daemon.New)

	// send to stdout by default
	RootCmd.SetArgs([]string{"-e"})

	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
