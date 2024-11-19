package lsp

import (
	"fmt"
	"io"
	"os/exec"
	"syscall"

	lang "github.com/hide-org/hide/pkg/lsp/v2/languages"
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
	Wait() error
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

func NewProcess(bin lang.Binary) (Process, error) {
	cmd := exec.Command(bin.Name, bin.Arguments...)
	cmd.Env = bin.EnvAsKeyVal()

	// Set SysProcAttr to create a new process group
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
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

func (p *ProcessImpl) Wait() error {
	return p.cmd.Wait()
}
