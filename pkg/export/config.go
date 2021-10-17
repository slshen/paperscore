package export

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	UserEmail     string `yaml:"user_email"`
	JSONKeyFile   string `yaml:"json_key_file"`
	SpreadsheetID string `yaml:"spreadsheet_id"`

	jsonKey []byte
}

func NewConfig() (*Config, error) {
	config := &Config{}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".softball")
	configFile := filepath.Join(dir, "config.yaml")
	dat, err := os.ReadFile(configFile)
	if err == nil {
		if err := yaml.Unmarshal(dat, config); err != nil {
			return nil, err
		}
		log.Default().Printf("Loaded config from %s", configFile)
	}
	config.jsonKey = []byte(os.Getenv("SOFTBALL_SHEET_JSON_KEY"))
	if len(config.jsonKey) == 0 && config.JSONKeyFile != "" {
		jsonKeyFile := filepath.Join(dir, "google_key.json")
		config.jsonKey, err = os.ReadFile(jsonKeyFile)
		if err != nil {
			return nil, err
		}
	}
	if len(config.jsonKey) == 0 {
		return nil, fmt.Errorf("no google service account key found")
	}
	if user := os.Getenv("SOFTBALL_SHEET_USER"); user != "" {
		config.UserEmail = user
	}
	if id := os.Getenv("SOFTBALL_SHEET_ID"); id != "" {
		config.SpreadsheetID = id
	}
	return config, nil
}
