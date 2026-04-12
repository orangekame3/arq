package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds arq configuration.
type Config struct {
	Root string `json:"root"`
}

// Path returns the config file path (~/.config/arq/config.json).
func Path() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "arq", "config.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "arq", "config.json")
}

// Load reads the config file. Returns zero-value Config if not found.
func Load() Config {
	data, err := os.ReadFile(Path())
	if err != nil {
		return Config{}
	}
	var c Config
	_ = json.Unmarshal(data, &c)
	return c
}

// Save writes the config file.
func Save(c Config) error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}
