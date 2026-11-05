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

// LoadEnv は環境変数と .env から設定を読む。出力先のディレクトリは作らない
func LoadEnv() *Config {
	// .env ファイルを読み込み（存在しなくてもエラーにしない）
	_ = godotenv.Load()

	return &Config{
		Org:       os.Getenv("SHIRABERU_ORG"),
		Format:    getEnvOrDefault("SHIRABERU_FORMAT", "markdown"),
		OutputDir: getEnvOrDefault("SHIRABERU_OUTPUT_DIR", "./output"),
	}
}

// Load は設定を読み、出力先のディレクトリを作る
func Load() (*Config, error) {
	cfg := LoadEnv()

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
