package main

import (
	"log"
	"net/http"
	"os"
	"testing"
)

func TestExecCheck(t *testing.T) {
	res, err := http.Get("https://google.com")
	if err != nil {
		t.Fatal(err)
		return
	}

	e := ExecCheck{
		command: "/usr/bin/sh -c 'echo $AVAIL_HTTP'",
		log:     log.New(os.Stderr, "exec-output", 0),
	}
	isUp, err := e.IsUp(res)
	if err != nil {
		t.Fatal(err)
		return
	}

	if !isUp {
		t.Fatal("Host is expected to be considered online!")
		return
	}
}

func TestShellCheck(t *testing.T) {
	res, err := http.Get("https://google.com")
	if err != nil {
		t.Fatal(err)
		return
	}

	e := ShellCheck{
		shell:  "/usr/bin/sh",
		script: "echo $AVAIL_HTTP",
		log:    log.New(os.Stderr, "shell-output", 0),
	}
	isUp, err := e.IsUp(res)
	if err != nil {
		t.Fatal(err)
		return
	}

	if !isUp {
		t.Fatal("Host is expected to be considered online!")
		return
	}
}
