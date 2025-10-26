package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

func NewConfig() (*Config, error) {
	var cfg Config
	v := viper.New()
	v.AddConfigPath("config")
	v.SetConfigName("config")
	v.SetConfigType("yml")
	err := v.ReadInConfig()
	if err != nil {
		slog.Error("fail to read config", slog.Any("error", err))
		return &cfg, err
	}
	err = v.Unmarshal(&cfg)
	if err != nil {
		slog.Error("unable to decode config into struct", slog.Any("error", err))
		return &cfg, err
	}
	return &cfg, nil
}
