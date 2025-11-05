package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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

type PidFlags struct {
	cfgPath    string
	optPid     int
	optPidFile string
}

func (this *PidFlags) SetFlags(f *flag.FlagSet) {
	f.StringVar(&this.cfgPath, "c", common.GetDefaultCfg(), "config file")
	f.IntVar(&this.optPid, "P", 0, "PID of the running daemon")
	f.StringVar(&this.optPidFile, "p", "", "PID file of the running daemon")
}

func (this *PidFlags) GetPid() (int, error) {
	if this.optPid != 0 {
		return this.optPid, nil
	} else {
		var pidFile string
		if this.optPidFile != "" {
			pidFile = this.optPidFile
		} else {
			cfg, err := config.ReadConfig(this.cfgPath)
			if err != nil {
				return 0, err
			}
			pidFile = cfg.GetPidFile()
		}
		pid, err := common.GetPid(pidFile)
		if err != nil {
			return 0, err
		}
		return pid, nil
	}
}

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
		fmt.Fprintln(os.Stderr, "  list      list sites")
		fmt.Fprintln(os.Stderr, "  schema    show http address of config's json schema")
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
	case "list":
		return this.list(f.Args()[1:])
	case "http":
		return this.http(f.Args()[1:])
	case "schema":
		return this.schema(f.Args()[1:])
	default:
		fmt.Fprintf(os.Stderr, "error: invalid command \"%s\"\n", cmd)
		return CODE_INVALID_INVOKATION
	}
}

func (this *Cli) run(args []string) int {
	f := flag.NewFlagSet("avail", flag.ExitOnError)
	help := f.Bool("h", false, "show help")
	cfgPath := f.String("c", common.GetDefaultCfg(), "config file")

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
	pf := PidFlags{}
	pf.SetFlags(f)

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

	pid, err := pf.GetPid()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return CODE_GENERAL_ERR
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

	statuses, err := info.SitesStatus(titles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return CODE_GENERAL_ERR
	}

	fmt.Println(statuses)
	return CODE_SUCCESS
}

func (this *Cli) list(args []string) int {
	f := flag.NewFlagSet("avail", flag.ExitOnError)
	help := f.Bool("h", false, "show help")
	pf := PidFlags{}
	pf.SetFlags(f)

	f.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  avail list")
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

	pid, err := pf.GetPid()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return CODE_GENERAL_ERR
	}

	info := NewInfo(pid)
	titles, err := info.Titles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return CODE_GENERAL_ERR
	}

	fmt.Println(strings.Join(titles, "\n"))
	return CODE_SUCCESS
}

func (this *Cli) http(args []string) int {
	f := flag.NewFlagSet("avail", flag.ExitOnError)
	help := f.Bool("h", false, "show help")
	pf := PidFlags{}
	pf.SetFlags(f)

	f.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  avail http <command> [arguments]")
		fmt.Fprintln(os.Stderr)

		fmt.Fprintln(os.Stderr, "Available Commands:")
		fmt.Fprintln(os.Stderr, "  status    get http status line")
		fmt.Fprintln(os.Stderr, "  header    get a header value")
		fmt.Fprintln(os.Stderr, "  body      get response body")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Description:")
		fmt.Fprintln(os.Stderr, "  reads a raw http response from the file set in AVAIL_HTTP and extracts parts of it")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  avail http status")
		fmt.Fprintln(os.Stderr, "  avail http header content-type")
		fmt.Fprintln(os.Stderr, "  avail http body")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		f.PrintDefaults()
	}

	f.Parse(args)

	if *help {
		f.Usage()

		return CODE_SUCCESS
	}

	if len(f.Args()) == 0 {
		return this.notEnoughArguments()
	}

	filePath, ok := os.LookupEnv("AVAIL_HTTP")
	if !ok {
		fmt.Fprintln(os.Stderr, "error: AVAIL_HTTP environment variable is not set")
		return CODE_GENERAL_ERR
	}
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return CODE_GENERAL_ERR
	}
	defer file.Close()
	buf := bufio.NewReader(file)
	res, err := http.ReadResponse(buf, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return CODE_GENERAL_ERR
	}
	defer res.Body.Close()

	cmd := f.Args()[0]
	switch cmd {
	case "status":
		return this.httpStatus(res, f.Args()[1:])
	case "header":
		return this.httpHeader(res, f.Args()[1:])
	case "body":
		return this.httpBody(res, f.Args()[1:])
	default:
		fmt.Fprintf(os.Stderr, "error: invalid command \"%s\"\n", cmd)
		return CODE_INVALID_INVOKATION
	}
}

func (this *Cli) httpStatus(res *http.Response, args []string) int {
	f := flag.NewFlagSet("avail", flag.ExitOnError)
	help := f.Bool("h", false, "show help")

	f.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  avail http status")
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

	fmt.Println(res.StatusCode)
	return CODE_SUCCESS
}

func (this *Cli) httpHeader(res *http.Response, args []string) int {
	f := flag.NewFlagSet("avail", flag.ExitOnError)
	help := f.Bool("h", false, "show help")

	f.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  avail schema")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		f.PrintDefaults()
	}

	f.Parse(args)

	if *help {
		f.Usage()

		return CODE_SUCCESS
	}

	if len(f.Args()) == 0 {
		return this.notEnoughArguments()
	}
	if len(f.Args()) != 1 {
		return this.extraArgument(f.Arg(1))
	}

	key := f.Arg(0)
	fmt.Println(res.Header.Get(key))
	return CODE_SUCCESS
}

func (this *Cli) httpBody(res *http.Response, args []string) int {
	f := flag.NewFlagSet("avail", flag.ExitOnError)
	help := f.Bool("h", false, "show help")

	f.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  avail body")
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

	_, err := io.Copy(os.Stdout, res.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		return CODE_GENERAL_ERR
	}

	return CODE_SUCCESS
}

func (this *Cli) schema(args []string) int {
	f := flag.NewFlagSet("avail", flag.ExitOnError)
	help := f.Bool("h", false, "show help")

	f.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  avail schema")
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

	fmt.Println(common.GetJsonSchemaAddress(VERSION))
	return CODE_SUCCESS
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
