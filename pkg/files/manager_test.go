package files_test

import (
	"context"
	"strings"
	"testing"

	"github.com/artmoskvin/hide/pkg/files"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/spf13/afero"
)

func TestReadFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "test.txt"
	content := "line1\nline2\nline3\n"
	afero.WriteFile(fs, path, []byte(content), 0644)

	fm := files.NewFileManager()
	actual, err := fm.ReadFile(context.Background(), fs, path)
	expected, _ := model.NewFile(path, content)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !actual.Equals(expected) {
		t.Errorf("Expected %+v, got %+v", expected, actual)
	}
}

func TestReadNonExistentFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "test.txt", []byte("line1\nline2\nline3\n"), 0644)

	fm := files.NewFileManager()
	_, err := fm.ReadFile(context.Background(), fs, "non-existent.txt")
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "file non-existent.txt not found") {
		t.Errorf("Expected error to contain 'file does not exist', got %s", err.Error())
	}
}

func TestFileManagerImpl_ApplyPatch_Success(t *testing.T) {
	tests := []struct {
		name     string
		patch    string
		expected model.File
	}{
		{
			name: "Apply patch to file",
			patch: `--- test.txt
+++ test.txt
@@ -1,10 +1,9 @@
 line1
-line2
+line20
 line3
-line4
+line40
 line5
-line6
 line7
 line8
-line9
 line10
+line11`,
			expected: model.File{
				Path: "test.txt",
				Lines: []model.Line{
					{Number: 1, Content: "line1"},
					{Number: 2, Content: "line20"},
					{Number: 3, Content: "line3"},
					{Number: 4, Content: "line40"},
					{Number: 5, Content: "line5"},
					{Number: 6, Content: "line7"},
					{Number: 7, Content: "line8"},
					{Number: 8, Content: "line10"},
					{Number: 9, Content: "line11"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filesystem := afero.NewMemMapFs()
			afero.WriteFile(filesystem, "test.txt", []byte("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n"), 0644)
			fm := files.NewFileManager()
			actual, err := fm.ApplyPatch(context.Background(), filesystem, "test.txt", tt.patch)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !actual.Equals(&tt.expected) {
				t.Errorf("Expected %+v, got %+v", tt.expected, actual)
			}
		})
	}
}

func TestFileManagerImpl_ApplyPatch_Failure(t *testing.T) {
	tests := []struct {
		name          string
		file          string
		patch         string
		expectedError string
	}{
		{
			name:          "File not found",
			file:          "not-found.txt",
			patch:         "",
			expectedError: "file not-found.txt not found",
		},
		{
			name: "Patch with multiple files",
			patch: `--- file1
+++ file1
@@ -1,3 +1,3 @@
 line1
-line2
+line20
 line3
--- file2
+++ file2
@@ -1,3 +1,3 @@
 line1
-line2
+line20
 line3
`,
			expectedError: "multiple files",
		},
		{
			name:          "Patch with no files",
			patch:         "",
			expectedError: "no files changed in patch",
		},
		{
			name: "Patch cannot be applied (no newline at end of file)",
			patch: `--- test.txt
+++ test.txt
@@ -1,3 +1,3 @@
 line1
-line2
+line20
 line3`,
			expectedError: "failed to apply patch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileSystem := afero.NewMemMapFs()
			afero.WriteFile(fileSystem, "test.txt", []byte("line1\nline2\nline3\n"), 0644)
			fm := files.NewFileManager()
			_, err := fm.ApplyPatch(context.Background(), fileSystem, tt.file, tt.patch)
			if err == nil {
				t.Fatalf("Expected error, got nil")
			}

			if !strings.Contains(strings.ToLower(err.Error()), tt.expectedError) {
				t.Errorf("Expected error to contain '%s', got %s", tt.expectedError, err.Error())
			}
		})
	}
}

func TestFileManagerImpl_UpdateLines_Success(t *testing.T) {
	tests := []struct {
		name     string
		lineDiff files.LineDiffChunk
		expected model.File
	}{
		{
			name: "Update 1 line",
			lineDiff: files.LineDiffChunk{
				StartLine: 1,
				EndLine:   2,
				Content:   "line11",
			},
			expected: model.File{
				Path: "test.txt",
				Lines: []model.Line{
					{Number: 1, Content: "line11"},
					{Number: 2, Content: "line2"},
					{Number: 3, Content: "line3"},
				},
			},
		},
		{
			name: "Update multiple lines",
			lineDiff: files.LineDiffChunk{
				StartLine: 1,
				EndLine:   3,
				Content:   "line11\nline12\n",
			},
			expected: model.File{
				Path: "test.txt",
				Lines: []model.Line{
					{Number: 1, Content: "line11"},
					{Number: 2, Content: "line12"},
					{Number: 3, Content: "line3"},
				},
			},
		},
		{
			name: "Add multiple lines at the end",
			lineDiff: files.LineDiffChunk{
				StartLine: 3,
				EndLine:   4,
				Content:   "line10\nline11\nline12\n",
			},
			expected: model.File{
				Path: "test.txt",
				Lines: []model.Line{
					{Number: 1, Content: "line1"},
					{Number: 2, Content: "line2"},
					{Number: 3, Content: "line10"},
					{Number: 4, Content: "line11"},
					{Number: 5, Content: "line12"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := files.NewFileManager()
			filesystem := afero.NewMemMapFs()
			afero.WriteFile(filesystem, "test.txt", []byte("line1\nline2\nline3\n"), 0644)
			actual, err := fm.UpdateLines(context.Background(), filesystem, "test.txt", tt.lineDiff)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !actual.Equals(&tt.expected) {
				t.Errorf("Expected %+v, got %+v", tt.expected, actual)
			}
		})
	}
}

func TestFileManagerImpl_UpdateLines_Failure(t *testing.T) {
	tests := []struct {
		name     string
		lineDiff files.LineDiffChunk
		expected string
	}{
		{
			name: "Start line > number of lines",
			lineDiff: files.LineDiffChunk{
				StartLine: 11,
				EndLine:   10,
				Content:   "line11",
			},
			expected: "Start line must be less than or equal to 3",
		},
		{
			name: "End line > number of lines + 1",
			lineDiff: files.LineDiffChunk{
				StartLine: 1,
				EndLine:   11,
				Content:   "line11",
			},
			expected: "End line must be less than or equal to 4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filesystem := afero.NewMemMapFs()
			afero.WriteFile(filesystem, "test.txt", []byte("line1\nline2\nline3\n"), 0644)
			fm := files.NewFileManager()
			_, err := fm.UpdateLines(context.Background(), filesystem, "test.txt", tt.lineDiff)
			if err == nil {
				t.Fatalf("Expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.expected) {
				t.Errorf("Expected error to contain '%s', got %s", tt.expected, err.Error())
			}
		})
	}
}

func TestUpdateFile_Success(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected model.File
	}{
		{
			name:     "Update file",
			content:  "line1\nline2\nline3\n",
			expected: model.File{Path: "test.txt", Lines: []model.Line{{Number: 1, Content: "line1"}, {Number: 2, Content: "line2"}, {Number: 3, Content: "line3"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filesystem := afero.NewMemMapFs()
			afero.WriteFile(filesystem, "test.txt", []byte("line11\nline12\n"), 0644)
			fm := files.NewFileManager()
			actual, err := fm.UpdateFile(context.Background(), filesystem, "test.txt", tt.content)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !actual.Equals(&tt.expected) {
				t.Errorf("Expected %+v, got %+v", tt.expected, actual)
			}
		})
	}
}

func TestUpdateFile_Failure(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "File not found",
			content:  "whatever",
			expected: "file test.txt not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filesystem := afero.NewMemMapFs()
			fm := files.NewFileManager()
			_, err := fm.UpdateFile(context.Background(), filesystem, "test.txt", tt.content)
			if err == nil {
				t.Fatalf("Expected error, got nil")
			}

			if !strings.Contains(strings.ToLower(err.Error()), tt.expected) {
				t.Errorf("Expected error to contain '%s', got %s", tt.expected, err.Error())
			}
		})
	}
}
