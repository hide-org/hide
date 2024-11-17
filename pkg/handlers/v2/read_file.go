package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hide-org/hide/pkg/files"
)

type ReadFileHandler struct {
	Files files.Service
}

func (h ReadFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filePath, err := GetFilePath(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid file path: %s", err), http.StatusBadRequest)
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

	file, err := h.Files.ReadFile(r.Context(), filePath)
	if err != nil {
		var fileNotFoundError *files.FileNotFoundError
		if errors.As(err, &fileNotFoundError) {
			http.Error(w, fileNotFoundError.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("failed to read file: %s", err), http.StatusInternalServerError)
		return
	}

	if startLinePresent || numLinesPresent {
		if startLinePresent {
			if startLine < 1 || startLine > len(file.Lines) {
				http.Error(w, fmt.Sprintf("start line must be between 1 and %d", len(file.Lines)), http.StatusBadRequest)
				return
			}
		} else {
			startLine = 1
		}

		endLine := len(file.Lines) + 1
		if numLinesPresent {
			endLine = startLine + numLines
			if endLine > len(file.Lines)+1 {
				endLine = len(file.Lines) + 1
			}
		}

		file = file.WithLineRange(startLine, endLine)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(file)
	return
}
