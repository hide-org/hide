package devcontainer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"

	"github.com/artmoskvin/hide/pkg/jsonc"
)

func ReadConfig(fileSystem fs.FS) ([]byte, error) {
	content, err := fs.ReadFile(fileSystem, ".devcontainer/devcontainer.json")

	if err == nil {
		return content, nil
	}

	content, err = fs.ReadFile(fileSystem, ".devcontainer.json")
	if err == nil {
		return content, nil
	}

	matches, err := fs.Glob(fileSystem, ".devcontainer/**/devcontainer.json")
	if err != nil {
		return nil, fmt.Errorf("Failed to glob search '.devcontainer/**/devcontainer.json': %w", err)
	}

	if len(matches) == 0 {
		return nil, errors.New("devcontainer.json not found")
	}

	if len(matches) > 1 {
		return nil, errors.New("multiple devcontainer.json found")
	}

	content, err = fs.ReadFile(fileSystem, matches[0])
	if err != nil {
		return nil, fmt.Errorf("Failed to read devcontainer.json: %w", err)
	}

	return content, nil
}

func ParseConfig(content []byte) (*Config, error) {
	devContainer := &Config{}
	if err := json.Unmarshal(jsonc.ToJSON(content), devContainer); err != nil {
		return nil, err
	}

	return devContainer, nil
}
