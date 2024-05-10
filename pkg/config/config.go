package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const CONFIG_FILE_NAME = "b3.json"

type Config struct {
	PostsGlob     []string           `json:"posts_glob"`
	OutDirPath    string             `json:"out_dir_path"`
	AssetsDirPath []string           `json:"assets_dir_path"`
	HomeLink      string             `json:"home_link"`
	HeaderLinks   []ConfigHeaderLink `json:"header_links"`
}

type ConfigHeaderLink struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func New(rootPath string) (Config, error) {
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
