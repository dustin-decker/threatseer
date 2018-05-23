package daemon

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	yaml "gopkg.in/yaml.v2"
)

// Config struct for server.yaml
type Config struct {
	ListenAddress     string `yaml:"listen_address"`
	NumberOfPipelines int    `yaml:"number_of_pipelines"`
}

// LoadConfigFromFile loads the config from a file
func LoadConfigFromFile() Config {
	// load server.yaml information
	filename := "config/daemon.yaml"
	var c Config
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.WithFields(log.Fields{"err": err}).Errorf("%s does not exist, aborting", filename)
	} else {
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Errorf("could not read %s", filename)
		}
		err = yaml.Unmarshal(bytes, &c)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Errorf("could not parse %s", filename)
		}
	}

	return c
}
