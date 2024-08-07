package handlers

import (
	"net/http"

	"github.com/artmoskvin/hide/pkg/files"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/spf13/afero"
)

type DeleteFileHandler struct {
	Manager     project.Manager
	FileManager files.FileManager
}

func (h DeleteFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectId := r.PathValue("id")
	filePath := r.PathValue("path")

	project, err := h.Manager.GetProject(projectId)

	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	err = h.FileManager.DeleteFile(r.Context(), afero.NewBasePathFs(afero.NewOsFs(), project.Path), filePath)

	if err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
