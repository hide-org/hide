package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artmoskvin/hide/pkg/files"
	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/artmoskvin/hide/pkg/project/mocks"
	"github.com/stretchr/testify/assert"
)

func TestListFilesHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name              string
		target            string
		mockListFilesFunc func(ctx context.Context, projectId string, showHidden bool, filter files.PatternFilter) ([]*model.File, error)
		wantStatusCode    int
		wantBody          string
	}{
		{
			name:   "successful listing",
			target: "/projects/123/files",
			mockListFilesFunc: func(ctx context.Context, projectId string, showHidden bool, filter files.PatternFilter) ([]*model.File, error) {
				return []*model.File{
					model.EmptyFile("file1.txt"),
					model.EmptyFile("file2.txt"),
				}, nil
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `[{"path":"file1.txt"},{"path":"file2.txt"}]`,
		},
		{
			name:   "successful listing with hidden",
			target: "/projects/123/files?showHidden",
			mockListFilesFunc: func(ctx context.Context, projectId string, showHidden bool, filter files.PatternFilter) ([]*model.File, error) {
				return []*model.File{
					model.EmptyFile("file1.txt"),
					model.EmptyFile("file2.txt"),
				}, nil
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `[{"path":"file1.txt"},{"path":"file2.txt"}]`,
		},
		{
			name:   "successful listing with filtering",
			target: "/projects/123/files?&include=*.txt&exclude=file1",
			mockListFilesFunc: func(ctx context.Context, projectId string, showHidden bool, filter files.PatternFilter) ([]*model.File, error) {
				// TODO: fix.
				return []*model.File{
					model.EmptyFile("file1.txt"),
					model.EmptyFile("file2.txt"),
					model.EmptyFile("file2.json"),
				}, nil
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `[{"path":"file2.txt"}]`,
		},
		{
			name:   "project not found",
			target: "/projects/123/files",
			mockListFilesFunc: func(ctx context.Context, projectId string, showHidden bool, filter files.PatternFilter) ([]*model.File, error) {
				return nil, project.NewProjectNotFoundError(projectId)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       "project 123 not found",
		},
		{
			name:   "internal server error",
			target: "/projects/123/files",
			mockListFilesFunc: func(ctx context.Context, projectId string, showHidden bool, filter files.PatternFilter) ([]*model.File, error) {
				return nil, errors.New("internal error")
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "Failed to list files: internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectManager := &mocks.MockProjectManager{
				ListFilesFunc: tt.mockListFilesFunc,
			}

			handler := handlers.ListFilesHandler{
				ProjectManager: mockProjectManager,
			}

			router := handlers.NewRouter().WithListFilesHandler(handler).Build()

			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.wantBody)
		})
	}
}
