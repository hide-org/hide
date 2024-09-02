package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/artmoskvin/hide/pkg/project/mocks"
	"github.com/stretchr/testify/assert"
)

func TestListTasksHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		mockManager    *mocks.MockProjectManager
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "project with tasks",
			projectID: "project-id",
			mockManager: &mocks.MockProjectManager{
				GetProjectFunc: func(ctx context.Context, projectId string) (model.Project, error) {
					return model.Project{
						Id: projectId,
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
			name:      "project with no tasks",
			projectID: "project-id",
			mockManager: &mocks.MockProjectManager{
				GetProjectFunc: func(ctx context.Context, projectId string) (model.Project, error) {
					return model.Project{
						Id: projectId,
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
			name:      "project not found",
			projectID: "project-id",
			mockManager: &mocks.MockProjectManager{
				GetProjectFunc: func(ctx context.Context, projectId string) (model.Project, error) {
					return model.Project{}, project.NewProjectNotFoundError(projectId)
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `project project-id not found`,
		},
		{
			name:      "internal server error",
			projectID: "project-id",
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
			req := httptest.NewRequest("GET", "/projects/"+tt.projectID+"/tasks", nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}
