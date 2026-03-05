package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port           string
	DeepSeekAPIKey string
	MaxFileSize    int64
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:        "8080",
		MaxFileSize: 10 * 1024 * 1024, // Default: 10MB.
	}

	if port := os.Getenv("PORT"); port != "" {
		cfg.Port = port
	}

	cfg.DeepSeekAPIKey = os.Getenv("DEEPSEEK_API_KEY")
	if cfg.DeepSeekAPIKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY environment variable is required")
	}

	if maxFileSizeStr := os.Getenv("MAX_FILE_SIZE"); maxFileSizeStr != "" {
		size, err := strconv.ParseInt(maxFileSizeStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid MAX_FILE_SIZE value %q: %w", maxFileSizeStr, err)
		}
		if size <= 0 {
			return nil, fmt.Errorf("MAX_FILE_SIZE must be positive, got %d", size)
		}
		cfg.MaxFileSize = size
	}

	return cfg, nil
}
