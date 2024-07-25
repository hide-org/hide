package mocks

import (
	"io/fs"

	"github.com/artmoskvin/hide/pkg/files"
	"github.com/spf13/afero"
)

// MockFileManager is a mock of the filemanager.FileManager interface for testing
type MockFileManager struct {
	CreateFileFunc  func(path string, content string) (files.File, error)
	ReadFileFunc    func(fileSystem fs.FS, path string, props files.ReadProps) (files.File, error)
	UpdateFileFunc  func(path string, content string) (files.File, error)
	DeleteFileFunc  func(path string) error
	ListFilesFunc   func(rootPath string) ([]files.File, error)
	ApplyPatchFunc  func(fileSystem afero.Fs, path string, patch string) (files.File, error)
	UpdateLinesFunc func(path string, lineDiffs []files.LineDiffChunk) (files.File, error)
}

func (m *MockFileManager) CreateFile(path string, content string) (files.File, error) {
	return m.CreateFileFunc(path, content)
}

func (m *MockFileManager) ReadFile(fileSystem fs.FS, path string, props files.ReadProps) (files.File, error) {
	return m.ReadFileFunc(fileSystem, path, props)
}

func (m *MockFileManager) UpdateFile(path string, content string) (files.File, error) {
	return m.UpdateFileFunc(path, content)
}

func (m *MockFileManager) DeleteFile(path string) error {
	return m.DeleteFileFunc(path)
}

func (m *MockFileManager) ListFiles(rootPath string) ([]files.File, error) {
	return m.ListFilesFunc(rootPath)
}

func (m *MockFileManager) ApplyPatch(fileSystem afero.Fs, path string, patch string) (files.File, error) {
	return m.ApplyPatchFunc(fileSystem, path, patch)
}

func (m *MockFileManager) UpdateLines(path string, lineDiffs []files.LineDiffChunk) (files.File, error) {
	return m.UpdateLinesFunc(path, lineDiffs)
}
