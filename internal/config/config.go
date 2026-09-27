package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Storage struct {
	AccountID       string `yaml:"account_id"`
	APIToken        string `yaml:"api_token"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
}

type Sqlite struct {
	APIToken   string `yaml:"api_token"`
	DatabaseID string `yaml:"database_id"`
}

type Cloudflare struct {
	AccountID string `yaml:"account_id"`
}

type Config struct {
	Env        string     `yaml:"env"`
	Storage    Storage    `yaml:"storage"`
	Sqlite     Sqlite     `yaml:"sqlite"`
	Cloudflare Cloudflare `yaml:"cloudflare"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	var missing []string

	if c.Env == "" {
		missing = append(missing, "env")
	}
	if c.Cloudflare.AccountID == "" {
		missing = append(missing, "cloudflare.account_id")
	}
	if c.Storage.APIToken == "" {
		missing = append(missing, "storage.api_token")
	}
	if c.Storage.AccessKeyID == "" {
		missing = append(missing, "storage.access_key_id")
	}
	if c.Storage.SecretAccessKey == "" {
		missing = append(missing, "storage.secret_access_key")
	}
	if c.Sqlite.APIToken == "" {
		missing = append(missing, "sqlite.api_token")
	}
	if c.Sqlite.DatabaseID == "" {
		missing = append(missing, "sqlite.database_id")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing fields: %s", strings.Join(missing, ", "))
	}
	return nil
}
