package handlers

import (
	"errors"
	"net/http"

	"github.com/artmoskvin/hide/pkg/project"
)

type DeleteFileHandler struct {
	ProjectManager project.Manager
}

func (h DeleteFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectId := r.PathValue("id")
	filePath := r.PathValue("path")

	if err := h.ProjectManager.DeleteFile(r.Context(), projectId, filePath); err != nil {
		var projectNotFoundError *project.ProjectNotFoundError
		if errors.As(err, &projectNotFoundError) {
			http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
