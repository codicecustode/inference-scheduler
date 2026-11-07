package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port       int
	NumGPUs    int
	VRAMPerGPU int
	RateLimit  int
	APIKey     string
}

func Load() *Config {
	return &Config{
		Port:       getEnvInt("PORT", 8080),
		NumGPUs:    getEnvInt("NUM_GPUS", 4),
		VRAMPerGPU: getEnvInt("VRAM_PER_GPU", 16384),
		RateLimit:  getEnvInt("RATE_LIMIT", 100),
		APIKey:     getEnvStr("API_KEY", "dev-key-change-me"),
	}
}

func getEnvStr(key, fallback string) string {

	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value

}

func getEnvInt(key string, fallback int) int {

	value, err := strconv.Atoi(os.Getenv(key))

	if err != nil {
		return fallback
	}

	return value

}
