package devcontainer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/artmoskvin/hide/pkg/jsonc"
)

func ParseDevContainerJSON(folder, relativePath string) (*DevContainerConfig, error) {
	path := ""
	if relativePath != "" {
		path = filepath.Join(filepath.ToSlash(folder), relativePath)
		_, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("devcontainer path %s doesn't exist: %w", path, err)
		}
	} else {
		path = filepath.Join(folder, ".devcontainer", "devcontainer.json")
		_, err := os.Stat(path)
		if err != nil {
			path = filepath.Join(folder, ".devcontainer.json")
			_, err = os.Stat(path)
			if err != nil {
				matches, err := filepath.Glob(filepath.ToSlash(filepath.Clean(folder)) + "/.devcontainer/**/devcontainer.json")
				if err != nil {
					return nil, err
				} else if len(matches) == 0 {
					return nil, nil
				}
			}
		}
	}

	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("make path absolute: %w", err)
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	devContainer := &DevContainerConfig{}
	err = json.Unmarshal(jsonc.ToJSON(bytes), devContainer)
	if err != nil {
		return nil, err
	}

	devContainer.Origin = path
	return devContainer, nil
}
