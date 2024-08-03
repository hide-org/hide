package languageserver

import (
	"context"
	"log"

	"github.com/artmoskvin/hide/pkg/model"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type Service interface {
	Initialize(ctx context.Context, params protocol.InitializeParams) (protocol.InitializeResult, error)
	NotifyInitialized(ctx context.Context) error
	NotifyDidOpen(project model.Project, file model.File) error
	NotifyDidChange(ctx context.Context, params protocol.DidChangeTextDocumentParams) error
	NotifyDidChangeWorkspaceFolders(ctx context.Context, params protocol.DidChangeWorkspaceFoldersParams) error
	// TODO: check if any LSP server supports this
	// PullDiagnostics(ctx context.Context, params DocumentDiagnosticParams) (DocumentDiagnosticReport, error)
	GetDiagnostics(projectId ProjectId, file model.File) []protocol.Diagnostic
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

// Initialize implements Service.
func (s *ServiceImpl) Initialize(ctx context.Context, params protocol.InitializeParams) (protocol.InitializeResult, error) {
	panic("unimplemented")
}

// NotifyDidChange implements Service.
func (s *ServiceImpl) NotifyDidChange(ctx context.Context, params protocol.DidChangeTextDocumentParams) error {
	panic("unimplemented")
}

// NotifyDidChangeWorkspaceFolders implements Service.
func (s *ServiceImpl) NotifyDidChangeWorkspaceFolders(ctx context.Context, params protocol.DidChangeWorkspaceFoldersParams) error {
	panic("unimplemented")
}

// NotifyDidOpen implements Service.
func (s *ServiceImpl) NotifyDidOpen(project model.Project, file model.File) error {
	languageId := s.languageDetector.DetectLanguage(file)
	client := s.getOrCreateLspClient(project, languageId)
	uri := PathToURI(file.Path)

	err := client.NotifyDidOpen(context.Background(), protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: languageId,
			Version:    1,
			Text:       file.Content,
		},
	})

	return err
}

// NotifyInitialized implements Service.
func (s *ServiceImpl) NotifyInitialized(ctx context.Context) error {
	panic("unimplemented")
}

func (s *ServiceImpl) GetDiagnostics(projectId ProjectId, file model.File) []protocol.Diagnostic {
	uri := PathToURI(file.Path)
	if diagnostics, ok := s.lspDiagnostics[projectId]; ok {
		return diagnostics[uri]
	}

	return nil
}

func (s *ServiceImpl) getOrCreateLspClient(project model.Project, languageId LanguageId) Client {
	projectId := project.Id

	if _, ok := s.lspClients[projectId]; !ok {
		s.lspClients[projectId] = make(map[LanguageId]Client)
	}

	if _, ok := s.lspClients[projectId][languageId]; !ok {
		// Debug
		log.Printf("Creating LSP client for project %s and language %s", projectId, languageId)

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
			// Debug
			log.Printf("Received diagnostics for %s in project %s", diagnostics.URI, projectId)

			if projectDiagnostics, ok := s.lspDiagnostics[projectId]; ok {
				var documentDiagnostics []protocol.Diagnostic

				if documentDiagnostics, ok := projectDiagnostics[diagnostics.URI]; ok {
					documentDiagnostics = append(documentDiagnostics, diagnostics.Diagnostics...)
				} else {
					documentDiagnostics = diagnostics.Diagnostics
				}

				projectDiagnostics[projectId] = documentDiagnostics
			} else {
				s.lspDiagnostics[projectId] = make(map[protocol.DocumentUri][]protocol.Diagnostic)
				s.lspDiagnostics[projectId][diagnostics.URI] = diagnostics.Diagnostics
			}
		}
	}
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
