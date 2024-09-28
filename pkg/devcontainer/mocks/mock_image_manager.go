package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/artmoskvin/hide/pkg/devcontainer"
)

var _ devcontainer.ImageManager = (*MockImageManager)(nil)

type MockImageManager struct {
	mock.Mock
}

func (m *MockImageManager) PullImage(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockImageManager) BuildImage(ctx context.Context, workingDir string, config devcontainer.Config) (string, error) {
	args := m.Called(ctx, workingDir, config)
	return args.String(0), args.Error(1)
}

func (m *MockImageManager) CheckLocalImage(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}
