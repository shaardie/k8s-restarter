package pkg

import (
	"fmt"
	"io/ioutil"
	"time"

	"k8s.io/apimachinery/pkg/util/yaml"
)

type Config struct {
	RestartInterval       time.Duration `json:"-"`
	RestartIntervalHelper string        `json:"restartInterval"`
	ExcludeNamespaces     []string      `json:"excludeNamespaces"`
	IncludeAnnotation     string        `json:"includeAnnotation"`
	ExcludeAnnotation     string        `json:"excludeAnnotation"`
}

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
	return cfg, nil
}
