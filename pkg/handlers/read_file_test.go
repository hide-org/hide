package handlers_test

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artmoskvin/hide/pkg/filemanager"
	filemanager_mocks "github.com/artmoskvin/hide/pkg/filemanager/mocks"
	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/project"
	project_mocks "github.com/artmoskvin/hide/pkg/project/mocks"
)

func TestReadFileHandler_Success(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		expectedFile filemanager.File
	}{
		{
			name:  "Read file with default params",
			query: "",
			expectedFile: filemanager.File{
				Path:    "test.txt",
				Content: "line1\nline2\nline3\n",
			},
		},
		{
			name:  "Read file with showLineNumbers=true",
			query: "showLineNumbers=true",
			expectedFile: filemanager.File{
				Path:    "test.txt",
				Content: "1:line1\n2:line2\n3:line3\n",
			},
		},
		{
			name:  "Read file with startLine=2",
			query: "startLine=2",
			expectedFile: filemanager.File{
				Path:    "test.txt",
				Content: "2:line2\n3:line3\n",
			},
		},
		{
			name:  "Read file with numLines=2",
			query: "numLines=2",
			expectedFile: filemanager.File{
				Path:    "test.txt",
				Content: "line1\nline2\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := &project_mocks.MockProjectManager{
				GetProjectFunc: func(projectId string) (project.Project, error) {
					return project.Project{}, nil
				},
			}

			mockFileManager := &filemanager_mocks.MockFileManager{
				ReadFileFunc: func(fileSystem fs.FS, path string, props filemanager.ReadProps) (filemanager.File, error) {
					return tt.expectedFile, nil
				},
			}

			handler := handlers.ReadFileHandler{Manager: mockManager, FileManager: mockFileManager}

			request, _ := http.NewRequest("GET", "/projects/123/files/test.txt?"+tt.query, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", response.Code)
			}

			var respFile filemanager.File
			if err := json.NewDecoder(response.Body).Decode(&respFile); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if respFile != tt.expectedFile {
				t.Errorf("Expected file %+v, got %+v", tt.expectedFile, respFile)
			}
		})
	}
}

func TestReadFileHandler_Fails_WithInvalidQueryParams(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		expectedCode int
	}{
		{
			name:         "Read file with invalid showLineNumbers param",
			query:        "showLineNumbers=invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Read file with invalid startLine param",
			query:        "startLine=invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Read file with invalid numLines param",
			query:        "numLines=invalid",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := &project_mocks.MockProjectManager{
				GetProjectFunc: func(projectId string) (project.Project, error) {
					return project.Project{}, nil
				},
			}

			handler := handlers.ReadFileHandler{Manager: mockManager, FileManager: filemanager.NewFileManager()}

			request, _ := http.NewRequest("GET", "/projects/123/files/test.txt?"+tt.query, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != tt.expectedCode {
				t.Errorf("Expected status %d, got %d", tt.expectedCode, response.Code)
				t.Errorf("Body: %s", response.Body.String())
			}
		})
	}
}

func TestReadFileHandler_Fails_WhenProjectNotFound(t *testing.T) {
	t.Run("Read file with invalid project ID", func(t *testing.T) {
		mockManager := &project_mocks.MockProjectManager{
			GetProjectFunc: func(projectId string) (project.Project, error) {
				return project.Project{}, errors.New("project not found")
			},
		}

		handler := handlers.ReadFileHandler{Manager: mockManager, FileManager: filemanager.NewFileManager()}

		request, _ := http.NewRequest("GET", "/projects/123/files/test.txt", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if response.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", response.Code)
		}
	})
}

func TestReadFileHandler_Fails_WhenReadFileFails(t *testing.T) {
	t.Run("Read file with invalid file path", func(t *testing.T) {
		mockManager := &project_mocks.MockProjectManager{
			GetProjectFunc: func(projectId string) (project.Project, error) {
				return project.Project{}, nil
			},
		}

		mockFileManager := &filemanager_mocks.MockFileManager{
			ReadFileFunc: func(fileSystem fs.FS, path string, props filemanager.ReadProps) (filemanager.File, error) {
				return filemanager.File{}, errors.New("file not found")
			},
		}

		handler := handlers.ReadFileHandler{Manager: mockManager, FileManager: mockFileManager}

		request, _ := http.NewRequest("GET", "/projects/123/files/invalid.txt", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if response.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", response.Code)
		}
	})
}
