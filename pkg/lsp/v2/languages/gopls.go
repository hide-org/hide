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

	protocol "github.com/tliron/glsp/protocol_3_16"
)

var _ Adapter = (*gopls)(nil)

type gopls struct{}

type goplsVersion struct {
	Version string `json:"Version"`
	Time    string `json:"Time"`
	Origin  struct {
		VCS    string `json:"VCS"`
		URL    string `json:"URL"`
		Subdir string `json:"Subdir"`
		Hash   string `json:"Hash"`
		Ref    string `json:"Ref"`
	} `json:"Origin"`
}

func (a *gopls) Name() ServerName {
	return "gopls"
}

func (a *gopls) FetchLatestServerVersion(ctx context.Context, delegate Delegate) (interface{}, error) {
	body, err := delegate.Get(ctx, "https://proxy.golang.org/golang.org/x/tools/gopls/@latest")
	if err != nil {
		return nil, err
	}

	var version goplsVersion
	if err := json.Unmarshal(body, &version); err != nil {
		return nil, err
	}

	return version, nil
}

func (a *gopls) FetchServerBinary(ctx context.Context, version interface{}, delegate Delegate) (*Binary, error) {
	var ver string

	if v, ok := version.(*goplsVersion); ok {
		ver = v.Version
	} else {
		ver = "latest"
	}

	installDir, err := delegate.MakeInstallPath(ctx, "gopls", ver)
	if err != nil {
		return nil, err
	}

	binaryName := "gopls"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	goplsPath := filepath.Join(installDir, binaryName)

	if exists, _ := checkVersion(goplsPath, ver); exists {
		return &Binary{
			Path:      goplsPath,
			Arguments: []string{"serve"},
			Env: map[string]string{
				"GOPATH": os.Getenv("GOPATH"),
				"GOROOT": os.Getenv("GOROOT"),
			},
		}, nil
	}

	cmd := exec.CommandContext(ctx, "go", "install", fmt.Sprintf("golang.org/x/tools/gopls@%s", ver))
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOBIN=%s", installDir),
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to install gopls: %s: %w", string(output), err)
	}

	return &Binary{
		Path:      goplsPath,
		Arguments: []string{"serve"},
		Env: map[string]string{
			"GOPATH": os.Getenv("GOPATH"),
			"GOROOT": os.Getenv("GOROOT"),
		},
	}, nil
}

func checkVersion(path string, wantVersion string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		return false, nil
	}

	cmd := exec.Command(path, "version")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	version := strings.TrimSpace(string(output)) //"gopls v0.16.2"
	if wantVersion == "latest" {
		return true, nil
	}

	return strings.Contains(version, wantVersion), nil
}

func (a *gopls) InitializationOptions(ctx context.Context, delegate Delegate) json.RawMessage {
	options := map[string]interface{}{
		"analyses": map[string]bool{
			"shadow": true, // Detect shadowed variables - helpful for catching subtle bugs
			"useany": true, // Suggests using 'any' instead of 'interface{}' for Go 1.18+
		},
		"staticcheck": true,            // Enable staticcheck analyzer - provides additional high-quality checks
		"memoryMode":  "DegradeClosed", // Better memory management for large projects
		"templateExtensions": []string{
			".tmpl",
			".gotmpl",
			".html.tmpl",
			".gohtml",
		},
	}

	// should always marshal
	out, _ := json.Marshal(options)
	return out
}

func (a *gopls) WorkspaceConfiguration(ctx context.Context, delegate Delegate) (json.RawMessage, error) {
	return nil, nil
}

func (a *gopls) CodeActions() ([]protocol.CodeActionKind, error) {
	return nil, nil
}
