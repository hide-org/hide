package lsp

import (
	"fmt"
	"io"
	"os/exec"
)

type Command struct {
	name string
	args []string
}

func NewCommand(name string, args []string) Command {
	return Command{name: name, args: args}
}

type Process interface {
	Start() error
	Stop() error
	ReadWriteCloser() io.ReadWriteCloser
}

type readWriteCloser struct {
	io.ReadCloser
	io.WriteCloser
}

func (rwc *readWriteCloser) Close() error {
	rerr := rwc.ReadCloser.Close()
	werr := rwc.WriteCloser.Close()
	if rerr != nil {
		return rerr
	}
	return werr
}

type ProcessImpl struct {
	cmd *exec.Cmd
	rwc io.ReadWriteCloser
}

func NewProcess(command Command) (Process, error) {
	cmd := exec.Command(command.name, command.args...)
	stdin, err := cmd.StdinPipe()

	if err != nil {
		return nil, fmt.Errorf("Failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return nil, fmt.Errorf("Failed to create stdout pipe: %w", err)
	}

	rwc := &readWriteCloser{stdout, stdin}

	return &ProcessImpl{cmd: cmd, rwc: rwc}, nil
}

func (p *ProcessImpl) Start() error {
	return p.cmd.Start()
}

func (p *ProcessImpl) Stop() error {
	if p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Kill()
}

func (p *ProcessImpl) ReadWriteCloser() io.ReadWriteCloser {
	return p.rwc
}
