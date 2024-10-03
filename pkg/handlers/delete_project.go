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

	result := <-h.Manager.DeleteProject(r.Context(), projectID)

	if result.IsFailure() {
		var projectNotFoundError *project.ProjectNotFoundError
		if errors.As(result.Error, &projectNotFoundError) {
			http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("Failed to delete project: %s", result.Error), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
