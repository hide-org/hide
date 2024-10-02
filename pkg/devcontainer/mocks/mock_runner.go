package mocks

import (
	"context"

	"github.com/hide-org/hide/pkg/devcontainer"
)

// MockDevContainerRunner is a mock of the devcontainer.Runner interface for testing
type MockDevContainerRunner struct {
	RunFunc  func(ctx context.Context, projectPath string, config devcontainer.Config) (string, error)
	StopFunc func(ctx context.Context, containerId string) error
	ExecFunc func(ctx context.Context, containerId string, command []string) (devcontainer.ExecResult, error)
}

func (m *MockDevContainerRunner) Run(ctx context.Context, projectPath string, config devcontainer.Config) (string, error) {
	return m.RunFunc(ctx, projectPath, config)
}

func (m *MockDevContainerRunner) Stop(ctx context.Context, containerId string) error {
	return m.StopFunc(ctx, containerId)
}

func (m *MockDevContainerRunner) Exec(ctx context.Context, containerId string, command []string) (devcontainer.ExecResult, error) {
	return m.ExecFunc(ctx, containerId, command)
}
