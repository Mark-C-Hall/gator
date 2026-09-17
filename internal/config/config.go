package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFile = ".gatorconfig.json"

type Config struct {
	DB_URL          string `json:"db_url,omitempty"`
	CurrentUserName string `json:"current_user_name,omitempty"`
}

func Read() (*Config, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return nil, fmt.Errorf("error getting config file path: %w", err)
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file: %w", err)
	}

	var config = &Config{}
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file json: %w", err)
	}

	return config, nil
}

func (c *Config) SetUser(username string) error {
	c.CurrentUserName = username
	err := c.write()
	if err != nil {
		return fmt.Errorf("error writing config: %w", err)
	}
	return nil
}

func (c *Config) write() error {
	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling json data: %w", err)
	}

	configFilePath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("error getting config file path: %w", err)
	}

	err = os.WriteFile(configFilePath, data, 0o644)
	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

func getConfigFilePath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to find homedir: %w", err)
	}

	configFilePath := filepath.Join(homedir, configFile)
	return configFilePath, nil
}
