package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/artmoskvin/hide/pkg/project"
	project_mocks "github.com/artmoskvin/hide/pkg/project/mocks"
)

func TestReadFileHandler_Success(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		wantFile model.File
	}{
		{
			name:   "Read file with default params",
			target: "/projects/123/files/test.txt",
			wantFile: model.File{
				Path: "test.txt",
				Lines: []model.Line{
					{Number: 1, Content: "line1"},
					{Number: 2, Content: "line2"},
					{Number: 3, Content: "line3"},
				},
			},
		},
		{
			name:   "Read file with startLine=2",
			target: "/projects/123/files/test.txt?startLine=2",
			wantFile: model.File{
				Path: "test.txt",
				Lines: []model.Line{
					{Number: 2, Content: "line2"},
					{Number: 3, Content: "line3"},
				},
			},
		},
		{
			name:   "Read file with numLines=2",
			target: "/projects/123/files/test.txt?numLines=2",
			wantFile: model.File{
				Path: "test.txt",
				Lines: []model.Line{
					{Number: 1, Content: "line1"},
					{Number: 2, Content: "line2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &model.File{
				Path: "test.txt",
				Lines: []model.Line{
					{Number: 1, Content: "line1"},
					{Number: 2, Content: "line2"},
					{Number: 3, Content: "line3"},
				},
			}
			mockManager := &project_mocks.MockProjectManager{
				ReadFileFunc: func(ctx context.Context, projectId string, path string) (*model.File, error) {
					return file, nil
				},
			}

			handler := handlers.ReadFileHandler{ProjectManager: mockManager}
			router := handlers.NewRouter().WithReadFileHandler(handler).Build()

			request, _ := http.NewRequest(http.MethodGet, tt.target, nil)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Errorf("want status 200, got %d", response.Code)
			}

			var respFile model.File
			if err := json.NewDecoder(response.Body).Decode(&respFile); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if !respFile.Equals(&tt.wantFile) {
				t.Errorf("want file %+v, got %+v", tt.wantFile, respFile)
			}
		})
	}
}

func TestReadFileHandler_Fails_WithInvalidQueryParams(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		wantCode int
	}{
		{
			name:     "Read file with invalid startLine param",
			target:   "/projects/123/files/test.txt?startLine=invalid",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Read file with invalid numLines param",
			target:   "/projects/123/files/test.txt?numLines=invalid",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := handlers.ReadFileHandler{}
			router := handlers.NewRouter().WithReadFileHandler(handler).Build()

			request, _ := http.NewRequest(http.MethodGet, tt.target, nil)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != tt.wantCode {
				t.Errorf("want status %d, got %d", tt.wantCode, response.Code)
				t.Errorf("Body: %s", response.Body.String())
			}
		})
	}
}

func TestReadFileHandler_Returns404_WhenProjectNotFound(t *testing.T) {
	t.Run("Read file with invalid project ID", func(t *testing.T) {
		mockManager := &project_mocks.MockProjectManager{
			ReadFileFunc: func(ctx context.Context, projectId string, path string) (*model.File, error) {
				return nil, project.NewProjectNotFoundError(projectId)
			},
		}

		handler := handlers.ReadFileHandler{ProjectManager: mockManager}
		router := handlers.NewRouter().WithReadFileHandler(handler).Build()

		request, _ := http.NewRequest(http.MethodGet, "/projects/123/files/test.txt", nil)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		if response.Code != http.StatusNotFound {
			t.Errorf("want status 404, got %d", response.Code)
		}
	})
}

func TestReadFileHandler_Returns500_WhenReadFileFails(t *testing.T) {
	t.Run("Read file with invalid file path", func(t *testing.T) {
		mockManager := &project_mocks.MockProjectManager{
			ReadFileFunc: func(ctx context.Context, projectId string, path string) (*model.File, error) {
				return nil, errors.New("file not found")
			},
		}

		handler := handlers.ReadFileHandler{ProjectManager: mockManager}
		router := handlers.NewRouter().WithReadFileHandler(handler).Build()

		request, _ := http.NewRequest(http.MethodGet, "/projects/123/files/invalid.txt", nil)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		if response.Code != http.StatusInternalServerError {
			t.Errorf("want status 500, got %d", response.Code)
		}
	})
}
