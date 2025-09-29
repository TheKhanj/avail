package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/thekhanj/avail/common"
	"github.com/thekhanj/avail/config"
)

type DaemonOption = func(ping *Daemon)

type Daemon struct {
	cfg *config.Config
	log *log.Logger
}

func NewDaemon(cfg *config.Config, opts ...DaemonOption) *Daemon {
	base := &Daemon{
		cfg: cfg,
		log: log.New(os.Stderr, "daemon: ", 0),
	}

	for _, o := range opts {
		o(base)
	}

	return base
}

func DaemonWithLog(log *log.Logger) DaemonOption {
	return func(d *Daemon) {
		d.log = log
	}
}

func (this *Daemon) Run(ctx context.Context) error {
	err := this.writePid()
	if err != nil {
		return err
	}
	defer this.cleanup()

	pings := make([]*Ping, len(this.cfg.Pings))
	for i, pingCfg := range this.cfg.Pings {
		ping, err := NewPingFromConfig(&pingCfg)
		if err != nil {
			return err
		}
		pings[i] = ping
	}

	this.runPings(ctx, pings)
	return nil
}

func (this *Daemon) runPings(ctx context.Context, pings []*Ping) {
	var wg sync.WaitGroup

	wg.Add(len(pings))
	for _, ping := range pings {
		go func() {
			defer wg.Done()

			ping.Run(ctx)
		}()
	}
	wg.Wait()
}

func (this *Daemon) cleanup() {
	err := os.Remove(this.cfg.GetPidFile())
	if err != nil {
		this.log.Println(err)
	}
	err = os.Remove(common.GetPidVarDir(syscall.Getpid()))
	if err != nil {
		this.log.Println(err)
	}
}

func (this *Daemon) writePid() error {
	pidFile := this.cfg.GetPidFile()

	stat, err := os.Stat(pidFile)
	if err == nil {
		if stat.IsDir() {
			return fmt.Errorf("PID file already exists and is a directory: %s", pidFile)
		}

		this.log.Printf("warning: PID file already exists: %s\n", pidFile)

		pid, err := common.GetPid(pidFile)
		if err != nil {
			return err
		}
		if common.ProcessExists(pid) {
			return fmt.Errorf("a process with PID %d already exists\n", pid)
		}
	}

	dir := filepath.Dir(pidFile)
	stat, err = os.Stat(dir)
	if err == nil && !stat.IsDir() {
		return fmt.Errorf("file already exists and is not a directory: %s", dir)
	}
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	return os.WriteFile(
		pidFile,
		[]byte(fmt.Sprintf("%d\n", syscall.Getpid())),
		0655,
	)
}
