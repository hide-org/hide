package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/artmoskvin/hide/pkg/lsp"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/artmoskvin/hide/pkg/project"
)

type SearchSymbolsHandler struct {
	pm         project.Manager
	lspService lsp.Service
}

func NewSearchSymbolsHandler(pm project.Manager, lspService lsp.Service) SearchSymbolsHandler {
	return SearchSymbolsHandler{pm: pm, lspService: lspService}
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

	proj, err := h.pm.GetProject(r.Context(), projectID)
	if err != nil {
		var projectNotFoundError *project.ProjectNotFoundError
		if errors.As(err, &projectNotFoundError) {
			http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("failed to get project: %s", err), http.StatusInternalServerError)
		return
	}

	symbols, err := h.lspService.GetWorkspaceSymbols(model.NewContextWithProject(r.Context(), &proj), query)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get workspace symbols: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(symbols)
}

func removeFilePrefix(fileURL string) (string, error) {
	u, err := url.Parse(fileURL)
	if err != nil {
		return "", err
	}
	if u.Scheme != "file" {
		return "", fmt.Errorf("not a file URL")
	}
	return filepath.FromSlash(u.Path), nil
}
