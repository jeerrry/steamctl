// Package sdk wraps the Steamworks SDK for achievement unlocking.
package sdk

import (
	"log/slog"
	"os"
	"strconv"
)

// SDK abstracts Steamworks operations so callers can swap live/stub
// implementations without importing steamworks directly.
type SDK interface {
	Init(appID uint32) error
	UnlockAchievement(name string) error
	Close()
}

// StubSDK logs actions instead of calling Steamworks.
// Used with --stub flag or in tests.
type StubSDK struct{}

func NewStubSDK() SDK {
	return &StubSDK{}
}

func (s *StubSDK) Init(appID uint32) error {
	slog.Info("stub: init", "app_id", appID)
	return nil
}

func (s *StubSDK) UnlockAchievement(name string) error {
	slog.Info("stub: unlock achievement", "achievement", name)
	return nil
}

func (s *StubSDK) Close() {}

func writeAppID(appID uint32) error {
	return os.WriteFile("steam_appid.txt", []byte(strconv.FormatUint(uint64(appID), 10)), 0644)
}
