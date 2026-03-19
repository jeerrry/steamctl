//go:build steam && windows

package sdk

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

var (
	steamDLL *syscall.DLL

	procRestartAppIfNecessary *syscall.Proc
	procInitFlat              *syscall.Proc
	procRunCallbacks          *syscall.Proc
	procSteamUserStats        *syscall.Proc
	procSetAchievement        *syscall.Proc
	procStoreStats            *syscall.Proc
)

func loadSteamDLL() error {
	if steamDLL != nil {
		return nil
	}

	dll, err := syscall.LoadDLL("steam_api64.dll")
	if err != nil {
		return fmt.Errorf("load steam_api64.dll: %w", err)
	}
	steamDLL = dll

	procs := []struct {
		ptr  **syscall.Proc
		name string
	}{
		{&procRestartAppIfNecessary, "SteamAPI_RestartAppIfNecessary"},
		{&procInitFlat, "SteamAPI_InitFlat"},
		{&procRunCallbacks, "SteamAPI_RunCallbacks"},
		{&procSteamUserStats, "SteamAPI_SteamUserStats_v013"},
		{&procSetAchievement, "SteamAPI_ISteamUserStats_SetAchievement"},
		{&procStoreStats, "SteamAPI_ISteamUserStats_StoreStats"},
	}

	for _, p := range procs {
		proc, err := dll.FindProc(p.name)
		if err != nil {
			return fmt.Errorf("find %s: %w", p.name, err)
		}
		*p.ptr = proc
	}

	return nil
}

// LiveSDK calls real Steamworks APIs via direct syscall to steam_api64.dll.
type LiveSDK struct {
	initialized bool
	userStats   uintptr
	stopTicker  chan struct{}
}

func NewLiveSDK() SDK {
	return &LiveSDK{}
}

func (l *LiveSDK) Init(appID uint32) error {
	if err := writeAppID(appID); err != nil {
		return fmt.Errorf("write steam_appid.txt: %w", err)
	}

	if err := loadSteamDLL(); err != nil {
		return err
	}

	// SteamAPI_RestartAppIfNecessary(uint32 appID) -> bool
	ret, _, _ := procRestartAppIfNecessary.Call(uintptr(appID))
	if ret != 0 {
		return fmt.Errorf("steam requested app restart for %d — relaunch through Steam", appID)
	}

	// SteamAPI_InitFlat(SteamErrMsg *errMsg) -> ESteamAPIInitResult
	var errMsg [1024]byte
	ret, _, _ = procInitFlat.Call(uintptr(unsafe.Pointer(&errMsg[0])))
	if ret != 0 {
		msg := cString(errMsg[:])
		if msg == "" {
			msg = "unknown error"
		}
		return fmt.Errorf("steamworks init failed (code %d): %s", ret, msg)
	}

	// SteamAPI_SteamUserStats_v013() -> ISteamUserStats*
	l.userStats, _, _ = procSteamUserStats.Call()
	if l.userStats == 0 {
		return fmt.Errorf("SteamUserStats interface not available")
	}

	l.initialized = true
	l.stopTicker = make(chan struct{})

	// Pump callbacks every 5s so Steam keeps showing us as "In-Game".
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				procRunCallbacks.Call()
			case <-l.stopTicker:
				return
			}
		}
	}()

	slog.Info("steamworks initialized", "app_id", appID)
	return nil
}

func (l *LiveSDK) UnlockAchievement(name string) error {
	if !l.initialized {
		return fmt.Errorf("sdk not initialized")
	}

	namePtr, err := syscall.BytePtrFromString(name)
	if err != nil {
		return fmt.Errorf("invalid achievement name: %w", err)
	}

	// ISteamUserStats::SetAchievement(self, const char *name) -> bool
	ret, _, _ := procSetAchievement.Call(l.userStats, uintptr(unsafe.Pointer(namePtr)))
	if ret == 0 {
		return fmt.Errorf("SetAchievement failed for %q", name)
	}

	// ISteamUserStats::StoreStats(self) -> bool
	ret, _, _ = procStoreStats.Call(l.userStats)
	if ret == 0 {
		return fmt.Errorf("StoreStats failed after unlocking %q", name)
	}

	procRunCallbacks.Call()

	slog.Info("achievement unlocked", "achievement", name)
	return nil
}

func (l *LiveSDK) Close() {
	if l.stopTicker != nil {
		close(l.stopTicker)
		l.stopTicker = nil
	}
	_ = os.Remove("steam_appid.txt")
	l.initialized = false
}

func writeAppID(appID uint32) error {
	return os.WriteFile("steam_appid.txt", []byte(strconv.FormatUint(uint64(appID), 10)), 0o644)
}

func cString(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
