package lang

import (
	"context"
	"encoding/json"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

type ServerName string

type Binary struct {
	// Path to the binary
	Path string
	// Command line arguments
	Arguments []string
	// Environment variables
	Env map[string]string
}

type Adapter interface {
	// Name returns the unique identifier for this language server
	Name() ServerName

	// FetchLatestServerVersion retrieves the latest available version info
	FetchLatestServerVersion(ctx context.Context, delegate Delegate) (interface{}, error)

	// FetchServerBinary downloads and prepares the language server binary
	FetchServerBinary(ctx context.Context, version interface{}, delegate Delegate) (*Binary, error)

	// InitializationOptions provides server-specific initialization options. See protocol.InitializeParams
	InitializationOptions(ctx context.Context, delegate Delegate) json.RawMessage

	// WorkspaceConfiguration configure the language server's behavior for the workspace. In devcontainer those are typically in customizations.
	//
	// Typically is applied with with lspCli.Notify(ctx, "workspace/didChangeConfiguration", protocol.DidChangeConfigurationParams{})
	WorkspaceConfiguration(ctx context.Context, delegate Delegate) (json.RawMessage, error)

	// CodeActions returns code actions supported by the LSP
	CodeActions() ([]protocol.CodeActionKind, error)
}
