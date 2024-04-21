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

func TestExecCmdHandler_Success(t *testing.T) {
	// Expected result
	expectedResult := project.CmdResult{StdOut: "Test output", StdErr: "Test error", ExitCode: 0}

	// Setup
	mockManager := &mocks.MockProjectManager{
		ExecCmdFunc: func(projectId string, req project.ExecCmdRequest) (project.CmdResult, error) {
			return expectedResult, nil
		},
	}

	handler := handlers.ExecCmdHandler{Manager: mockManager}

	requestBody := project.ExecCmdRequest{Cmd: "test command"}
	body, _ := json.Marshal(requestBody)
	request, _ := http.NewRequest("POST", "/projects/123/exec", bytes.NewBuffer(body))
	response := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(response, request)

	// Verify
	if response.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, response.Code)
	}

	var respResult project.CmdResult
	if err := json.NewDecoder(response.Body).Decode(&respResult); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if respResult != expectedResult {
		t.Errorf("Unexpected result returned: %+v", respResult)
	}
}

func TestExecCmdHandler_Failure(t *testing.T) {
	// Setup
	mockManager := &mocks.MockProjectManager{
		ExecCmdFunc: func(projectId string, req project.ExecCmdRequest) (project.CmdResult, error) {
			return project.CmdResult{}, errors.New("Test error")
		},
	}

	handler := handlers.ExecCmdHandler{Manager: mockManager}

	requestBody := project.ExecCmdRequest{Cmd: "test command"}
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

func TestExecCmdHandler_BadRequest(t *testing.T) {
	// Setup
	mockManager := &mocks.MockProjectManager{}

	handler := handlers.ExecCmdHandler{Manager: mockManager}

	request, _ := http.NewRequest("POST", "/projects/123/exec", bytes.NewBuffer([]byte("invalid json")))
	response := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(response, request)

	// Verify
	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}
