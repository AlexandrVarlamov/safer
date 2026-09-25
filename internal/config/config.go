package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
    AccountID       string `env:"ACCOUNT_ID,required"`
    APIToken        string `env:"API_TOKEN,required"`
    AccessKeyID     string `env:"ACCESS_KEY_ID,required"`
    SecretAccessKey string `env:"SECRET_ACCESS_KEY,required"`
}

func Load() (*Config, error) {
    if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
        return nil, fmt.Errorf("load .env: %w", err)
    }
    var cfg Config
    if err := env.Parse(&cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }
    return &cfg, nil
}