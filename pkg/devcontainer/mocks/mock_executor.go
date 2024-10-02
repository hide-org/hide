package mocks

import (
	"io"

	"github.com/stretchr/testify/mock"

	"github.com/hide-org/hide/pkg/devcontainer"
)

var _ devcontainer.Executor = (*MockExecutor)(nil)

type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Run(command []string, dir string, stdout, stderr io.Writer) error {
	args := m.Called(command, dir, stdout, stderr)
	return args.Error(0)
}
