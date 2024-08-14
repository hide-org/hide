package model

import (
	"bufio"
	"fmt"
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
	var content strings.Builder

	for _, line := range f.Lines {
		content.WriteString(line.Content)
		content.WriteString("\n")
	}

	return content.String()
}

func (f *File) GetLine(lineNumber int) Line {
	if lineNumber < 1 || lineNumber > len(f.Lines) {
		return Line{}
	}

	return f.Lines[lineNumber-1]
}

func (f *File) GetLineRange(start, end int) []Line {
	if start < 1 {
		start = 1
	}

	if end > len(f.Lines) {
		end = len(f.Lines)
	}

	return f.Lines[start-1 : end]
}

func (f *File) WithLineRange(start, end int) *File {
	return &File{Path: f.Path, Lines: f.GetLineRange(start, end)}
}

func (f *File) ReplaceLineRange(start, end int, content string) (*File, error) {
	replacement, err := NewLines(content)
	if err != nil {
		return f, err
	}

	newLength := len(f.Lines) - (end - start) + len(replacement)
	result := make([]Line, newLength)

	copy(result, f.Lines[:start])
	copy(result[start:], replacement)
	copy(result[start+len(replacement):], f.Lines[end:])

	for i := start - 1; i < len(result); i++ {
		result[i].Number = i + 1
	}

	return &File{Path: f.Path, Lines: result}, nil
}

func NewFile(path string, content string) (*File, error) {
	lines, err := NewLines(content)

	if err != nil {
		return nil, fmt.Errorf("Failed to create lines from content: %w", err)
	}

	return &File{Path: path, Lines: lines}, nil
}

func NewLines(content string) ([]Line, error) {
	var lines []Line

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNumber := 1

	for scanner.Scan() {
		lines = append(lines, Line{Number: lineNumber, Content: scanner.Text()})
		lineNumber++
	}

	return lines, scanner.Err()
}

func EmptyFile(path string) *File {
	return &File{Path: path, Lines: []Line{}}
}
