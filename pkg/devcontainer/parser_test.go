package devcontainer_test

import (
	"encoding/json"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/hide-org/hide/pkg/devcontainer"
	"github.com/hide-org/hide/pkg/jsonc"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name     string
		content  devcontainer.File
		expected *devcontainer.Config
	}{
		{
			name: "empty",
			content: devcontainer.File{Path: "config.json", Content: []byte(`{
		}`)},
			expected: &devcontainer.Config{},
		},
		{
			name: "full",
			content: devcontainer.File{Path: "config.json", Content: []byte(`{
	"dockerComposeFile": "docker-compose.yml",
	"service": "app",
	"runServices": ["app", "db"],
	"workspaceFolder": "/workspace"
}`)},
			expected: &devcontainer.Config{
				DockerComposeProps: devcontainer.DockerComposeProps{
					DockerComposeFile: []string{"docker-compose.yml"},
					Service:           "app",
					RunServices:       []string{"app", "db"},
				},
				DockerImageProps: devcontainer.DockerImageProps{
					WorkspaceFolder: "/workspace",
				},
			},
		},
		{
			name: "full with comments",
			content: devcontainer.File{Path: "config.json", Content: []byte(`{
	// Required when using Docker Compose.
	"dockerComposeFile": "docker-compose.yml",
	// Required when using Docker Compose.
	"service": "app",
	"runServices": ["app", "db"],
	"workspaceFolder": "/workspace"
}`)},
			expected: &devcontainer.Config{
				DockerComposeProps: devcontainer.DockerComposeProps{
					DockerComposeFile: []string{"docker-compose.yml"},
					Service:           "app",
					RunServices:       []string{"app", "db"},
				},
				DockerImageProps: devcontainer.DockerImageProps{
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
		content  devcontainer.File
		expected *devcontainer.DockerImageProps
	}{
		{
			name: "image",
			content: devcontainer.File{Path: "config.json", Content: []byte(`{
	"image": "node:14",
}`)},
			expected: &devcontainer.DockerImageProps{
				Image: "node:14",
			},
		},
		{
			name: "dockerfile",
			content: devcontainer.File{Path: "config.json", Content: []byte(`{
	"dockerfile": "Dockerfile",
	"context": ".",
}`)},
			expected: &devcontainer.DockerImageProps{
				Dockerfile: "Dockerfile",
				Context:    ".",
			},
		},
		{
			name: "build",
			content: devcontainer.File{Path: "config.json", Content: []byte(`{
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
}`)},
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
			},
		},
		{
			name: "appPort string",
			content: devcontainer.File{Path: "config.json", Content: []byte(`{
	"appPort": "3000",
}`)},
			expected: &devcontainer.DockerImageProps{
				AppPort: []int{3000},
			},
		},
		{
			name: "appPort int",
			content: devcontainer.File{Path: "config.json", Content: []byte(`{
	"appPort": 3000,
}`)},
			expected: &devcontainer.DockerImageProps{
				AppPort: []int{3000},
			},
		},
		{
			name: "appPort array",
			content: devcontainer.File{Path: "config.json", Content: []byte(`{
	"appPort": [3000, 3001],
}`)},
			expected: &devcontainer.DockerImageProps{
				AppPort: []int{3000, 3001},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseDockerImageProps(tt.content.Content)
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
		content  devcontainer.File
		expected *devcontainer.Config
	}{
		{
			name: "ignored",
			content: devcontainer.File{Path: "config.json", Content: []byte(`{
	"customizations": {
		// Configure properties specific to VS Code.
		"vscode": {
			// Set *default* container specific settings.json values on container create.
			"settings": {},
			"extensions": ["streetsidesoftware.code-spell-checker"],
		}
	}
}`)},
			expected: &devcontainer.Config{
				GeneralProperties: devcontainer.GeneralProperties{
					Customizations: devcontainer.Customizations{},
				},
			},
		},
		{
			name: "hide with tasks",
			content: devcontainer.File{Path: "config.json", Content: []byte(`{
	"customizations": {
		"hide": {
			"tasks": [
				{
					"alias": "test-task",
					"command": "echo test"
				}
			]
		}
	}
}`)},
			expected: &devcontainer.Config{
				GeneralProperties: devcontainer.GeneralProperties{
					Customizations: devcontainer.Customizations{
						Hide: &devcontainer.HideCustomization{
							Tasks: []devcontainer.Task{
								{
									Alias:   "test-task",
									Command: "echo test",
								},
							},
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
		expected   devcontainer.File
	}{
		{
			name: ".devcontainer/devcontainer.json",
			fileSystem: fstest.MapFS{
				".devcontainer/devcontainer.json": {Data: []byte(`{
	"key": "value"
}`)},
			},
			expected: devcontainer.File{Path: ".devcontainer/devcontainer.json", Content: []byte(`{
	"key": "value"
}`)},
		},
		{
			name: ".devcontainer.json",
			fileSystem: fstest.MapFS{
				".devcontainer.json": {Data: []byte(`{
	"key": "value"
}`)},
			},
			expected: devcontainer.File{Path: ".devcontainer.json", Content: []byte(`{
	"key": "value"
}`)},
		},
		{
			name: ".devcontainer/test-folder/devcontainer.json",
			fileSystem: fstest.MapFS{
				".devcontainer/test-folder/devcontainer.json": {Data: []byte(`{
	"key": "value"
}`)},
			},
			expected: devcontainer.File{Path: ".devcontainer/test-folder/devcontainer.json", Content: []byte(`{
	"key": "value"
}`)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := devcontainer.FindConfig(tt.fileSystem)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !actual.Equals(&tt.expected) {
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
