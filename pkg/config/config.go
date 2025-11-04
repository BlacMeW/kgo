package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server struct {
		Port     string `yaml:"port" json:"port"`
		Host     string `yaml:"host" json:"host"`
		LogLevel string `yaml:"logLevel" json:"logLevel"`
	} `yaml:"server" json:"server"`

	Kubernetes struct {
		Kubeconfig string `yaml:"kubeconfig" json:"kubeconfig"`
		Context    string `yaml:"context" json:"context"`
		Namespace  string `yaml:"namespace" json:"namespace"`
	} `yaml:"kubernetes" json:"kubernetes"`

	UI struct {
		Theme       string `yaml:"theme" json:"theme"`
		AutoRefresh int    `yaml:"autoRefresh" json:"autoRefresh"`
		MaxLogs     int    `yaml:"maxLogs" json:"maxLogs"`
	} `yaml:"ui" json:"ui"`

	Features struct {
		EnableMetrics bool `yaml:"enableMetrics" json:"enableMetrics"`
		EnableExec    bool `yaml:"enableExec" json:"enableExec"`
		EnableLogs    bool `yaml:"enableLogs" json:"enableLogs"`
	} `yaml:"features" json:"features"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	config := &Config{}

	// Server defaults
	config.Server.Port = "8080"
	config.Server.Host = "0.0.0.0"
	config.Server.LogLevel = "info"

	// Kubernetes defaults
	config.Kubernetes.Kubeconfig = ""
	config.Kubernetes.Context = ""
	config.Kubernetes.Namespace = "default"

	// UI defaults
	config.UI.Theme = "dark"
	config.UI.AutoRefresh = 30
	config.UI.MaxLogs = 1000

	// Features defaults
	config.Features.EnableMetrics = true
	config.Features.EnableExec = true
	config.Features.EnableLogs = true

	return config
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	if configPath == "" {
		// Try to find config file in current directory or home directory
		possiblePaths := []string{
			"./kgo.yaml",
			"./kgo.yml",
			"./config.yaml",
			"./config.yml",
			filepath.Join(os.Getenv("HOME"), ".kgo.yaml"),
			filepath.Join(os.Getenv("HOME"), ".kgo.yml"),
			filepath.Join(os.Getenv("HOME"), ".config", "kgo", "config.yaml"),
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}
	}

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %v", configPath, err)
		}

		// Try YAML first
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %v", configPath, err)
		}
	}

	return config, nil
}

// SaveConfig saves configuration to file
func (c *Config) SaveConfig(configPath string) error {
	if configPath == "" {
		configPath = "./kgo.yaml"
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}
