package handlers

import (
	"errors"
	"net/http"

	"github.com/hide-org/hide/pkg/files"
)

type DeleteFileHandler struct {
	Files files.Service
}

func (h DeleteFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filePath, err := GetFilePath(r)
	if err != nil {
		http.Error(w, "invalid file path", http.StatusBadRequest)
		return
	}

	if err := h.Files.DeleteFile(r.Context(), filePath); err != nil {
		var fileNotFoundError *files.FileNotFoundError
		if errors.As(err, &fileNotFoundError) {
			http.Error(w, fileNotFoundError.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, "failed to delete file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return 
}
