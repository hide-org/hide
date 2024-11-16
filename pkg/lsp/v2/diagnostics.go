package lsp

import (
	"sync"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// In memory store for diagnostics. Applies mutex locking for concurrent access.
type DiagnosticsStore struct {
	diagnostics map[protocol.DocumentUri][]protocol.Diagnostic
	mu          sync.Mutex
}

func (d *DiagnosticsStore) Get(uri protocol.DocumentUri) ([]protocol.Diagnostic, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	diagnostics, ok := d.diagnostics[uri]
	return diagnostics, ok
}

func (d *DiagnosticsStore) Set(uri protocol.DocumentUri, diagnostics []protocol.Diagnostic) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.diagnostics[uri] = diagnostics
	return
}



func NewDiagnosticsStore() *DiagnosticsStore {
	return &DiagnosticsStore{
		diagnostics: make(map[protocol.DocumentUri][]protocol.Diagnostic),
	}
}
