package gitignore_test

import (
	"testing"

	"github.com/hide-org/hide/pkg/gitignore"
	"github.com/hide-org/hide/pkg/gitignore/mocks"
	"github.com/stretchr/testify/assert"
)

func TestMatcherImpl_Match(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		isDir          bool
		expectedPath   []string
		expectedResult bool
	}{
		{
			name:           "match file",
			path:           "file.txt",
			isDir:          false,
			expectedPath:   []string{"file.txt"},
			expectedResult: true,
		},
		{
			name:           "match directory",
			path:           "dir",
			isDir:          true,
			expectedPath:   []string{"dir"},
			expectedResult: true,
		},
		{
			name:           "no match",
			path:           "file.go",
			isDir:          false,
			expectedPath:   []string{"file.go"},
			expectedResult: false,
		},
		{
			name:           "match with absolute path",
			path:           "/root/file.txt",
			isDir:          false,
			expectedPath:   []string{"root", "file.txt"},
			expectedResult: true,
		},
		{
			name:           "match with relative path",
			path:           "dir/file.txt",
			isDir:          false,
			expectedPath:   []string{"dir", "file.txt"},
			expectedResult: true,
		},
		{
			name: "match with cleaned path",
			path: "dir/../file.txt",
			isDir: false,
			expectedPath: []string{"file.txt"},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMatcher := mocks.NewMockGitignoreMatcher()
			mockMatcher.On("Match", tt.expectedPath, tt.isDir).Return(tt.expectedResult)

			m := gitignore.NewMatcher(mockMatcher)
			got, err := m.Match(tt.path, tt.isDir)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, got)

			mockMatcher.AssertCalled(t, "Match", tt.expectedPath, tt.isDir)
			mockMatcher.AssertNumberOfCalls(t, "Match", 1)
		})
	}
}
