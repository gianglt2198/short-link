package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	Name        string `mapstructure:"name"`
	Port        string `mapstructure:"port"`
	BaseURL     string `mapstructure:"base_url"`
	Environment string `mapstructure:"environment"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

func NewConfig() (*Config, error) {
	// Load .env into OS env (best-effort; ignored if file absent)
	_ = godotenv.Load()

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	// dot-separated viper keys → underscore-separated env vars
	// e.g. server.base_url → SERVER__BASE_URL, database.url → DATABASE__URL
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}

var Module = fx.Provide(NewConfig)
