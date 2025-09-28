package common

import (
	"context"
	"fmt"
	"os"
	"os/signal"
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
