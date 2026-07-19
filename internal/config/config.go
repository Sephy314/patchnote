package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	configDirName  = "patchnote"
	configFileName = "config"
	configType     = "yaml"
)

// Config holds the application configuration.
type Config struct {
	Provider    string  `mapstructure:"provider"`
	Model       string  `mapstructure:"model"`
	Temperature float64 `mapstructure:"temperature"`
	Language    string  `mapstructure:"language"`
	APIKey      string  `mapstructure:"api_key"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Provider:    "groq",
		Model:       "llama-3.3-70b-versatile",
		Temperature: 0.2,
		Language:    "english",
	}
}

// Dir returns the configuration directory path (~/.config/patchnote).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", configDirName), nil
}

// Path returns the full path to the configuration file.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName+"."+configType), nil
}

// Load reads the configuration file and returns a Config.
// If the file does not exist, it returns default values.
func Load() (*Config, error) {
	cfgPath, err := Path()
	if err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigFile(cfgPath)
	v.SetConfigType(configType)

	v.SetDefault("provider", "groq")
	v.SetDefault("model", "llama-3.3-70b-versatile")
	v.SetDefault("temperature", 0.2)
	v.SetDefault("language", "english")

	if err := v.ReadInConfig(); err != nil {
		if isConfigNotFound(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to disk.
func Save(cfg *Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	v := viper.New()
	v.SetConfigFile(filepath.Join(dir, configFileName+"."+configType))
	v.SetConfigType(configType)

	v.Set("provider", cfg.Provider)
	v.Set("model", cfg.Model)
	v.Set("temperature", cfg.Temperature)
	v.Set("language", cfg.Language)
	v.Set("api_key", cfg.APIKey)

	if err := v.WriteConfigAs(filepath.Join(dir, configFileName+"."+configType)); err != nil {
		return fmt.Errorf("cannot write config file: %w", err)
	}

	return nil
}

func isConfigNotFound(err error) bool {
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		return true
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return true
	}
	return false
}
