package files

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/artmoskvin/hide/pkg/model"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/spf13/afero"
)

type FileManager interface {
	CreateFile(ctx context.Context, fs afero.Fs, path, content string) (*model.File, error)
	ReadFile(ctx context.Context, fs afero.Fs, path string) (*model.File, error)
	UpdateFile(ctx context.Context, fs afero.Fs, path, content string) (*model.File, error)
	DeleteFile(ctx context.Context, fs afero.Fs, path string) error
	ListFiles(ctx context.Context, fs afero.Fs, showHidden bool, filter PatternFilter) ([]*model.File, error)
	ApplyPatch(ctx context.Context, fs afero.Fs, path, patch string) (*model.File, error)
	UpdateLines(ctx context.Context, fs afero.Fs, path string, lineDiff LineDiffChunk) (*model.File, error)
}

type FileManagerImpl struct{}

func NewFileManager() FileManager {
	return &FileManagerImpl{}
}

func (fm *FileManagerImpl) CreateFile(ctx context.Context, fs afero.Fs, path, content string) (*model.File, error) {
	exists, err := fileExists(fs, path)
	if err != nil {
		return nil, fmt.Errorf("Failed to check if file %s exists: %w", path, err)
	}

	if exists {
		return nil, NewFileAlreadyExistsError(path)
	}

	file, err := model.NewFile(path, content)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(file.Path)
	if err := fs.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("Failed to create directory %s: %w", dir, err)
	}

	if err := writeFile(fs, file); err != nil {
		return nil, fmt.Errorf("Failed to write file %s: %w", path, err)
	}

	return file, nil
}

func (fm *FileManagerImpl) ReadFile(ctx context.Context, fs afero.Fs, path string) (*model.File, error) {
	exists, err := fileExists(fs, path)
	if err != nil {
		return nil, fmt.Errorf("Failed to check if file %s exists: %w", path, err)
	}

	if !exists {
		return nil, NewFileNotFoundError(path)
	}

	return readFile(fs, path)
}

func (fm *FileManagerImpl) UpdateFile(ctx context.Context, fs afero.Fs, path, content string) (*model.File, error) {
	exists, err := fileExists(fs, path)
	if err != nil {
		return nil, fmt.Errorf("Failed to check if file %s exists: %w", path, err)
	}

	if !exists {
		return nil, NewFileNotFoundError(path)
	}

	file, err := model.NewFile(path, content)
	if err != nil {
		return nil, fmt.Errorf("Failed to create file: %w", err)
	}

	if err := writeFile(fs, file); err != nil {
		return nil, fmt.Errorf("Failed to write file %s: %w", path, err)
	}

	return readFile(fs, path)
}

func (fm *FileManagerImpl) DeleteFile(ctx context.Context, fs afero.Fs, path string) error {
	exists, err := fileExists(fs, path)
	if err != nil {
		return fmt.Errorf("Failed to check if file %s exists: %w", path, err)
	}

	if !exists {
		return NewFileNotFoundError(path)
	}
	return fs.Remove(path)
}

type PatternFilter struct {
	Include []string
	Exclude []string
}

func (p PatternFilter) Keep(path string) (ok bool, err error) {
	basePath := filepath.Base(path)

	exclude, err := p.shouldExclude(basePath)
	if err != nil {
		return false, err
	}
	if exclude {
		return false, nil
	}

	include, err := p.shouldInclude(basePath)
	if err != nil {
		return false, err
	}
	if !include {
		return false, nil
	}

	return true, nil
}

func (p PatternFilter) shouldInclude(basePath string) (ok bool, err error) {
	if len(p.Include) == 0 {
		return true, nil
	}

	for _, pattern := range p.Include {
		matched, err := filepath.Match(pattern, basePath)
		if err != nil {
			return false, fmt.Errorf("Error include matching pattern %s: %w", pattern, err)
		}
		if matched {
			return matched, nil
		}
	}

	return false, nil
}

