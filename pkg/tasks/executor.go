package task

import (
	"errors"
	"fmt"
	"io"
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

	cmd := exec.Command(cmnd, args...)
	cmd.Dir = dir

	err = cmd.Run()
	if errors.Is(err, &exec.ExitError{}) {
		v := err.(*exec.ExitError)
		result.ExitCode = v.ExitCode()
	}
	if err != nil {
		return result, err
	}

	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		return result, err
	}
	stdOutStr, err := readAll(stdOut)
	if err != nil {
		return result, err
	}
	result.StdOut = stdOutStr

	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return result, err
	}
	stdErrStr, err := readAll(stdErr)
	if err != nil {
		return result, err
	}
	result.StdErr = stdErrStr

	return
}

func readAll(src io.Reader) (string, error) {
	out, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func readOutput(src io.Reader, dst io.Writer) error {
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}
