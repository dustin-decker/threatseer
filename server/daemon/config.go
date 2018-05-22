package daemon

import (
	"io/ioutil"
	"log"
	"os"

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
		log.Fatalf("%s does not exist, aborting", filename)
	} else {
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("could not read %s, got %s", filename, err.Error())
		}
		err = yaml.Unmarshal(bytes, &c)
		if err != nil {
			log.Fatalf("could not parse %s, got %s", filename, err.Error())
		}
	}

	return c
}