func (p PatternFilter) shouldExclude(basePath string) (ok bool, err error) {
	if len(p.Exclude) == 0 {
		return false, nil
	}

	for _, pattern := range p.Exclude {
		matched, err := filepath.Match(pattern, basePath)
		if err != nil {
			return false, fmt.Errorf("Error exclude matching pattern %s: %w", pattern, err)
		}
		if matched {
			return matched, nil
		}
	}

	return false, nil
}

func (fm *FileManagerImpl) ListFiles(ctx context.Context, fs afero.Fs, showHidden bool, filter PatternFilter) ([]*model.File, error) {
	var files []*model.File

	err := afero.Walk(fs, "/", func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return errors.New("context cancelled")
		default:
		}

		if err != nil {
			return fmt.Errorf("Error walking file tree on path %s: %w", path, err)
		}

		if !showHidden && isHidden(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		ok, err := filter.Keep(path)
		if err != nil {
			return err
		}
		if !ok {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() {
			file, err := readFile(fs, path)
			if err != nil {
				return fmt.Errorf("Error reading file %s: %w", path, err)
			}

			files = append(files, file)
		}

		return nil
	})

	return files, err
}

func (fm *FileManagerImpl) ApplyPatch(ctx context.Context, fs afero.Fs, path, patch string) (*model.File, error) {
	exists, err := fileExists(fs, path)
	if err != nil {
		return nil, fmt.Errorf("Failed to check if file %s exists: %w", path, err)
	}

	if !exists {
		return nil, NewFileNotFoundError(path)
	}

	file, err := readFile(fs, path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file %s: %w", path, err)
	}

	files, _, err := gitdiff.Parse(strings.NewReader(patch))
	if err != nil {
		return nil, fmt.Errorf("Failed to parse patch: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("No files changed in patch")
	}

	if len(files) > 1 {
		return nil, fmt.Errorf("Patch cannot contain multiple files")
	}

	var output bytes.Buffer

	if err := gitdiff.Apply(&output, strings.NewReader(file.GetContent()), files[0]); err != nil {
		return nil, fmt.Errorf("Failed to apply patch to %s: %w\n%s", path, err, patch)
	}

	if err := afero.WriteFile(fs, path, output.Bytes(), 0o644); err != nil {
		return nil, fmt.Errorf("Failed to write file %s after applying patch: %w", path, err)
	}

	return readFile(fs, path)
}

func (fm *FileManagerImpl) UpdateLines(ctx context.Context, fs afero.Fs, path string, lineDiff LineDiffChunk) (*model.File, error) {
	exists, err := fileExists(fs, path)
	if err != nil {
		return nil, fmt.Errorf("Failed to check if file %s exists: %w", path, err)
	}

	if !exists {
		return nil, NewFileNotFoundError(path)
	}

	file, err := readFile(fs, path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file %s: %w", path, err)
	}

	numLines := len(file.Lines)

	if lineDiff.StartLine == lineDiff.EndLine {
		return nil, fmt.Errorf("Start line must be less than end line")
	}

	if lineDiff.StartLine > numLines {
		return nil, fmt.Errorf("Start line must be less than or equal to %d", numLines)
	}

	if lineDiff.EndLine > numLines+1 {
		return nil, fmt.Errorf("End line must be less than or equal to %d", numLines+1)
	}

	file, err = file.ReplaceLineRange(lineDiff.StartLine, lineDiff.EndLine, lineDiff.Content)
	if err != nil {
		return nil, fmt.Errorf("Failed to replace lines: %w", err)
	}

	if err := writeFile(fs, file); err != nil {
		return nil, fmt.Errorf("Failed to write file %s: %w", path, err)
	}

	return readFile(fs, path)
}

func readFile(fs afero.Fs, path string) (*model.File, error) {
	content, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, err
	}

	return model.NewFile(path, string(content))
}

func writeFile(fs afero.Fs, file *model.File) error {
	return afero.WriteFile(fs, file.Path, []byte(file.GetContent()), 0o644)
}

func fileExists(fs afero.Fs, path string) (bool, error) {
	return afero.Exists(fs, path)
}

func isHidden(path string) bool {
	name := filepath.Base(path)
	return strings.HasPrefix(name, ".") && name != "." && name != ".."
}
