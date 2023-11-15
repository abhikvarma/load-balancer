package utils

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type LBStrategy int

const (
	RoundRobin LBStrategy = iota
	LeastConnected
)

func GetLBStrategy(strategy string) LBStrategy {
	switch strategy {
	case "least-connected":
		return LeastConnected
	default:
		return RoundRobin
	}
}

type Config struct {
	Port                     int      `yaml:"lb_port"`
	Backends                 []string `yaml:"backends"`
	Strategy                 string   `yaml:"strategy"`
	HealthCheckIntervalInSec int      `yaml:"health_check_interval_in_sec"`
	MaxRetries               int      `yaml:"max_retries"`
}

const DEFAULT_MAX_RETRIES int = 3
const DEFAULT_HEALTH_CHECK_INTERVAL_IN_SEC int = 20

func GetLBConfig() (*Config, error) {
	var config Config
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}

	if len(config.Backends) == 0 {
		return nil, errors.New("at least one backend host expected, none provided")
	}

	if config.Port == 0 {
		return nil, errors.New("load balancer port not found")
	}

	if config.HealthCheckIntervalInSec <= 0 || config.HealthCheckIntervalInSec > 600 {
		Logger.Warn(fmt.Sprintf("health_check_interval_in_sec outside the range (0, 600], reverting to the default of %d", DEFAULT_HEALTH_CHECK_INTERVAL_IN_SEC))
		config.HealthCheckIntervalInSec = DEFAULT_HEALTH_CHECK_INTERVAL_IN_SEC
	}

	if config.MaxRetries <= 0 || config.MaxRetries > 10 {
		Logger.Warn(fmt.Sprintf("max_retries outside the range (0, 10], reverting to the default of %d", DEFAULT_MAX_RETRIES))
		config.MaxRetries = DEFAULT_MAX_RETRIES
	}

	return &config, nil

}
