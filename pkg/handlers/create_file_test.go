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
		expectedStatus int
		expectedFile   *model.File
	}{
		{
			name: "Success",
			createFileFunc: func(ctx context.Context, projectId, path, content string) (*model.File, error) {
				return model.NewFile("/test/path", "test content")
			},
			requestBody:    handlers.CreateFileRequest{Path: "/test/path", Content: "test content"},
			expectedStatus: http.StatusCreated,
			expectedFile:   func() *model.File { f, _ := model.NewFile("/test/path", "test content"); return f }(),
		},
		{
			name: "ProjectNotFound",
			createFileFunc: func(ctx context.Context, projectId, path, content string) (*model.File, error) {
				return nil, project.NewProjectNotFoundError(projectId)
			},
			requestBody:    handlers.CreateFileRequest{Path: "/test/path", Content: "test content"},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "FileAlreadyExists",
			createFileFunc: func(ctx context.Context, projectId, path, content string) (*model.File, error) {
				return nil, files.NewFileAlreadyExistsError(path)
			},
			requestBody:    handlers.CreateFileRequest{Path: "/test/path", Content: "test content"},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "InternalServerError",
			createFileFunc: func(ctx context.Context, projectId, path, content string) (*model.File, error) {
				return nil, errors.New("Test error")
			},
			requestBody:    handlers.CreateFileRequest{Path: "/test/path", Content: "test content"},
			expectedStatus: http.StatusInternalServerError,
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
			request, _ := http.NewRequest("POST", "/projects/123/files", bytes.NewBuffer(body))
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, response.Code)
			}

			if tc.expectedFile != nil {
				var respFile model.File
				if err := json.NewDecoder(response.Body).Decode(&respFile); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if !respFile.Equals(tc.expectedFile) {
					t.Errorf("Unexpected file returned: %+v", respFile)
				}
			}
		})
	}
}

func TestCreateFileHandler_InvalidPayload(t *testing.T) {
	mockManager := &mocks.MockProjectManager{}

	handler := handlers.CreateFileHandler{ProjectManager: mockManager}
	router := handlers.NewRouter().WithCreateFileHandler(handler).Build()

	request, _ := http.NewRequest("POST", "/projects/123/files", bytes.NewBuffer([]byte("invalid json")))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}
