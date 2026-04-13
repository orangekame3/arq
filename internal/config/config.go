package config

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds arq configuration.
type Config struct {
	Root      string          `toml:"root"`
	Translate TranslateConfig `toml:"translate"`
	Summarize SummarizeConfig `toml:"summarize"`
}

// TranslateConfig holds LLM translation settings.
type TranslateConfig struct {
	Enabled  bool   `toml:"enabled"`  // auto-translate on get
	Provider string `toml:"provider"` // "anthropic", "openai", or "openrouter"
	Model    string `toml:"model"`    // e.g. "gpt-4o-mini", "claude-haiku-4-5-20251001"
	Lang     string `toml:"lang"`     // target language (default: "Japanese")
	APIKey   string `toml:"api_key"`  // optional, falls back to env var
}

// SummarizeConfig holds LLM summarization settings.
// Falls back to TranslateConfig values for provider, api_key, and lang if not set.
type SummarizeConfig struct {
	Enabled  bool   `toml:"enabled"`  // auto-summarize on get
	Provider string `toml:"provider"` // "anthropic", "openai", or "openrouter"
	Model    string `toml:"model"`    // e.g. "gpt-4o", "claude-sonnet-4-5-20241022"
	Lang     string `toml:"lang"`     // target language (falls back to translate.lang)
	APIKey   string `toml:"api_key"`  // optional, falls back to translate.api_key
	Prompt   string `toml:"prompt"`   // custom instruction prompt (overrides default)
}

// ValidProviders lists all supported LLM providers.
var ValidProviders = []string{"openai", "anthropic", "openrouter"}

// IsValidProvider checks if a provider name is supported.
func IsValidProvider(p string) bool {
	for _, v := range ValidProviders {
		if v == p {
			return true
		}
	}
	return false
}

// Path returns the config file path (~/.config/arq/config.toml).
func Path() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "arq", "config.toml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "arq", "config.toml")
}

// Load reads the config file. Returns zero-value Config if not found.
func Load() Config {
	var c Config
	_, _ = toml.DecodeFile(Path(), &c)
	return c
}

// Save writes the config file.
func Save(c Config) error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(c); err != nil {
		return err
	}
	return os.WriteFile(p, buf.Bytes(), 0o644)
}
