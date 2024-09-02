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
		projectID          string
		filePath           string
		mockDeleteFileFunc func(ctx context.Context, projectId, path string) error
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:      "successful deletion",
			projectID: "123",
			filePath:  "test.txt",
			mockDeleteFileFunc: func(ctx context.Context, projectId, path string) error {
				return nil
			},
			expectedStatusCode: http.StatusNoContent,
			expectedBody:       "",
		},
		{
			name:      "invalid file path",
			projectID: "123",
			filePath:  "",
			mockDeleteFileFunc: func(ctx context.Context, projectId, path string) error {
				return nil
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid file path",
		},
		{
			name:      "project not found",
			projectID: "123",
			filePath:  "test.txt",
			mockDeleteFileFunc: func(ctx context.Context, projectId, path string) error {
				return project.NewProjectNotFoundError(projectId)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "project 123 not found",
		},
		{
			name:      "file not found",
			projectID: "123",
			filePath:  "test.txt",
			mockDeleteFileFunc: func(ctx context.Context, projectId, path string) error {
				return files.NewFileNotFoundError(path)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "file test.txt not found",
		},
		{
			name:      "internal server error",
			projectID: "123",
			filePath:  "test.txt",
			mockDeleteFileFunc: func(ctx context.Context, projectId, path string) error {
				return errors.New("internal error")
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Failed to delete file",
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

			req := httptest.NewRequest("DELETE", "/projects/"+tt.projectID+"/files/"+tt.filePath, nil)
			rr := httptest.NewRecorder()

			router := handlers.NewRouter().WithDeleteFileHandler(handler).Build()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}
