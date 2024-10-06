package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hide-org/hide/pkg/project"
)

type DeleteProjectHandler struct {
	Manager project.Manager
}

func (h DeleteProjectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectID, err := getProjectID(r)
	if err != nil {
		http.Error(w, "invalid project ID", http.StatusBadRequest)
		return
	}

	if err := h.Manager.DeleteProject(r.Context(), projectID); err != nil {
		var projectNotFoundError *project.ProjectNotFoundError
		if errors.As(err, &projectNotFoundError) {
			http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("Failed to delete project: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
