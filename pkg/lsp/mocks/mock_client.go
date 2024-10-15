package mocks

import (
	"context"

	"github.com/hide-org/hide/pkg/lsp"
	"github.com/stretchr/testify/mock"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var _ lsp.Client = (*MockClient)(nil)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) GetWorkspaceSymbols(ctx context.Context, params protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]protocol.SymbolInformation), args.Error(1)
}

func (m *MockClient) Initialize(ctx context.Context, params protocol.InitializeParams) (protocol.InitializeResult, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(protocol.InitializeResult), args.Error(1)
}

func (m *MockClient) NotifyInitialized(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClient) NotifyDidOpen(ctx context.Context, params protocol.DidOpenTextDocumentParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockClient) NotifyDidClose(ctx context.Context, params protocol.DidCloseTextDocumentParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockClient) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClient) GetDocumentOutline(ctx context.Context, params protocol.DocumentSymbolParams) ([]protocol.DocumentSymbol, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]protocol.DocumentSymbol), args.Error(1)
}
