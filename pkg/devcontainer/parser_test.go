package devcontainer_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/jsonc"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected *devcontainer.Config
	}{
		{
			name: "empty",
			content: []byte(`{
		}`),
			expected: &devcontainer.Config{},
		},
		{
			name: "full",
			content: []byte(`{
			"dockerComposeFile": "docker-compose.yml",
			"service": "app",
			"runServices": ["app", "db"],
			"workspaceFolder": "/workspace"
		}`),
			expected: &devcontainer.Config{
				DockerComposeProps: devcontainer.DockerComposeProps{
					DockerComposeFile: []string{"docker-compose.yml"},
					Service:           "app",
					RunServices:       []string{"app", "db"},
				},
				GeneralProperties: devcontainer.GeneralProperties{
					WorkspaceFolder: "/workspace",
				},
			},
		},
		{
			name: "full with comments",
			content: []byte(`{
			// Required when using Docker Compose.
			"dockerComposeFile": "docker-compose.yml",
			// Required when using Docker Compose.
			"service": "app",
			"runServices": ["app", "db"],
			"workspaceFolder": "/workspace"
		}`),
			expected: &devcontainer.Config{
				DockerComposeProps: devcontainer.DockerComposeProps{
					DockerComposeFile: []string{"docker-compose.yml"},
					Service:           "app",
					RunServices:       []string{"app", "db"},
				},
				GeneralProperties: devcontainer.GeneralProperties{
					WorkspaceFolder: "/workspace",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := devcontainer.ParseConfig(tt.content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !actual.Equals(tt.expected) {
				t.Errorf("expected: %+v, actual: %+v", tt.expected, actual)
			}
		})
	}
}

func TestDockerImagePropsUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected *devcontainer.DockerImageProps
	}{
		{
			name: "image",
			content: []byte(`{
	"image": "node:14",
	"workspaceMount": "/workspace"
}`),
			expected: &devcontainer.DockerImageProps{
				Image:          "node:14",
				WorkspaceMount: "/workspace",
			},
		},
		{
			name: "dockerfile",
			content: []byte(`{
	"dockerfile": "Dockerfile",
	"context": ".",
	"workspaceMount": "/workspace"
}`),
			expected: &devcontainer.DockerImageProps{
				Dockerfile:     "Dockerfile",
				Context:        ".",
				WorkspaceMount: "/workspace",
			},
		},
		{
			name: "build",
			content: []byte(`{
	"build": {
		"dockerfile": "Dockerfile",
		"context": ".",
		"args": {
			"NODE_ENV": "development"
		},
		"options": ["--no-cache"],
		"target": "development",
		"cacheFrom": ["node:14"]
	},
	"workspaceMount": "/workspace"
}`),
			expected: &devcontainer.DockerImageProps{
				Build: &devcontainer.BuildProps{
					Dockerfile: "Dockerfile",
					Context:    ".",
					Args: args(map[string]string{
						"NODE_ENV": "development",
					}),
					Options:   []string{"--no-cache"},
					Target:    "development",
					CacheFrom: []string{"node:14"},
				},
				WorkspaceMount: "/workspace",
			},
		},
		{
			name: "appPort string",
			content: []byte(`{
	"appPort": "3000",
	"workspaceMount": "/workspace"
}`),
			expected: &devcontainer.DockerImageProps{
				AppPort:        []int{3000},
				WorkspaceMount: "/workspace",
			},
		},
		{
			name: "appPort int",
			content: []byte(`{
	"appPort": 3000,
	"workspaceMount": "/workspace"
}`),
			expected: &devcontainer.DockerImageProps{
				AppPort:        []int{3000},
				WorkspaceMount: "/workspace",
			},
		},
		{
			name: "appPort array",
			content: []byte(`{
	"appPort": [3000, 3001],
	"workspaceMount": "/workspace"
}`),
			expected: &devcontainer.DockerImageProps{
				AppPort:        []int{3000, 3001},
				WorkspaceMount: "/workspace",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseDockerImageProps(tt.content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !actual.Equals(tt.expected) {
				t.Errorf("expected: %+v, actual: %+v", tt.expected, actual)
			}
		})
	}
}

func parseDockerImageProps(content []byte) (*devcontainer.DockerImageProps, error) {
	config := &devcontainer.DockerImageProps{}
	if err := json.Unmarshal(jsonc.ToJSON(content), config); err != nil {
		return nil, err
	}

	return config, nil
}

func TestConfigWithCustomizations(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected *devcontainer.Config
	}{
		{
			name: "vscode",
			content: []byte(`{
	"customizations": {
		// Configure properties specific to VS Code.
		"vscode": {
			// Set *default* container specific settings.json values on container create.
			"settings": {},
			"extensions": ["streetsidesoftware.code-spell-checker"],
		}
	}
}`),
			expected: &devcontainer.Config{
				GeneralProperties: devcontainer.GeneralProperties{
					Customizations: map[string]map[string]any{
						"vscode": {
							"settings": map[string]any{},
							// interface{} because unmarshaller doesn't know the type of the value
							"extensions": []interface{}{"streetsidesoftware.code-spell-checker"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := devcontainer.ParseConfig(tt.content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !actual.Equals(tt.expected) {
				t.Errorf("expected: %+v, actual: %+v", tt.expected, actual)
			}
		})
	}
}

func TestReadConfig(t *testing.T) {
	tests := []struct {
		name       string
		fileSystem fs.FS
		expected   []byte
	}{
		{
			name: ".devcontainer/devcontainer.json",
			fileSystem: fstest.MapFS{
				".devcontainer/devcontainer.json": {Data: []byte(`{
	"key": "value"
}`)},
			},
			expected: []byte(`{
	"key": "value"
}`),
		},
		{
			name: ".devcontainer.json",
			fileSystem: fstest.MapFS{
				".devcontainer.json": {Data: []byte(`{
	"key": "value"
}`)},
			},
			expected: []byte(`{
	"key": "value"
}`),
		},
		{
			name: ".devcontainer/test-folder/devcontainer.json",
			fileSystem: fstest.MapFS{
				".devcontainer/test-folder/devcontainer.json": {Data: []byte(`{
	"key": "value"
}`)},
			},
			expected: []byte(`{
	"key": "value"
}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := devcontainer.FindConfig(tt.fileSystem)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !bytes.Equal(actual, tt.expected) {
				t.Errorf("expected: %+v, actual: %+v", tt.expected, actual)
			}
		})
	}
}

func TestReadConfigFails(t *testing.T) {
	tests := []struct {
		name       string
		fileSystem fs.FS
	}{
		{
			name: "no devcontainer.json",
			fileSystem: fstest.MapFS{
				"test-file": {Data: []byte("test")},
			},
		},
		{
			name: "more than one devcontainer.json",
			fileSystem: fstest.MapFS{
				".devcontainer/test-folder-1/devcontainer.json": {Data: []byte(`{
	"key": "value"
}`)},
				".devcontainer/test-folder-2/devcontainer.json": {Data: []byte(`{
	"key": "value"
}`)},
			},
		},
		{
			name: "too deep subfolder structure",
			fileSystem: fstest.MapFS{
				".devcontainer/test-folder/test-subfolder/devcontainer.json": {Data: []byte(`{
	"key": "value"
}`)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := devcontainer.FindConfig(tt.fileSystem)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func args(args map[string]string) map[string]*string {
	result := make(map[string]*string)

	for key, value := range args {
		result[key] = &value
	}

	return result
}
