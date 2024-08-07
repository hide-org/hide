package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/artmoskvin/hide/pkg/project"
	"github.com/gorilla/mux"
)

const key = "id"

type FileResponse struct {
	Path string `json:"path"`
}

type ListFilesHandler struct {
	ProjectManager project.Manager
}

func (h ListFilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, ok := vars[key]
	if !ok {
		http.Error(w, "invalid project ID", http.StatusBadRequest)
	}

	// projectId := r.PathValue("id")
	files, err := h.ProjectManager.ListFiles(r.Context(), projectID)
	if err != nil {
		var projectNotFoundError *project.ProjectNotFoundError
		if errors.As(err, &projectNotFoundError) {
			http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to list files", http.StatusInternalServerError)
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
