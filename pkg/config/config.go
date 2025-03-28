package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const CONFIG_FILE_NAME = "b3.json"

type Config struct {
	PostsGlob                []string           `json:"posts_glob"`
	AssetsToUploadRegexp     string             `json:"assets_to_upload_regexp"`
	OutDirPath               string             `json:"out_dir_path"`
	AssetsDirPath            []string           `json:"assets_dir_path"`
	DotEnvPath               string             `json:"dot_env_path"`
	HomeLink                 string             `json:"home_link"`
	HeaderLinks              []ConfigHeaderLink `json:"header_links"`
	DocTitle                 string             `json:"doc_title"`
	DocDescription           string             `json:"doc_description"`
	StripHtmlExtInProdLinks  bool               `json:"strip_html_ext_in_prod_links"`
	TrimPostOgDescriptionsAt int                `json:"trim_post_og_descriptions_at"` // -1 to not trim
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

	cfg := Config{
		TrimPostOgDescriptionsAt: -1,
	}
	err = json.Unmarshal(data, &cfg)

	if err != nil {
		return Config{}, fmt.Errorf("failed to parse b3 configuration file: %v", err)
	}

	return cfg, nil
}
