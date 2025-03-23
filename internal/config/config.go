package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env           string        `yaml:"env"  env:"ENV" env-default:"local" env-required:"true"`
	AccessTTL     time.Duration `yaml:"access_ttl" env-required:"true"`
	AccessSecret  string        `yaml:"access_secret" env-required:"true"`
	RefreshTTL    time.Duration `yaml:"refresh_ttl" env-required:"true"`
	RefreshSecret string        `yaml:"refresh_secret" env-required:"true"`
	GRPCConfig    `yaml:"grpc"`
	Database      `yaml:"database"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-required:"true"`
}

type Database struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password" env-default:"postgres"`
	Name     string `yaml:"name" env-default:"mydb"`
	SSLMode  string `yaml:"ssl_mode" env-default:"disable"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is required")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file: '%s' does not exist", configPath)
	}

	var config Config
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	return &config
}
