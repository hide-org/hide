package outline

import (
	"context"
	"path/filepath"

	"github.com/hide-org/hide/pkg/lsp/v2"
	"github.com/hide-org/hide/pkg/model"
)

type Service interface {
	// TODO: can we move model here?
	Get(ctx context.Context, path string) (lsp.DocumentOutline, error)
}

type ServiceImpl struct {
	workspaceDir string
	lsp          lsp.Service
}

func NewService(lsp lsp.Service, workspaceDir string) Service {
	return &ServiceImpl{lsp: lsp, workspaceDir: workspaceDir}
}

func (s *ServiceImpl) Get(ctx context.Context, path string) (lsp.DocumentOutline, error) {
	return s.lsp.GetDocumentOutline(ctx, model.File{Path: "file://" + filepath.Join(s.workspaceDir, path)})
}
