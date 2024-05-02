package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/artmoskvin/hide/pkg/project/mocks"
)

func TestCreateTaskHandler_Success(t *testing.T) {
	// Expected result
	expectedResult := project.TaskResult{StdOut: "Test output", StdErr: "Test error", ExitCode: 0}

	// Setup
	mockManager := &mocks.MockProjectManager{
		CreateTaskFunc: func(projectId string, req project.TaskRequest) (project.TaskResult, error) {
			return expectedResult, nil
		},
	}

	handler := handlers.CreateTaskHandler{Manager: mockManager}

	requestBody := project.TaskRequest{Command: new(string)}
	*requestBody.Command = "test command"

	body, _ := json.Marshal(requestBody)
	request, _ := http.NewRequest("POST", "/projects/123/exec", bytes.NewBuffer(body))
	response := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(response, request)

	// Verify
	if response.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, response.Code)
	}

	var respResult project.TaskResult
	if err := json.NewDecoder(response.Body).Decode(&respResult); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if respResult != expectedResult {
		t.Errorf("Unexpected result returned: %+v", respResult)
	}
}

func TestCreateTaskHandler_Failure(t *testing.T) {
	// Setup
	mockManager := &mocks.MockProjectManager{
		CreateTaskFunc: func(projectId string, req project.TaskRequest) (project.TaskResult, error) {
			return project.TaskResult{}, errors.New("Test error")
		},
	}

	handler := handlers.CreateTaskHandler{Manager: mockManager}

	requestBody := project.TaskRequest{Command: new(string)}
	*requestBody.Command = "test command"

	body, _ := json.Marshal(requestBody)
	request, _ := http.NewRequest("POST", "/projects/123/exec", bytes.NewBuffer(body))
	response := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(response, request)

	// Verify
	if response.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}
}

func TestCreateTaskHandler_BadRequest(t *testing.T) {
	// Setup
	mockManager := &mocks.MockProjectManager{}

	handler := handlers.CreateTaskHandler{Manager: mockManager}

	request, _ := http.NewRequest("POST", "/projects/123/exec", bytes.NewBuffer([]byte("invalid json")))
	response := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(response, request)

	// Verify
	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}
