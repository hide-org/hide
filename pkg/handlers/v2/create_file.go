package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hide-org/hide/pkg/files"
)

type CreateFileRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type CreateFileHandler struct {
	Files files.Service
}

func (h CreateFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request CreateFileRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("failed parsing request body: %s", err), http.StatusBadRequest)
		return
	}

	file, err := h.Files.CreateFile(r.Context(), request.Path, request.Content)
	if err != nil {
		var fileAlreadyExistsError *files.FileAlreadyExistsError
		if errors.As(err, &fileAlreadyExistsError) {
			http.Error(w, fileAlreadyExistsError.Error(), http.StatusConflict)
			return
		}

		http.Error(w, fmt.Sprintf("failed to create file: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(file)
	return 
}
