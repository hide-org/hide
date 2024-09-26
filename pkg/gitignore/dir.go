package gitignore

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/plumbing/format/gitignore"
	"github.com/spf13/afero"
)

const (
	commentPrefix = "#"
	coreSection   = "core"
	eol           = "\n"
	excludesfile  = "excludesfile"
	gitDir        = ".git"
	gitignoreFile = ".gitignore"
	gitconfigFile = ".gitconfig"
	systemFile    = "/etc/gitconfig"
)

// ReadPatterns reads gitignore patterns recursively traversing through the directory
// structure. The result is in the ascending order of priority (last higher).
func ReadPatterns(fs afero.Fs, path []string) (ps []gitignore.Pattern, err error) {
	ps, _ = readIgnoreFile(fs, path, gitignoreFile)

	var fis []os.FileInfo
	fis, err = readDir(filepath.Join(path...))
	if err != nil {
		return
	}

	for _, fi := range fis {
		if fi.IsDir() && fi.Name() != gitDir {
			var subps []gitignore.Pattern
			subps, err = ReadPatterns(fs, append(path, fi.Name()))
			if err != nil {
				return
			}

			if len(subps) > 0 {
				ps = append(ps, subps...)
			}
		}
	}

	return
}

// readIgnoreFile reads a specific git ignore file.
func readIgnoreFile(fs afero.Fs, path []string, ignoreFile string) (ps []gitignore.Pattern, err error) {
	f, err := fs.Open(filepath.Join(append(path, ignoreFile)...))
	if err == nil {
		defer f.Close()

		if data, err := io.ReadAll(f); err == nil {
			for _, s := range strings.Split(string(data), eol) {
				if !strings.HasPrefix(s, commentPrefix) && len(strings.TrimSpace(s)) > 0 {
					ps = append(ps, gitignore.ParsePattern(s, path))
				}
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	return
}

func readDir(dir string) ([]os.FileInfo, error) {
	abs, err := filepath.Abs(dir) // TODO: check if that corresponds to func (fs *BoundOS) abs(filename string) (string, error) {} in billy
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}

	infos := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		fi, err := entry.Info()
		if err != nil {
			return nil, err
		}
		infos = append(infos, fi)
	}
	return infos, nil
}
