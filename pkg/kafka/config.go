package _kafka

import (
	"fmt"
)

type Config struct {
	Brokers []string `json:"brokers" yaml:"brokers"`
	Group   string   `json:"group" yaml:"group"`
	Topics  []string `json:"topics" yaml:"topics"`
}

func DefaultConfig() *Config {
	return &Config{
		Brokers: []string{"localhost:9092"},
		Group:   "test-group",
		Topics:  []string{"test-topic"},
	}
}

func (c *Config) Validate() error {
	if len(c.Brokers) == 0 {
		return fmt.Errorf("brokers is required")
	}
	if c.Group == "" {
		return fmt.Errorf("group is required")
	}
	if len(c.Topics) == 0 {
		return fmt.Errorf("topics is required")
	}

	return nil
}
