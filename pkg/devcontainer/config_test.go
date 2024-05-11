package devcontainer_test

import (
	"encoding/json"
	"github.com/artmoskvin/hide/pkg/devcontainer"
	"testing"
)

func TestMountUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     []byte
		expected devcontainer.Mount
	}{
		{
			name: "string",
			json: []byte(`"type=bind,source=/test_source,target=/test_target"`),
			expected: devcontainer.Mount{
				Type:        "bind",
				Source:      "/test_source",
				Destination: "/test_target",
			},
		},
		{
			name: "object",
			json: []byte(`{"source":"/test_source","target":"/test_target","type":"volume"}`),
			expected: devcontainer.Mount{
				Type:        "volume",
				Source:      "/test_source",
				Destination: "/test_target",
			},
		},
		{
			name: "incomplete object",
			json: []byte(`{"source":"/test_source","target":"/test_target"}`),
			expected: devcontainer.Mount{
				Source:      "/test_source",
				Destination: "/test_target",
			},
		},
		{
			name: "object with wrong type",
			json: []byte(`{"source":"/test_source","target":"/test_target","type":123}`),
			expected: devcontainer.Mount{
				Source:      "/test_source",
				Destination: "/test_target",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var test devcontainer.Mount
			err := json.Unmarshal(tt.json, &test)
			if err != nil {
				t.Fatalf("Failed to unmarshal json: %v", err)
			}

			if test != tt.expected {
				t.Fatalf("Expected %v, got %v", tt.expected, test)
			}
		})
	}
}
