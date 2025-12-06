package config

import (
	"os"
)

type Config struct {
	OuraAPIKey   string
	ProcessorURL string
	LogLevel     string
}

func Load() *Config {
	return &Config{
		OuraAPIKey:   os.Getenv("OURA_API_KEY"),
		ProcessorURL: os.Getenv("PROCESSOR_URL"),
		LogLevel:     os.Getenv("LOG_LEVEL"),
	}
}
