package filemanager

import (
	"fmt"
	"os"
)

type File struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type FileManager interface {
	CreateFile(path string, content string) (File, error)
	ReadFile(path string) (File, error)
	UpdateFile(path string, content string) (File, error)
	DeleteFile(path string) error
}

type FileManagerImpl struct{}

func NewFileManager() FileManager {
	return &FileManagerImpl{}
}

func (fm *FileManagerImpl) CreateFile(path string, content string) (File, error) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return File{}, fmt.Errorf("Failed to create file: %w", err)
	}

	return File{Path: path, Content: content}, nil
}

func (fm *FileManagerImpl) ReadFile(path string) (File, error) {
	fmt.Println("Reading file", path)
	content, err := os.ReadFile(path)

	if err != nil {
		return File{}, fmt.Errorf("Failed to read file: %w", err)
	}

	return File{Path: path, Content: string(content)}, nil
}

func (fm *FileManagerImpl) UpdateFile(path string, content string) (File, error) {
	return fm.CreateFile(path, content)
}

func (fm *FileManagerImpl) DeleteFile(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("Failed to delete file: %w", err)
	}

	return nil
}
