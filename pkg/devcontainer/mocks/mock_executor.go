package mocks

import "io"

// MockExecutor is a mock of the util.Executor interface for testing
type MockExecutor struct {
	RunFunc func(command []string, dir string, stdout, stderr io.Writer) error
}

func (m *MockExecutor) Run(command []string, dir string, stdout, stderr io.Writer) error {
	return m.RunFunc(command, dir, stdout, stderr)
}
