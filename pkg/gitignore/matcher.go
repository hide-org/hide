package gitignore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/plumbing/format/gitignore"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
)

type Matcher interface {
	Match(path string, isDir bool) (bool, error)
}

func NewMatcher(matcher gitignore.Matcher) Matcher {
	return MatcherImpl{matcher: matcher}
}

type MatcherImpl struct {
	matcher gitignore.Matcher
}

func (m MatcherImpl) Match(path string, isDir bool) (bool, error) {
	path = filepath.Clean(path)

	if strings.HasPrefix(path, "/") {
		relPath, err := filepath.Rel("/", path)
		if err != nil {
			return false, fmt.Errorf("failed to get relative path of %s: %w", path, err)
		}

		path = relPath
	}

	return m.matcher.Match(strings.Split(path, string(os.PathSeparator)), isDir), nil
}

type MatcherFactory interface {
	NewMatcher(fs afero.Fs) (Matcher, error)
}

type MatcherFactoryImpl struct{}

func NewMatcherFactory() MatcherFactory {
	return &MatcherFactoryImpl{}
}

func (mf MatcherFactoryImpl) NewMatcher(fs afero.Fs) (Matcher, error) {
	ps, err := ReadPatterns(fs, nil)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("Created gitignore matcher with %d patterns", len(ps))
	m := NewMatcher(gitignore.NewMatcher(ps))
	return m, nil
}
