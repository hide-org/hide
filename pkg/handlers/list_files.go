package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/artmoskvin/hide/pkg/filemanager"
	"github.com/artmoskvin/hide/pkg/project"
)

type FileResponse struct {
	Path string `json:"path"`
}

type ListFilesHandler struct {
	ProjectManager project.Manager
	FileManager    filemanager.FileManager
}

func (h ListFilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectId := r.PathValue("id")
	project, err := h.ProjectManager.GetProject(projectId)

	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	files, err := h.FileManager.ListFiles(project.Path)

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
