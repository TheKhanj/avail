package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/thekhanj/avail/common"
	"github.com/thekhanj/avail/config"
)

var VERSION = "dev"

const (
	CODE_SUCCESS int = iota
	CODE_GENERAL_ERR
	CODE_INVALID_CONFIG
	CODE_INVALID_INVOKATION
	CODE_INITIALIZATION_FAILED
)

type Cli struct {
	args []string
}

func (this *Cli) Exec() int {
	f := flag.NewFlagSet("avail", flag.ExitOnError)
	help := f.Bool("h", false, "show help")
	version := f.Bool("v", false, "show version")

	f.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  avail -h")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Available Commands:")
		fmt.Fprintln(os.Stderr, "  run       run the daemon")
		fmt.Fprintln(os.Stderr, "  status    show status of sites")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		f.PrintDefaults()
	}

	f.Parse(this.args)

	if *help {
		f.Usage()

		return CODE_SUCCESS
	}

	if *version {
		fmt.Println(VERSION)

		return CODE_SUCCESS
	}

	if len(f.Args()) == 0 {
		return this.notEnoughArguments()
	}

	cmd := f.Args()[0]
	switch cmd {
	case "run":
		return this.run(f.Args()[1:])
	case "status":
		return this.status(f.Args()[1:])
	default:
		fmt.Fprintf(os.Stderr, "error: invalid command \"%s\"\n", cmd)
		return CODE_INVALID_INVOKATION
	}
}

func (this *Cli) run(args []string) int {
	f := flag.NewFlagSet("avail", flag.ExitOnError)
	help := f.Bool("h", false, "show help")
	cfgPath := f.String("c", this.getDefaultCfg(), "config file")

	f.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  run -c avail.json")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		f.PrintDefaults()
	}

	f.Parse(args)

	if *help {
		f.Usage()

		return CODE_SUCCESS
	}

	if len(f.Args()) != 0 {
		return this.extraArgument(f.Arg(0))
	}

	cfg, err := config.ReadConfig(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return CODE_INVALID_CONFIG
	}

	ctx := common.NewSignalCtx(context.Background())
	d := NewDaemon(cfg)

	err = d.Run(ctx)
	if err != nil {
		log.Printf("error: %v", err)
		return CODE_GENERAL_ERR
	}

	return CODE_SUCCESS
}

func (this *Cli) status(args []string) int {
	// TODO: implement this
	return CODE_SUCCESS
}

func (this *Cli) getDefaultCfg() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(os.Getenv("HOME"), ".config")
	}

	locations := [...]string{
		"avail.json",
		filepath.Join(configHome, "avail.json"),
		"/etc/avail.json",
	}

	for _, path := range locations {
		_, err := os.Stat(path)
		if err == nil {
			return path
		}
	}

	return locations[0]
}

func (this *Cli) notEnoughArguments() int {
	fmt.Fprintln(os.Stderr, "error: not enough arguments.")
	fmt.Fprintln(os.Stderr, "Try '-h' for help.")

	return CODE_INVALID_INVOKATION
}

func (this *Cli) extraArgument(arg string) int {
	fmt.Fprintf(os.Stderr, "error: extra argument: %s\n", arg)
	fmt.Fprintln(os.Stderr, "Try '-h' for help.")

	return CODE_INVALID_INVOKATION
}

func main() {
	c := Cli{args: os.Args[1:]}
	os.Exit(c.Exec())
}
