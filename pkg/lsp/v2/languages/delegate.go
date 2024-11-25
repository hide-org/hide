package lang

import (
	"context"
	"io"
	"net/http"
	"path/filepath"

	"github.com/spf13/afero"
)

type Delegate interface {
	Get(ctx context.Context, uri string) ([]byte, error)

	ProjectRootPath() string
	ReadFile(ctx context.Context, path string) ([]byte, error)

	MakeInstallPath(ctx context.Context, lspName string, version string) (path string, error error)
	Exist(ctx context.Context, path string) bool
}

func NewDefaultDelegate(fs afero.Fs, cli http.Client, rootDir string, binDir string) *defaultDelegate {
	return &defaultDelegate{
		fs:      fs,
		cli:     cli,
		rootDir: rootDir,
		binDir:  binDir,
	}
}

type defaultDelegate struct {
	cli     http.Client
	fs      afero.Fs
	rootDir string
	binDir  string
}

func (d *defaultDelegate) Get(ctx context.Context, uri string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	// TODO: check for status
	resp, err := d.cli.Do(req)
	if err != nil {
		return nil, err
	}

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (d *defaultDelegate) ProjectRootPath() string {
	return d.rootDir
}

func (d *defaultDelegate) ReadFile(ctx context.Context, path string) ([]byte, error) {
	return afero.ReadFile(d.fs, path)
}

func (d *defaultDelegate) MakeInstallPath(ctx context.Context, lspName string, version string) (path string, err error) {
	dir := filepath.Join(d.binDir, lspName, version)
	if err := d.fs.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func (d *defaultDelegate) Exist(ctx context.Context, path string) bool {
	if _, err := d.fs.Stat(path); err != nil {
		return false
	}
	return true
}
