package lsp

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// TODO: update docs
// In memory store for diagnostics. Applies mutex locking for concurrent access.
type DiagnosticsService struct {
	diagnostics map[ProjectId]map[protocol.DocumentUri][]protocol.Diagnostic
	listeners   map[ProjectId]map[LanguageId]*listenerInfo
	mu          sync.RWMutex
}

type listenerInfo struct {
	cancel  context.CancelFunc
	channel chan protocol.PublishDiagnosticsParams
}

func NewDiagnosticsService() *DiagnosticsService {
	return &DiagnosticsService{
		diagnostics: make(map[ProjectId]map[protocol.DocumentUri][]protocol.Diagnostic),
		listeners:   make(map[ProjectId]map[LanguageId]*listenerInfo),
	}
}

func (d *DiagnosticsService) Get(projectId ProjectId, uri protocol.DocumentUri) ([]protocol.Diagnostic, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if diagnostics, ok := d.diagnostics[projectId]; ok {
		if diagnostics, ok := diagnostics[uri]; ok {
			return diagnostics, true
		}
	}

	return nil, false
}

func (d *DiagnosticsService) set(projectId ProjectId, uri protocol.DocumentUri, diagnostics []protocol.Diagnostic) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.diagnostics[projectId]; !ok {
		d.diagnostics[projectId] = make(map[protocol.DocumentUri][]protocol.Diagnostic)
	}

	d.diagnostics[projectId][uri] = diagnostics
}

func (d *DiagnosticsService) StartListener(projectId ProjectId, languageId LanguageId, channel chan protocol.PublishDiagnosticsParams) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.listeners[projectId]; !exists {
		d.listeners[projectId] = make(map[LanguageId]*listenerInfo)
	}

	// TODO: should we cancel or return error?
	// Cancel existing listener if any
	if info, exists := d.listeners[projectId][languageId]; exists {
		info.cancel()
		close(info.channel)
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.listeners[projectId][languageId] = &listenerInfo{
		cancel:  cancel,
		channel: channel,
	}

	go d.listen(ctx, projectId, languageId, channel)
}

func (d *DiagnosticsService) StopListener(projectId ProjectId, languageId LanguageId) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if projectListeners, exists := d.listeners[projectId]; exists {
		if info, exists := projectListeners[languageId]; exists {
			info.cancel()
			close(info.channel)
			delete(projectListeners, languageId)
		}
		if len(projectListeners) == 0 {
			delete(d.listeners, projectId)
		}
	}
}

func (d *DiagnosticsService) listen(ctx context.Context, projectId ProjectId, languageId LanguageId, channel chan protocol.PublishDiagnosticsParams) {
	for {
		select {
		case <-ctx.Done():
			return
		case diagnostics, ok := <-channel:
			if !ok {
				return
			}
			d.set(projectId, diagnostics.URI, diagnostics.Diagnostics)
			log.Debug().
				Str("projectId", string(projectId)).
				Str("languageId", string(languageId)).
				Str("uri", string(diagnostics.URI)).
				Msgf("Received diagnostics: %d", len(diagnostics.Diagnostics))
		}
	}
}

func (d *DiagnosticsService) DeleteAllForProject(projectId ProjectId) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.diagnostics, projectId)
	if projectListeners, exists := d.listeners[projectId]; exists {
		for _, info := range projectListeners {
			info.cancel()
			close(info.channel)
		}
		delete(d.listeners, projectId)
	}
}

