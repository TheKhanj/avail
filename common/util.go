package common

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

func GetBaseVarDir() string {
	if runtime.GOOS == "windows" {
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = os.TempDir()
		}
		return filepath.Join(base, "avail")
	}

	uid := syscall.Getuid()
	if uid == 0 {
		return filepath.Join("/var/run/avail")
	}
	return filepath.Join(
		"/var/run/user",
		fmt.Sprintf("%d", uid), "avail",
	)
}

func GetPidVarDir(pid int) string {
	return filepath.Join(GetBaseVarDir(), fmt.Sprintf("%d", pid))
}

func NewSignalCtx(
	ctx context.Context,
) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)
	signal.Notify(stop, syscall.SIGINT)

	go func() {
		select {
		case <-ctx.Done():
			cancel()
		case <-stop:
			cancel()
		}
	}()

	return ctx
}

func GetPid(pidFile string) (int, error) {
	b, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(strings.TrimSpace(string(b)))
}

func ProcessExists(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		// Unreachable on Unix systems
		return true
	}
	err = proc.Signal(syscall.Signal(0))

	return err == nil
}
