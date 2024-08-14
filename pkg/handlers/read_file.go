package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/artmoskvin/hide/pkg/project"
)

type ReadFileHandler struct {
	ProjectManager project.Manager
}

func (h ReadFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectID, err := getProjectID(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid project ID: %s", err), http.StatusBadRequest)
		return
	}

	filePath, err := getFilePath(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid file path: %s", err), http.StatusBadRequest)
		return
	}

	queryParams := r.URL.Query()

	startLine, startLinePresent, err := parseIntQueryParam(queryParams, "startLine")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	numLines, numLinesPresent, err := parseIntQueryParam(queryParams, "numLines")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, err := h.ProjectManager.ReadFile(r.Context(), projectID, filePath)

	if err != nil {
		var projectNotFoundError *project.ProjectNotFoundError
		if errors.As(err, &projectNotFoundError) {
			http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("Failed to read file: %s", err), http.StatusInternalServerError)
		return
	}

	if startLinePresent || numLinesPresent {
		if startLinePresent {
			if startLine < 1 || startLine > len(file.Lines) {
				http.Error(w, fmt.Sprintf("Start line must be between 1 and %d", len(file.Lines)), http.StatusBadRequest)
				return
			}
		} else {
			startLine = 1
		}

		endLine := len(file.Lines)
		if numLinesPresent {
			endLine = startLine + numLines - 1
			if endLine > len(file.Lines) {
				endLine = len(file.Lines)
			}
		}

		file = file.WithLineRange(startLine, endLine)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(file)
}
