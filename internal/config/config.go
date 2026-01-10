package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	Org       string
	Format    string
	OutputDir string
}

func Load() (*Config, error) {
	// .env ファイルを読み込み（存在しなくてもエラーにしない）
	_ = godotenv.Load()

	cfg := &Config{
		Org:       os.Getenv("SHIRABERU_ORG"),
		Format:    getEnvOrDefault("SHIRABERU_FORMAT", "markdown"),
		OutputDir: getEnvOrDefault("SHIRABERU_OUTPUT_DIR", "./output"),
	}

	if cfg.OutputDir != "" {
		if len(cfg.OutputDir) >= 2 && cfg.OutputDir[:2] == "~/" {
			cfg.OutputDir = filepath.Join(os.Getenv("HOME"), cfg.OutputDir[2:])
		}
		if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
