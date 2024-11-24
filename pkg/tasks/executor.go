package tasks

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"

	"github.com/rs/zerolog/log"
)

type Executor interface {
	Run(command []string, dir string) (result Result, err error)
}

type ExecutorImpl struct{}

func NewExecutorImpl() Executor {
	return &ExecutorImpl{}
}

func (e *ExecutorImpl) Run(command []string, dir string) (result Result, err error) {
	log.Debug().Msgf("> %s", command)

	if len(command) == 0 {
		return result, fmt.Errorf("command is empty")
	}

	cmnd := command[0]

	var args []string
	if len(command) > 0 {
		args = command[1:]
	}

	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	cmd := exec.Command(cmnd, args...)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return Result{StdOut: stdout.String(), StdErr: stderr.String(), ExitCode: exitError.ExitCode()}, nil
	}
	if err != nil {
		return result, err
	}

	return Result{StdOut: stdout.String(), StdErr: stderr.String(), ExitCode: 0}, nil
}
