package config

import "time"

// Config for threatseer
type Config struct {
	ListenAddress     string `config:"listen_address"`
	NumberOfPipelines uint   `config:"number_of_pipelines"`
	// Profile Engine options
	ProfileBuildingDuration     time.Duration `config:"profile_building_duration"`
	ProfileEventFilterCacheSize uint          `config:"profile_event_filter_cache_size"`
	// TLS options
	TLSEnabled        bool   `config:"tls_enabled"`
	TLSRootCAPath     string `config:"tls_root_ca_path"`
	TLSServerKeyPath  string `config:"tls_server_key_path"`
	TLSServerCertPath string `config:"tls_server_cert_path"`
}

// DefaultConfig threatseer config
var DefaultConfig = Config{
	ListenAddress: "0.0.0.0:8081",
	// NumberOfPipelines=0 means we use runtime.NumProcs()
	NumberOfPipelines:       0,
	ProfileBuildingDuration: 60 * time.Minute,
	// 8000000 events consumes about 10MB of RAM. Don't go lower.
	ProfileEventFilterCacheSize: 10000000,
	TLSEnabled:                  false,
}
