package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/lsp"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/artmoskvin/hide/pkg/project/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSearchSymbolsHandler_ServeHTTP(t *testing.T) {
	symbols := []lsp.SymbolInfo{
		newSymbolInfo("symbol1", "kind1", "path1"),
		newSymbolInfo("symbol2", "kind2", "path2"),
	}

	symbolJson, _ := json.Marshal(symbols[0])
	symbolsJson, _ := json.Marshal(symbols)

	tests := []struct {
		name              string
		target            string
		mockSearchSymbols func(ctx context.Context, projectId model.ProjectId, query string) ([]lsp.SymbolInfo, error)
		wantStatusCode    int
		wantBody          string
	}{
		{
			name:   "success",
			target: "/projects/123/search?type=symbol&query=test-query",
			mockSearchSymbols: func(ctx context.Context, projectId model.ProjectId, query string) ([]lsp.SymbolInfo, error) {
				return symbols, nil
			},
			wantStatusCode: http.StatusOK,
			wantBody:       string(symbolsJson),
		},
		{
			name:   "success with limit",
			target: "/projects/123/search?type=symbol&query=test-query&limit=1",
			mockSearchSymbols: func(ctx context.Context, projectId model.ProjectId, query string) ([]lsp.SymbolInfo, error) {
				return symbols[:1], nil
			},
			wantStatusCode: http.StatusOK,
			wantBody:       string(symbolJson),
		},
		{
			name:   "success with limit higher than number of symbols",
			target: fmt.Sprintf("/projects/123/search?type=symbol&query=test-query&limit=%d", len(symbols)+1),
			mockSearchSymbols: func(ctx context.Context, projectId model.ProjectId, query string) ([]lsp.SymbolInfo, error) {
				return symbols, nil
			},
			wantStatusCode: http.StatusOK,
			wantBody:       string(symbolsJson),
		},
		{
			name:   "invalid limit",
			target: "/projects/123/search?type=symbol&query=test-query&limit=invalid",
			mockSearchSymbols: func(ctx context.Context, projectId model.ProjectId, query string) ([]lsp.SymbolInfo, error) {
				return nil, nil
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "invalid limit invalid",
		},
		{
			name:   "limit too large",
			target: fmt.Sprintf("/projects/123/search?type=symbol&query=test-query&limit=%d", handlers.MaxLimit+1),
			mockSearchSymbols: func(ctx context.Context, projectId model.ProjectId, query string) ([]lsp.SymbolInfo, error) {
				return nil, nil
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "limit must be between 1 and 100",
		},
		{
			name:   "limit too small",
			target: fmt.Sprintf("/projects/123/search?type=symbol&query=test-query&limit=%d", handlers.MinLimit-1),
			mockSearchSymbols: func(ctx context.Context, projectId model.ProjectId, query string) ([]lsp.SymbolInfo, error) {
				return nil, nil
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "limit must be between 1 and 100",
		},
		{
			name:   "project not found",
			target: "/projects/123/search?type=symbol&query=test-query",
			mockSearchSymbols: func(ctx context.Context, projectId model.ProjectId, query string) ([]lsp.SymbolInfo, error) {
				return nil, project.NewProjectNotFoundError(projectId)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       "project 123 not found",
		},
		{
			name:   "internal server error",
			target: "/projects/123/search?type=symbol&query=test-query",
			mockSearchSymbols: func(ctx context.Context, projectId model.ProjectId, query string) ([]lsp.SymbolInfo, error) {
				return nil, errors.New("internal error")
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "failed to get workspace symbols: internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := &mocks.MockProjectManager{
				SearchSymbolsFunc: tt.mockSearchSymbols,
			}

			handler := handlers.NewSearchSymbolsHandler(mockManager)
			router := handlers.NewRouter().WithSearchSymbolsHandler(handler).Build()

			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.wantBody)
		})
	}
}

func newSymbolInfo(name, kind, path string) lsp.SymbolInfo {
	return lsp.SymbolInfo{
		Name: name,
		Kind: kind,
		Location: lsp.Location{
			Path: path,
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      1,
					Character: 1,
				},
				End: lsp.Position{
					Line:      2,
					Character: 2,
				},
			},
		},
	}
}
