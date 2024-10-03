package devcontainer_test

import (
	"encoding/json"
	"maps"
	"slices"
	"testing"

	"github.com/hide-org/hide/pkg/devcontainer"
)

type stringArrayTestStruct struct {
	TestField devcontainer.StringArray `json:"test_field"`
}

func (tj *stringArrayTestStruct) Equals(other stringArrayTestStruct) bool {
	return slices.Equal(tj.TestField, other.TestField)
}

func TestStringArray_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     []byte
		expected stringArrayTestStruct
	}{
		{
			name: "single string",
			json: []byte(`{"test_field":"test"}`),
			expected: stringArrayTestStruct{
				TestField: devcontainer.StringArray{"test"},
			},
		},
		{
			name: "array of strings",
			json: []byte(`{"test_field":["test1", "test2"]}`),
			expected: stringArrayTestStruct{
				TestField: devcontainer.StringArray{"test1", "test2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var test stringArrayTestStruct
			err := json.Unmarshal(tt.json, &test)
			if err != nil {
				t.Fatalf("Failed to unmarshal json: %v", err)
			}

			if !test.Equals(tt.expected) {
				t.Fatalf("Expected %d elements, got %d", len(tt.expected.TestField), len(test.TestField))
			}
		})
	}
}

func TestStringArray_UnmarshalJSON_Fails(t *testing.T) {
	tests := []struct {
		name string
		json []byte
	}{
		{
			name: "invalid json",
			json: []byte(`{test_field":"test"}`),
		},
		{
			name: "unsupported type",
			json: []byte(`{"test_field":1}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var test stringArrayTestStruct
			err := json.Unmarshal(tt.json, &test)
			if err == nil {
				t.Fatalf("Expected error, got nil")
			}
		})
	}
}

type lifecycleCommandTestStruct struct {
	TestField devcontainer.LifecycleCommand `json:"test_field"`
}

func (tj *lifecycleCommandTestStruct) Equals(other lifecycleCommandTestStruct) bool {
	return maps.EqualFunc(tj.TestField, other.TestField, slices.Equal)
}

func TestLifecycleCommand_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     []byte
		expected lifecycleCommandTestStruct
	}{
		{
			name: "single string",
			json: []byte(`{"test_field":"test"}`),
			expected: lifecycleCommandTestStruct{
				TestField: devcontainer.LifecycleCommand{
					"": {devcontainer.DefaultShell, "-c", "test"},
				},
			},
		},
		{
			name: "array of strings",
			json: []byte(`{"test_field":["test1", "test2"]}`),
			expected: lifecycleCommandTestStruct{
				TestField: devcontainer.LifecycleCommand{
					"": {"test1", "test2"},
				},
			},
		},
		{
			name: "map of strings",
			json: []byte(`{"test_field":{"key1":"test1", "key2":"test2"}}`),
			expected: lifecycleCommandTestStruct{
				TestField: devcontainer.LifecycleCommand{
					"key1": {devcontainer.DefaultShell, "-c", "test1"},
					"key2": {devcontainer.DefaultShell, "-c", "test2"},
				},
			},
		},
		{
			name: "map of arrays",
			json: []byte(`{"test_field":{"key1":["test1", "test2"], "key2":["test3", "test4"]}}`),
			expected: lifecycleCommandTestStruct{
				TestField: devcontainer.LifecycleCommand{
					"key1": {"test1", "test2"},
					"key2": {"test3", "test4"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var test lifecycleCommandTestStruct
			err := json.Unmarshal(tt.json, &test)
			if err != nil {
				t.Fatalf("Failed to unmarshal json: %v", err)
			}

			if !test.Equals(tt.expected) {
				t.Fatalf("Expected %d elements, got %d", len(tt.expected.TestField), len(test.TestField))
			}
		})
	}
}

func TestLifecycleCommand_UnmarshalJSON_Fails(t *testing.T) {
	tests := []struct {
		name string
		json []byte
	}{
		{
			name: "invalid json",
			json: []byte(`{test_field":"test"}`),
		},
		{
			name: "unsupported type",
			json: []byte(`{"test_field":1}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var test lifecycleCommandTestStruct
			err := json.Unmarshal(tt.json, &test)
			if err == nil {
				t.Fatalf("Expected error, got nil")
			}
		})
	}
}
