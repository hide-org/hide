package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hide-org/hide/pkg/handlers"
	"github.com/hide-org/hide/pkg/model"
	"github.com/hide-org/hide/pkg/project"
	"github.com/hide-org/hide/pkg/project/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCreateProjectHandler(t *testing.T) {
	tests := []struct {
		name              string
		createProjectFunc func(ctx context.Context, req project.CreateProjectRequest) (*model.Project, error)
		request           project.CreateProjectRequest
		wantStatusCode    int
		wantProject       *model.Project
		wantError         string
	}{
		{
			name: "successful creation",
			createProjectFunc: func(ctx context.Context, req project.CreateProjectRequest) (*model.Project, error) {
				return &model.Project{ID: "123", Path: "/test/path"}, nil
			},
			request: project.CreateProjectRequest{
				Repository: model.Repository{
					Url: "https://github.com/example/repo.git",
				},
			},
			wantStatusCode: http.StatusCreated,
			wantProject:    &model.Project{ID: "123", Path: "/test/path"},
		},
		{
			name: "failed creation",
			createProjectFunc: func(ctx context.Context, req project.CreateProjectRequest) (*model.Project, error) {
				return nil, errors.New("Test error")
			},
			request: project.CreateProjectRequest{
				Repository: model.Repository{
					Url: "https://github.com/example/repo.git",
				},
			},
			wantStatusCode: http.StatusInternalServerError,
			wantError:      "Failed to create project: Test error",
		},
		{
			name: "validation error",
			request: project.CreateProjectRequest{
				Repository: model.Repository{},
			},
			wantStatusCode: http.StatusBadRequest,
			wantError:      "Validation error: Key: 'CreateProjectRequest.Repository' Error:Field validation for 'Repository' failed on the 'required' tag",
		},
		{
			name: "url without protocol",
			request: project.CreateProjectRequest{
				Repository: model.Repository{
					Url: "github.com/django/django",
				},
			},
			wantStatusCode: http.StatusBadRequest,
			wantError:      "Validation error",
		},
		{
			name: "url with `file` protocol",
			createProjectFunc: func(ctx context.Context, req project.CreateProjectRequest) (*model.Project, error) {
				return &model.Project{ID: "123", Path: "/test/path"}, nil
			},
			request: project.CreateProjectRequest{
				Repository: model.Repository{
					Url: "file:///code/django/django",
				},
			},
			wantStatusCode: http.StatusCreated,
			wantProject:    &model.Project{ID: "123", Path: "/test/path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := &mocks.MockProjectManager{
				CreateProjectFunc: tt.createProjectFunc,
			}

			handler := handlers.CreateProjectHandler{Manager: mockManager, Validator: validator.New(validator.WithRequiredStructEnabled())}
			router := handlers.NewRouter().WithCreateProjectHandler(handler).Build()

			body, _ := json.Marshal(tt.request)
			request, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(body))
			response := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(response, request)

			// Verify
			if response.Code != tt.wantStatusCode {
				t.Errorf("want status %d, got %d", tt.wantStatusCode, response.Code)
			}

			if tt.wantProject != nil {
				var respProject model.Project
				if err := json.NewDecoder(response.Body).Decode(&respProject); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if !reflect.DeepEqual(respProject, *tt.wantProject) {
					t.Errorf("want project %+v, got %+v", tt.wantProject, respProject)
				}
			}

			if tt.wantError != "" {
				assert.Contains(t, response.Body.String(), tt.wantError)
			}

		})
	}
}

func TestCreateProjectHandler_BadRequest(t *testing.T) {
	// Setup
	mockManager := &mocks.MockProjectManager{}
	handler := handlers.CreateProjectHandler{Manager: mockManager}
	router := handlers.NewRouter().WithCreateProjectHandler(handler).Build()

	request, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer([]byte("invalid json")))
	response := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(response, request)

	// Verify
	if response.Code != http.StatusBadRequest {
		t.Errorf("want status %d, got %d", http.StatusBadRequest, response.Code)
	}
}
