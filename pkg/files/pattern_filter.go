package files

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/gobwas/glob"
)

type PatternFilter struct {
	Include []string
	Exclude []string
}

func (p PatternFilter) keep(path string, info fs.FileInfo) (ok bool, err error) {
	exclude, err := p.shouldExclude(path, info)
	if err != nil {
		return false, err
	}
	if exclude {
		return false, nil
	}

	include, err := p.shouldInclude(path, info)
	if err != nil {
		return false, err
	}
	if !include {
		return false, nil
	}

	return true, nil
}

func (p PatternFilter) shouldInclude(path string, info fs.FileInfo) (ok bool, err error) {
	// always include directories
	if len(p.Include) == 0 || info.IsDir() {
		return true, nil
	}

	for _, pattern := range p.Include {
		g, err := glob.Compile(pattern)
		if err != nil {
			return false, fmt.Errorf("Error include matching pattern %s: %w", pattern, err)
		}
		if g.Match(path) {
			return true, nil
		}
	}

	return false, nil
}

func (p PatternFilter) shouldExclude(path string, info fs.FileInfo) (ok bool, err error) {
	if len(p.Exclude) == 0 {
		return false, nil
	}

	for _, pattern := range p.Exclude {
		g, err := glob.Compile(pattern)
		if err != nil {
			return false, fmt.Errorf("Error exclude matching pattern %s: %w", pattern, err)
		}

		if g.Match(path) {
			if info.IsDir() {
				// exclude whole directory
				return false, filepath.SkipDir
			}
			return true, nil
		}
	}

	return false, nil
}
