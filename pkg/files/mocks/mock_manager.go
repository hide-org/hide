package mocks

import (
	"io/fs"

	"github.com/artmoskvin/hide/pkg/files"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/spf13/afero"
)

// MockFileManager is a mock of the filemanager.FileManager interface for testing
type MockFileManager struct {
	CreateFileFunc  func(path string, content string) (model.File, error)
	ReadFileFunc    func(fileSystem fs.FS, path string, props files.ReadProps) (model.File, error)
	UpdateFileFunc  func(fileSystem afero.Fs, path string, content string) (model.File, error)
	DeleteFileFunc  func(path string) error
	ListFilesFunc   func(rootPath string) ([]model.File, error)
	ApplyPatchFunc  func(fileSystem afero.Fs, path string, patch string) (model.File, error)
	UpdateLinesFunc func(filesystem afero.Fs, path string, lineDiff files.LineDiffChunk) (model.File, error)
}

func (m *MockFileManager) CreateFile(path string, content string) (model.File, error) {
	return m.CreateFileFunc(path, content)
}

func (m *MockFileManager) ReadFile(fileSystem fs.FS, path string, props files.ReadProps) (model.File, error) {
	return m.ReadFileFunc(fileSystem, path, props)
}

func (m *MockFileManager) UpdateFile(fileSystem afero.Fs, path string, content string) (model.File, error) {
	return m.UpdateFileFunc(fileSystem, path, content)
}

func (m *MockFileManager) DeleteFile(path string) error {
	return m.DeleteFileFunc(path)
}

func (m *MockFileManager) ListFiles(rootPath string) ([]model.File, error) {
	return m.ListFilesFunc(rootPath)
}

func (m *MockFileManager) ApplyPatch(fileSystem afero.Fs, path string, patch string) (model.File, error) {
	return m.ApplyPatchFunc(fileSystem, path, patch)
}

func (m *MockFileManager) UpdateLines(filesystem afero.Fs, path string, lineDiff files.LineDiffChunk) (model.File, error) {
	return m.UpdateLinesFunc(filesystem, path, lineDiff)
}
