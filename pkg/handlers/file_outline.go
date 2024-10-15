package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hide-org/hide/pkg/project"
)

type DocumentOutline struct {
	ProjectManager project.Manager
}

func (h DocumentOutline) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectID, err := getProjectID(r)
	if err != nil {
		http.Error(w, "invalid project ID", http.StatusBadRequest)
		return
	}

	filePath, err := GetFilePath(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid file path: %s", err), http.StatusBadRequest)
		return
	}

	// TODO: rename to document outline
	outline, err := h.ProjectManager.DocumentOutline(r.Context(), projectID, filePath)
	if err != nil {
		var projectNotFoundError *project.ProjectNotFoundError
		if errors.As(err, &projectNotFoundError) {
			http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("Failed to create file outline: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(outline)
}
