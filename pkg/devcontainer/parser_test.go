package devcontainer_test

import (
	"github.com/artmoskvin/hide/pkg/devcontainer"
	"testing"
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
					DockerComposeFile: "docker-compose.yml",
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
					DockerComposeFile: "docker-compose.yml",
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

func TestDockerComposePropsEquals(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected *devcontainer.DockerComposeProps
	}{
		{
			name: "full",
			content: []byte(`{
	"dockerComposeFile": "docker-compose.yml", 
	"service": "app", 
	"runServices": ["app", "db"], 
	"workspaceFolder": "/workspace"
}`),
			expected: &devcontainer.DockerComposeProps{
				DockerComposeFile: "docker-compose.yml",
				Service:           "app",
				RunServices:       []string{"app", "db"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := devcontainer.ParseDockerComposeConfig(tt.content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !actual.Equals(tt.expected) {
				t.Errorf("expected: %+v, actual: %+v", tt.expected, actual)
			}
		})
	}
}
