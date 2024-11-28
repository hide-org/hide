package lang

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rs/zerolog/log"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var _ Adapter = (*tsserver)(nil)

type tsserver struct{}

type tsVersion struct {
	Version string `json:"version"`
}

// checkNodeRequirements verifies that Node.js and npm are available
func checkNodeRequirements(ctx context.Context) error {
	// Check node version
	nodeCmd := exec.CommandContext(ctx, "node", "--version")
	output, err := nodeCmd.Output()
	if err != nil {
		return fmt.Errorf("node.js is required but not available: %w", err)
	}
	nodeVersion := strings.TrimSpace(string(output))

	// Ensure minimum Node.js version (example: v14.0.0)
	version := strings.TrimPrefix(nodeVersion, "v")
	if !isVersionSufficient(version, "14.0.0") {
		return fmt.Errorf("node.js version %s is too old, minimum required is 14.0.0", version)
	}
	log.Debug().Msgf("Found Node.js version: %s", nodeVersion)

	// Check npm version
	npmBin := "npm"
	if runtime.GOOS == "windows" {
		npmBin += ".cmd"
	}
	npmCmd := exec.CommandContext(ctx, npmBin, "--version")
	output, err = npmCmd.Output()
	if err != nil {
		return fmt.Errorf("npm is required but not available: %w", err)
	}
	npmVersion := strings.TrimSpace(string(output))
	log.Debug().Msgf("Found npm version: %s", npmVersion)

	return nil
}

// Helper function to compare versions
func isVersionSufficient(current, minimum string) bool {
	currentParts := strings.Split(current, ".")
	minimumParts := strings.Split(minimum, ".")

	for i := 0; i < len(minimumParts); i++ {
		if i >= len(currentParts) {
			return false
		}
		if currentParts[i] < minimumParts[i] {
			return false
		}
		if currentParts[i] > minimumParts[i] {
			return true
		}
	}
	return true
}

func (a *tsserver) Name() ServerName {
	return "typescript-language-server"
}

func (a *tsserver) FetchServerBinary(ctx context.Context, version interface{}, delegate Delegate) (*Binary, error) {
	if err := checkNodeRequirements(ctx); err != nil {
		log.Error().Err(err).Msg("typescript-language-server: node requirements not met")
		return nil, err
	}

	var ver string
	if v, ok := version.(tsVersion); ok {
		ver = v.Version
	} else {
		ver = "latest"
	}

	installDir, err := delegate.MakeInstallPath(ctx, "typescript-language-server", ver)
	if err != nil {
		log.Error().Err(err).Msgf("typescript-language-server: failed to make install path %s", ver)
		return nil, err
	}

	serverPath := filepath.Join(installDir, "node_modules", ".bin", "typescript-language-server")
	if runtime.GOOS == "windows" {
		serverPath += ".cmd"
	}

	npmBin := "npm"
	if runtime.GOOS == "windows" {
		npmBin += ".cmd"
	}

	if !delegate.Exist(ctx, serverPath) {
		// Create a package.json if it doesn't exist
		pkgJSON := map[string]interface{}{
			"name":    "ts-lsp-install",
			"version": "1.0.0",
			"private": true,
		}
		pkgJSONBytes, _ := json.Marshal(pkgJSON)
		pkgJSONPath := filepath.Join(installDir, "package.json")
		if err := os.WriteFile(pkgJSONPath, pkgJSONBytes, 0644); err != nil {
			return nil, fmt.Errorf("failed to create package.json: %w", err)
		}

		// Install with all dependencies
		cmd := exec.CommandContext(ctx, npmBin, "install", "--save",
			"typescript-language-server@"+ver,
			"typescript@latest",
			"@typescript/vfs@latest",                      // Virtual file system support
			"@typescript-eslint/typescript-estree@latest", // For parsing
		)
		cmd.Dir = installDir // Set working directory to installDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("failed to install typescript-language-server: %s: %w", string(output), err)
		}
	}

	env := map[string]string{
		"NODE_PATH": filepath.Join(installDir, "node_modules"),
		"PATH":      os.Getenv("PATH"),
		// Enable logging for debugging
		"TSS_LOG": "-level verbose -file " + filepath.Join(installDir, "tsserver.log"),
		"DEBUG":   "typescript-language-server:*",
	}

	return &Binary{
		Name: a.Name(),
		Path: serverPath,
		Arguments: []string{
			"--stdio",              // Use stdio for communication
			"--log-level", "debug", // Enable debug logging
		},
		Env: env,
	}, nil
}

func (a *tsserver) FetchLatestServerVersion(ctx context.Context, delegate Delegate) (interface{}, error) {
	body, err := delegate.Get(ctx, "https://registry.npmjs.org/typescript-language-server/latest")
	if err != nil {
		log.Error().Err(err).Msg("typescript-language-server: failed to get version")
		return nil, err
	}

	var version tsVersion
	if err := json.Unmarshal(body, &version); err != nil {
		log.Error().Err(err).Msg("typescript-language-server: failed to unmarshal version")
		return nil, err
	}

	log.Debug().Msgf("typescript-language-server: got version %+v", version)
	return version, nil
}

func (a *tsserver) InitializationOptions(ctx context.Context, delegate Delegate) json.RawMessage {
	options := map[string]interface{}{
		"documentSymbols":                   true, // Enable document outline/symbols
		"hierarchicalDocumentSymbolSupport": true, // Enable hierarchical symbols
		"diagnostics":                       true, // Enable diagnostics
		"completions":                       true, // Enable completions
		"codeActions":                       true, // Enable code actions
		"hover":                             true, // Enable hover
		"implementation":                    true, // Enable go to implementation
		"references":                        true, // Enable find references
		"definition":                        true, // Enable go to definition
		"tsserver": map[string]interface{}{
			"maxTsServerMemory":        4096,
			"enableProjectDiagnostics": true,
		},
	}

	out, _ := json.Marshal(options)
	return out
}

func (a *tsserver) WorkspaceConfiguration(ctx context.Context, delegate Delegate) (json.RawMessage, error) {
	return nil, nil
}

func (a *tsserver) CodeActions() ([]protocol.CodeActionKind, error) {
	return []protocol.CodeActionKind{
		"quickfix",               // Standard quick fixes
		"refactor",               // Generic refactoring actions
		"refactor.extract",       // Extract code to function/variable
		"refactor.inline",        // Inline code from function/variable
		"refactor.rewrite",       // Rewrite code structure
		"source",                 // Source code modifications
		"source.organizeImports", // Organize imports
	}, nil
}

func (a *tsserver) Languages() []LanguageID {
	return []LanguageID{JavaScript, TypeScript, TSX, JSX}
}
