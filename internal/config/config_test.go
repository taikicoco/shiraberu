package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear environment variables
	_ = os.Unsetenv("SHIRABERU_ORG")
	_ = os.Unsetenv("SHIRABERU_FORMAT")
	_ = os.Unsetenv("SHIRABERU_OUTPUT_DIR")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Check defaults
	if cfg.Format != "markdown" {
		t.Errorf("Format: got %q, want %q", cfg.Format, "markdown")
	}
	if cfg.OutputDir != "./output" {
		t.Errorf("OutputDir: got %q, want %q", cfg.OutputDir, "./output")
	}
	if cfg.Org != "" {
		t.Errorf("Org: got %q, want empty", cfg.Org)
	}
}

func TestLoad_WithEnvVars(t *testing.T) {
	// Save and restore env vars
	origOrg := os.Getenv("SHIRABERU_ORG")
	origFormat := os.Getenv("SHIRABERU_FORMAT")
	origOutputDir := os.Getenv("SHIRABERU_OUTPUT_DIR")
	defer func() {
		_ = os.Setenv("SHIRABERU_ORG", origOrg)
		_ = os.Setenv("SHIRABERU_FORMAT", origFormat)
		_ = os.Setenv("SHIRABERU_OUTPUT_DIR", origOutputDir)
	}()

	// Set test values
	_ = os.Setenv("SHIRABERU_ORG", "test-org")
	_ = os.Setenv("SHIRABERU_FORMAT", "html")
	tmpDir := t.TempDir()
	_ = os.Setenv("SHIRABERU_OUTPUT_DIR", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Org != "test-org" {
		t.Errorf("Org: got %q, want %q", cfg.Org, "test-org")
	}
	if cfg.Format != "html" {
		t.Errorf("Format: got %q, want %q", cfg.Format, "html")
	}
	if cfg.OutputDir != tmpDir {
		t.Errorf("OutputDir: got %q, want %q", cfg.OutputDir, tmpDir)
	}
}

func TestLoad_TildeExpansion(t *testing.T) {
	origHome := os.Getenv("HOME")
	origOutputDir := os.Getenv("SHIRABERU_OUTPUT_DIR")
	defer func() {
		_ = os.Setenv("HOME", origHome)
		_ = os.Setenv("SHIRABERU_OUTPUT_DIR", origOutputDir)
	}()

	tmpDir := t.TempDir()
	_ = os.Setenv("HOME", tmpDir)
	_ = os.Setenv("SHIRABERU_OUTPUT_DIR", "~/shiraberu-output")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "shiraberu-output")
	if cfg.OutputDir != expected {
		t.Errorf("OutputDir: got %q, want %q", cfg.OutputDir, expected)
	}

	// Check directory was created
	if _, err := os.Stat(cfg.OutputDir); os.IsNotExist(err) {
		t.Errorf("OutputDir was not created")
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		want         string
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_KEY_NOT_SET",
			defaultValue: "default",
			envValue:     "",
			want:         "default",
		},
		{
			name:         "returns env value when set",
			key:          "TEST_KEY_SET",
			defaultValue: "default",
			envValue:     "custom",
			want:         "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				_ = os.Setenv(tt.key, tt.envValue)
				defer func() { _ = os.Unsetenv(tt.key) }()
			} else {
				_ = os.Unsetenv(tt.key)
			}

			got := getEnvOrDefault(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvOrDefault(%q, %q): got %q, want %q", tt.key, tt.defaultValue, got, tt.want)
			}
		})
	}
}
