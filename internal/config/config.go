package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	JobScraper  JobScraperConfig
	TelexAPIKey string
}

type JobScraperConfig struct {
	RAPID_API_KEY      string
	RAPID_API_HOST     string
	RAPID_API_BASE_URL string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port: getEnv("PORT", "8080"),
		JobScraper: JobScraperConfig{
			RAPID_API_KEY:      os.Getenv("RAPID_API_KEY"),
			RAPID_API_HOST:     os.Getenv("RAPID_API_HOST"),
			RAPID_API_BASE_URL: os.Getenv("RAPID_API_BASE_URL"),
		},
		TelexAPIKey: os.Getenv("TELEX_API_KEY"),
	}

	// Validate required fields
	if cfg.JobScraper.RAPID_API_KEY == "" {
		return nil, fmt.Errorf("RAPID_API_KEY is required")
	}
	if cfg.JobScraper.RAPID_API_HOST == "" {
		return nil, fmt.Errorf("RAPID_API_HOST is required")
	}
	if cfg.JobScraper.RAPID_API_BASE_URL == "" {
		return nil, fmt.Errorf("RAPID_API_BASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
