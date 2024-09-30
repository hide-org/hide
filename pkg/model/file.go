package model

import (
	"strings"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

type Line struct {
	Number  int    `json:"number"`
	Content string `json:"content"`
}

type File struct {
	Path  string `json:"path"`
	Lines []Line `json:"lines"`
	// NOTE: should diagnostics be part of a line?
	Diagnostics []protocol.Diagnostic `json:"diagnostics,omitempty"`
}

func (f *File) Equals(other *File) bool {
	if f == nil && other == nil {
		return true
	}

	if f == nil || other == nil {
		return false
	}

	if f.Path != other.Path || len(f.Lines) != len(other.Lines) {
		return false
	}

	for i := range f.Lines {
		if f.Lines[i] != other.Lines[i] {
			return false
		}
	}

	// NOTE: I'm not sure that diagnostics should be part of comparison, so I'm leaving it out for now

	return true
}

func (f *File) GetContent() string {
	lines := make([]string, len(f.Lines))
	for i, line := range f.Lines {
		lines[i] = line.Content
	}
	return strings.Join(lines, "\n")
}

func (f *File) GetContentBytes() []byte {
	return []byte(f.GetContent())
}

// GetLine returns the line with the given line number. Line numbers are 1-based.
func (f *File) GetLine(lineNumber int) Line {
	if lineNumber < 1 || lineNumber > len(f.Lines) {
		return Line{}
	}

	return f.Lines[lineNumber-1]
}

// GetLineRange returns the lines between start and end (exclusive). Line numbers are 1-based.
func (f *File) GetLineRange(start, end int) []Line {
	// Convert to 0-based indexing
	start -= 1
	end -= 1

	if start < 0 {
		start = 0
	}

	if end > len(f.Lines) {
		end = len(f.Lines)
	}

	return f.Lines[start:end]
}

// WithLineRange returns a new File with the lines between start and end (exclusive). Line numbers are 1-based.
func (f *File) WithLineRange(start, end int) *File {
	return &File{Path: f.Path, Lines: f.GetLineRange(start, end)}
}

func (f *File) WithPath(path string) *File {
	return &File{Path: path, Lines: f.Lines}
}

// ReplaceLineRange replaces the lines between start and end (exclusive) with the given content. Line numbers are 1-based.
func (f *File) ReplaceLineRange(start, end int, content string) (*File, error) {
	if start == end {
		return f, nil
	}

	replacement := NewLines(content)

	newLength := len(f.Lines) - (end - start) + len(replacement)
	result := make([]Line, newLength)

	// Convert to 0-based indexing
	start -= 1
	end -= 1

	copy(result, f.Lines[:start])
	copy(result[start:], replacement)
	if end < len(f.Lines) {
		copy(result[start+len(replacement):], f.Lines[end:])
	}

	for i := start; i < len(result); i++ {
		result[i].Number = i + 1
	}

	return &File{Path: f.Path, Lines: result}, nil
}

// NewFile creates a new File from the given path and content. Content is split into lines. Line numbers are 1-based.
func NewFile(path string, content string) *File {
	return &File{Path: path, Lines: NewLines(content)}
}

// NewLines splits the given content into lines. Line numbers are 1-based.
func NewLines(content string) []Line {
	var lines []Line
	for i, v := range strings.Split(content, "\n") {
		lines = append(lines, Line{Number: i + 1, Content: v})
	}

	return lines
}

func EmptyFile(path string) *File {
	return &File{Path: path, Lines: []Line{}}
}
