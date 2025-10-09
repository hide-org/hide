package lang

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

var _ Adapter = (*jdtls)(nil)

type jdtls struct{}

type jdtlsVersion struct {
	Version string `json:"version"`
}

func (a *jdtls) Name() ServerName {
	return "jdtls"
}

func (a *jdtls) FetchLatestServerVersion(ctx context.Context, delegate Delegate) (interface{}, error) {
	// jdtls does not expose a simple version endpoint. We rely on the binary
	// being present on PATH and treat it as "latest".
	return jdtlsVersion{Version: "latest"}, nil
}

func (a *jdtls) FetchServerBinary(ctx context.Context, version interface{}, delegate Delegate) (*Binary, error) {
	path, err := exec.LookPath("jdtls")
	if err != nil {
		return nil, fmt.Errorf("jdtls language server not found in PATH: %w", err)
	}

	dataDir := filepath.Join(delegate.ProjectRootPath(), ".jdtls")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create jdtls data directory: %w", err)
	}

	return &Binary{
		Name: a.Name(),
		Path: path,
		Arguments: []string{
			"-data", dataDir,
		},
	}, nil
}

func (a *jdtls) InitializationOptions(ctx context.Context, delegate Delegate) json.RawMessage {
	return nil
}

func (a *jdtls) WorkspaceConfiguration(ctx context.Context, delegate Delegate) (json.RawMessage, error) {
	return nil, nil
}

func (a *jdtls) CodeActions() ([]protocol.CodeActionKind, error) {
	return nil, nil
}

func (a *jdtls) Languages() []LanguageID {
	return []LanguageID{Java}
}
