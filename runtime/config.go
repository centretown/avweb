package runtime

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	Actions   map[string]*Action `json:"actions"`
	Locations []*Location        `json:"locations"`
}

func NewConfig() (cfg *Config) {
	cfg = &Config{
		Actions:   make(map[string]*Action),
		Locations: make([]*Location, 0),
	}
	return cfg
}

func (cfg *Config) Read(filename string) error {
	fs, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fs.Close()

	buf, err := io.ReadAll(fs)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, cfg)
}

func (cfg *Config) Write(filename string) error {
	buf, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	var perm os.FileMode = 0660
	err = os.WriteFile(filename, buf, perm)
	return err
}
