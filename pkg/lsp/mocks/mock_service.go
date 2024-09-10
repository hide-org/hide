package mocks

import (
	"context"

	"github.com/artmoskvin/hide/pkg/lsp"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/stretchr/testify/mock"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

type MockLspService struct {
	mock.Mock
}

func (m *MockLspService) StartServer(ctx context.Context, languageId lsp.LanguageId) error {
	args := m.Called(ctx, languageId)
	return args.Error(0)
}

func (m *MockLspService) StopServer(ctx context.Context, languageId lsp.LanguageId) error {
	args := m.Called(ctx, languageId)
	return args.Error(0)
}

func (m *MockLspService) GetWorkspaceSymbols(ctx context.Context, query string, symbolFilter lsp.SymbolFilter) ([]lsp.SymbolInfo, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]lsp.SymbolInfo), args.Error(1)
}

func (m *MockLspService) NotifyDidOpen(ctx context.Context, file model.File) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockLspService) NotifyDidClose(ctx context.Context, file model.File) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockLspService) GetDiagnostics(ctx context.Context, file model.File) ([]protocol.Diagnostic, error) {
	args := m.Called(ctx, file)
	return args.Get(0).([]protocol.Diagnostic), args.Error(1)
}

func (m *MockLspService) CleanupProject(ctx context.Context, projectId lsp.ProjectId) error {
	args := m.Called(ctx, projectId)
	return args.Error(0)
}
