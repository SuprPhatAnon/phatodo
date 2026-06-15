package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	LocalDirName  = ".phatodo"
	LocalFileName = "config.json"
)

type LocalConfig struct {
	APIURL       string `json:"api_url"`
	WorkspaceID  string `json:"workspace_id"`
	ProjectID    string `json:"project_id"`
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
}

func DefaultLocalConfig() LocalConfig {
	return LocalConfig{
		APIURL:       "http://localhost:8080",
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "replace-me",
		AccessSecret: "replace-me",
	}
}

func LocalPath(workdir string) string {
	return filepath.Join(workdir, LocalDirName, LocalFileName)
}

func WriteLocal(workdir string, cfg LocalConfig) (string, error) {
	path := LocalPath(workdir)
	if _, err := os.Stat(path); err == nil {
		return path, fmt.Errorf("%s already exists", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return path, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return path, err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return path, err
	}
	data = append(data, '\n')

	return path, os.WriteFile(path, data, 0o600)
}

func ReadLocal(workdir string) (LocalConfig, string, error) {
	path := LocalPath(workdir)
	data, err := os.ReadFile(path)
	if err != nil {
		return LocalConfig{}, path, err
	}

	var cfg LocalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return LocalConfig{}, path, err
	}
	return cfg, path, nil
}
