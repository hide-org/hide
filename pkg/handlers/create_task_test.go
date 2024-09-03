package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/artmoskvin/hide/pkg/project/mocks"
)

func TestCreateTaskHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func() *mocks.MockProjectManager
		requestBody    handlers.TaskRequest
		expectedStatus int
		expectedResult *project.TaskResult
	}{
		{
			name: "command success",
			setupMock: func() *mocks.MockProjectManager {
				return &mocks.MockProjectManager{
					CreateTaskFunc: func(ctx context.Context, projectId string, command string) (project.TaskResult, error) {
						return project.TaskResult{StdOut: "Test output", StdErr: "Test error", ExitCode: 0}, nil
					},
				}
			},
			requestBody: handlers.TaskRequest{
				Command: func() *string { s := "test command"; return &s }(),
			},
			expectedStatus: http.StatusOK,
			expectedResult: &project.TaskResult{StdOut: "Test output", StdErr: "Test error", ExitCode: 0},
		},
		{
			name: "alias success",
			setupMock: func() *mocks.MockProjectManager {
				return &mocks.MockProjectManager{
					ResolveTaskAliasFunc: func(ctx context.Context, projectId string, alias string) (devcontainer.Task, error) {
						return devcontainer.Task{Command: "resolved command"}, nil
					},
					CreateTaskFunc: func(ctx context.Context, projectId string, command string) (project.TaskResult, error) {
						return project.TaskResult{StdOut: "Test output", StdErr: "Test error", ExitCode: 0}, nil
					},
				}
			},
			requestBody: handlers.TaskRequest{
				Alias: func() *string { s := "test alias"; return &s }(),
			},
			expectedStatus: http.StatusOK,
			expectedResult: &project.TaskResult{StdOut: "Test output", StdErr: "Test error", ExitCode: 0},
		},
		{
			name: "failure",
			setupMock: func() *mocks.MockProjectManager {
				return &mocks.MockProjectManager{
					CreateTaskFunc: func(ctx context.Context, projectId string, command string) (project.TaskResult, error) {
						return project.TaskResult{}, errors.New("Test error")
					},
				}
			},
			requestBody: handlers.TaskRequest{
				Command: func() *string { s := "test command"; return &s }(),
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := tt.setupMock()

			handler := handlers.CreateTaskHandler{Manager: mockManager}
			router := handlers.NewRouter().WithCreateTaskHandler(handler).Build()

			body, _ := json.Marshal(tt.requestBody)
			request, _ := http.NewRequest("POST", "/projects/123/tasks", bytes.NewBuffer(body))
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, response.Code)
				if tt.expectedStatus == http.StatusInternalServerError {
					t.Errorf("Body: %s", response.Body.String())
				}
			}

			if tt.expectedStatus == http.StatusOK {
				var respResult project.TaskResult
				if err := json.NewDecoder(response.Body).Decode(&respResult); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if respResult != *tt.expectedResult {
					t.Errorf("Unexpected result returned: %+v", respResult)
				}
			}
		})
	}
}

func TestCreateTaskHandler_BadRequest(t *testing.T) {
	// Setup
	mockManager := &mocks.MockProjectManager{}

	handler := handlers.CreateTaskHandler{Manager: mockManager}
	router := handlers.NewRouter().WithCreateTaskHandler(handler).Build()

	// No request body
	request, _ := http.NewRequest("POST", "/projects/123/tasks", bytes.NewBuffer([]byte("invalid json")))
	response := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(response, request)

	// Verify
	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestTaskRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		request  handlers.TaskRequest
		expected error
	}{
		{
			name:     "Empty request",
			request:  handlers.TaskRequest{},
			expected: errors.New("either command or alias must be provided"),
		},
		{
			name:     "Both command and alias",
			request:  handlers.TaskRequest{Command: new(string), Alias: new(string)},
			expected: errors.New("only one of command or alias must be provided"),
		},
		{
			name:     "Command provided",
			request:  handlers.TaskRequest{Command: new(string)},
			expected: nil,
		},
		{
			name:     "Alias provided",
			request:  handlers.TaskRequest{Alias: new(string)},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if err == nil && tt.expected != nil {
				t.Errorf("Expected error %v, got nil", tt.expected)
			} else if err != nil && tt.expected == nil {
				t.Errorf("Unexpected error %v", err)
			} else if err != nil && tt.expected != nil && err.Error() != tt.expected.Error() {
				t.Errorf("Expected error %v, got %v", tt.expected, err)
			}
		})
	}
}
