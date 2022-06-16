package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"k8s.io/apimachinery/pkg/util/yaml"
)

// Config represents the configuration of this service
type Config struct {
	ReconcilationInterval       time.Duration `json:"-"`
	ReconcilationIntervalHelper string        `json:"reconcilationInterval"`
	RestartInterval             time.Duration `json:"-"`
	RestartIntervalHelper       string        `json:"restartInterval"`
	Include                     Matcher       `json:"include"`
	Exclude                     Matcher       `json:"exclude"`
}

type Matcher struct {
	Enabled   bool       `json:"enabled"`
	Selectors []Selector `json:"selectors"`
}

type Selector struct {
	Namespace   string            `json:"namespace"`
	MatchLabels map[string]string `json:"matchLabels"`
}

// GetConfig reads and parses the configuration from the configuration file
func GetConfig(cf string) (*Config, error) {
	cfg := &Config{}
	content, err := ioutil.ReadFile(cf)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config file %v, %w", cf, err)
	}
	err = yaml.Unmarshal(content, cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config file %v, %w", cf, err)
	}
	if cfg.RestartIntervalHelper != "" {
		d, err := time.ParseDuration(cfg.RestartIntervalHelper)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse duration %v in config file %v, %w", cfg.RestartIntervalHelper, cf, err)

		}
		cfg.RestartInterval = d
	}
	if cfg.ReconcilationIntervalHelper != "" {
		d, err := time.ParseDuration(cfg.ReconcilationIntervalHelper)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse duration %v in config file %v, %w", cfg.ReconcilationIntervalHelper, cf, err)

		}
		cfg.ReconcilationInterval = d
	}
	return cfg, nil
}
