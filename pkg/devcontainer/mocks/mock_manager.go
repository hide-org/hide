package mocks

import (
	"github.com/artmoskvin/hide/pkg/devcontainer"
)

// MockDevContainerManager is a mock type for devcontainer.Manager
type MockDevContainerManager struct {
	StartContainerFunc         func(projectPath string, config devcontainer.Config) (devcontainer.Container, error)
	FindContainerByProjectFunc func(projectId string) (devcontainer.Container, error)
	StopContainerFunc          func(containerId string) error
	ExecFunc                   func(containerId string, projectPath string, command string) (devcontainer.ExecResult, error)
}

func (m *MockDevContainerManager) StartContainer(projectPath string, config devcontainer.Config) (devcontainer.Container, error) {
	return m.StartContainerFunc(projectPath, config)
}

func (m *MockDevContainerManager) FindContainerByProject(projectId string) (devcontainer.Container, error) {
	return m.FindContainerByProjectFunc(projectId)
}

func (m *MockDevContainerManager) StopContainer(containerId string) error {
	return m.StopContainerFunc(containerId)
}

func (m *MockDevContainerManager) Exec(containerId string, projectPath string, command string) (devcontainer.ExecResult, error) {
	return m.ExecFunc(containerId, projectPath, command)
}
