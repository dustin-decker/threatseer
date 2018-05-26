package config

// Config for threatseer
type Config struct {
	ListenAddress     string `config:"listen_address"`
	NumberOfPipelines int    `config:"number_of_pipelines"`
}

// DefaultConfig threatseer config
var DefaultConfig = Config{
	ListenAddress: "0.0.0.0:8081",
	// NumberOfPipelines=0 means we use runtime.NumProcs()
	NumberOfPipelines: 0,
}
