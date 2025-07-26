package main

import (
	"log"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SoundPath    string
	Categories   Categories
	ReplaceWords map[string]string `yaml:"replace_words"`
	MaxMemory    int               `yaml:"max_memory"`
}

func LoadConfig(soundPath string) (Config, error) {
	out := Config{SoundPath: soundPath, MaxMemory: 10485760}
	pathTo := path.Join(soundPath, "config.yaml")
	configBytes, err := os.ReadFile(pathTo)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("> WARNING: Config at %s not found.", pathTo)
			return out, nil
		}
		return out, err
	}
	return out, yaml.Unmarshal(configBytes, &out)
}
