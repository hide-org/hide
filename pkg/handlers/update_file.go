package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/artmoskvin/hide/pkg/filemanager"
	"github.com/artmoskvin/hide/pkg/project"
)

type UpdateFileRequest struct {
	Content string `json:"content"`
}

type UpdateFileHandler struct {
	Manager     project.Manager
	FileManager filemanager.FileManager
}

func (h UpdateFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectId := r.PathValue("id")
	filePath := r.PathValue("path")

	var request UpdateFileRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed parsing request body", http.StatusBadRequest)
		return
	}

	project, err := h.Manager.GetProject(projectId)

	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	fullPath := filepath.Join(project.Path, filePath)
	file, err := h.FileManager.UpdateFile(fullPath, request.Content)

	if err != nil {
		http.Error(w, "Failed to update file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(file)
}
