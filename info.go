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

type SiteStatusList []*SiteStatus

func (this *SiteStatusList) Apply(opts ...SiteStatusOption) {
	for _, s := range *this {
		s.Apply(opts...)
	}
}

func (this SiteStatusList) String() string {
	lines := make([]string, len(this))

	for i, s := range this {
		lines[i] = s.String()
	}

	return strings.Join(lines, "\n")
}

func (this *SiteStatusList) MaxTitleLength(opts ...SiteStatusOption) int {
	max := 0
	for _, s := range *this {
		if len(s.title) > max {
			max = len(s.title)
		}
	}
	return max
}

var _ fmt.Stringer = (*SiteStatusList)(nil)

type SiteStatusOption = func(*SiteStatus)

func SiteStatusWithTitleLength(titleLength int) SiteStatusOption {
	return func(s *SiteStatus) {
		s.titleLength = titleLength
	}
}

type SiteStatus struct {
	title       string
	titleLength int

	Latency int64
	Health  bool
}

func (this *SiteStatus) Apply(opts ...SiteStatusOption) {
	for _, o := range opts {
		o(this)
	}
}

func (this SiteStatus) String() string {
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

	titleLength := this.titleLength
	if titleLength == 0 {
		titleLength = 20
	}

	return fmt.Sprintf(
		"%s%-"+strconv.Itoa(titleLength+1)+"s%s %s%s%s (latency: %s%d ms%s)",
		titleColor, this.title+":", colorReset,
		healthColor, health, colorReset,
		latencyColor, this.Latency, colorReset,
	)
}

var _ fmt.Stringer = (*SiteStatus)(nil)

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

func (this *Info) SitesStatus(titles []string) (SiteStatusList, error) {
	ret := make(SiteStatusList, len(titles))
	for i, title := range titles {
		s, err := this.SiteStatus(title)
		if err != nil {
			return nil, err
		}
		ret[i] = &s
	}

	titleLength := ret.MaxTitleLength()
	ret.Apply(SiteStatusWithTitleLength(titleLength))

	return ret, nil
}

func (this *Info) SiteStatus(title string) (SiteStatus, error) {
	ret := SiteStatus{title: title}
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
