package gitignore

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/plumbing/format/gitignore"
	"github.com/spf13/afero"
)

const (
	commentPrefix   = "#"
	coreSection     = "core"
	eol             = "\n"
	excludesfile    = "excludesfile"
	gitDir          = ".git"
	gitignoreFile   = ".gitignore"
	gitconfigFile   = ".gitconfig"
	systemFile      = "/etc/gitconfig"
	infoExcludeFile = gitDir + "/info/exclude"
)

// ReadPatterns reads gitignore patterns recursively traversing through the directory
// structure. The result is in the ascending order of priority (last higher).
func ReadPatterns(fs afero.Fs, path []string) ([]gitignore.Pattern, error) {
	path = path[:len(path):len(path)]

	ps, err := readIgnoreFile(fs, path, infoExcludeFile)
	if err != nil {
		return nil, err
	}

	subps, err := readIgnoreFile(fs, path, gitignoreFile)
	if err != nil {
		return nil, err
	}
	ps = append(ps, subps...)

	var fis []os.FileInfo
	fis, err = readDir(fs, filepath.Join(path...))
	if err != nil {
		return nil, err
	}

	for _, fi := range fis {
		if fi.IsDir() && fi.Name() != gitDir {
			var subps []gitignore.Pattern
			subps, err := ReadPatterns(fs, append(path, fi.Name()))
			if err != nil {
				return nil, err
			}

			if len(subps) > 0 {
				ps = append(ps, subps...)
			}
		}
	}

	return ps, nil
}

// readIgnoreFile reads a specific .gitignore file
func readIgnoreFile(fs afero.Fs, path []string, ignoreFile string) (ps []gitignore.Pattern, err error) {
	f, err := fs.Open(filepath.Join(append(path, ignoreFile)...))
	if err != nil {
		// .gitignore does not exist along this path
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	for _, s := range strings.Split(string(data), eol) {
		if !strings.HasPrefix(s, commentPrefix) && len(strings.TrimSpace(s)) > 0 {
			ps = append(ps, gitignore.ParsePattern(s, path))
		}
	}

	return ps, nil
}

func readDir(fs afero.Fs, dir string) ([]os.FileInfo, error) {
	d, err := fs.Open(dir)
	if err != nil {
		return nil, err
	}

	return d.Readdir(0)
}
