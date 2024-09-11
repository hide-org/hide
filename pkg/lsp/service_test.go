package lsp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/artmoskvin/hide/pkg/lsp"
	"github.com/artmoskvin/hide/pkg/lsp/mocks"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestService_GetWorkspaceSymbols(t *testing.T) {
	aSymbol := protocol.SymbolInformation{Name: "test-name", Kind: protocol.SymbolKindClass, Location: protocol.Location{URI: "file:///test/project/test-uri", Range: protocol.Range{Start: protocol.Position{Line: 0, Character: 0}, End: protocol.Position{Line: 1, Character: 1}}}}

	includeKind := protocol.SymbolKindClass
	includedSymbol := aSymbol

	excludeKind := protocol.SymbolKindField
	excludedSymbol := protocol.SymbolInformation{Name: "test-name", Kind: excludeKind, Location: protocol.Location{URI: "file:///test/project/test-uri", Range: protocol.Range{Start: protocol.Position{Line: 0, Character: 0}, End: protocol.Position{Line: 1, Character: 1}}}}

	tests := []struct {
		name         string
		ctx          context.Context
		query        string
		symbolFilter lsp.SymbolFilter
		mockSetup    func(*mocks.MockClientPool)
		wantSymbols  []lsp.SymbolInfo
		wantErr      string
	}{
		{
			name:  "success",
			ctx:   model.NewContextWithProject(context.Background(), &model.Project{Id: "project-id", Path: "/test/project"}),
			query: "test-query",
			mockSetup: func(m *mocks.MockClientPool) {
				client := &mocks.MockClient{}
				client.On("GetWorkspaceSymbols", mock.MatchedBy(isContext), protocol.WorkspaceSymbolParams{Query: "test-query"}).Return([]protocol.SymbolInformation{aSymbol}, nil)

				m.On("GetAllForProject", "project-id").Return(map[lsp.LanguageId]lsp.Client{lsp.LanguageId("test-lang"): client}, true)
			},
			wantSymbols: []lsp.SymbolInfo{{Name: "test-name", Kind: "Class", Location: lsp.Location{Path: "test-uri", Range: lsp.Range{Start: lsp.Position{Line: 1, Character: 0}, End: lsp.Position{Line: 2, Character: 1}}}}},
		},
		{
			name:      "project not in context",
			ctx:       context.Background(),
			query:     "test-query",
			mockSetup: func(m *mocks.MockClientPool) {},
			wantErr:   "project not found in context",
		},
		{
			name:        "client not found",
			ctx:         model.NewContextWithProject(context.Background(), &model.Project{Id: "project-id", Path: "/test/project"}),
			query:       "test-query",
			mockSetup:   func(m *mocks.MockClientPool) { m.On("GetAllForProject", "project-id").Return(nil, false) },
			wantSymbols: nil,
			wantErr:     "",
		},
		{
			name: "context cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(model.NewContextWithProject(context.Background(), &model.Project{Id: "project-id", Path: "/test/project"}))
				cancel()

				return ctx
			}(),
			query: "test-query",
			mockSetup: func(m *mocks.MockClientPool) {
				client := &mocks.MockClient{}
				m.On("GetAllForProject", "project-id").Return(map[lsp.LanguageId]lsp.Client{lsp.LanguageId("test-lang"): client}, true)
			},
			wantSymbols: nil,
			wantErr:     "context cancelled",
		},
		{
			name:  "client error",
			ctx:   model.NewContextWithProject(context.Background(), &model.Project{Id: "project-id", Path: "/test/project"}),
			query: "test-query",
			mockSetup: func(m *mocks.MockClientPool) {
				client := &mocks.MockClient{}
				client.On("GetWorkspaceSymbols", mock.MatchedBy(isContext), protocol.WorkspaceSymbolParams{Query: "test-query"}).Return(nil, errors.New("test-error"))

				m.On("GetAllForProject", "project-id").Return(map[lsp.LanguageId]lsp.Client{lsp.LanguageId("test-lang"): client}, true)
			},
			wantErr: "test-error",
		},
		{
			name:  "exclude symbols",
			ctx:   model.NewContextWithProject(context.Background(), &model.Project{Id: "project-id", Path: "/test/project"}),
			query: "test-query",
			mockSetup: func(m *mocks.MockClientPool) {
				client := &mocks.MockClient{}
				client.On("GetWorkspaceSymbols", mock.MatchedBy(isContext), protocol.WorkspaceSymbolParams{Query: "test-query"}).Return([]protocol.SymbolInformation{aSymbol, excludedSymbol}, nil)

				m.On("GetAllForProject", "project-id").Return(map[lsp.LanguageId]lsp.Client{lsp.LanguageId("test-lang"): client}, true)
			},
			symbolFilter: lsp.NewExcludeSymbolFilter(excludeKind),
			wantSymbols:  []lsp.SymbolInfo{{Name: "test-name", Kind: "Class", Location: lsp.Location{Path: "test-uri", Range: lsp.Range{Start: lsp.Position{Line: 1, Character: 0}, End: lsp.Position{Line: 2, Character: 1}}}}},
		},
		{
			name:  "include symbols",
			ctx:   model.NewContextWithProject(context.Background(), &model.Project{Id: "project-id", Path: "/test/project"}),
			query: "test-query",
			mockSetup: func(m *mocks.MockClientPool) {
				client := &mocks.MockClient{}
				client.On("GetWorkspaceSymbols", mock.MatchedBy(isContext), protocol.WorkspaceSymbolParams{Query: "test-query"}).Return([]protocol.SymbolInformation{includedSymbol}, nil)

				m.On("GetAllForProject", "project-id").Return(map[lsp.LanguageId]lsp.Client{lsp.LanguageId("test-lang"): client}, true)
			},
			symbolFilter: lsp.NewIncludeSymbolFilter(includeKind),
			wantSymbols:  []lsp.SymbolInfo{{Name: "test-name", Kind: "Class", Location: lsp.Location{Path: "test-uri", Range: lsp.Range{Start: lsp.Position{Line: 1, Character: 0}, End: lsp.Position{Line: 2, Character: 1}}}}},
		},
		{
			name:  "fail to remove file prefix",
			ctx:   model.NewContextWithProject(context.Background(), &model.Project{Id: "project-id", Path: "/test/project"}),
			query: "test-query",
			mockSetup: func(m *mocks.MockClientPool) {
				client := &mocks.MockClient{}
				client.On("GetWorkspaceSymbols", mock.MatchedBy(isContext), protocol.WorkspaceSymbolParams{Query: "test-query"}).Return([]protocol.SymbolInformation{{Name: "test-name", Kind: protocol.SymbolKindClass, Location: protocol.Location{URI: "/test/project/test-uri", Range: protocol.Range{Start: protocol.Position{Line: 0, Character: 0}, End: protocol.Position{Line: 1, Character: 1}}}}}, nil)

				m.On("GetAllForProject", "project-id").Return(map[lsp.LanguageId]lsp.Client{lsp.LanguageId("test-lang"): client}, true)
			},
			wantErr: "failed to remove file prefix from URI",
		},
		{
			name:  "fail to get relative path of file",
			ctx:   model.NewContextWithProject(context.Background(), &model.Project{Id: "project-id", Path: "invalid-path"}),
			query: "test-query",
			mockSetup: func(m *mocks.MockClientPool) {
				client := &mocks.MockClient{}
				client.On("GetWorkspaceSymbols", mock.MatchedBy(isContext), protocol.WorkspaceSymbolParams{Query: "test-query"}).Return([]protocol.SymbolInformation{aSymbol}, nil)

				m.On("GetAllForProject", "project-id").Return(map[lsp.LanguageId]lsp.Client{lsp.LanguageId("test-lang"): client}, true)
			},
			wantErr: "failed to get relative path of file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClientPool := &mocks.MockClientPool{}
			tt.mockSetup(mockClientPool)

			service := lsp.NewService(nil, nil, nil, mockClientPool)

			symbols, err := service.GetWorkspaceSymbols(tt.ctx, tt.query, tt.symbolFilter)

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSymbols, symbols)
			}

		})
	}
}

func isContext(ctx interface{}) bool {
	_, ok := ctx.(context.Context)
	return ok
}
