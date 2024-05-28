package filemanager

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type File struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type FileManager interface {
	CreateFile(path string, content string) (File, error)
	ReadFile(fileSystem fs.FS, path string) (File, error)
	UpdateFile(path string, content string) (File, error)
	DeleteFile(path string) error
	ListFiles(rootPath string) ([]File, error)
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

func (fm *FileManagerImpl) ReadFile(fileSystem fs.FS, path string) (File, error) {
	log.Println("Reading file", path)
	content, err := fs.ReadFile(fileSystem, path)

	if err != nil {
		log.Printf("Failed to open file %s: %s", path, err)
		return File{}, fmt.Errorf("Failed to open file: %w", err)
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

func (fm *FileManagerImpl) ListFiles(rootPath string) ([]File, error) {
	log.Println("Listing files in", rootPath)

	var files []File

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error walking directory %s on path %s: %s", rootPath, path, err)
			return fmt.Errorf("Error walking directory %s on path %s: %w", rootPath, path, err)
		}

		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			log.Printf("Error getting relative path from %s to %s: %s", rootPath, path, err)
			return fmt.Errorf("Error getting relative path from %s to %s: %w", rootPath, path, err)
		}

		if !info.IsDir() {
			file, err := fm.ReadFile(os.DirFS(rootPath), relativePath)
			if err != nil {
				log.Printf("Error reading file %s: %s", path, err)
				return fmt.Errorf("Error reading file %s: %w", path, err)
			}

			files = append(files, file)
		}

		return nil
	})

	return files, err
}
