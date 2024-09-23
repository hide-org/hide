package files

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

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

func expandPatterns(patterns []string) []string {
	patternz := []string{}

	for _, pattern := range patterns {
		if !strings.Contains(pattern, "/") {
			// glob to files and dirs
			patternz = append(
				patternz,
				fmt.Sprintf("**/%s", pattern),
				fmt.Sprintf("%s", pattern),
				fmt.Sprintf("**/%s/**", pattern),
				fmt.Sprintf("%s/**", pattern),
			)

			// TODO: can pattern satisfy several conditions?
			continue
		}

		if strings.HasSuffix(pattern, "/") {
			// glob to dirs only
			patternz = append(
				patternz,
				fmt.Sprintf("**/%s**", pattern),
				fmt.Sprintf("%s**", pattern),
				// TODO: does this match on empty dirs?
			)

			continue
		}

		if strings.HasPrefix(pattern, "/") {
			// remove prefix because we match on relative path
			patternz = append(
				patternz,
				pattern[len("/"):],
				// TODO: should we include children dirs?
			)

			continue
		}

		if strings.HasPrefix(pattern, "**/") {
			// glob to files and dirs
			patternz = append(
				patternz,
				pattern,
				fmt.Sprintf("%s/**", pattern),
				fmt.Sprintf("%s", pattern[len("**/"):]),
				fmt.Sprintf("%s/**", pattern[len("**/"):]),
			)

			continue
		}

		patternz = append(patternz, pattern)
	}

	return patternz
}

func NewPatternFilter(include []string, exclude []string) PatternFilter {
	incl := []string{}
	excl := []string{}

	if include != nil {
		incl = expandPatterns(include)
	}

	if exclude != nil {
		excl = expandPatterns(exclude)
	}

	return PatternFilter{Include: incl, Exclude: excl}
}
