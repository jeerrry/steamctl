package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_FileNotExist(t *testing.T) {
	t.Parallel()

	s, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(s.Games) != 0 {
		t.Errorf("expected empty games map, got %d entries", len(s.Games))
	}
}

func TestSaveAndLoad(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(time.Second)
	s := &Store{
		Games: map[int]*GameState{
			440: {
				AppID:            440,
				Name:             "Team Fortress 2",
				AchievementIndex: 5,
				NextUnlockAt:     now.Add(30 * time.Minute),
				StartedAt:        now,
			},
		},
	}

	path := filepath.Join(t.TempDir(), "sub", "state.json")

	if err := s.Save(path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	gs := loaded.GetGame(440)
	if gs == nil {
		t.Fatal("GetGame(440) returned nil")
	}
	if gs.Name != "Team Fortress 2" {
		t.Errorf("Name = %q, want Team Fortress 2", gs.Name)
	}
	if gs.AchievementIndex != 5 {
		t.Errorf("AchievementIndex = %d, want 5", gs.AchievementIndex)
	}
}

func TestSetGame_And_ResetGame(t *testing.T) {
	t.Parallel()

	s := &Store{Games: make(map[int]*GameState)}

	s.SetGame(&GameState{AppID: 730, Name: "CS2"})
	if s.GetGame(730) == nil {
		t.Fatal("expected game 730 to exist")
	}

	s.ResetGame(730)
	if s.GetGame(730) != nil {
		t.Fatal("expected game 730 to be removed")
	}
}

func TestSave_AtomicWrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	s := &Store{Games: map[int]*GameState{
		1: {AppID: 1, Name: "Test"},
	}}

	if err := s.Save(path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify no .tmp file remains
	_, err := os.Stat(path + ".tmp")
	if !os.IsNotExist(err) {
		t.Error("expected .tmp file to not exist after save")
	}

	// Verify actual file exists and is valid JSON
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.GetGame(1) == nil {
		t.Error("expected game 1 to exist after atomic save")
	}
}
