package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/thekhanj/avail/exec"
)

type Check interface {
	IsUp(res *http.Response) (bool, error)
}

type StatusCheck struct{}

func (this *StatusCheck) IsUp(res *http.Response) (bool, error) {
	return 200 <= res.StatusCode && res.StatusCode < 300, nil
}

var _ Check = (*StatusCheck)(nil)

type ExecCheck struct {
	stdin   io.Reader
	command string
	log     *log.Logger
}

func (this *ExecCheck) IsUp(res *http.Response) (bool, error) {
	name, err := this.writeRawHttp(res)
	if err != nil {
		return false, err
	}
	defer os.Remove(name)

	e, err := exec.New(
		exec.WithShlex(this.command),
		exec.WithEnv(fmt.Sprintf("AVAIL_HTTP=%s", name)),
		func(e *exec.Exec) error {
			if this.log != nil {
				return exec.WithLogger(this.log)(e)
			}

			return nil
		},
		func(e *exec.Exec) error {
			if this.stdin != nil {
				return exec.WithStdin(this.stdin)(e)
			}

			return nil
		},
	)
	if err != nil {
		return false, err
	}

	exitCode, err := e.Run()
	if err != nil {
		return false, err
	}

	return exitCode == 0, nil
}

func (this *ExecCheck) writeRawHttp(res *http.Response) (string, error) {
	tmp, err := os.CreateTemp(os.TempDir(), "avail-http-*")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	err = res.Write(tmp)
	if err != nil {
		return "", err
	}

	return tmp.Name(), nil
}

var _ Check = (*ExecCheck)(nil)

type ShellCheck struct {
	shell  string
	script string
	log    *log.Logger
}

func (this *ShellCheck) IsUp(res *http.Response) (bool, error) {
	e := ExecCheck{}
	e.stdin = bytes.NewReader([]byte(this.script))
	e.command = this.shell
	e.log = this.log

	return e.IsUp(res)
}

var _ Check = (*ShellCheck)(nil)
