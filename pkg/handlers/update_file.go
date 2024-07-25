package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/artmoskvin/hide/pkg/files"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/spf13/afero"
)

type UpdateType string

const (
	Udiff     UpdateType = "udiff"
	LineDiff  UpdateType = "linediff"
	Overwrite UpdateType = "overwrite"
)

type UdiffRequest struct {
	Patch string `json:"patch"`
}

type LineDiffRequest struct {
	LineDiff files.LineDiffChunk `json:"lineDiff"`
}

type OverwriteRequest struct {
	Content string `json:"content"`
}

type UpdateFileRequest struct {
	Type      UpdateType        `json:"type"`
	Udiff     *UdiffRequest     `json:"udiff,omitempty"`
	LineDiff  *LineDiffRequest  `json:"lineDiff,omitempty"`
	Overwrite *OverwriteRequest `json:"overwrite,omitempty"`
}

func (r *UpdateFileRequest) Validate() error {
	if r.Type == "" {
		return errors.New("type must be provided")
	}

	switch r.Type {
	case Udiff:
		if r.Udiff == nil {
			return errors.New("udiff must be provided")
		}
	case LineDiff:
		if r.LineDiff == nil {
			return errors.New("lineDiff must be provided")
		}
	case Overwrite:
		if r.Overwrite == nil {
			return errors.New("overwrite must be provided")
		}
	default:
		return fmt.Errorf("invalid type: %s", r.Type)
	}

	return nil
}

type UpdateFileHandler struct {
	Manager     project.Manager
	FileManager files.FileManager
}

func (h UpdateFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectId := r.PathValue("id")
	filePath := r.PathValue("path")

	var request UpdateFileRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed parsing request body", http.StatusBadRequest)
		return
	}

	if err := request.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err), http.StatusBadRequest)
		return
	}

	project, err := h.Manager.GetProject(projectId)

	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	fileSystem := afero.NewBasePathFs(afero.NewOsFs(), project.Path)

	var file files.File

	switch request.Type {
	case Udiff:
		updatedFile, err := h.FileManager.ApplyPatch(fileSystem, filePath, request.Udiff.Patch)
		if err != nil {
			http.Error(w, "Failed to update file", http.StatusInternalServerError)
			return
		}
		file = updatedFile
	case LineDiff:
		updatedFile, err := h.FileManager.UpdateLines(fileSystem, filePath, request.LineDiff.LineDiff)
		if err != nil {
			http.Error(w, "Failed to update file", http.StatusInternalServerError)
			return
		}
		file = updatedFile
	case Overwrite:
		updatedFile, err := h.FileManager.UpdateFile(filePath, request.Overwrite.Content)
		if err != nil {
			http.Error(w, "Failed to update file", http.StatusInternalServerError)
			return
		}
		file = updatedFile
	default:
		http.Error(w, "Invalid request: type must be provided", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(file)
}
