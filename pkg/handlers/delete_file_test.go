package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artmoskvin/hide/pkg/files"
	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/artmoskvin/hide/pkg/project/mocks"
	"github.com/stretchr/testify/assert"
)

func TestDeleteFileHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name               string
		target             string
		mockDeleteFileFunc func(ctx context.Context, projectId, path string) error
		wantStatusCode     int
		wantBody           string
	}{
		{
			name:   "successful deletion",
			target: "/projects/123/files/test.txt",
			mockDeleteFileFunc: func(ctx context.Context, projectId, path string) error {
				return nil
			},
			wantStatusCode: http.StatusNoContent,
			wantBody:       "",
		},
		{
			name:   "invalid file path",
			target: "/projects/123/files/",
			mockDeleteFileFunc: func(ctx context.Context, projectId, path string) error {
				return nil
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "Invalid file path",
		},
		{
			name:   "project not found",
			target: "/projects/123/files/test.txt",
			mockDeleteFileFunc: func(ctx context.Context, projectId, path string) error {
				return project.NewProjectNotFoundError(projectId)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       "project 123 not found",
		},
		{
			name:   "file not found",
			target: "/projects/123/files/test.txt",
			mockDeleteFileFunc: func(ctx context.Context, projectId, path string) error {
				return files.NewFileNotFoundError(path)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       "file test.txt not found",
		},
		{
			name:   "internal server error",
			target: "/projects/123/files/test.txt",
			mockDeleteFileFunc: func(ctx context.Context, projectId, path string) error {
				return errors.New("internal error")
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "Failed to delete file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := &mocks.MockProjectManager{
				DeleteFileFunc: tt.mockDeleteFileFunc,
			}

			handler := handlers.DeleteFileHandler{
				ProjectManager: mockPM,
			}

			req := httptest.NewRequest(http.MethodDelete, tt.target, nil)
			rr := httptest.NewRecorder()

			router := handlers.NewRouter().WithDeleteFileHandler(handler).Build()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.wantBody)
		})
	}
}
