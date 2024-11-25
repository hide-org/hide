package workspaces

import (
	"context"
	"fmt"
	"strings"
)

type SshOption struct {
	key   string
	value string
}

func NewSshOption(key string, value string) SshOption {
	return SshOption{key: key, value: value}
}

type Workspace interface {
	ID() string
	SSHIsReady(ctx context.Context) (bool, error)
	Ssh(ctx context.Context, command string, opts ...SshOption) (string, error)
	Stop(ctx context.Context) error
}

func GetArchitecture(ctx context.Context, workspace Workspace) (string, error) {
	result, err := workspace.Ssh(ctx, "uname -m")
	if err != nil {
		return "", fmt.Errorf("failed to get container architecture: %w", err)
	}

	arch := strings.TrimSpace(result)
	switch arch {
	case "x86_64":
		return "amd64", nil
	case "aarch64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("unsupported architecture: %s", arch)
	}
}
