package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hide-org/hide/pkg/outline"
)

type DocumentOutline struct {
	Outline outline.Service
}

func (h DocumentOutline) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filePath, err := GetFilePath(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid file path: %s", err), http.StatusBadRequest)
		return
	}

	outline, err := h.Outline.Get(r.Context(), filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create file outline: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(outline)
	return
}
