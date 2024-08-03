package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/artmoskvin/hide/pkg/project/mocks"
	"github.com/artmoskvin/hide/pkg/result"
)

const repoUrl = "https://github.com/example/repo.git"

func TestCreateProjectHandler_Success(t *testing.T) {
	// Expected project
	expectedProject := model.Project{Id: "123", Path: "/test/path"}

	// Setup
	mockManager := &mocks.MockProjectManager{
		CreateProjectFunc: func(req project.CreateProjectRequest) <-chan result.Result[model.Project] {
			ch := make(chan result.Result[model.Project], 1)
			ch <- result.Success(expectedProject)
			return ch
		},
	}

	handler := handlers.CreateProjectHandler{Manager: mockManager}

	requestBody := project.CreateProjectRequest{Repository: project.Repository{Url: repoUrl}}
	body, _ := json.Marshal(requestBody)
	request, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer(body))
	response := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(response, request)

	// Verify
	if response.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, response.Code)
	}

	var respProject model.Project
	if err := json.NewDecoder(response.Body).Decode(&respProject); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !reflect.DeepEqual(respProject, expectedProject) {
		t.Errorf("Unexpected project returned: %+v", respProject)
	}
}

func TestCreateProjectHandler_Failure(t *testing.T) {
	// Setup
	mockManager := &mocks.MockProjectManager{
		CreateProjectFunc: func(req project.CreateProjectRequest) <-chan result.Result[model.Project] {
			ch := make(chan result.Result[model.Project], 1)
			ch <- result.Failure[model.Project](errors.New("Test error"))
			return ch
		},
	}

	handler := handlers.CreateProjectHandler{Manager: mockManager}

	requestBody := project.CreateProjectRequest{Repository: project.Repository{Url: repoUrl}}
	body, _ := json.Marshal(requestBody)
	request, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer(body))
	response := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(response, request)

	// Verify
	if response.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}
}

func TestCreateProjectHandler_BadRequest(t *testing.T) {
	// Setup
	mockManager := &mocks.MockProjectManager{}

	handler := handlers.CreateProjectHandler{Manager: mockManager}

	request, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer([]byte("invalid json")))
	response := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(response, request)

	// Verify
	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}
