package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/artmoskvin/hide/pkg/files"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/spf13/afero"
)

type FileResponse struct {
	Path string `json:"path"`
}

type ListFilesHandler struct {
	ProjectManager project.Manager
	FileManager    files.FileManager
}

func (h ListFilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectId := r.PathValue("id")
	project, err := h.ProjectManager.GetProject(projectId)

	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	files, err := h.FileManager.ListFiles(r.Context(), afero.NewBasePathFs(afero.NewOsFs(), project.Path))

	if err != nil {
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
