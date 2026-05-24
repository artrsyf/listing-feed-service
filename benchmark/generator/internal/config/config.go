package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Users      int `yaml:"users"`
	Orders     int `yaml:"orders"`
	OrderItems int `yaml:"order_items"`
	Products   int `yaml:"products"`
	Categories int `yaml:"categories"`

	BatchSize int `yaml:"batch_size"`
	Workers   int `yaml:"workers"`

	TimeRangeDays int `yaml:"time_range_days"`

	Seed int64 `yaml:"seed"`
}

func Default() Config {
	return Config{
		Users:      1000000,
		Orders:     10000000,
		OrderItems: 50000000,
		Products:   1000000,
		Categories: 10000,

		BatchSize: 50000,
		Workers:   8,

		TimeRangeDays: 365,

		Seed: 42,
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		return &cfg, fmt.Errorf("cannot read config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return &cfg, fmt.Errorf("invalid yaml: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return &cfg, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Users <= 0 {
		return fmt.Errorf("users must be > 0")
	}
	if c.Orders <= 0 {
		return fmt.Errorf("orders must be > 0")
	}
	if c.OrderItems <= 0 {
		return fmt.Errorf("order_items must be > 0")
	}
	if c.OrderItems < c.Orders {
		return fmt.Errorf("order_items must be >= orders so every order can have at least one item")
	}
	if c.Products <= 0 {
		return fmt.Errorf("products must be > 0")
	}
	if c.Categories <= 0 {
		return fmt.Errorf("categories must be > 0")
	}
	if c.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be > 0")
	}
	if c.Workers <= 0 {
		return fmt.Errorf("workers must be > 0")
	}
	if c.TimeRangeDays <= 0 {
		return fmt.Errorf("time_range_days must be > 0")
	}

	return nil
}
