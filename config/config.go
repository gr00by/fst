package config

import (
	"encoding/json"
	"errors"
	"os"
	"os/user"
	"path"
)

const fileName = ".fst.cfg"

var (
	errConfigLoad = errors.New("failed to load config file, use `fst config` to run configuration setup")
	errConfigSave = errors.New("failed to save config file")
)

// Config stores config file structure.
type Config struct {
	AWSCredentials AWSCredentials `json:"aws_credentials"`
}

// AWSCredentials stores the AWS credentials.
type AWSCredentials struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

// SaveToFile saves configuration data to file.
func SaveToFile(cfg *Config) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path.Join(usr.HomeDir, fileName), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return errConfigSave
	}
	defer file.Close()

	if err = json.NewEncoder(file).Encode(cfg); err != nil {
		return errConfigSave
	}
	return nil
}

// LoadFromFile loads configuration data from file.
func LoadFromFile() (*Config, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path.Join(usr.HomeDir, fileName))
	if err != nil {
		return nil, errConfigLoad
	}
	defer file.Close()

	cfg := &Config{}
	if err = json.NewDecoder(file).Decode(cfg); err != nil {
		return nil, errConfigLoad
	}

	return cfg, nil
}
