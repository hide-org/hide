package lang

import (
	"context"
)

type Delegate interface {
	Get(ctx context.Context, uri string) ([]byte, error)
	ProjectRootPath() string
	ReadFile(ctx context.Context, path string) ([]byte, error)

	MakeInstallPath(ctx context.Context, lspName string, version string) (path string, error error)
}

// TODO implement
