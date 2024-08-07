package mocks

import (
	"context"

	"github.com/artmoskvin/hide/pkg/files"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/spf13/afero"
)

// MockFileManager is a mock of the filemanager.FileManager interface for testing
type MockFileManager struct {
	CreateFileFunc  func(ctx context.Context, fs afero.Fs, path, content string) (model.File, error)
	ReadFileFunc    func(ctx context.Context, fs afero.Fs, path string, props files.ReadProps) (model.File, error)
	UpdateFileFunc  func(ctx context.Context, fs afero.Fs, path, content string) (model.File, error)
	DeleteFileFunc  func(ctx context.Context, fs afero.Fs, path string) error
	ListFilesFunc   func(ctx context.Context, fs afero.Fs) ([]model.File, error)
	ApplyPatchFunc  func(ctx context.Context, fs afero.Fs, path, patch string) (model.File, error)
	UpdateLinesFunc func(ctx context.Context, fs afero.Fs, path string, lineDiff files.LineDiffChunk) (model.File, error)
}

func (m *MockFileManager) CreateFile(ctx context.Context, fs afero.Fs, path, content string) (model.File, error) {
	return m.CreateFileFunc(ctx, fs, path, content)
}

func (m *MockFileManager) ReadFile(ctx context.Context, fs afero.Fs, path string, props files.ReadProps) (model.File, error) {
	return m.ReadFileFunc(ctx, fs, path, props)
}

func (m *MockFileManager) UpdateFile(ctx context.Context, fs afero.Fs, path, content string) (model.File, error) {
	return m.UpdateFileFunc(ctx, fs, path, content)
}

func (m *MockFileManager) DeleteFile(ctx context.Context, fs afero.Fs, path string) error {
	return m.DeleteFileFunc(ctx, fs, path)
}

func (m *MockFileManager) ListFiles(ctx context.Context, fs afero.Fs) ([]model.File, error) {
	return m.ListFilesFunc(ctx, fs)
}

func (m *MockFileManager) ApplyPatch(ctx context.Context, fs afero.Fs, path, patch string) (model.File, error) {
	return m.ApplyPatchFunc(ctx, fs, path, patch)
}

func (m *MockFileManager) UpdateLines(ctx context.Context, fs afero.Fs, path string, lineDiff files.LineDiffChunk) (model.File, error) {
	return m.UpdateLinesFunc(ctx, fs, path, lineDiff)
}
