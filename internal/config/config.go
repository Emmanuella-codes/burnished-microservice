package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServerAddress string
	OpenAIKey     string
	Debug         bool
	TempDir       string
	MaxFileSize   int64
}

func Load() (*Config, error) {
	serverPort := getEnvWithDefault("SERVER_PORT", "8080")
	serverHost := getEnvWithDefault("SERVER_HOST", "0.0.0.0")

	maxFileSizeStr := getEnvWithDefault("MAX_FILE_SIZE_MB","10")
	maxFileSizeMB, err := strconv.ParseInt(maxFileSizeStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_FILE_SIZE_MB: %v", err)
	}

	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	cfg := &Config{
		ServerAddress: fmt.Sprintf("%s:%s", serverHost, serverPort),
		OpenAIKey:     openAIKey,
		Debug: 				 getEnvWithDefault("DEBUG", "false") == "true",
		TempDir: 			 getEnvWithDefault("TEMP_DIR", "/tmp/burnished"),
		MaxFileSize:   maxFileSizeMB * 1024 * 1024,
	}

	if err := os.MkdirAll(cfg.TempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %v", err)
	}
	return cfg, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
