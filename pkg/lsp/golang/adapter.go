package golang

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hide-org/hide/pkg/lsp"
)

type Adapter struct{}

type GoplsVersion struct {
	Version string `json:"version"`
}

func (a *Adapter) FetchServerBinary(ctx context.Context, version interface{}, delegate lsp.Delegate) (*lsp.Binary, error) {
	var goplsVersion string

	if v, ok := version.(*GoplsVersion); ok {
		goplsVersion = v.Version
	} else {
		goplsVersion = "latest"
	}

	installDir, err := delegate.MakeInstallPath(ctx, "gopls", goplsVersion)
	if err != nil {
		return nil, err
	}

	binaryName := "gopls"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	goplsPath := filepath.Join(installDir, binaryName)

	if exists, _ := checkVersion(goplsPath, goplsVersion); exists {
		return &lsp.Binary{
			Path:      goplsPath,
			Arguments: []string{"serve"},
			Env: map[string]string{
				"GOPATH": os.Getenv("GOPATH"),
				"GOROOT": os.Getenv("GOROOT"),
			},
		}, nil
	}

	cmd := exec.CommandContext(ctx, "go", "install", fmt.Sprintf("golang.org/x/tools/gopls@%s", goplsVersion))
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOBIN=%s", installDir),
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to install gopls: %s: %w", string(output), err)
	}

	return &lsp.Binary{
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

	version := strings.TrimSpace(string(output)) //"gopls v0.11.0"
	if wantVersion == "latest" {
		return true, nil
	}

	return strings.Contains(version, wantVersion), nil
}
