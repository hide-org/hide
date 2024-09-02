package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/artmoskvin/hide/pkg/project/mocks"
	"github.com/stretchr/testify/assert"
)

func TestListFilesHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name                string
		projectID           string
		showHidden          string
		mockListFilesFunc   func(ctx context.Context, projectId string, showHidden bool) ([]*model.File, error)
		expectedStatusCode  int
		expectedBody        string
	}{
		{
			name:      "successful listing",
			projectID: "123",
			showHidden: "false",
			mockListFilesFunc: func(ctx context.Context, projectId string, showHidden bool) ([]*model.File, error) {
				return []*model.File{
					model.EmptyFile("file1.txt"),
					model.EmptyFile("file2.txt"),
				}, nil
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `[{"path":"file1.txt"},{"path":"file2.txt"}]`,
		},
		{
			name:      "project not found",
			projectID: "123",
			showHidden: "false",
			mockListFilesFunc: func(ctx context.Context, projectId string, showHidden bool) ([]*model.File, error) {
				return nil, project.NewProjectNotFoundError(projectId)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "project 123 not found",
		},
		{
			name:      "internal server error",
			projectID: "123",
			showHidden: "false",
			mockListFilesFunc: func(ctx context.Context, projectId string, showHidden bool) ([]*model.File, error) {
				return nil, errors.New("internal error")
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Failed to list files: internal error",
		},
		{
			name:      "invalid showHidden query parameter",
			projectID: "123",
			showHidden: "invalid",
			mockListFilesFunc: func(ctx context.Context, projectId string, showHidden bool) ([]*model.File, error) {
				return nil, nil
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid `showHidden` query parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectManager := &mocks.MockProjectManager{
				ListFilesFunc: tt.mockListFilesFunc,
			}

			handler := handlers.ListFilesHandler{
				ProjectManager: mockProjectManager,
			}

			router := handlers.NewRouter().WithListFilesHandler(handler).Build()

			req := httptest.NewRequest("GET", "/projects/"+tt.projectID+"/files?showHidden="+tt.showHidden, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}
