package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/thekhanj/avail/common"
	"github.com/thekhanj/avail/config"
)

type PingOption = func(ping *Ping)

func NewPingFromConfig(cfg *config.Ping) (*Ping, error) {
	interval, err := time.ParseDuration(string(cfg.Interval))
	if err != nil {
		return nil, err
	}
	timeout, err := time.ParseDuration(string(cfg.Timeout))
	if err != nil {
		return nil, err
	}

	client := http.DefaultClient
	if cfg.Proxy != nil {
		client, err = NewProxiedHttpClient(string(*cfg.Proxy))
		if err != nil {
			return nil, err
		}
	}

	checkCfg, err := cfg.GetCheck()
	if err != nil {
		return nil, err
	}
	check, err := NewCheckFromConfig(checkCfg)
	if err != nil {
		return nil, err
	}

	return NewPing(
		cfg.Title, cfg.Url,
		PingWithInterval(interval),
		PingWithTimeout(timeout),
		PingWithClient(client),
		PingWithCheck(check),
	)
}

func NewPing(title, url string, opts ...PingOption) (*Ping, error) {
	base := &Ping{
		url:   url,
		title: title,

		log:      log.New(os.Stderr, title+": ", 0),
		interval: time.Second * 5,
		timeout:  time.Second * 30,
		path:     filepath.Join(common.GetPidVarDir(syscall.Getpid()), title),
		client:   http.DefaultClient,

		ch:         make(chan struct{}),
		wasHealthy: false,

		check:     &StatusCheck{},
		firstTime: true,
	}

	for _, o := range opts {
		o(base)
	}

	stat, err := os.Stat(base.path)
	if err == nil && !stat.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", base.path)
	}
	if err != nil {
		err = os.MkdirAll(base.path, 0755)
		if err != nil {
			return nil, err
		}
	}

	return base, nil
}

func PingWithCheck(check Check) PingOption {
	return func(ping *Ping) {
		ping.check = check
	}
}

func PingWithInterval(interval time.Duration) PingOption {
	return func(ping *Ping) {
		ping.interval = interval
	}
}

func PingWithTimeout(timeout time.Duration) PingOption {
	return func(ping *Ping) {
		ping.timeout = timeout
	}
}

func PingWithLog(log *log.Logger) PingOption {
	return func(ping *Ping) {
		ping.log = log
	}
}

func PingWithPath(path string) PingOption {
	return func(ping *Ping) {
		ping.path = path
	}
}

func PingWithClient(client *http.Client) PingOption {
	return func(ping *Ping) {
		ping.client = client
	}
}

type Ping struct {
	url   string
	path  string
	title string

	interval time.Duration
	timeout  time.Duration
	client   *http.Client

	log *log.Logger

	wasHealthy bool
	ch         chan struct{}

	check     Check
	firstTime bool
}

func (this *Ping) Run(ctx context.Context) {
	defer this.cleanup()

	this.log.Printf(
		"running HTTP ping on \"%s\" (path: \"%s\")...\n",
		this.url, this.path,
	)
	defer this.log.Printf(
		"running HTTP ping on \"%s\" (path: \"%s\") done\n",
		this.url, this.path,
	)

	go this.schedule(ctx)

	for range this.ch {
		err := this.checkAvailability(ctx)
		if err != nil {
			this.log.Println(err)
		}
	}
}

func (this *Ping) cleanup() {
	err := os.RemoveAll(this.path)
	if err != nil {
		this.log.Println(err)
	}
}

func (this *Ping) checkAvailability(ctx context.Context) error {
	reqCtx, cancel := context.WithTimeout(ctx, this.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", this.url, nil)
	if err != nil {
		this.update(0, false)
		return err
	}

	before := time.Now()
	res, err := this.client.Do(req)
	after := time.Now()
	latency := after.UnixMilli() - before.UnixMilli()
	if err != nil {
		this.update(latency, false)
		return err
	}

	isUp, err := this.check.IsUp(res)
	if err != nil || !isUp {
		this.update(latency, false)
		return err
	}

	this.update(latency, true)
	return nil
}

func (this *Ping) update(latency int64, health bool) {
	if health {
		this.log.Printf("GET request succeeded (latency: %d ms)\n", latency)
	} else {
		this.log.Printf("GET request failed (latency: %d ms)\n", latency)
	}

	err := os.WriteFile(
		filepath.Join(this.path, "latency"),
		[]byte(fmt.Sprintf("%d\n", latency)),
		0644,
	)
	if err != nil {
		this.log.Println(err)
	}

	if this.firstTime || this.wasHealthy != health {
		this.firstTime = false
		content := "0\n"
		if health {
			content = "1\n"
		}

		err := os.WriteFile(
			filepath.Join(this.path, "health"), []byte(content), 0644,
		)
		if err != nil {
			this.log.Println(err)
		}

		this.wasHealthy = health
	}
}

func (this *Ping) schedule(ctx context.Context) {
	defer close(this.ch)

	for {
		this.ch <- struct{}{}

		select {
		case <-ctx.Done():
			return
		case <-time.After(this.interval):
			continue
		}
	}
}
