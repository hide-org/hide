package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hide-org/hide/pkg/lsp/v2"
	"github.com/hide-org/hide/pkg/symbols"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type searchSymbolsOptions struct {
	maxLimit     int
	limit        int
	symbolFilter lsp.SymbolFilter
}

type SearchSymbolsHandler struct {
	symbols symbols.Service
	opts    *searchSymbolsOptions
}

type SearchSymbolsOptions func(opts *searchSymbolsOptions)

func SearchSymbolsMaxLimit(n int) SearchSymbolsOptions {
	return func(opts *searchSymbolsOptions) {
		opts.maxLimit = n
	}
}

func SearchSymbolsLimit(n int) SearchSymbolsOptions {
	return func(opts *searchSymbolsOptions) {
		opts.limit = n
	}
}

func IncludeSymbols(symbols ...protocol.SymbolKind) SearchSymbolsOptions {
	return func(opts *searchSymbolsOptions) {
		opts.symbolFilter = lsp.NewIncludeSymbolFilter(symbols...)
	}
}

func ExcludedSymbols(symbols ...protocol.SymbolKind) SearchSymbolsOptions {
	return func(opts *searchSymbolsOptions) {
		opts.symbolFilter = lsp.NewExcludeSymbolFilter(symbols...)
	}
}

func NewSearchSymbolsHandler(symbols symbols.Service, opts ...SearchSymbolsOptions) SearchSymbolsHandler {
	options := &searchSymbolsOptions{
		maxLimit:     100,
		limit:        10,
		symbolFilter: lsp.NewExcludeSymbolFilter(protocol.SymbolKindField, protocol.SymbolKindVariable),
	}

	for _, o := range opts {
		o(options)
	}

	h := SearchSymbolsHandler{
		symbols: symbols,
		opts:    options,
	}

	return h
}

func (h SearchSymbolsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "query not specified", http.StatusBadRequest)
		return
	}

	limit := h.opts.limit

	if r.URL.Query().Has("limit") {
		limitParam := r.URL.Query().Get("limit")
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid limit %s: %s", limitParam, err), http.StatusBadRequest)
			return
		}

		if limit <= 0 || limit > h.opts.maxLimit {
			http.Error(w, fmt.Sprintf("limit must be between 1 and %d", h.opts.maxLimit), http.StatusBadRequest)
			return
		}
	}

	// TODO: check how we handle nil response
	symbols, err := h.symbols.Search(r.Context(), query, h.opts.symbolFilter)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get workspace symbols: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(symbols[:min(limit, len(symbols))])
	return
}
