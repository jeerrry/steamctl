//go:build steam && linux

package sdk

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>

static void *lib;

static bool (*fn_RestartAppIfNecessary)(uint32_t);
static int  (*fn_InitFlat)(char*);
static void (*fn_RunCallbacks)(void);
static void*(*fn_SteamUserStats)(void);
static bool (*fn_SetAchievement)(void*, const char*);
static bool (*fn_StoreStats)(void*);

static const char* load_steam(void) {
	lib = dlopen("./libsteam_api.so", RTLD_NOW);
	if (!lib) lib = dlopen("libsteam_api.so", RTLD_NOW);
	if (!lib) return dlerror();

	fn_RestartAppIfNecessary = dlsym(lib, "SteamAPI_RestartAppIfNecessary");
	fn_InitFlat              = dlsym(lib, "SteamAPI_InitFlat");
	fn_RunCallbacks          = dlsym(lib, "SteamAPI_RunCallbacks");
	fn_SteamUserStats        = dlsym(lib, "SteamAPI_SteamUserStats_v013");
	fn_SetAchievement        = dlsym(lib, "SteamAPI_ISteamUserStats_SetAchievement");
	fn_StoreStats            = dlsym(lib, "SteamAPI_ISteamUserStats_StoreStats");

	if (!fn_RestartAppIfNecessary || !fn_InitFlat || !fn_RunCallbacks ||
	    !fn_SteamUserStats || !fn_SetAchievement || !fn_StoreStats) {
		return "missing required symbol in libsteam_api.so";
	}
	return NULL;
}

static bool call_restart(uint32_t appID)     { return fn_RestartAppIfNecessary(appID); }
static int  call_init(char *errMsg)          { return fn_InitFlat(errMsg); }
static void call_run_callbacks(void)         { fn_RunCallbacks(); }
static void* call_user_stats(void)           { return fn_SteamUserStats(); }
static bool call_set_achievement(void *s, const char *name) { return fn_SetAchievement(s, name); }
static bool call_store_stats(void *s)        { return fn_StoreStats(s); }
*/
import "C"

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
	"unsafe"
)

type LiveSDK struct {
	initialized bool
	userStats   unsafe.Pointer
	stopTicker  chan struct{}
}

func NewLiveSDK() SDK {
	return &LiveSDK{}
}

func (l *LiveSDK) Init(appID uint32) error {
	if err := writeAppID(appID); err != nil {
		return fmt.Errorf("write steam_appid.txt: %w", err)
	}

	cerr := C.load_steam()
	if cerr != nil {
		return fmt.Errorf("load libsteam_api.so: %s", C.GoString(cerr))
	}

	if C.call_restart(C.uint32_t(appID)) {
		return fmt.Errorf("steam requested app restart for %d — relaunch through Steam", appID)
	}

	var errMsg [1024]C.char
	ret := C.call_init(&errMsg[0])
	if ret != 0 {
		return fmt.Errorf("steamworks init failed (code %d): %s", ret, C.GoString(&errMsg[0]))
	}

	l.userStats = C.call_user_stats()
	if l.userStats == nil {
		return fmt.Errorf("SteamUserStats interface not available")
	}

	l.initialized = true
	l.stopTicker = make(chan struct{})

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				C.call_run_callbacks()
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

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if !C.call_set_achievement(l.userStats, cname) {
		return fmt.Errorf("SetAchievement failed for %q", name)
	}

	if !C.call_store_stats(l.userStats) {
		return fmt.Errorf("StoreStats failed after unlocking %q", name)
	}

	C.call_run_callbacks()

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

