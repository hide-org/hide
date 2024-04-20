package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/artmoskvin/hide/pkg/project/mocks"
)

// TODO: how to mock DevContainerCli?
//
//	func TestCreateProjectHandler(t *testing.T) {
//		req, err := http.NewRequest("POST", "/project", nil)
//		if err != nil {
//			t.Fatal(err)
//		}
//
//		rr := httptest.NewRecorder()
//		handler := http.HandlerFunc(handlers.CreateProject)
//
//		handler.ServeHTTP(rr, req)
//
//		if rr.Code != http.StatusCreated {
//			t.Errorf("CreateProject() status = %v, want %v", rr.Code, http.StatusCreated)
//		}
//	}
func TestCreateProjectHandler_ServeHTTP_Success(t *testing.T) {
	// Setup
	mockManager := &mocks.MockProjectManager{
		CreateProjectFunc: func(req project.CreateProjectRequest) (project.Project, error) {
			return project.Project{Id: "123", Path: "/test/path"}, nil
		},
	}

	handler := handlers.CreateProjectHandler{Manager: mockManager}
	server := httptest.NewServer(handler)
	defer server.Close()

	requestBody := project.CreateProjectRequest{RepoUrl: "https://github.com/example/repo.git"}
	body, _ := json.Marshal(requestBody)
	request, _ := http.NewRequest("POST", server.URL, bytes.NewBuffer(body))
	response := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(response, request)

	// Verify
	if response.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, response.Code)
	}

	var respProject project.Project
	if err := json.NewDecoder(response.Body).Decode(&respProject); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if respProject.Id != "123" || respProject.Path != "/test/path" {
		t.Errorf("Unexpected project returned: %+v", respProject)
	}
}

// func TestCreateProjectHandler_MethodNotAllowed(t *testing.T) {
// 	req, err := http.NewRequest("GET", "/project", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	rr := httptest.NewRecorder()
// 	createProjectHandler := handlers.CreateProjectHandler{Manager: projectManager}
// 	handler := http.Handler(createProjectHandler)
//
// 	handler.ServeHTTP(rr, req)
//
// 	if rr.Code != http.StatusMethodNotAllowed {
// 		t.Errorf("CreateProject() status = %v, want %v", rr.Code, http.StatusMethodNotAllowed)
// 	}
// }
