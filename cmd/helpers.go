package cmd

import (
	"fmt"
	"os"
	"runtime"
)

func homeDir() (string, error) {
	if runtime.GOOS == "windows" {
		home := os.Getenv("USERPROFILE")
		if home != "" {
			return home, nil
		}
		return "", fmt.Errorf("USERPROFILE not set")
	}
	home := os.Getenv("HOME")
	if home != "" {
		return home, nil
	}
	return "", fmt.Errorf("HOME not set")
}
