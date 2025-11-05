package exec

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"os/exec"

	"github.com/google/shlex"
)

type Option = func(e *Exec) error

var ErrEmptyCommand = errors.New("Empty command")

func WithLogger(l *log.Logger) Option {
	return func(e *Exec) error {
		e.log = l
		return nil
	}
}

func WithCommand(command string, args ...string) Option {
	return func(e *Exec) error {
		e.command = command
		e.args = args
		return nil
	}
}

func WithShlex(command string) Option {
	return func(e *Exec) error {
		parts, err := shlex.Split(command)
		if err != nil {
			return err
		}

		if len(parts) == 0 {
			return ErrEmptyCommand
		}

		e.command = parts[0]
		e.args = parts[1:]
		return nil
	}
}

func New(opts ...Option) (*Exec, error) {
	e := &Exec{
		command: "",
		args:    []string{},
		log:     nil,
	}

	var err error
	for _, opt := range opts {
		err = opt(e)
		if err != nil {
			return nil, err
		}
	}

	if e.command == "" {
		return nil, ErrEmptyCommand
	}

	return e, nil
}

type Exec struct {
	command string
	args    []string
	log     *log.Logger
}

func (this *Exec) RunContext(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, this.command, this.args...)

	return this.run(cmd)
}

func (this *Exec) Run() error {
	cmd := exec.Command(this.command, this.args...)

	return this.run(cmd)
}

func (this *Exec) run(cmd *exec.Cmd) error {
	if this.log == nil {
		err := cmd.Start()
		if err != nil {
			return err
		}
	} else {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}

		err = cmd.Start()
		if err != nil {
			return err
		}

		go this.flush("stderr", stderr)
		go this.flush("stdout", stdout)
	}

	return cmd.Wait()
}

func (this *Exec) flush(name string, r io.ReadCloser) {
	defer r.Close()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		this.log.Printf("[%s]: %s", name, line)
	}
}
