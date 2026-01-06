package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Org       string `yaml:"org"`
	Period    string `yaml:"period"`
	Format    string `yaml:"format"`
	OutputDir string `yaml:"output_dir"`
}

func Load() (*Config, error) {
	cfg := &Config{
		Period:    "today",
		Format:    "markdown",
		OutputDir: "./output",
	}

	paths := []string{
		filepath.Join(os.Getenv("HOME"), ".config", "shiraberu", "config.yaml"),
		filepath.Join(os.Getenv("HOME"), ".shiraberu.yaml"),
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
		break
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
