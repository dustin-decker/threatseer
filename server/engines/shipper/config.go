package shipper

import "time"

// Config has custom options for Threatseer
type Config struct {
	Interval time.Duration `config:"interval"`
	Type     string
}

// DefaultConfig contains defaults for custom options
var DefaultConfig = Config{
	Interval: 5 * time.Second,
	Type:     "threatseer",
}
