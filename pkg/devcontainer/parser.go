package devcontainer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/artmoskvin/hide/pkg/jsonc"
)

func ReadConfig(folder, relativePath string) ([]byte, error) {
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

	return bytes, nil
}

func ParseConfig(content []byte) (*Config, error) {
	devContainer := &Config{}
	if err := json.Unmarshal(jsonc.ToJSON(content), devContainer); err != nil {
		return nil, err
	}

	return devContainer, nil
}

func ParseDockerComposeConfig(content []byte) (*DockerComposeProps, error) {
	config := &DockerComposeProps{}
	if err := json.Unmarshal(jsonc.ToJSON(content), config); err != nil {
		return nil, err
	}

	return config, nil
}
