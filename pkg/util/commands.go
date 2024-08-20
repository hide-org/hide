package util

import (
	"fmt"
	"io"
	"os/exec"

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

func (e *ExecutorImpl) Run(command []string, dir string, stdout, stderr io.Writer) error {
	log.Debug().Msgf("> %s", command)

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = dir
	cmdStdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Error().Err(err).Msg("Error creating StdoutPipe")
		return fmt.Errorf("Error creating StdoutPipe: %w", err)
	}

	cmdStderr, err := cmd.StderrPipe()
	if err != nil {
		log.Error().Err(err).Msg("Error creating StderrPipe")
		return fmt.Errorf("Error creating StderrPipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		log.Error().Err(err).Msg("Error starting command")
		return fmt.Errorf("Error starting command: %w", err)
	}

	// Stream the command output
	// TODO: should we use logger instead?
	go ReadOutput(cmdStdout, stdout)
	go ReadOutput(cmdStderr, stderr)

	// Wait for the command to complete
	// TODO: do async wait
	if err := cmd.Wait(); err != nil {
		log.Error().Err(err).Msg("Error waiting for command")
		return fmt.Errorf("Error waiting for command: %w", err)
	}

	return nil
}

func ReadOutput(src io.Reader, dst io.Writer) error {
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}
