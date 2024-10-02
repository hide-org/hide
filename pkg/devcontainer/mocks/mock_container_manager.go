package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/hide-org/hide/pkg/devcontainer"
)

var _ devcontainer.ContainerManager = (*MockContainerManager)(nil)

type MockContainerManager struct {
	mock.Mock
}

func (m *MockContainerManager) CreateContainer(ctx context.Context, image string, projectPath string, config devcontainer.Config) (string, error) {
	args := m.Called(ctx, image, projectPath, config)
	return args.String(0), args.Error(1)
}

func (m *MockContainerManager) StartContainer(ctx context.Context, containerId string) error {
	args := m.Called(ctx, containerId)
	return args.Error(0)
}

func (m *MockContainerManager) StopContainer(ctx context.Context, containerId string) error {
	args := m.Called(ctx, containerId)
	return args.Error(0)
}

func (m *MockContainerManager) Exec(ctx context.Context, containerId string, command []string) (devcontainer.ExecResult, error) {
	args := m.Called(ctx, containerId, command)
	return args.Get(0).(devcontainer.ExecResult), args.Error(1)
}
