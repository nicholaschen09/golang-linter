package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Rules       map[string]RuleConfig `yaml:"rules"`
	Cache       CacheConfig           `yaml:"cache"`
	Output      OutputConfig          `yaml:"output"`
	Concurrency int                   `yaml:"concurrency"`
	EnableAll   bool                  `yaml:"enable_all"`
}

type RuleConfig struct {
	Enabled  bool              `yaml:"enabled"`
	Severity string            `yaml:"severity,omitempty"`
	Options  map[string]any    `yaml:"options,omitempty"`
}

type CacheConfig struct {
	Enabled bool   `yaml:"enabled"`
	Dir     string `yaml:"dir"`
}

type OutputConfig struct {
	Format string `yaml:"format"`
	Color  bool   `yaml:"color"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Rules:       make(map[string]RuleConfig),
		Cache:       CacheConfig{Enabled: true, Dir: filepath.Join(home, ".cache", "glint")},
		Output:      OutputConfig{Format: "text", Color: true},
		Concurrency: runtime.NumCPU(),
		EnableAll:   true,
	}
}

var configNames = []string{".glint.yml", ".glint.yaml", "glint.yml", "glint.yaml"}

func Load(dir string) (*Config, error) {
	for _, name := range configNames {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		return parse(data)
	}
	return DefaultConfig(), nil
}

func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	return parse(data)
}

func parse(data []byte) (*Config, error) {
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = runtime.NumCPU()
	}
	if cfg.Cache.Dir == "" {
		home, _ := os.UserHomeDir()
		cfg.Cache.Dir = filepath.Join(home, ".cache", "glint")
	}
	return cfg, nil
}

func WriteDefault(path string) error {
	cfg := DefaultConfig()
	cfg.EnableAll = false
	cfg.Rules = map[string]RuleConfig{
		"unchecked-error":        {Enabled: true, Severity: "error"},
		"nil-deref":              {Enabled: true, Severity: "error"},
		"shadow-var":             {Enabled: true, Severity: "warning"},
		"naming-convention":      {Enabled: true, Severity: "warning"},
		"import-order":           {Enabled: true, Severity: "info"},
		"line-length":            {Enabled: true, Severity: "warning", Options: map[string]any{"max": 120}},
		"prealloc-slice":         {Enabled: true, Severity: "warning"},
		"unnecessary-conversion": {Enabled: true, Severity: "warning"},
		"hardcoded-secret":       {Enabled: true, Severity: "error"},
		"sql-injection":          {Enabled: true, Severity: "error"},
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}
