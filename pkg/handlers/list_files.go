package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/artmoskvin/hide/pkg/project"
)

type FileResponse struct {
	Path string `json:"path"`
}

type ListFilesHandler struct {
	ProjectManager project.Manager
}

func (h ListFilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectID, err := getProjectID(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid project ID: %s", err), http.StatusBadRequest)
		return
	}

	showHidden, err := parseBoolQueryParam(r.URL.Query(), "showHidden", false)

	if err != nil {
		http.Error(w, fmt.Sprintf("invalid showHidden query parameter: %s", err), http.StatusBadRequest)
		return
	}

	files, err := h.ProjectManager.ListFiles(r.Context(), projectID, showHidden)
	if err != nil {
		var projectNotFoundError *project.ProjectNotFoundError
		if errors.As(err, &projectNotFoundError) {
			http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("Failed to list files: %s", err), http.StatusInternalServerError)
		return
	}

	var fileResponses []FileResponse

	for _, file := range files {
		fileResponses = append(fileResponses, FileResponse{Path: file.Path})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fileResponses)
}
