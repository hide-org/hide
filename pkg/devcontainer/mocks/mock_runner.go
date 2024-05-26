package mocks

import "github.com/artmoskvin/hide/pkg/devcontainer"

// MockDevContainerRunner is a mock of the devcontainer.Runner interface for testing
type MockDevContainerRunner struct {
	RunFunc  func(projectPath string, config devcontainer.Config) (string, error)
	StopFunc func(containerId string) error
	ExecFunc func(containerId string, command []string) (devcontainer.ExecResult, error)
}

func (m *MockDevContainerRunner) Run(projectPath string, config devcontainer.Config) (string, error) {
	return m.RunFunc(projectPath, config)
}

func (m *MockDevContainerRunner) Stop(containerId string) error {
	return m.StopFunc(containerId)
}

func (m *MockDevContainerRunner) Exec(containerId string, command []string) (devcontainer.ExecResult, error) {
	return m.ExecFunc(containerId, command)
}
