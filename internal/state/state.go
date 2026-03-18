// Package state provides JSON-based persistence for tracking achievement unlock progress.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Store holds the full persistent state across all games.
type Store struct {
	Games map[int]*GameState `json:"games"` // keyed by app ID
}

// GameState tracks unlock progress for a single game.
type GameState struct {
	AppID            int       `json:"app_id"`
	Name             string    `json:"name"`
	AchievementIndex int       `json:"achievement_index"` // next achievement to unlock
	NextUnlockAt     time.Time `json:"next_unlock_at"`
	Completed        bool      `json:"completed"`
	StartedAt        time.Time `json:"started_at"`
	CompletedAt      time.Time `json:"completed_at,omitempty"`
}

// Load reads state from disk. Returns an empty store if the file doesn't exist.
func Load(path string) (*Store, error) {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return nil, fmt.Errorf("resolving state path: %w", err)
		}
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Store{Games: make(map[int]*GameState)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading state %s: %w", path, err)
	}

	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("decoding state: %w", err)
	}
	if s.Games == nil {
		s.Games = make(map[int]*GameState)
	}
	return &s, nil
}

// Save writes the store to disk atomically, creating parent directories as needed.
func (s *Store) Save(path string) error {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return fmt.Errorf("resolving state path: %w", err)
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding state: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("writing temp state: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("renaming state file: %w", err)
	}
	return nil
}

// GetGame returns the state for a game, or nil if not tracked.
func (s *Store) GetGame(appID int) *GameState {
	return s.Games[appID]
}

// SetGame creates or updates state for a game.
func (s *Store) SetGame(gs *GameState) {
	s.Games[gs.AppID] = gs
}

// ResetGame removes state for a game.
func (s *Store) ResetGame(appID int) {
	delete(s.Games, appID)
}

// DefaultPath returns ~/.steamctl/state.json.
func DefaultPath() (string, error) {
	home, err := homeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".steamctl", "state.json"), nil
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
