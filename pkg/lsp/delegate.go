package lsp

import (
	"context"
	"net/http"
)

type Delegate interface {
	Get(ctx context.Context, url string) (*http.Response, error)
	// ProjectRootPath() string
	// ReadFile(ctx context.Context, path string) ([]byte, error)

	MakeInstallPath(ctx context.Context, lspName string, version string) (path string, error error)
}
