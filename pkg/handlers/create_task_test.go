package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hide-org/hide/pkg/devcontainer"
	"github.com/hide-org/hide/pkg/handlers"
	"github.com/hide-org/hide/pkg/project"
	"github.com/hide-org/hide/pkg/project/mocks"
)

func TestCreateTaskHandler(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func() *mocks.MockProjectManager
		requestBody handlers.TaskRequest
		wantStatus  int
		wantResult  *project.TaskResult
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
			wantStatus: http.StatusOK,
			wantResult: &project.TaskResult{StdOut: "Test output", StdErr: "Test error", ExitCode: 0},
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
			wantStatus: http.StatusOK,
			wantResult: &project.TaskResult{StdOut: "Test output", StdErr: "Test error", ExitCode: 0},
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
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := tt.setupMock()

			handler := handlers.CreateTaskHandler{Manager: mockManager}
			router := handlers.NewRouter().WithCreateTaskHandler(handler).Build()

			body, _ := json.Marshal(tt.requestBody)
			request, _ := http.NewRequest(http.MethodPost, "/projects/123/tasks", bytes.NewBuffer(body))
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, response.Code)
				if tt.wantStatus == http.StatusInternalServerError {
					t.Errorf("Body: %s", response.Body.String())
				}
			}

			if tt.wantStatus == http.StatusOK {
				var respResult project.TaskResult
				if err := json.NewDecoder(response.Body).Decode(&respResult); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if respResult != *tt.wantResult {
					t.Errorf("want result %+v, got %+v", tt.wantResult, respResult)
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
	request, _ := http.NewRequest(http.MethodPost, "/projects/123/tasks", bytes.NewBuffer([]byte("invalid json")))
	response := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(response, request)

	// Verify
	if response.Code != http.StatusBadRequest {
		t.Errorf("want status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestTaskRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request handlers.TaskRequest
		want    error
	}{
		{
			name:    "Empty request",
			request: handlers.TaskRequest{},
			want:    errors.New("either command or alias must be provided"),
		},
		{
			name:    "Both command and alias",
			request: handlers.TaskRequest{Command: new(string), Alias: new(string)},
			want:    errors.New("only one of command or alias must be provided"),
		},
		{
			name:    "Command provided",
			request: handlers.TaskRequest{Command: new(string)},
			want:    nil,
		},
		{
			name:    "Alias provided",
			request: handlers.TaskRequest{Alias: new(string)},
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if err == nil && tt.want != nil {
				t.Errorf("want error %v, got nil", tt.want)
			} else if err != nil && tt.want == nil {
				t.Errorf("want error %v, got %v", tt.want, err)
			} else if err != nil && tt.want != nil && err.Error() != tt.want.Error() {
				t.Errorf("want error %v, got %v", tt.want, err)
			}
		})
	}
}
