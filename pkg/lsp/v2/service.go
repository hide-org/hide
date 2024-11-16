package lsp

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/hide-org/hide/pkg/model"
	"github.com/rs/zerolog/log"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type (
	LanguageId     = string
	ProjectRoot    = string
	LspClientStore = map[LanguageId]Client
	LspDiagnostics = map[protocol.DocumentUri][]protocol.Diagnostic
)

var LspServerExecutables = map[LanguageId]Command{
	Go:         NewCommand("gopls", []string{}),
	Python:     NewCommand("pyright-langserver", []string{"--stdio"}),
	JavaScript: NewCommand("typescript-language-server", []string{"--stdio"}),
	TypeScript: NewCommand("typescript-language-server", []string{"--stdio"}),
}

type Service interface {
	StartServer(ctx context.Context, languageId LanguageId) error
	StopServer(ctx context.Context, languageId LanguageId) error
	GetWorkspaceSymbols(ctx context.Context, query string, symbolFilter SymbolFilter) ([]SymbolInfo, error)
	GetDocumentOutline(ctx context.Context, file model.File) (DocumentOutline, error)
	NotifyDidOpen(ctx context.Context, file model.File) error
	NotifyDidClose(ctx context.Context, file model.File) error
	// TODO: check if any LSP server supports this
	// PullDiagnostics(ctx context.Context, params DocumentDiagnosticParams) (DocumentDiagnosticReport, error)
	GetDiagnostics(ctx context.Context, file model.File) ([]protocol.Diagnostic, error)
	Cleanup(ctx context.Context) error
}

type ServiceImpl struct {
	languageDetector     LanguageDetector
	clientPool           ClientPool
	diagnosticsStore     *DiagnosticsStore
	lspServerExecutables map[LanguageId]Command
	// TODO: can we pass the root URI as url.URL?
	rootURI              string // example: "file:///workspace"
}

// StartServer implements Service.
func (s *ServiceImpl) StartServer(ctx context.Context, languageId LanguageId) error {
	command, ok := s.lspServerExecutables[languageId]
	if !ok {
		return NewLanguageNotSupportedError(languageId)
	}

	// Start the language server
	process, err := NewProcess(command)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create language server process")
		return fmt.Errorf("Failed to create language server process: %w", err)
	}

	if err := process.Start(); err != nil {
		log.Error().Err(err).Msg("Failed to start language server")
		return fmt.Errorf("Failed to start language server: %w", err)
	}

	// Create a client for the language server
	client, diagnostics := NewClient(process)

	// Initialize the language server
	root := DocumentURI(s.rootURI)
	initResult, err := client.Initialize(ctx, protocol.InitializeParams{
		RootURI: &root,
		Capabilities: protocol.ClientCapabilities{
			TextDocument: &protocol.TextDocumentClientCapabilities{
				Synchronization: &protocol.TextDocumentSyncClientCapabilities{
					DynamicRegistration: boolPointer(true),
				},
				DocumentSymbol: &protocol.DocumentSymbolClientCapabilities{
					HierarchicalDocumentSymbolSupport: boolPointer(true),
				},
			},
		},
		WorkspaceFolders: []protocol.WorkspaceFolder{
			{
				URI: root,
				// Name: project.Id, TODO: remove or use root
			},
		},
	})
	if err != nil {
		log.Error().Str("languageId", languageId).Err(err).Msg("Failed to initialize language server")
		return fmt.Errorf("failed to initialize language server: %w", err)
	}

	log.Debug().Str("languageId", languageId).Msg("Initialized language server")

	// Check capabilities
	if opt, ok := initResult.Capabilities.TextDocumentSync.(protocol.TextDocumentSyncOptions); ok {
		log.Debug().Str("languageId", languageId).Msgf("LSP server supports open/close file: %t", *opt.OpenClose)
		log.Debug().Str("languageId", languageId).Msgf("LSP server supports change notifications: %v", *opt.Change)
	}

	// Notify that initialized
	if err := client.NotifyInitialized(ctx); err != nil {
		log.Error().Err(err).Str("languageId", languageId).Msg("Failed to notify initialized")
		return fmt.Errorf("Failed to notify initialized: %w", err)
	}

	s.clientPool.Set(languageId, client)

	// TODO: kill this goroutine when the project is deleted
	go s.listenForDiagnostics(diagnostics)
	return nil
}

func (s *ServiceImpl) StopServer(ctx context.Context, languageId LanguageId) error {
	client, ok := s.getClient(ctx, languageId)

	if !ok {
		log.Warn().Str("languageId", languageId).Msg("LSP client not found")
		return nil
	}

	if err := client.Shutdown(ctx); err != nil {
		log.Error().Err(err).Str("languageId", languageId).Msg("Failed to stop language server")
		return fmt.Errorf("Failed to stop language server: %w", err)
	}

	s.clientPool.Delete(languageId)

	return nil
}

