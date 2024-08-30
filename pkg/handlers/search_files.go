package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/artmoskvin/hide/pkg/model"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/rs/zerolog/log"
)

const (
	queryKey = "query"
)

type SearchFilesHandler struct {
	ProjectManager project.Manager
}

func (h SearchFilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectID, err := getProjectID(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid project ID: %s", err), http.StatusBadRequest)
		return
	}
	log.Info().Msgf("project id %s", projectID)

	q := r.URL.Query().Get(queryKey)
	if q == "" {
		http.Error(w, "Query not specified", http.StatusBadRequest)
		return
	}

	files, err := h.ProjectManager.ListFiles(r.Context(), projectID, false)
	if err != nil {
		var projectNotFoundError *project.ProjectNotFoundError
		if errors.As(err, &projectNotFoundError) {
			http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("Failed to list files: %s", err), http.StatusInternalServerError)
		return
	}

	// for case insensitive search
	q = strings.ToLower(q)

	resultFiles := make([]model.File, 0)
	for _, file := range files {
		resultLines := make([]model.Line, 0)

		for _, line := range file.Lines {
			if strings.Contains(strings.ToLower(line.Content), q) {
				resultLines = append(resultLines, line)
			}
		}

		if len(resultLines) != 0 {
			resultFiles = append(resultFiles, model.File{
				Path:  file.Path,
				Lines: resultLines,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resultFiles)
}
