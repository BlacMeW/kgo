package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Server.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", config.Server.Port)
	}

	if config.UI.Theme != "dark" {
		t.Errorf("Expected default theme dark, got %s", config.UI.Theme)
	}

	if !config.Features.EnableMetrics {
		t.Error("Expected metrics to be enabled by default")
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	configContent := `
server:
  port: "9090"
  logLevel: "debug"
kubernetes:
  namespace: "test-ns"
ui:
  theme: "light"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load the config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check values
	if config.Server.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", config.Server.Port)
	}

	if config.Server.LogLevel != "debug" {
		t.Errorf("Expected log level debug, got %s", config.Server.LogLevel)
	}

	if config.Kubernetes.Namespace != "test-ns" {
		t.Errorf("Expected namespace test-ns, got %s", config.Kubernetes.Namespace)
	}

	if config.UI.Theme != "light" {
		t.Errorf("Expected theme light, got %s", config.UI.Theme)
	}
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "saved-config.yaml")

	config := DefaultConfig()
	config.Server.Port = "9999"
	config.UI.Theme = "light"

	err := config.SaveConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load it back
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.Server.Port != "9999" {
		t.Errorf("Expected saved port 9999, got %s", loadedConfig.Server.Port)
	}

	if loadedConfig.UI.Theme != "light" {
		t.Errorf("Expected saved theme light, got %s", loadedConfig.UI.Theme)
	}
}
