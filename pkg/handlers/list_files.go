package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hide-org/hide/pkg/project"
)

type FileInfo struct {
	Path string `json:"path"`
}

type ListFilesHandler struct {
	ProjectManager project.Manager
}

func (h ListFilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectID, err := getProjectID(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid project ID: %s", err), http.StatusBadRequest)
		return
	}

	opts := getListFilesOptions(r)

	files, err := h.ProjectManager.ListFiles(r.Context(), projectID, opts...)
	if err != nil {
		var projectNotFoundError *project.ProjectNotFoundError
		if errors.As(err, &projectNotFoundError) {
			http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("Failed to list files: %s", err), http.StatusInternalServerError)
		return
	}

	if getAcceptFormat(r) == "text/plain" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(files.String()))
		return
	}

	var response []FileInfo

	for _, file := range files {
		response = append(response, FileInfo{Path: file.Path})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
