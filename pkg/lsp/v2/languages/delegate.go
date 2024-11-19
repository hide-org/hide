package lang

import (
	"context"
	"errors"
	"fmt"
)

type Delegate interface {
	Get(ctx context.Context, uri string) ([]byte, error)

	ProjectRootPath() string
	ReadFile(ctx context.Context, path string) ([]byte, error)

	MakeInstallPath(ctx context.Context, lspName string, version string) (path string, error error)
}

func NewDefaultDelegate() *defaultDelegate {
	return &defaultDelegate{}
}

type defaultDelegate struct{}

func (*defaultDelegate) Get(ctx context.Context, uri string) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (defaultDelegate) ProjectRootPath() string {
	return ""
}

func (defaultDelegate) ReadFile(ctx context.Context, path string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (defaultDelegate) MakeInstallPath(ctx context.Context, lspName string, version string) (path string, error error) {
	return "", errors.New("not implemented")
}
