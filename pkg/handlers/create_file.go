package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/artmoskvin/hide/pkg/filemanager"
	"github.com/artmoskvin/hide/pkg/project"
)

type CreateFileRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type CreateFileHandler struct {
	Manager     project.Manager
	FileManager filemanager.FileManager
}

func (h CreateFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectId := r.PathValue("id")
	var request CreateFileRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed parsing request body", http.StatusBadRequest)
		return
	}

	project, err := h.Manager.GetProject(projectId)

	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	fullPath := filepath.Join(project.Path, request.Path)
	file, err := h.FileManager.CreateFile(fullPath, request.Content)

	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(file)
}
