package config

import "time"

type Config struct {
	Server   ServerConfig   `yaml:"Server"`
	Postgres PostgresConfig `yaml:"Postgres"`
	JWT      JWTConfig      `yaml:"JWT"`
	Pepper   string         `yaml:"Pepper"`
	Telegram TelegramConfig `yaml:"Telegram"`
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}
type Client struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"DBName"`
	SSLMode  string `yaml:"sslMode"`
	PgDriver string `yaml:"pgDriver"`
}

type ServerConfig struct {
	AppVersion string `yaml:"appVersion"`
	Host       string `yaml:"host" validate:"required"`
	Port       string `yaml:"port" validate:"required"`
}

type TelegramConfig struct {
	Token      string        `yaml:"token"`
	APIBaseURL string        `yaml:"api_base_url"`
	Timeout    time.Duration `yaml:"timeout"`
}
