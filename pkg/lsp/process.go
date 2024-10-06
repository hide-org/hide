package lsp

import (
	"fmt"
	"io"
	"os/exec"
	"syscall"
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
	cmd  *exec.Cmd
	rwc  io.ReadWriteCloser
	pgid int
}

func NewProcess(command Command) (Process, error) {
	cmd := exec.Command(command.name, command.args...)

	// Set up process group
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

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
	err := p.cmd.Start()
	if err != nil {
		return err
	}

	// Get the process group ID
	p.pgid, err = syscall.Getpgid(p.cmd.Process.Pid)
	if err != nil {
		return fmt.Errorf("Failed to get process group ID: %w", err)
	}

	return nil
}

func (p *ProcessImpl) Stop() error {
	if p.cmd.Process == nil {
		return nil
	}
	// Kill the entire process group
	err := syscall.Kill(-p.pgid, syscall.SIGKILL)
	if err != nil {
		return fmt.Errorf("Failed to kill process group: %w", err)
	}

	// return p.cmd.Wait()
	return nil
}

func (p *ProcessImpl) ReadWriteCloser() io.ReadWriteCloser {
	return p.rwc
}
