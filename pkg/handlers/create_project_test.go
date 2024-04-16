package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artmoskvin/hide/pkg/handlers"
)

func TestCreateProjectHandler(t *testing.T) {
	req, err := http.NewRequest("POST", "/project", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.CreateProject)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("CreateProject() status = %v, want %v", rr.Code, http.StatusCreated)
	}
}

func TestCreateProjectHandler_MethodNotAllowed(t *testing.T) {
	req, err := http.NewRequest("GET", "/project", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.CreateProject)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("CreateProject() status = %v, want %v", rr.Code, http.StatusMethodNotAllowed)
	}
}
