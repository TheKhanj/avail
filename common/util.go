package common

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

func GetVarDir() string {
	pid := syscall.Getpid()

	if runtime.GOOS == "windows" {
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = os.TempDir()
		}
		return filepath.Join(base, "avail", fmt.Sprintf("%d", pid))
	}

	uid := syscall.Getuid()
	if uid == 0 {
		return filepath.Join("/var/run/avail", fmt.Sprintf("%d", pid))
	}
	return filepath.Join(
		"/var/run/user",
		fmt.Sprintf("%d", uid), "avail", fmt.Sprintf("%d", pid),
	)
}
