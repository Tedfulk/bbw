package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
	Session  string `mapstructure:"session"`
}

// LoadConfig loads the configuration from the user's home directory
func LoadConfig() (*Config, error) {
	configHome, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Join(configHome, ".config", "bbw")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "config.yaml")

	// Create config file if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		file, err := os.Create(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create config file: %w", err)
		}
		file.Close()
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves the configuration to disk
func SaveConfig(config *Config) error {
	viper.Set("email", config.Email)
	viper.Set("password", config.Password)
	viper.Set("session", config.Session)

	return viper.WriteConfig()
} 