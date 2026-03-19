//go:build !steam

package sdk

import "fmt"

// NewLiveSDK returns an SDK that fails on Init when built without the "steam"
// build tag. Build with `go build -tags steam` to enable real Steamworks calls.
func NewLiveSDK() SDK {
	return &unavailableSDK{}
}

type unavailableSDK struct{}

func (u *unavailableSDK) Init(uint32) error {
	return fmt.Errorf("steamworks not available — rebuild with: go build -tags steam")
}

func (u *unavailableSDK) UnlockAchievement(string) error {
	return fmt.Errorf("steamworks not available")
}

func (u *unavailableSDK) Close() {}
