package satoken

import "time"

// Config authorization configuration parameters
type Config struct {
	TokenName         string
	Timeout           time.Duration
	ActiveTimeout     time.Duration
	IsConcurrent      bool
	IsShare           bool
	TokenStyle        string
	MaxLoginCount     int
	MaxTryTimes       int
	DataRefreshPeriod int
	AutoRenew         bool
}

// NewDefaultConfig create to default config
func NewDefaultConfig() *Config {
	return &Config{
		TokenName:     "satoken",
		TokenStyle:    "uuid",
		Timeout:       30 * time.Minute,
		ActiveTimeout: -1,
	}
}
