package app

import (
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SoundPath           string
	RandomSoundInterval int64
	RandomSounds        []string
	Categories          Categories
}

func LoadConfig(soundPath string) (Config, error) {
	pathTo := path.Join(soundPath, "config.yaml")
	configBytes, err := os.ReadFile(pathTo)
	if err != nil {
		return Config{}, err
	}
	out := Config{}
	err = yaml.Unmarshal(configBytes, &out)
	if err != nil {
		return Config{}, err
	}
	out.SoundPath = soundPath
	if out.RandomSoundInterval <= 0 {
		out.RandomSoundInterval = 800
	}
	return out, nil
}
