package task

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

type Executor interface {
	Run(command []string, dir string, stdout, stderr io.Writer) error
}

type ExecutorImpl struct {
}

func NewExecutorImpl() Executor {
	return &ExecutorImpl{}
}

func (e *ExecutorImpl) Run(command []string, dir string) (stdout io.ReadCloser, stderr io.ReadCloser, err error) {
	log.Debug().Msgf("> %s", command)

	if len(command) == 0 {
		return nil, nil, fmt.Errorf("command is empty")
	}
	cmnd := command[0]
	var args []string
	if len(command) > 0 {
		args = command[1:]
	}

	cmd := exec.Command(cmnd, args...)
	cmd.Dir = dir

	err := cmd.Run()
	if errors.Is(err, &exec.ExitError{}) {
		v := err.(*exec.ExitError)
		exitCode := v.ExitCode()
		stdErr := string(v.Stderr)
	}
	// cmdStdout, err := cmd.StdoutPipe()

	// if err != nil {
	// 	log.Error().Err(err).Msg("Error creating StdoutPipe")
	// 	return nil, nil, fmt.Errorf("Error creating StdoutPipe: %w", err)
	// }

	// cmdStderr, err := cmd.StderrPipe()
	// if err != nil {
	// 	log.Error().Err(err).Msg("Error creating StderrPipe")
	// 	return nil, nil, fmt.Errorf("Error creating StderrPipe: %w", err)
	// }

	// if err := cmd.Start(); err != nil {
	// 	log.Error().Err(err).Msg("Error starting command")
	// 	return nil, nil, fmt.Errorf("Error starting command: %w", err)
	// }

	// // Wait for the command to complete
	// // TODO: do async wait
	// if err := cmd.Wait(); err != nil {
	// 	log.Error().Err(err).Msg("Error waiting for command")
	// 	return nil, nil, fmt.Errorf("Error waiting for command: %w", err)
	// }

	// return cmdStdout, cmdStderr, nil
}

func readOutput(src io.Reader, dst io.Writer) error {
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}
