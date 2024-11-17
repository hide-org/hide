package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hide-org/hide/pkg/files"
	"github.com/hide-org/hide/pkg/model"
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
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
	Content   string `json:"content"`
}

type OverwriteRequest struct {
	Content string `json:"content"`
}

type UpdateFileRequest struct {
	Type      UpdateType        `json:"type"`
	Udiff     *UdiffRequest     `json:"udiff,omitempty"`
	LineDiff  *LineDiffRequest  `json:"linediff,omitempty"`
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

		if r.LineDiff.StartLine == r.LineDiff.EndLine {
			return errors.New("start line must be less than end line")
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
	Files files.Service
}

func (h UpdateFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filePath, err := GetFilePath(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid file path: %s", err), http.StatusBadRequest)
		return
	}

	var request UpdateFileRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("failed parsing request body: %s", err), http.StatusBadRequest)
		return
	}

	if err := request.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %s", err), http.StatusBadRequest)
		return
	}

	var file *model.File

	switch request.Type {
	case Udiff:
		updatedFile, err := h.Files.ApplyPatch(r.Context(), filePath, request.Udiff.Patch)

		if err != nil {
			var fileNotFoundError *files.FileNotFoundError
			if errors.As(err, &fileNotFoundError) {
				http.Error(w, fileNotFoundError.Error(), http.StatusNotFound)
				return
			}

			http.Error(w, fmt.Sprintf("failed to update file: %s", err), http.StatusInternalServerError)
			return
		}
		file = updatedFile
	case LineDiff:
		lineDiff := request.LineDiff
		updatedFile, err := h.Files.UpdateLines(r.Context(), filePath, files.LineDiffChunk{StartLine: lineDiff.StartLine, EndLine: lineDiff.EndLine, Content: lineDiff.Content})
		if err != nil {
			var fileNotFoundError *files.FileNotFoundError
			if errors.As(err, &fileNotFoundError) {
				http.Error(w, fileNotFoundError.Error(), http.StatusNotFound)
				return
			}

			http.Error(w, fmt.Sprintf("failed to update file: %s", err), http.StatusInternalServerError)
			return
		}
		file = updatedFile
	case Overwrite:
		updatedFile, err := h.Files.UpdateFile(r.Context(), filePath, request.Overwrite.Content)
		if err != nil {
			var fileNotFoundError *files.FileNotFoundError
			if errors.As(err, &fileNotFoundError) {
				http.Error(w, fileNotFoundError.Error(), http.StatusNotFound)
				return
			}

			http.Error(w, fmt.Sprintf("failed to update file: %s", err), http.StatusInternalServerError)
			return
		}
		file = updatedFile
	default:
		http.Error(w, "invalid request: type must be provided", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(file)
	return
}
