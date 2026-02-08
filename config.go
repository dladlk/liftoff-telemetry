package main

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	General struct {
		Save        bool  `toml:"save"`
		SaveEachNth int32 `toml:"saveEachNth"`
	} `toml:"general"`
	Log struct {
		LogToFile bool `toml:"logToFile"`
		Debug     bool `toml:"debug"`
	} `toml:"log"`
}

func LoadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := Config{}

	if err := toml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
