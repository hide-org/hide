package files_test

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/artmoskvin/hide/pkg/files"
)

func TestNewReadProps(t *testing.T) {
	tests := []struct {
		name     string
		props    []files.ReadPropsSetter
		expected files.ReadProps
	}{
		{
			name: "ShowLineNumbers",
			props: []files.ReadPropsSetter{
				func(props *files.ReadProps) {
					props.ShowLineNumbers = true
				},
			},
			expected: files.ReadProps{
				ShowLineNumbers: true,
				StartLine:       files.DefaultStartLine,
				NumLines:        files.DefaultNumLines,
			},
		},
		{
			name: "StartLine",
			props: []files.ReadPropsSetter{
				func(props *files.ReadProps) {
					props.StartLine = 10
				},
			},
			expected: files.ReadProps{
				ShowLineNumbers: files.DefaultShowLineNumbers,
				StartLine:       10,
				NumLines:        files.DefaultNumLines,
			},
		},
		{
			name: "NumLines",
			props: []files.ReadPropsSetter{
				func(props *files.ReadProps) {
					props.NumLines = 20
				},
			},
			expected: files.ReadProps{
				ShowLineNumbers: files.DefaultShowLineNumbers,
				StartLine:       files.DefaultStartLine,
				NumLines:        20,
			},
		},
		{
			name: "All",
			props: []files.ReadPropsSetter{
				func(props *files.ReadProps) {
					props.ShowLineNumbers = true
					props.StartLine = 10
					props.NumLines = 20
				},
			},
			expected: files.ReadProps{
				ShowLineNumbers: true,
				StartLine:       10,
				NumLines:        20,
			},
		},
		{
			name:  "Default",
			props: []files.ReadPropsSetter{},
			expected: files.ReadProps{
				ShowLineNumbers: files.DefaultShowLineNumbers,
				StartLine:       files.DefaultStartLine,
				NumLines:        files.DefaultNumLines,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := files.NewReadProps(tt.props...)
			if actual != tt.expected {
				t.Errorf("Expected %+v, got %+v", tt.expected, actual)
			}
		})
	}
}

func TestFileManagerImpl_ReadFile_Success(t *testing.T) {
	files := fstest.MapFS{
		"test.txt": {Data: []byte("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")},
	}

	tests := []struct {
		name     string
		fs       fs.FS
		filePath string
		props    files.ReadPropsSetter
		expected files.File
	}{
		{
			name:     "ShowLineNumbers = true",
			fs:       files,
			filePath: "test.txt",
			props: func(props *files.ReadProps) {
				props.ShowLineNumbers = true
				props.StartLine = 2
				props.NumLines = 3
			},
			expected: files.File{
				Path:    "test.txt",
				Content: "2:line2\n3:line3\n4:line4\n",
			},
		},
		{
			name:     "ShowLineNumbers = false",
			fs:       files,
			filePath: "test.txt",
			props: func(props *files.ReadProps) {
				props.ShowLineNumbers = false
				props.StartLine = 4
				props.NumLines = 4
			},
			expected: files.File{
				Path:    "test.txt",
				Content: "line4\nline5\nline6\nline7\n",
			},
		},
		{
			name:     "NumLines = 0",
			fs:       files,
			filePath: "test.txt",
			props: func(props *files.ReadProps) {
				props.ShowLineNumbers = true
				props.StartLine = 2
				props.NumLines = 0
			},
			expected: files.File{
				Path:    "test.txt",
				Content: "",
			},
		},
		{
			name:     "If StartLine + NumLines > number of lines then show all lines",
			fs:       files,
			filePath: "test.txt",
			props: func(props *files.ReadProps) {
				props.ShowLineNumbers = true
				props.StartLine = 5
				props.NumLines = 10
			},
			expected: files.File{
				Path:    "test.txt",
				Content: " 5:line5\n 6:line6\n 7:line7\n 8:line8\n 9:line9\n10:line10\n",
			},
		},
		{
			name:     "Line numbers are padded with spaces",
			fs:       files,
			filePath: "test.txt",
			props: func(props *files.ReadProps) {
				props.ShowLineNumbers = true
				props.StartLine = 5
				props.NumLines = 10
			},
			expected: files.File{
				Path:    "test.txt",
				Content: " 5:line5\n 6:line6\n 7:line7\n 8:line8\n 9:line9\n10:line10\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := files.NewFileManager()
			actual, err := fm.ReadFile(tt.fs, tt.filePath, files.NewReadProps(tt.props))
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if actual != tt.expected {
				t.Errorf("Expected %+v, got %+v", tt.expected, actual)
			}
		})
	}
}

func TestFileManagerImpl_ReadFile_Failure(t *testing.T) {
	files := fstest.MapFS{
		"test.txt": {Data: []byte("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")},
	}

	tests := []struct {
		name     string
		fs       fs.FS
		filePath string
		props    files.ReadPropsSetter
		expected string
	}{
		{
			name:     "StartLine < 1",
			fs:       files,
			filePath: "test.txt",
			props: func(props *files.ReadProps) {
				props.StartLine = 0
			},
			expected: "Start line must be greater than or equal to 1",
		},
		{
			name:     "StartLine > number of lines",
			fs:       files,
			filePath: "test.txt",
			props: func(props *files.ReadProps) {
				props.StartLine = 11
			},
			expected: "Start line must be less than or equal to 10",
		},
		{
			name:     "NumLines < 0",
			fs:       files,
			filePath: "test.txt",
			props: func(props *files.ReadProps) {
				props.NumLines = -1
			},
			expected: "Number of lines must be greater than or equal to 0",
		},
		{
			name:     "Failed to read file",
			fs:       fstest.MapFS{},
			filePath: "test.txt",
			props:    func(props *files.ReadProps) {},
			expected: "Failed to open file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := files.NewFileManager()
			_, err := fm.ReadFile(tt.fs, tt.filePath, files.NewReadProps(tt.props))
			if err == nil {
				t.Fatalf("Expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.expected) {
				t.Errorf("Expected error to contain '%s', got %s", tt.expected, err.Error())
			}
		})
	}
}