func (s *ServiceImpl) GetWorkspaceSymbols(ctx context.Context, query string, symbolFilter SymbolFilter) ([]SymbolInfo, error) {
	clients := s.getClients(ctx)
	if len(clients) == 0 {
		log.Warn().Msg("LSP client not found")
		return nil, nil
	}

	symbols := []protocol.SymbolInformation{}
	for _, client := range clients {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result, err := client.GetWorkspaceSymbols(ctx, protocol.WorkspaceSymbolParams{Query: query})
		if err != nil {
			return nil, err
		}

		for _, symbol := range result {
			if symbolFilter.shouldExcludeSymbol(symbol) {
				continue
			}

			if symbolFilter.shouldIncludeSymbol(symbol) {
				symbols = append(symbols, symbol)
			}
		}
	}

	result := make([]SymbolInfo, 0, len(symbols))
	for _, symbol := range symbols {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		symbolPath, err := removeFilePrefix(symbol.Location.URI)
		if err != nil {
			log.Error().Err(err).Str("URI", symbol.Location.URI).Msg("failed to remove file prefix from URI")
			return nil, fmt.Errorf("failed to remove file prefix from URI: %w", err)
		}

		url, err := url.Parse(symbolPath)
		if err != nil {
			log.Error().Err(err).Str("path", symbolPath).Msg("failed to parse path")
			return nil, fmt.Errorf("failed to parse path: %w", err)
		}

		relativePath, err := filepath.Rel(url.Path, symbolPath)
		if err != nil {
			log.Error().Err(err).Str("base", url.Path).Str("path", symbolPath).Msg("failed to get relative path of file")
			return nil, fmt.Errorf("failed to get relative path of file: %w", err)
		}

		result = append(result, SymbolInfo{
			Name: symbol.Name,
			Kind: symbolKindToString(symbol.Kind),
			// NOTE: LSP uses 0-based line numbers, but Hide uses 1-based. Characters remain 0-based.
			Location: Location{Path: relativePath, Range: Range{Start: Position{Line: int(symbol.Location.Range.Start.Line) + 1, Character: int(symbol.Location.Range.Start.Character)}, End: Position{Line: int(symbol.Location.Range.End.Line) + 1, Character: int(symbol.Location.Range.End.Character)}}},
		})
	}

	return result, nil
}

func (s *ServiceImpl) GetDocumentOutline(ctx context.Context, file model.File) (DocumentOutline, error) {
	lang := s.languageDetector.DetectLanguage(&file)

	cli, ok := s.getClient(ctx, lang)
	if !ok {
		return DocumentOutline{}, NewLanguageServerNotFoundError(lang)
	}

	symbols, err := cli.GetDocumentSymbols(ctx, protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: DocumentURI(file.Path),
		},
	})
	if err != nil {
		return DocumentOutline{}, err
	}

	return documentOutlineFrom(symbols, file.Path), nil
}

// NotifyDidClose implements Service.
func (s *ServiceImpl) NotifyDidClose(ctx context.Context, file model.File) error {
	languageId := s.languageDetector.DetectLanguage(&file)
	client, ok := s.getClient(ctx, languageId)

	if !ok {
		log.Warn().Str("languageId", languageId).Msg("LSP client not found")
		return NewLanguageServerNotFoundError(languageId)
	}

	err := client.NotifyDidClose(ctx, protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: DocumentURI(file.Path),
		},
	})

	return err
}

// NotifyDidOpen implements Service.
func (s *ServiceImpl) NotifyDidOpen(ctx context.Context, file model.File) error {
	languageId := s.languageDetector.DetectLanguage(&file)
	client, ok := s.getClient(ctx, languageId)

	if !ok {
		log.Warn().Str("languageId", languageId).Msg("LSP client not found")
		return NewLanguageServerNotFoundError(languageId)
	}

	err := client.NotifyDidOpen(ctx, protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     DocumentURI(file.Path),
			Version: 1,
			Text:    file.GetContent(),
		},
	})

	return err
}

func (s *ServiceImpl) GetDiagnostics(ctx context.Context, file model.File) ([]protocol.Diagnostic, error) {
	uri := DocumentURI(file.Path)
	if diagnostics, ok := s.diagnosticsStore.Get(uri); ok {
		return diagnostics, nil
	}

	return nil, nil
}

func (s *ServiceImpl) Cleanup(ctx context.Context) error {
	for _, client := range s.clientPool.GetAll() {
		if err := client.Shutdown(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *ServiceImpl) getClient(ctx context.Context, languageId LanguageId) (Client, bool) {
	client, ok := s.clientPool.Get(languageId)
	return client, ok
}

func (s *ServiceImpl) getClients(ctx context.Context) []Client {
	clients := make([]Client, 0)
	for _, client := range s.clientPool.GetAll() {
		clients = append(clients, client)
	}

	return clients
}

func (s *ServiceImpl) listenForDiagnostics(channel <-chan protocol.PublishDiagnosticsParams) {
	log.Debug().Msg("Start listening")

	// reads fro chanel util it is closed
	for diagnostics := range channel {
		log.Debug().Str("uri", diagnostics.URI).Msg("Received diagnostics")
		log.Debug().Str("uri", diagnostics.URI).Msgf("Diagnostics: %+v", diagnostics.Diagnostics)

		s.updateDiagnostics(diagnostics)
	}

	log.Debug().Msg("Done listening")
}

func (s *ServiceImpl) updateDiagnostics(diagnostics protocol.PublishDiagnosticsParams) {
	s.diagnosticsStore.Set(diagnostics.URI, diagnostics.Diagnostics)
}

func DocumentURI(pathURI string) protocol.DocumentUri {
	return protocol.DocumentUri(pathURI)
}

func NewService(languageDetector LanguageDetector, lspServerExecutables map[LanguageId]Command, diagnosticsStore *DiagnosticsStore, clientPool ClientPool) Service {
	return &ServiceImpl{
		languageDetector:     languageDetector,
		clientPool:           clientPool,
		diagnosticsStore:     diagnosticsStore,
		lspServerExecutables: lspServerExecutables,
	}
}

func boolPointer(b bool) *bool {
	return &b
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
