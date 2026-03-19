// Package sdk wraps the Steamworks SDK for achievement unlocking.
package sdk

import "log/slog"

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

// NewStubSDK returns a StubSDK that logs instead of calling Steamworks.
func NewStubSDK() SDK {
	return &StubSDK{}
}

// Init logs the app ID without contacting Steam.
func (s *StubSDK) Init(appID uint32) error {
	slog.Info("stub: init", "app_id", appID)
	return nil
}

// UnlockAchievement logs the achievement name without contacting Steam.
func (s *StubSDK) UnlockAchievement(name string) error {
	slog.Info("stub: unlock achievement", "achievement", name)
	return nil
}

// Close is a no-op for StubSDK.
func (s *StubSDK) Close() {}
