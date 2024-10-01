package gitignore

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/plumbing/format/gitignore"
	"github.com/spf13/afero"
)

type MatcherSuite struct {
	afero.Fs // git repository root
}

func (s *MatcherSuite) SetUpTest(t *testing.T) {
	// setup generic git repository root
	fs := afero.NewMemMapFs()

	// gitignore from .git
	mkdirAll(t, fs, ".git/info", os.ModePerm)
	createFile(t, fs, ".git/info/exclude", "exclude.crlf\n")

	// gitignore from root file
	createFile(t, fs, ".gitignore", "vendor/g*/\nignore.crlf\nignore_dir\n**/*.txt")

	// gitignore from vendor folder
	mkdirAll(t, fs, "vendor", os.ModePerm)
	createFile(t, fs, "vendor/.gitignore", "!github.com/\n")

	// gitignore from ignore_dir
	mkdirAll(t, fs, "ignore_dir", os.ModePerm)
	createFile(t, fs, "ignore_dir/.gitignore", "!file\n")
	createFile(t, fs, "ignore_dir/file", "")
	createFile(t, fs, "ignore_dir/otherfile", "")

	// other files
	mkdirAll(t, fs, "another", os.ModePerm)
	mkdirAll(t, fs, "exclude.crlf", os.ModePerm)
	mkdirAll(t, fs, "ignore.crlf", os.ModePerm)
	mkdirAll(t, fs, "vendor/github.com", os.ModePerm)
	mkdirAll(t, fs, "vendor/gopkg.in", os.ModePerm)

	// gitignore in sub-dirs with other files
	mkdirAll(t, fs, "multiple/sub/ignores/first", os.ModePerm)
	createFile(t, fs, "multiple/sub/ignores/first/.gitignore", "ignore_dir\n")
	mkdirAll(t, fs, "multiple/sub/ignores/first/ignore_dir", os.ModePerm)
	createFile(t, fs, "multiple/sub/ignores/first/ignore_dir/file", "")

	mkdirAll(t, fs, "multiple/sub/ignores/second", os.ModePerm)
	createFile(t, fs, "multiple/sub/ignores/second/.gitignore", "ignore_dir\n")
	mkdirAll(t, fs, "multiple/sub/ignores/second/ignore_dir", os.ModePerm)
	createFile(t, fs, "multiple/sub/ignores/second/ignore_dir/file", "")

	mkdirAll(t, fs, "globs", os.ModePerm)
	createFile(t, fs, "globs/something.txt", "")

	s.Fs = fs
}

func TestReadPatterns(t *testing.T) {
	suite := &MatcherSuite{}
	suite.SetUpTest(t)

	ps, err := ReadPatterns(suite.Fs, make([]string, 0, 5))
	if err != nil {
		t.Fatal(err)
	}

	wantN := 8
	if n := len(ps); n != wantN {
		t.Fatalf("wrong pattern length: got %d, want %d", n, wantN)
	}

	m := gitignore.NewMatcher(ps)

	for _, test := range []struct {
		path      []string
		isDir     bool
		wantMatch bool
	}{
		{
			path:      []string{"exclude.crlf"},
			isDir:     true,
			wantMatch: true,
		},
		{
			path:      []string{"ignore.crlf"},
			isDir:     true,
			wantMatch: true,
		},
		{
			path:      []string{"vendor", "gopkg.in"},
			isDir:     true,
			wantMatch: true,
		},
		{
			path:      []string{"ignore_dir"},
			isDir:     true,
			wantMatch: true,
		},
		{
			path:      []string{"ignore_dir", "file"},
			isDir:     false,
			wantMatch: true,
		},
		{
			path:      []string{"ignore_dir", "otherfile"},
			isDir:     false,
			wantMatch: true,
		},
		{
			path:      []string{"vendor", "github.com"},
			isDir:     true,
			wantMatch: false,
		},
		{
			path:      []string{"multiple", "sub", "ignores", "first", "ignore_dir"},
			isDir:     true,
			wantMatch: true,
		},
		{
			path:      []string{"multiple", "sub", "ignores", "first", "ignore_dir", "file"},
			isDir:     false,
			wantMatch: true,
		},
		{
			path:      []string{"multiple", "sub", "ignores", "second", "ignore_dir"},
			isDir:     true,
			wantMatch: true,
		},
		{
			path:      []string{"multiple", "sub", "ignores", "second", "ignore_dir", "file"},
			isDir:     false,
			wantMatch: true,
		},
		{
			path:      []string{"globs", "something.txt"},
			isDir:     false,
			wantMatch: true,
		},
	} {
		t.Run(filepath.Join(test.path...), func(t *testing.T) {
			if gotMatch := m.Match(test.path, test.isDir); gotMatch != test.wantMatch {
				t.Fatalf("want match %v, got %v", test.wantMatch, gotMatch)
			}
		})
	}
}

func mkdirAll(t *testing.T, fs afero.Fs, path string, parm fs.FileMode) {
	err := fs.MkdirAll(path, parm)
	if err != nil {
		t.Fatalf("failed to mkdir: %s", err)
	}
}

func createFile(t *testing.T, fs afero.Fs, path, content string) {
	f, err := fs.Create(path)
	if err != nil {
		t.Fatalf("failed to create path %s: %s", path, err)
	}
	defer f.Close()

	if _, err := f.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write content in path %s: err %s", path, err)
	}
}
