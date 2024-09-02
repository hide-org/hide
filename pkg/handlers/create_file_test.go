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
	"github.com/gorilla/mux"
)

func TestCreateFileHandler_Success(t *testing.T) {
	expectedFile, _ := model.NewFile("/test/path", "test content")

	mockManager := &mocks.MockProjectManager{
		CreateFileFunc: func(ctx context.Context, projectId, path, content string) (*model.File, error) {
			return expectedFile, nil
		},
	}

	router := mux.NewRouter()
	handler := handlers.CreateFileHandler{ProjectManager: mockManager}
	router.Handle("/projects/{id}/files", handler).Methods("POST")

	requestBody := handlers.CreateFileRequest{Path: "/test/path", Content: "test content"}
	body, _ := json.Marshal(requestBody)
	request, _ := http.NewRequest("POST", "/projects/123/files", bytes.NewBuffer(body))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, response.Code)
	}

	var respFile model.File
	if err := json.NewDecoder(response.Body).Decode(&respFile); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !respFile.Equals(expectedFile) {
		t.Errorf("Unexpected file returned: %+v", respFile)
	}
}

func TestCreateFileHandler_ProjectNotFound(t *testing.T) {
	mockManager := &mocks.MockProjectManager{
		CreateFileFunc: func(ctx context.Context, projectId, path, content string) (*model.File, error) {
			return nil, project.NewProjectNotFoundError(projectId)
		},
	}

	router := mux.NewRouter()
	handler := handlers.CreateFileHandler{ProjectManager: mockManager}
	router.Handle("/projects/{id}/files", handler).Methods("POST")

	requestBody := handlers.CreateFileRequest{Path: "/test/path", Content: "test content"}
	body, _ := json.Marshal(requestBody)
	request, _ := http.NewRequest("POST", "/projects/123/files", bytes.NewBuffer(body))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestCreateFileHandler_FileAlreadyExists(t *testing.T) {
	mockManager := &mocks.MockProjectManager{
		CreateFileFunc: func(ctx context.Context, projectId, path, content string) (*model.File, error) {
			return nil, files.NewFileAlreadyExistsError(path)
		},
	}

	router := mux.NewRouter()
	handler := handlers.CreateFileHandler{ProjectManager: mockManager}
	router.Handle("/projects/{id}/files", handler).Methods("POST")

	requestBody := handlers.CreateFileRequest{Path: "/test/path", Content: "test content"}
	body, _ := json.Marshal(requestBody)
	request, _ := http.NewRequest("POST", "/projects/123/files", bytes.NewBuffer(body))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, response.Code)
	}
}

func TestCreateFileHandler_InvalidPayload(t *testing.T) {
	mockManager := &mocks.MockProjectManager{}

	router := mux.NewRouter()
	handler := handlers.CreateFileHandler{ProjectManager: mockManager}
	router.Handle("/projects/{id}/files", handler).Methods("POST")

	request, _ := http.NewRequest("POST", "/projects/123/files", bytes.NewBuffer([]byte("invalid json")))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestCreateFileHandler_InternalServerError(t *testing.T) {
	mockManager := &mocks.MockProjectManager{
		CreateFileFunc: func(ctx context.Context, projectId, path, content string) (*model.File, error) {
			return nil, errors.New("Test error")
		},
	}

	router := mux.NewRouter()
	handler := handlers.CreateFileHandler{ProjectManager: mockManager}
	router.Handle("/projects/{id}/files", handler).Methods("POST")

	requestBody := handlers.CreateFileRequest{Path: "/test/path", Content: "test content"}
	body, _ := json.Marshal(requestBody)
	request, _ := http.NewRequest("POST", "/projects/123/files", bytes.NewBuffer(body))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}
}
