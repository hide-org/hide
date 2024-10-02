package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hide-org/hide/pkg/handlers"
	"github.com/hide-org/hide/pkg/project"
	"github.com/hide-org/hide/pkg/project/mocks"
	"github.com/hide-org/hide/pkg/result"
	"github.com/stretchr/testify/assert"
)

func TestDeleteProjectHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name                  string
		target                string
		mockDeleteProjectFunc func(ctx context.Context, projectId string) <-chan result.Empty
		wantStatusCode        int
		wantBody              string
	}{
		{
			name:   "successful deletion",
			target: "/projects/123",
			mockDeleteProjectFunc: func(ctx context.Context, projectId string) <-chan result.Empty {
				ch := make(chan result.Empty, 1)
				ch <- result.EmptySuccess()
				return ch
			},
			wantStatusCode: http.StatusNoContent,
			wantBody:       "",
		},
		{
			name:   "project not found",
			target: "/projects/123",
			mockDeleteProjectFunc: func(ctx context.Context, projectId string) <-chan result.Empty {
				ch := make(chan result.Empty, 1)
				ch <- result.EmptyFailure(project.NewProjectNotFoundError(projectId))
				return ch
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       "project 123 not found\n",
		},
		{
			name:   "internal server error",
			target: "/projects/123",
			mockDeleteProjectFunc: func(ctx context.Context, projectId string) <-chan result.Empty {
				ch := make(chan result.Empty, 1)
				ch <- result.EmptyFailure(errors.New("internal error"))
				return ch
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "Failed to delete project: internal error\n",
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

			req := httptest.NewRequest(http.MethodDelete, tt.target, nil)
			rr := httptest.NewRecorder()

			router := handlers.NewRouter().WithDeleteProjectHandler(handler).Build()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
			assert.Equal(t, tt.wantBody, rr.Body.String())
		})
	}
}
