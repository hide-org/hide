package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hide-org/hide/pkg/project"
)

type CreateProjectHandler struct {
	Manager project.Manager
}

func (h CreateProjectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request project.CreateProjectRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed parsing request body", http.StatusBadRequest)
		return
	}

	result := <-h.Manager.CreateProject(r.Context(), request)

	if result.IsFailure() {
		http.Error(w, fmt.Sprintf("Failed to create project: %s", result.Error), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result.Get())
}
