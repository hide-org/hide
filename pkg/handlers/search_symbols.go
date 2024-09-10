package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/artmoskvin/hide/pkg/project"
)

const MinLimit = 1
const MaxLimit = 100
const DefaultLimit = 10

type SearchSymbolsHandler struct {
	pm project.Manager
}

func NewSearchSymbolsHandler(pm project.Manager) SearchSymbolsHandler {
	return SearchSymbolsHandler{pm: pm}
}

func (h SearchSymbolsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectID, err := getProjectID(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid project ID: %s", err), http.StatusBadRequest)
		return
	}

	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "query not specified", http.StatusBadRequest)
		return
	}

	limit := DefaultLimit

	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		limit, err = strconv.Atoi(limitParam)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid limit %s: %s", limitParam, err), http.StatusBadRequest)
			return
		}

		if limit < MinLimit || limit > MaxLimit {
			http.Error(w, "limit must be between 1 and 100", http.StatusBadRequest)
			return
		}
	}

	symbols, err := h.pm.SearchSymbols(r.Context(), projectID, query)
	if err != nil {
		var projectNotFoundError *project.ProjectNotFoundError
		if errors.As(err, &projectNotFoundError) {
			http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("failed to get workspace symbols: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(symbols[:min(limit, len(symbols))])
}
