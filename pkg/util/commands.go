package util

import (
	"fmt"
	"io"
	"log"
	"os/exec"
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
	log.Println("> ", command)

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = dir
	cmdStdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Printf("Error creating StdoutPipe: %v\n", err)
		return fmt.Errorf("Error creating StdoutPipe: %w", err)
	}

	cmdStderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Error creating StderrPipe: %v\n", err)
		return fmt.Errorf("Error creating StderrPipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Error starting command: %v\n", err)
		return fmt.Errorf("Error starting command: %w", err)
	}

	// Stream the command output
	// TODO: should we use logger instead?
	go StreamOutput(cmdStdout, stdout)
	go StreamOutput(cmdStderr, stderr)

	// Wait for the command to complete
	// TODO: do async wait
	if err := cmd.Wait(); err != nil {
		log.Printf("Error waiting for command: %v\n", err)
		return fmt.Errorf("Error waiting for command: %w", err)
	}

	return nil
}

func StreamOutput(src io.Reader, dst io.Writer) error {
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("Error streaming output: %w", err)
	}

	return nil
}
