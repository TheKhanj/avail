package main

import (
	"context"
	"flag"
	"fmt"
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
		fmt.Fprintln(os.Stderr, "  avail run")
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
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return CODE_INVALID_CONFIG
	}

	ctx := common.NewSignalCtx(context.Background())
	d := NewDaemon(cfg)

	err = d.Run(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		return CODE_GENERAL_ERR
	}

	return CODE_SUCCESS
}

func (this *Cli) status(args []string) int {
	f := flag.NewFlagSet("avail", flag.ExitOnError)
	help := f.Bool("h", false, "show help")
	cfgPath := f.String("c", this.getDefaultCfg(), "config file")
	optPid := f.Int("P", 0, "PID of the running daemon")
	optPidFile := f.String("p", "", "PID file of the running daemon")

	f.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  avail status")
		fmt.Fprintln(os.Stderr, "  avail status <title...>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		f.PrintDefaults()
	}

	f.Parse(args)

	if *help {
		f.Usage()

		return CODE_SUCCESS
	}

	var err error

	var pid int
	if *optPid != 0 {
		pid = *optPid
	} else {
		var pidFile string
		if *optPidFile != "" {
			pidFile = *optPidFile
		} else {
			cfg, err := config.ReadConfig(*cfgPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return CODE_INVALID_CONFIG
			}
			pidFile = cfg.GetPidFile()
		}
		pid, err = common.GetPid(pidFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return CODE_GENERAL_ERR
		}
	}

	info := NewInfo(pid)

	titles := f.Args()
	if len(titles) == 0 {
		titles, err = info.Titles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return CODE_GENERAL_ERR
		}
	}

	statuses := make([]TitleStatus, len(titles))
	for i, title := range titles {
		s, err := info.Status(title)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return CODE_GENERAL_ERR
		}
		statuses[i] = s
	}

	for _, s := range statuses {
		fmt.Println(s)
	}

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
	fmt.Fprintf(os.Stderr, "error: extra argument: %v\n", arg)
	fmt.Fprintln(os.Stderr, "Try '-h' for help.")

	return CODE_INVALID_INVOKATION
}

func main() {
	c := Cli{args: os.Args[1:]}
	os.Exit(c.Exec())
}
