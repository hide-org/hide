package lsp

import (
	"sync"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// In memory store for diagnostics. Applies mutex locking for concurrent access.
type DiagnosticsStore struct {
	diagnostics map[ProjectId]map[protocol.DocumentUri][]protocol.Diagnostic
	mu          sync.Mutex
}

func (d *DiagnosticsStore) Get(projectId ProjectId, uri protocol.DocumentUri) ([]protocol.Diagnostic, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if diagnostics, ok := d.diagnostics[projectId]; ok {
		if diagnostics, ok := diagnostics[uri]; ok {
			return diagnostics, true
		}
	}

	return nil, false
}

func (d *DiagnosticsStore) Set(projectId ProjectId, uri protocol.DocumentUri, diagnostics []protocol.Diagnostic) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.diagnostics[projectId]; !ok {
		d.diagnostics[projectId] = make(map[protocol.DocumentUri][]protocol.Diagnostic)
	}

	d.diagnostics[projectId][uri] = diagnostics
}

func (d *DiagnosticsStore) DeleteAllForProject(projectId ProjectId) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.diagnostics, projectId)
}

func NewDiagnosticsStore() *DiagnosticsStore {
	return &DiagnosticsStore{
		diagnostics: make(map[ProjectId]map[protocol.DocumentUri][]protocol.Diagnostic),
	}
}
