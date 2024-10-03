package files_test

import (
	"context"
	"strings"
	"testing"

	"github.com/artmoskvin/hide/pkg/files"
	"github.com/artmoskvin/hide/pkg/gitignore/mocks"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
)

func TestReadFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "test.txt"
	content := "line1\nline2\nline3\n"
	afero.WriteFile(fs, path, []byte(content), 0o644)

	fm := files.NewFileManager(nil)
	actual, err := fm.ReadFile(context.Background(), fs, path)
	expected := model.NewFile(path, content)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !actual.Equals(expected) {
		t.Errorf("Expected %+v, got %+v", expected, actual)
	}
}

func TestReadNonExistentFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "test.txt", []byte("line1\nline2\nline3\n"), 0o644)

	fm := files.NewFileManager(nil)
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
			afero.WriteFile(filesystem, "test.txt", []byte("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n"), 0o644)
			fm := files.NewFileManager(nil)
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
			afero.WriteFile(fileSystem, "test.txt", []byte("line1\nline2\nline3\n"), 0o644)
			fm := files.NewFileManager(nil)
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
		expected *model.File
	}{
		{
			name: "Update 1 line",
			lineDiff: files.LineDiffChunk{
				StartLine: 1,
				EndLine:   2,
				Content:   "line11",
			},
			expected: model.NewFile("test.txt", "line11\nline2\nline3\n"),
		},
		{
			name: "Update multiple lines",
			lineDiff: files.LineDiffChunk{
				StartLine: 1,
				EndLine:   3,
				Content:   "line11\nline12",
			},
			expected: model.NewFile("test.txt", "line11\nline12\nline3\n"),
		},
		{
			name: "Add multiple lines at the end",
			lineDiff: files.LineDiffChunk{
				StartLine: 3,
				EndLine:   4,
				Content:   "line10\nline11\nline12",
			},
			expected: model.NewFile("test.txt", "line1\nline2\nline10\nline11\nline12\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := files.NewFileManager(nil)
			filesystem := afero.NewMemMapFs()
			afero.WriteFile(filesystem, "test.txt", []byte("line1\nline2\nline3\n"), 0o644)
			actual, err := fm.UpdateLines(context.Background(), filesystem, "test.txt", tt.lineDiff)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !actual.Equals(tt.expected) {
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
			expected: "Start line must be less than or equal to 4",
		},
		{
			name: "End line > number of lines + 1",
			lineDiff: files.LineDiffChunk{
				StartLine: 1,
				EndLine:   11,
				Content:   "line11",
			},
			expected: "End line must be less than or equal to 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filesystem := afero.NewMemMapFs()
			afero.WriteFile(filesystem, "test.txt", []byte("line1\nline2\nline3\n"), 0o644)
			fm := files.NewFileManager(nil)
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
		expected *model.File
	}{
		{
			name:     "Update file",
			content:  "line1\nline2\nline3\n",
			expected: model.NewFile("test.txt", "line1\nline2\nline3\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filesystem := afero.NewMemMapFs()
			afero.WriteFile(filesystem, "test.txt", []byte("line11\nline12\n"), 0o644)
			fm := files.NewFileManager(nil)
			actual, err := fm.UpdateFile(context.Background(), filesystem, "test.txt", tt.content)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !actual.Equals(tt.expected) {
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
			fm := files.NewFileManager(nil)
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

func TestListFile(t *testing.T) {
	// RUN test
	for _, tt := range []struct {
		name      string
		fs        afero.Fs
		mockSetup func(*mocks.MockMatcherFactory)
		opts      []files.ListFileOption
		wantFile  []*model.File
	}{
		{
			name: "all files",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/hello.txt",
						content: "Hi there\n",
					},
					{
						path:    "/something/something.txt",
						content: "something1\nsomething2\nsomething3\n",
					},
					{
						path:    "/something/items.json",
						content: `["a1","a2"]`,
					},
					{
						path:    "/node_modules/module1/file.js",
						content: `import tmp`,
					},
					{
						path:    "/node_modules/module2/file.js",
						content: `import tmp`,
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			wantFile: []*model.File{
				model.EmptyFile("hello.txt"),
				model.EmptyFile("node_modules/module1/file.js"),
				model.EmptyFile("node_modules/module2/file.js"),
				model.EmptyFile("something/items.json"),
				model.EmptyFile("something/something.txt"),
			},
		},
		{
			name: "with exclude filter",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/hello.txt",
						content: "Hi there\n",
					},
					{
						path:    "/something/something.txt",
						content: "something1\nsomething2\nsomething3\n",
					},
					{
						path:    "/something/items.json",
						content: `["a1","a2"]`,
					},
					{
						path:    "/node_modules/module1/file.js",
						content: `import tmp`,
					},
					{
						path:    "/node_modules/module2/file.js",
						content: `import tmp`,
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"*something*"},
					Exclude: []string{"*.json", "node_modules"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("something/something.txt"),
			},
		},
		{
			name: "match directory anywhere",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/logs/debug.log",
						content: "content",
					},
					{
						path:    "/logs/monday/foo.bar",
						content: "content",
					},
					{
						path:    "/build/logs/debug.log",
						content: "content",
					},
					{
						path:    "/node_modules/module1/file.js",
						content: "content",
					},
					{
						path:    "/node_modules/module2/file.js",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"**/logs/**"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("build/logs/debug.log"),
				model.EmptyFile("logs/debug.log"),
				model.EmptyFile("logs/monday/foo.bar"),
			},
		},
		{
			name: "match directory with file",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/logs/debug.log",
						content: "content",
					},
					{
						path:    "/logs/build/debug.log",
						content: "content",
					},
					{
						path:    "/build/logs/debug.log",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"**/logs/debug.log"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("build/logs/debug.log"),
				model.EmptyFile("logs/debug.log"),
			},
		},
		{
			name: "match file extension",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/logs/debug.log",
						content: "content",
					},
					{
						path:    "/logs/build/debug.log",
						content: "content",
					},
					{
						path:    "/build/logs/debug.log",
						content: "content",
					},
					{
						path:    "/node_modules/module1/file.js",
						content: "content",
					},
					{
						path:    "/node_modules/module2/file.js",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"*.log"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("build/logs/debug.log"),
				model.EmptyFile("logs/build/debug.log"),
				model.EmptyFile("logs/debug.log"),
			},
		},
		{
			name: "match file extension with exclude",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/logs/debug.log",
						content: "content",
					},
					{
						path:    "/logs/build/debug.log",
						content: "content",
					},
					{
						path:    "/build/logs/debug.log",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"*.log"},
					Exclude: []string{"**/build/**"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("logs/debug.log"),
			},
		},
		{
			name: "include files only from root",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/debug.log",
						content: "content",
					},
					{
						path:    "/build/debug.log",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"/debug.log"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("debug.log"),
			},
		},
		{
			name: "include files by name",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/debug.log",
						content: "content",
					},
					{
						path:    "/build/debug.log",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"**/debug.log"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("build/debug.log"),
				model.EmptyFile("debug.log"),
			},
		},
		{
			name: "include files with question mark",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/debug0.log",
						content: "content",
					},
					{
						path:    "/debug1.log",
						content: "content",
					},
					{
						path:    "/debug10.log",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"**/debug?.log"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("debug0.log"),
				model.EmptyFile("debug1.log"),
			},
		},
		{
			name: "include files with numeric range",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/debug0.log",
						content: "content",
					},
					{
						path:    "/debug1.log",
						content: "content",
					},
					{
						path:    "/debug10.log",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"**/debug[0-9].log"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("debug0.log"),
				model.EmptyFile("debug1.log"),
			},
		},
		{
			name: "include files with character set",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/debug0.log",
						content: "content",
					},
					{
						path:    "/debug1.log",
						content: "content",
					},
					{
						path:    "/debug2.log",
						content: "content",
					},
					{
						path:    "/debug10.log",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"**/debug[01].log"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("debug0.log"),
				model.EmptyFile("debug1.log"),
			},
		},
		{
			name: "include files with negated character set",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/debug0.log",
						content: "content",
					},
					{
						path:    "/debug1.log",
						content: "content",
					},
					{
						path:    "/debug2.log",
						content: "content",
					},
					{
						path:    "/debug3.log",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"/debug[!01].log"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("debug2.log"),
				model.EmptyFile("debug3.log"),
			},
		},
		{
			name: "include files with alphabetical range",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/debug0.log",
						content: "content",
					},
					{
						path:    "/debug1.log",
						content: "content",
					},
					{
						path:    "/debuga.log",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"/debug[a-z].log"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("debuga.log"),
			},
		},
		{
			name: "include files and directories",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/logs.txt",
						content: "content",
					},
					{
						path:    "/logs/debug.log",
						content: "content",
					},
					{
						path:    "/logs/latest/foo.bar",
						content: "content",
					},
					{
						path:    "/build/logs.txt",
						content: "content",
					},
					{
						path:    "/build/logs/debug.log",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"**/logs*", "**/logs/**"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("build/logs/debug.log"),
				model.EmptyFile("build/logs.txt"),
				model.EmptyFile("logs/debug.log"),
				model.EmptyFile("logs/latest/foo.bar"),
				model.EmptyFile("logs.txt"),
			},
		},
		{
			name: "include only directories",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/logs.txt",
						content: "content",
					},
					{
						path:    "/logs/debug.log",
						content: "content",
					},
					{
						path:    "/logs/latest/foo.bar",
						content: "content",
					},
					{
						path:    "/build/logs.txt",
						content: "content",
					},
					{
						path:    "/build/logs/debug.log",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"**/logs/**"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("build/logs/debug.log"),
				model.EmptyFile("logs/debug.log"),
				model.EmptyFile("logs/latest/foo.bar"),
			},
		},
		{
			name: "include with double asterisk",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/logs/debug.log",
						content: "content",
					},
					{
						path:    "/logs/monday/debug.log",
						content: "content",
					},
					{
						path:    "/logs/monday/pm/debug.log",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"/logs/**/debug.log"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("logs/debug.log"),
				model.EmptyFile("logs/monday/debug.log"),
				model.EmptyFile("logs/monday/pm/debug.log"),
			},
		},
		{
			name: "include with wildcard in directory",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/logs/debug.log",
						content: "content",
					},
					{
						path:    "/logs/monday/debug.log",
						content: "content",
					},
					{
						path:    "/logs/monday/pm/debug.log",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"/logs/*day/debug.log"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("logs/monday/debug.log"),
			},
		},
		{
			name: "include file with directory",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/logs/debug.log",
						content: "content",
					},
					{
						path:    "/debug.log",
						content: "content",
					},
					{
						path:    "/build/logs/debug.log",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithFilter(files.PatternFilter{
					Include: []string{"/logs/debug.log"},
				}),
			},
			wantFile: []*model.File{
				model.EmptyFile("logs/debug.log"),
			},
		},
		{
			name: "with content",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/test1.txt",
						content: "content-1",
					},
					{
						path:    "/test2.txt",
						content: "content-2",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithContent(),
			},
			wantFile: []*model.File{
				model.NewFile("test1.txt", "content-1"),
				model.NewFile("test2.txt", "content-2"),
			},
		},
		{
			name: "with hidden",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/.hidden.txt",
						content: "content",
					},
					{
						path:    "/file.txt",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			opts: []files.ListFileOption{
				files.ListFilesWithShowHidden(),
			},
			wantFile: []*model.File{
				model.EmptyFile(".hidden.txt"),
				model.EmptyFile("file.txt"),
			},
		},
		{
			name: "with gitignore match",
			fs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				for _, file := range []struct {
					path    string
					content string
				}{
					{
						path:    "/ignore.txt",
						content: "content",
					},
					{
						path:    "/do-not-ignore.txt",
						content: "content",
					},
				} {
					if err := afero.WriteFile(fs, file.path, []byte(file.content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return fs
			}(),
			mockSetup: func(m *mocks.MockMatcherFactory) {
				mockMatcher := mocks.NewMockMatcher()
				mockMatcher.On("Match", "/do-not-ignore.txt", mock.Anything).Return(false, nil)
				mockMatcher.On("Match", "/ignore.txt", mock.Anything).Return(true, nil)
				mockMatcher.On("Match", mock.Anything, mock.Anything).Return(false, nil)
				m.On("NewMatcher", mock.Anything).Return(mockMatcher, nil)
			},
			wantFile: []*model.File{
				model.EmptyFile("do-not-ignore.txt"),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mockGitignoreFactory := mocks.NewMockMatcherFactory()
			tt.mockSetup(mockGitignoreFactory)
			fm := files.NewFileManager(mockGitignoreFactory)

			files, err := fm.ListFiles(context.Background(), tt.fs, tt.opts...)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(files, tt.wantFile); diff != "" {
				t.Fatalf("got diff %s", diff)
			}
		})
	}
}
