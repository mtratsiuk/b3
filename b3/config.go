package b3

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const CONFIG_FILE_NAME = "b3.json"

type Config struct {
	Posts []string `json:"posts"`
}

func NewConfig(rootPath string) (Config, error) {
	data, err := os.ReadFile(filepath.Join(rootPath, CONFIG_FILE_NAME))

	if err != nil {
		return Config{}, fmt.Errorf("failed to read b3 configuration file: %v", err)
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)

	if err != nil {
		return Config{}, fmt.Errorf("failed to parse b3 configuration file: %v", err)
	}

	return cfg, nil
}
