package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hide-org/hide/pkg/devcontainer"
	"github.com/hide-org/hide/pkg/handlers"
	"github.com/hide-org/hide/pkg/model"
	"github.com/hide-org/hide/pkg/project"
	"github.com/hide-org/hide/pkg/project/mocks"
	"github.com/stretchr/testify/assert"
)

func TestListTasksHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		target         string
		mockManager    *mocks.MockProjectManager
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "project with tasks",
			target: "/projects/project-id/tasks",
			mockManager: &mocks.MockProjectManager{
				GetProjectFunc: func(ctx context.Context, projectId string) (model.Project, error) {
					return model.Project{
						ID: projectId,
						Config: model.Config{
							DevContainerConfig: devcontainer.Config{
								GeneralProperties: devcontainer.GeneralProperties{
									Customizations: devcontainer.Customizations{
										Hide: &devcontainer.HideCustomization{
											Tasks: []devcontainer.Task{
												{Alias: "task1", Command: "command1"},
												{Alias: "task2", Command: "command2"},
											},
										},
									},
								},
							},
						},
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"alias":"task1","command":"command1"},{"alias":"task2","command":"command2"}]`,
		},
		{
			name:   "project with no tasks",
			target: "/projects/project-id/tasks",
			mockManager: &mocks.MockProjectManager{
				GetProjectFunc: func(ctx context.Context, projectId string) (model.Project, error) {
					return model.Project{
						ID: projectId,
						Config: model.Config{
							DevContainerConfig: devcontainer.Config{
								GeneralProperties: devcontainer.GeneralProperties{
									Customizations: devcontainer.Customizations{
										Hide: &devcontainer.HideCustomization{
											Tasks: []devcontainer.Task{},
										},
									},
								},
							},
						},
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[]`,
		},
		{
			name:   "project not found",
			target: "/projects/project-id/tasks",
			mockManager: &mocks.MockProjectManager{
				GetProjectFunc: func(ctx context.Context, projectId string) (model.Project, error) {
					return model.Project{}, project.NewProjectNotFoundError(projectId)
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `project project-id not found`,
		},
		{
			name:   "internal server error",
			target: "/projects/project-id/tasks",
			mockManager: &mocks.MockProjectManager{
				GetProjectFunc: func(ctx context.Context, projectId string) (model.Project, error) {
					return model.Project{}, errors.New("internal error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `Failed to get project: internal error`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := handlers.ListTasksHandler{Manager: tt.mockManager}
			router := handlers.NewRouter().WithListTasksHandler(handler).Build()
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}
