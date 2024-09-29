package gitignore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/plumbing/format/gitignore"
	"github.com/spf13/afero"
)

type MatcherSuite struct {
	GFS afero.Fs // git repository root
}

func (s *MatcherSuite) SetUpTest(t *testing.T) {
	// setup generic git repository root
	fs := afero.NewMemMapFs()

	// gitignore from .git
	err := fs.MkdirAll(".git/info", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	f, err := fs.Create(".git/info/exclude")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("exclude.crlf\n"))
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	// gitignore from root file
	f, err = fs.Create(".gitignore")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("vendor/g*/\n"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("ignore.crlf\n"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("ignore_dir\n"))
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	// gitignore from vendor folder
	err = fs.MkdirAll("vendor", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	f, err = fs.Create("vendor/.gitignore")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("!github.com/\n"))
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	// gitignore from ignore_dir
	err = fs.MkdirAll("ignore_dir", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	f, err = fs.Create("ignore_dir/.gitignore")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("!file\n"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = fs.Create("ignore_dir/file")
	if err != nil {
		t.Fatal(err)
	}
	_, err = fs.Create("ignore_dir/otherfile")
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	// other files
	err = fs.MkdirAll("another", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	err = fs.MkdirAll("exclude.crlf", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	err = fs.MkdirAll("ignore.crlf", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	err = fs.MkdirAll("vendor/github.com", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	err = fs.MkdirAll("vendor/gopkg.in", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	// gitignore in sub-dirs with other files
	err = fs.MkdirAll("multiple/sub/ignores/first", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	err = fs.MkdirAll("multiple/sub/ignores/second", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	f, err = fs.Create("multiple/sub/ignores/first/.gitignore")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("ignore_dir\n"))
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	f, err = fs.Create("multiple/sub/ignores/second/.gitignore")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("ignore_dir\n"))
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = fs.MkdirAll("multiple/sub/ignores/first/ignore_dir", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	err = fs.MkdirAll("multiple/sub/ignores/second/ignore_dir", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	s.GFS = fs
}

func TestReadPatterns(t *testing.T) {
	suite := &MatcherSuite{}
	suite.SetUpTest(t)

	checkPatterns := func(ps []gitignore.Pattern) {
		if n := len(ps); n != 8 {
			t.Fatalf("wrong pattern length: got %d, want %d", n, 7)
		}

		m := gitignore.NewMatcher(ps)

		for _, v := range []struct {
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
				wantMatch: false,
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
				path:      []string{"multiple", "sub", "ignores", "second", "ignore_dir"},
				isDir:     true,
				wantMatch: true,
			},
		} {
			if gotMatch := m.Match(v.path, v.isDir); gotMatch != v.wantMatch {
				t.Fatalf("failed on path %s: want match %v, got %v", filepath.Join(v.path...), v.wantMatch, gotMatch)
			}
		}
	}

	ps, err := ReadPatterns(suite.GFS, nil)
	if err != nil {
		t.Fatal(err)
	}
	checkPatterns(ps)

	// passing an empty slice with capacity to check we don't hit a bug where the extra capacity is reused incorrectly
	ps, err = ReadPatterns(suite.GFS, make([]string, 0, 6))
	if err != nil {
		t.Fatal(err)
	}
	checkPatterns(ps)
}
