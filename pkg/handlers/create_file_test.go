package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artmoskvin/hide/pkg/files"
	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/artmoskvin/hide/pkg/project/mocks"
)

func TestCreateFileHandler(t *testing.T) {
	testCases := []struct {
		name           string
		createFileFunc func(ctx context.Context, projectId, path, content string) (*model.File, error)
		requestBody    handlers.CreateFileRequest
		wantStatus     int
		wantFile       *model.File
	}{
		{
			name: "Success",
			createFileFunc: func(ctx context.Context, projectId, path, content string) (*model.File, error) {
				return model.NewFile("/test/path", "test content"), nil
			},
			requestBody: handlers.CreateFileRequest{Path: "/test/path", Content: "test content"},
			wantStatus:  http.StatusCreated,
			wantFile:    func() *model.File { return model.NewFile("/test/path", "test content") }(),
		},
		{
			name: "ProjectNotFound",
			createFileFunc: func(ctx context.Context, projectId, path, content string) (*model.File, error) {
				return nil, project.NewProjectNotFoundError(projectId)
			},
			requestBody: handlers.CreateFileRequest{Path: "/test/path", Content: "test content"},
			wantStatus:  http.StatusNotFound,
		},
		{
			name: "FileAlreadyExists",
			createFileFunc: func(ctx context.Context, projectId, path, content string) (*model.File, error) {
				return nil, files.NewFileAlreadyExistsError(path)
			},
			requestBody: handlers.CreateFileRequest{Path: "/test/path", Content: "test content"},
			wantStatus:  http.StatusConflict,
		},
		{
			name: "InternalServerError",
			createFileFunc: func(ctx context.Context, projectId, path, content string) (*model.File, error) {
				return nil, errors.New("Test error")
			},
			requestBody: handlers.CreateFileRequest{Path: "/test/path", Content: "test content"},
			wantStatus:  http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockManager := &mocks.MockProjectManager{
				CreateFileFunc: tc.createFileFunc,
			}

			handler := handlers.CreateFileHandler{ProjectManager: mockManager}
			router := handlers.NewRouter().WithCreateFileHandler(handler).Build()

			body, _ := json.Marshal(tc.requestBody)
			request, _ := http.NewRequest(http.MethodPost, "/projects/123/files", bytes.NewBuffer(body))
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, response.Code)
			}

			if tc.wantFile != nil {
				var respFile model.File
				if err := json.NewDecoder(response.Body).Decode(&respFile); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if !respFile.Equals(tc.wantFile) {
					t.Errorf("Unwant file returned: %+v", respFile)
				}
			}
		})
	}
}

func TestCreateFileHandler_InvalidPayload(t *testing.T) {
	mockManager := &mocks.MockProjectManager{}

	handler := handlers.CreateFileHandler{ProjectManager: mockManager}
	router := handlers.NewRouter().WithCreateFileHandler(handler).Build()

	request, _ := http.NewRequest(http.MethodPost, "/projects/123/files", bytes.NewBuffer([]byte("invalid json")))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("want status %d, got %d", http.StatusBadRequest, response.Code)
	}
}
