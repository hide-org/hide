package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/artmoskvin/hide/pkg/project/mocks"
	"github.com/artmoskvin/hide/pkg/result"
	"github.com/stretchr/testify/assert"
)

func TestDeleteProjectHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name                    string
		projectID               string
		mockDeleteProjectFunc   func(ctx context.Context, projectId string) <-chan result.Empty
		expectedStatusCode      int
		expectedBody            string
	}{
		{
			name:      "successful deletion",
			projectID: "123",
			mockDeleteProjectFunc: func(ctx context.Context, projectId string) <-chan result.Empty {
				ch := make(chan result.Empty, 1)
				ch <- result.EmptySuccess()
				return ch
			},
			expectedStatusCode: http.StatusNoContent,
			expectedBody:       "",
		},
		{
			name:      "project not found",
			projectID: "123",
			mockDeleteProjectFunc: func(ctx context.Context, projectId string) <-chan result.Empty {
				ch := make(chan result.Empty, 1)
				ch <- result.EmptyFailure(project.NewProjectNotFoundError(projectId))
				return ch
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "project 123 not found\n",
		},
		{
			name:      "internal server error",
			projectID: "123",
			mockDeleteProjectFunc: func(ctx context.Context, projectId string) <-chan result.Empty {
				ch := make(chan result.Empty, 1)
				ch <- result.EmptyFailure(errors.New("internal error"))
				return ch
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Failed to delete project: internal error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := &mocks.MockProjectManager{
				DeleteProjectFunc: tt.mockDeleteProjectFunc,
			}

			handler := handlers.DeleteProjectHandler{
				Manager: mockPM,
			}

			req := httptest.NewRequest("DELETE", "/projects/"+tt.projectID, nil)
			rr := httptest.NewRecorder()

			router := handlers.NewRouter().WithDeleteProjectHandler(handler).Build()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())
		})
	}
}
