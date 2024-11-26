package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hide-org/hide/pkg/project"
)

type GetProjectHandler struct {
	Manager project.Manager
}

func (h GetProjectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectID, err := getProjectID(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid project ID: %s", err), http.StatusBadRequest)
		return
	}

	p, err := h.Manager.GetProject(r.Context(), projectID)
	if err != nil {
		var projectNotFoundError *project.ProjectNotFoundError
		if errors.As(err, &projectNotFoundError) {
			http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("failed to get project: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(p); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %s", err), http.StatusInternalServerError)
		return
	}
} 