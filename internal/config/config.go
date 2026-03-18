package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Steam SteamConfig `toml:"steam"`
	Games []Game      `toml:"games"`
}

type SteamConfig struct {
	APIKey  string `toml:"api_key"`
	SteamID string `toml:"steam_id"`
}

type Game struct {
	AppID              int      `toml:"app_id"`
	Name               string   `toml:"name"`
	Idle               bool     `toml:"idle"`
	UnlockAchievements bool     `toml:"unlock_achievements"`
	TimeRange          Duration `toml:"time_range"`
	MinInterval        Duration `toml:"min_interval"`
	MaxInterval        Duration `toml:"max_interval"`
}

// Duration wraps time.Duration for TOML string parsing (e.g. "30h", "10m").
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", string(text), err)
	}
	return nil
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.Duration.String()), nil
}

func Load(path string) (*Config, error) {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return nil, fmt.Errorf("resolving config path: %w", err)
		}
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Steam.APIKey == "" {
		return fmt.Errorf("steam.api_key is required")
	}
	if c.Steam.SteamID == "" {
		return fmt.Errorf("steam.steam_id is required")
	}
	for i, g := range c.Games {
		if g.AppID == 0 {
			return fmt.Errorf("games[%d].app_id is required", i)
		}
		if g.UnlockAchievements {
			if g.TimeRange.Duration <= 0 {
				return fmt.Errorf("games[%d].time_range must be positive when unlock_achievements is true", i)
			}
			if g.MinInterval.Duration <= 0 {
				return fmt.Errorf("games[%d].min_interval must be positive when unlock_achievements is true", i)
			}
			if g.MaxInterval.Duration <= 0 {
				return fmt.Errorf("games[%d].max_interval must be positive when unlock_achievements is true", i)
			}
			if g.MinInterval.Duration > g.MaxInterval.Duration {
				return fmt.Errorf("games[%d].min_interval must be <= max_interval", i)
			}
		}
	}
	return nil
}

// Save writes the config to the given path, creating parent directories as needed.
func (c *Config) Save(path string) error {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return fmt.Errorf("resolving config path: %w", err)
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	if err := enc.Encode(c); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// DefaultPath returns ~/.steamctl/config.toml.
func DefaultPath() (string, error) {
	home, err := homeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".steamctl", "config.toml"), nil
}

func homeDir() (string, error) {
	if runtime.GOOS == "windows" {
		if h := os.Getenv("USERPROFILE"); h != "" {
			return h, nil
		}
		return "", fmt.Errorf("USERPROFILE not set")
	}
	if h := os.Getenv("HOME"); h != "" {
		return h, nil
	}
	return "", fmt.Errorf("HOME not set")
}
