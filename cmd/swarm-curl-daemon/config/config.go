package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type DaemonConfig struct {
	Addr  string `yaml:"addr" json:"addr"`
	Token string `yaml:"token" json:"token"`
}

func Load(raw []byte) (*DaemonConfig, error) {
	var cfg DaemonConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadFromFile(filePath string) (*DaemonConfig, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return Load(raw)
}
