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
	ID              string  `json:"id"`
	Image           string  `json:"image"`
	CompilationCmd  string  `json:"compilationCmd"`
	ExecutionCmd    string  `json:"executionCmd"`
	ReportPath      string  `json:"reportPath"`
	ReportExtractor string  `json:"reportExtractor"`
	Memory          int64   `json:"memory"`
	CPU             float32 `json:"cpu"`
	ReadOnly        bool    `json:"readOnly"`
	DiskSize        string  `json:"diskSize"`
	Timeout         int     `json:"timeout"`
}

type Config struct {
	ContainerConfig       []ContainerConfig `json:"containerConfig"`
	HostCleanupIntervalS  int64             `json:"hostCleanupIntervalMs"`
	CacheCleanupIntervalS int64             `json:"cacheCleanupIntervalMs"`
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

func GetContainerConfig(id string) *ContainerConfig {
	var containerConf ContainerConfig
	for _, c := range Conf.ContainerConfig {
		if id == c.ID {
			containerConf = c
			break
		}
	}
	return &containerConf
}
