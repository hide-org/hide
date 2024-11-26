package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hide-org/hide/pkg/project"
)

type GetProjectsHandler struct {
	Manager project.Manager
}

func (h GetProjectsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projects, err := h.Manager.GetProjects(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get projects: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(projects); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %s", err), http.StatusInternalServerError)
		return
	}
} 