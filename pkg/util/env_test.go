package util_test

import (
	"os"
	"testing"
	"testing/fstest"

	"github.com/artmoskvin/hide/pkg/util"
)

func TestLoadEnv(t *testing.T) {
	// Create a mock file system
	mockFS := fstest.MapFS{
		".env": &fstest.MapFile{
			Data: []byte("PORT=8080\nDATABASE_URL=postgres://user:pass@localhost/db"),
		},
	}

	// Load the environment variables
	err := util.LoadEnv(mockFS, ".env")
	if err != nil {
		t.Fatalf("loadEnv failed: %v", err)
	}

	// Check if the environment variables were set correctly
	tests := []struct {
		key      string
		expected string
	}{
		{"PORT", "8080"},
		{"DATABASE_URL", "postgres://user:pass@localhost/db"},
	}

	for _, tt := range tests {
		if got := os.Getenv(tt.key); got != tt.expected {
			t.Errorf("os.Getenv(%q) = %q, want %q", tt.key, got, tt.expected)
		}
	}
}
