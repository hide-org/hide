package files

import "fmt"

type FileNotFoundError struct {
	path string
}

func (e FileNotFoundError) Error() string {
	return fmt.Sprintf("file %s not found", e.path)
}

func NewFileNotFoundError(path string) *FileNotFoundError {
	return &FileNotFoundError{path: path}
}

type FileAlreadyExistsError struct {
	path string
}

func (e FileAlreadyExistsError) Error() string {
	return fmt.Sprintf("file %s already exists", e.path)
}

func NewFileAlreadyExistsError(path string) *FileAlreadyExistsError {
	return &FileAlreadyExistsError{path: path}
}
