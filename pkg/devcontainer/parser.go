package devcontainer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/hide-org/hide/pkg/jsonc"
)

type File struct {
	Path    string
	Content []byte
}

func (f *File) Equals(other *File) bool {
	if f == nil && other == nil {
		return true
	}

	if f == nil || other == nil {
		return false
	}

	return f.Path == other.Path && bytes.Equal(f.Content, other.Content)
}

func FindConfig(fileSystem fs.FS) (File, error) {
	content, err := fs.ReadFile(fileSystem, ".devcontainer/devcontainer.json")

	if err == nil {
		return File{Path: ".devcontainer/devcontainer.json", Content: content}, nil
	}

	content, err = fs.ReadFile(fileSystem, ".devcontainer.json")
	if err == nil {
		return File{Path: ".devcontainer.json", Content: content}, nil
	}

	matches, err := fs.Glob(fileSystem, ".devcontainer/**/devcontainer.json")
	if err != nil {
		return File{}, fmt.Errorf("Failed to glob search '.devcontainer/**/devcontainer.json': %w", err)
	}

	if len(matches) == 0 {
		return File{}, errors.New("devcontainer.json not found")
	}

	if len(matches) > 1 {
		return File{}, errors.New("multiple devcontainer.json found")
	}

	content, err = fs.ReadFile(fileSystem, matches[0])
	if err != nil {
		return File{}, fmt.Errorf("Failed to read devcontainer.json: %w", err)
	}

	return File{Path: matches[0], Content: content}, nil
}

func ParseConfig(configFile File) (*Config, error) {
	config := &Config{}
	if err := json.Unmarshal(jsonc.ToJSON(configFile.Content), config); err != nil {
		return nil, err
	}

	config.Path = filepath.Dir(configFile.Path)

	return config, nil
}
