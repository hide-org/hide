package devcontainer

import (
	"context"
	"fmt"
	"strings"
)

type SshOption struct {
	key   string
	value string
}

// TODO: refactor using the options pattern
func NewSshOption(key string, value string) SshOption {
	return SshOption{key: key, value: value}
}

type Container interface {
	ID() string
	Stop(ctx context.Context) error
	Exec(ctx context.Context, command []string) (ExecResult, error)
	ExecDetached(ctx context.Context, command []string) (string, error)
	Ssh(ctx context.Context, command []string, opts ...SshOption) (string, error)
}

func GetArchitecture(ctx context.Context, container Container) (string, error) {
	result, err := container.Exec(ctx, []string{"uname", "-m"})
	if err != nil {
		return "", fmt.Errorf("failed to get container architecture: %w", err)
	}

	arch := strings.TrimSpace(result.StdOut)
	switch arch {
	case "x86_64":
		return "amd64", nil
	case "aarch64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("unsupported architecture: %s", arch)
	}
}
