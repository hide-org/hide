package mocks

import (
	"io/fs"

	"github.com/artmoskvin/hide/pkg/filemanager"
)

// MockFileManager is a mock of the filemanager.FileManager interface for testing
type MockFileManager struct {
	CreateFileFunc func(path string, content string) (filemanager.File, error)
	ReadFileFunc   func(fileSystem fs.FS, path string, props filemanager.ReadProps) (filemanager.File, error)
	UpdateFileFunc func(path string, content string) (filemanager.File, error)
	DeleteFileFunc func(path string) error
	ListFilesFunc  func(rootPath string) ([]filemanager.File, error)
}

func (m *MockFileManager) CreateFile(path string, content string) (filemanager.File, error) {
	return m.CreateFileFunc(path, content)
}

func (m *MockFileManager) ReadFile(fileSystem fs.FS, path string, props filemanager.ReadProps) (filemanager.File, error) {
	return m.ReadFileFunc(fileSystem, path, props)
}

func (m *MockFileManager) UpdateFile(path string, content string) (filemanager.File, error) {
	return m.UpdateFileFunc(path, content)
}

func (m *MockFileManager) DeleteFile(path string) error {
	return m.DeleteFileFunc(path)
}

func (m *MockFileManager) ListFiles(rootPath string) ([]filemanager.File, error) {
	return m.ListFilesFunc(rootPath)
}
