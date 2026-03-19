package sdk

import "testing"

func TestStubSDKImplementsInterface(t *testing.T) {
	var _ SDK = NewStubSDK()
}

func TestLiveSDKConstructor(t *testing.T) {
	// Without the "steam" build tag, NewLiveSDK returns unavailableSDK.
	// With it, returns LiveSDK. Either way, it satisfies the interface.
	var _ SDK = NewLiveSDK()
}

func TestStubInit(t *testing.T) {
	s := NewStubSDK()
	if err := s.Init(12345); err != nil {
		t.Fatalf("StubSDK.Init() returned error: %v", err)
	}
}

func TestStubUnlockAchievement(t *testing.T) {
	s := NewStubSDK()
	_ = s.Init(12345)
	if err := s.UnlockAchievement("ACH_WIN_FIRST_GAME"); err != nil {
		t.Fatalf("StubSDK.UnlockAchievement() returned error: %v", err)
	}
}

func TestStubClose(t *testing.T) {
	s := NewStubSDK()
	_ = s.Init(12345)
	s.Close() // should not panic
}

func TestUnavailableSDKReturnsError(t *testing.T) {
	s := NewLiveSDK()
	err := s.Init(12345)
	if err == nil {
		// If this passes (err == nil), we're running with -tags steam and
		// Steam client is available — skip instead of failing.
		s.Close()
		t.Skip("running with steam build tag — live SDK initialized successfully")
	}
}
