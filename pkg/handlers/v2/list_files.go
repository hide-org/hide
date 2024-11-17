package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hide-org/hide/pkg/files"
)

type FileInfo struct {
	Path string `json:"path"`
}

type ListFilesHandler struct {
	Files files.Service
}

func (h ListFilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	opts := getListFilesOptions(r)

	filez, err := h.Files.ListFiles(r.Context(), opts...)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list files: %s", err), http.StatusInternalServerError)
		return
	}

	if getAcceptFormat(r) == "text/plain" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(filez.String()))
		return
	}

	var response []FileInfo

	for _, file := range filez {
		response = append(response, FileInfo{Path: file.Path})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return
}
