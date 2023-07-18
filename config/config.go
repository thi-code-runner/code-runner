package config

import (
	"encoding/json"
	"log"
	"os"
)

var Conf *Config

type ConfigManager struct {
	Path   string
	Config *Config
}
type ContainerConfig struct {
	ID                     string   `json:"id"`
	Image                  string   `json:"image"`
	CompilationCmd         string   `json:"compilationCmd"`
	ExecutionCmd           string   `json:"executionCmd"`
	ReserveContainerAmount int      `json:"reserveContainerAmount"`
	Add                    []string `json:"add"`
	ReportPath             string   `json:"reportPath"`
	ReportExtractor        string   `json:"reportExtractor"`
}

type Config struct {
	ContainerConfig []ContainerConfig `json:"containerConfig"`
}

func NewConfigManager(path string) *ConfigManager {
	return &ConfigManager{Path: path, Config: &Config{}}
}

func (cr *ConfigManager) ReadConfig() {
	log.Printf("reading config file: %s", cr.Path)
	b, err := os.ReadFile(cr.Path)
	if err != nil {
		log.Fatalf("could not read config file\n%q", err)
	}
	var config Config
	err = json.Unmarshal(b, &config)
	if err != nil {
		log.Fatalf("could not read config file\n%q", err)
	}
	cr.Config = &config
	Conf = &config
}
