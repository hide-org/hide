package lsp

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/artmoskvin/hide/pkg/model"
	"github.com/rs/zerolog/log"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type Service interface {
	NotifyDidOpen(ctx context.Context, file model.File) error
	NotifyDidClose(ctx context.Context, file model.File) error
	// TODO: check if any LSP server supports this
	// PullDiagnostics(ctx context.Context, params DocumentDiagnosticParams) (DocumentDiagnosticReport, error)
	GetDiagnostics(ctx context.Context, file model.File) []protocol.Diagnostic
	StopClient(ctx context.Context, file model.File) error
	Cleanup(ctx context.Context) error
	CleanupProject(ctx context.Context, projectId ProjectId) error
}

type ProjectId = string
type LanguageId = string
type ProjectRoot = string
type LspClientStore = map[ProjectId]map[LanguageId]Client
type LspDiagnostics = map[ProjectId]map[protocol.DocumentUri][]protocol.Diagnostic

type ServiceImpl struct {
	lspClients             LspClientStore
	lspClientFactoryMethod func(LanguageId, ProjectRoot, chan protocol.PublishDiagnosticsParams) Client
	lspDiagnostics         LspDiagnostics
	languageDetector       LanguageDetector
}

// NotifyDidClose implements Service.
func (s *ServiceImpl) NotifyDidClose(ctx context.Context, file model.File) error {
	project, ok := model.ProjectFromContext(ctx)

	if !ok {
		log.Error().Msg("Project not found in context")
		return fmt.Errorf("Project not found in context")
	}

	languageId := s.languageDetector.DetectLanguage(file)
	client := s.getOrCreateLspClient(*project, languageId)
	err := client.NotifyDidClose(ctx, protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: PathToURI(file.Path),
		},
	})

	return err
}

// NotifyDidOpen implements Service.
func (s *ServiceImpl) NotifyDidOpen(ctx context.Context, file model.File) error {
	project, ok := model.ProjectFromContext(ctx)

	if !ok {
		log.Error().Msg("Project not found in context")
		return fmt.Errorf("Project not found in context")
	}

	languageId := s.languageDetector.DetectLanguage(file)
	client := s.getOrCreateLspClient(*project, languageId)
	fullPath := filepath.Join(project.Path, file.Path)

	err := client.NotifyDidOpen(ctx, protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        PathToURI(fullPath),
			LanguageID: languageId,
			Version:    1,
			Text:       file.Content,
		},
	})

	return err
}

func (s *ServiceImpl) GetDiagnostics(ctx context.Context, file model.File) []protocol.Diagnostic {
	project, ok := model.ProjectFromContext(ctx)

	if !ok {
		log.Error().Msg("Project not found in context")
		return nil
	}

	uri := PathToURI(filepath.Join(project.Path, file.Path))
	if diagnostics, ok := s.lspDiagnostics[project.Id]; ok {
		return diagnostics[uri]
	}

	return nil
}

func (s *ServiceImpl) StopClient(ctx context.Context, file model.File) error {
	project, ok := model.ProjectFromContext(ctx)

	if !ok {
		log.Error().Msg("Project not found in context")
		return fmt.Errorf("Project not found in context")
	}

	languageId := s.languageDetector.DetectLanguage(file)

	if _, ok := s.lspClients[project.Id]; !ok {
		return nil
	}

	if _, ok := s.lspClients[project.Id][languageId]; !ok {
		return nil
	}

	if err := s.lspClients[project.Id][languageId].StopServer(); err != nil {
		return err
	}

	delete(s.lspClients[project.Id], languageId)

	return nil
}

func (s *ServiceImpl) Cleanup(ctx context.Context) error {
	for projectId := range s.lspClients {
		if err := s.CleanupProject(ctx, projectId); err != nil {
			return err
		}
	}

	return nil
}

func (s *ServiceImpl) CleanupProject(ctx context.Context, projectId ProjectId) error {
	panic("not implemented")
}

