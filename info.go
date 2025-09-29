package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/thekhanj/avail/common"
	"golang.org/x/term"
)

type TitleStatus struct {
	title string

	Latency int64
	Health  bool
}

func (this TitleStatus) String() string {
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	healthColor := ""
	latencyColor := ""
	titleColor := ""
	colorReset := ""

	if isTTY {
		titleColor = ""
		latencyColor = "\x1b[36m"

		if this.Health {
			healthColor = "\x1b[1m\x1b[32m"
		} else {
			healthColor = "\x1b[1m\x1b[31m"
		}

		colorReset = "\x1b[0m"
	}

	health := "OK"
	if !this.Health {
		health = "FAILED"
	}

	return fmt.Sprintf(
		"%s%-20s%s %s%s%s (latency: %s%d ms%s)",
		titleColor, this.title+":", colorReset,
		healthColor, health, colorReset,
		latencyColor, this.Latency, colorReset,
	)
}

var _ fmt.Stringer = (*TitleStatus)(nil)

func NewInfo(pid int) *Info {
	return &Info{pid}
}

type Info struct {
	pid int
}

func (this *Info) Titles() ([]string, error) {
	pings := filepath.Join(common.GetPidVarDir(this.pid))

	titles := make([]string, 0)

	entries, err := os.ReadDir(pings)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			titles = append(titles, entry.Name())
		}
	}
	return titles, nil
}

func (this *Info) Status(title string) (TitleStatus, error) {
	ret := TitleStatus{title: title}
	dir := filepath.Join(common.GetPidVarDir(this.pid), title)

	latencyFile := filepath.Join(dir, "latency")
	b, err := os.ReadFile(latencyFile)
	if err != nil {
		return ret, err
	}
	latency, err := strconv.ParseInt(
		strings.TrimSpace(string(b)),
		10, 64,
	)
	if err != nil {
		return ret, err
	}

	healthFile := filepath.Join(dir, "health")
	b, err = os.ReadFile(healthFile)
	if err != nil {
		return ret, err
	}
	health, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		return ret, err
	}

	ret.Latency = latency
	if health == 0 {
		ret.Health = false
	} else {
		ret.Health = true
	}

	return ret, nil
}
