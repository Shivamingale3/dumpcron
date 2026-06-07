package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	BackupRoot    string `yaml:"backup_root"`
	RetentionDays int    `yaml:"retention_days"`
	Jobs          []Job  `yaml:"jobs"`
}

type Job struct {
	Name      string   `yaml:"name"`
	Type      string   `yaml:"type"`
	Host      string   `yaml:"host"`
	Port      int      `yaml:"port"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
	Databases []string `yaml:"databases"`
	Time      string   `yaml:"time"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