func (s *ServiceImpl) getOrCreateLspClient(project model.Project, languageId LanguageId) Client {
	projectId := project.Id

	if _, ok := s.lspClients[projectId]; !ok {
		s.lspClients[projectId] = make(map[LanguageId]Client)
	}

	if _, ok := s.lspClients[projectId][languageId]; !ok {
		log.Debug().Str("projectId", projectId).Str("languageId", languageId).Msg("Creating LSP client")

		diagnosticsChannel := make(chan protocol.PublishDiagnosticsParams)
		s.lspClients[projectId][languageId] = s.lspClientFactoryMethod(languageId, project.Path, diagnosticsChannel)
		go s.listenForDiagnostics(projectId, diagnosticsChannel)
	}

	return s.lspClients[projectId][languageId]
}

func (s *ServiceImpl) listenForDiagnostics(projectId ProjectId, channel chan protocol.PublishDiagnosticsParams) {
	for {
		select {
		case diagnostics := <-channel:
			log.Debug().Str("projectId", projectId).Str("uri", diagnostics.URI).Msg("Received diagnostics")
			log.Debug().Str("projectId", projectId).Str("uri", diagnostics.URI).Msgf("Diagnostics: %+v", diagnostics.Diagnostics)

			s.updateDiagnostics(projectId, diagnostics)
		}
	}
}

func (s *ServiceImpl) updateDiagnostics(projectId ProjectId, diagnostics protocol.PublishDiagnosticsParams) {
	if _, ok := s.lspDiagnostics[projectId]; !ok {
		s.lspDiagnostics[projectId] = make(map[protocol.DocumentUri][]protocol.Diagnostic)
	}

	s.lspDiagnostics[projectId][diagnostics.URI] = diagnostics.Diagnostics
}

func PathToURI(path string) protocol.DocumentUri {
	return protocol.DocumentUri("file://" + path)
}

func NewService(lspClientFactoryMethod func(LanguageId, ProjectRoot, chan protocol.PublishDiagnosticsParams) Client, languageDetector LanguageDetector) Service {
	return &ServiceImpl{
		lspClients:             make(LspClientStore),
		lspClientFactoryMethod: lspClientFactoryMethod,
		lspDiagnostics:         make(LspDiagnostics),
		languageDetector:       languageDetector,
	}
}

func ClientFactoryMethod(languageId, projectRoot string, diagnosticsChannel chan protocol.PublishDiagnosticsParams) Client {
	ctx := context.Background()

	// Define the lsp server executable based on the languageId (currently only "go" is supported)
	lspServerExecutable := "gopls"

	// Start the language server
	process, err := NewProcess(lspServerExecutable)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create language server process")
	}

	if err := process.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start language server")
	}

	// TODO: fix me
	// defer process.Stop()

	// Create a client for the language server
	client := NewClient(ctx, process, diagnosticsChannel)

	// Initialize the language server
	root := PathToURI(projectRoot)
	_, err = client.Initialize(context.Background(), protocol.InitializeParams{
		RootURI: &root,
		Capabilities: protocol.ClientCapabilities{
			TextDocument: &protocol.TextDocumentClientCapabilities{
				Synchronization: &protocol.TextDocumentSyncClientCapabilities{
					DynamicRegistration: boolPointer(true),
				},
			},
		},
		// WorkspaceFolders: []protocol.WorkspaceFolder{
		// 	{
		// 		URI:  root,
		// 		Name: "hide",
		// 	},
		// },
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize language server")
	}

	log.Debug().Str("languageId", languageId).Msg("Initialized language server")

	// if opt, ok := initResult.Capabilities.TextDocumentSync.(protocol.TextDocumentSyncOptions); ok {
	// 	log.Printf("Support open/close file: %t", *opt.OpenClose)
	// 	log.Printf("Support change notifications: %v", *opt.Change)
	// }

	// Notify that initialized
	if err := client.NotifyInitialized(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to notify initialized")
	}

	return client
}

func boolPointer(b bool) *bool {
	return &b
}
