package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const validTOML = `
[steam]
api_key = "TESTKEY123"
steam_id = "00000000000000000"

[[games]]
app_id = 3764200
name = "RE Requiem"
idle = true
unlock_achievements = true
time_range = "30h"
min_interval = "10m"
max_interval = "2h"

[[games]]
app_id = 440
name = "Team Fortress 2"
idle = true
unlock_achievements = false
`

func TestLoad(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(validTOML), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Steam.APIKey != "TESTKEY123" {
		t.Errorf("APIKey = %q, want TESTKEY123", cfg.Steam.APIKey)
	}
	if cfg.Steam.SteamID != "00000000000000000" {
		t.Errorf("SteamID = %q, want 00000000000000000", cfg.Steam.SteamID)
	}
	if len(cfg.Games) != 2 {
		t.Fatalf("len(Games) = %d, want 2", len(cfg.Games))
	}

	g := cfg.Games[0]
	if g.AppID != 3764200 {
		t.Errorf("Games[0].AppID = %d, want 3764200", g.AppID)
	}
	if g.TimeRange.Duration != 30*time.Hour {
		t.Errorf("Games[0].TimeRange = %v, want 30h", g.TimeRange.Duration)
	}
	if g.MinInterval.Duration != 10*time.Minute {
		t.Errorf("Games[0].MinInterval = %v, want 10m", g.MinInterval.Duration)
	}
	if g.MaxInterval.Duration != 2*time.Hour {
		t.Errorf("Games[0].MaxInterval = %v, want 2h", g.MaxInterval.Duration)
	}
}

func TestLoad_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name: "missing api_key",
			content: "[steam]\n" +
				"steam_id = \"123\"\n",
			wantErr: "api_key is required",
		},
		{
			name: "missing steam_id",
			content: "[steam]\n" +
				"api_key = \"KEY\"\n",
			wantErr: "steam_id is required",
		},
		{
			name: "missing app_id",
			content: "[steam]\n" +
				"api_key = \"KEY\"\n" +
				"steam_id = \"123\"\n" +
				"\n" +
				"[[games]]\n" +
				"name = \"Test\"\n",
			wantErr: "app_id is required",
		},
		{
			name: "min > max interval",
			content: "[steam]\n" +
				"api_key = \"KEY\"\n" +
				"steam_id = \"123\"\n" +
				"\n" +
				"[[games]]\n" +
				"app_id = 1\n" +
				"unlock_achievements = true\n" +
				"time_range = \"10h\"\n" +
				"min_interval = \"3h\"\n" +
				"max_interval = \"1h\"\n",
			wantErr: "min_interval must be <= max_interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := filepath.Join(dir, "config.toml")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}

			_, err := Load(path)
			if err == nil {
				t.Fatal("Load() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Steam: SteamConfig{
			APIKey:  "ROUNDTRIP",
			SteamID: "765611980",
		},
		Games: []Game{
			{
				AppID:              999,
				Name:               "Test Game",
				Idle:               true,
				UnlockAchievements: true,
				TimeRange:          Duration{24 * time.Hour},
				MinInterval:        Duration{5 * time.Minute},
				MaxInterval:        Duration{1 * time.Hour},
			},
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "config.toml")

	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.Steam.APIKey != "ROUNDTRIP" {
		t.Errorf("APIKey = %q, want ROUNDTRIP", loaded.Steam.APIKey)
	}
	if len(loaded.Games) != 1 || loaded.Games[0].AppID != 999 {
		t.Errorf("unexpected games: %+v", loaded.Games)
	}
}
