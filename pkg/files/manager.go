package files

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/hide-org/hide/pkg/gitignore"
	"github.com/hide-org/hide/pkg/model"
	"github.com/spf13/afero"
)

type FileManager interface {
	CreateFile(ctx context.Context, fs afero.Fs, path, content string) (*model.File, error)
	ReadFile(ctx context.Context, fs afero.Fs, path string) (*model.File, error)
	UpdateFile(ctx context.Context, fs afero.Fs, path, content string) (*model.File, error)
	DeleteFile(ctx context.Context, fs afero.Fs, path string) error
	ListFiles(ctx context.Context, fs afero.Fs, opts ...ListFileOption) ([]*model.File, error)
	ApplyPatch(ctx context.Context, fs afero.Fs, path, patch string) (*model.File, error)
	UpdateLines(ctx context.Context, fs afero.Fs, path string, lineDiff LineDiffChunk) (*model.File, error)
}

type FileManagerImpl struct {
	gitignoreFactory gitignore.MatcherFactory
}

func NewFileManager(factory gitignore.MatcherFactory) FileManager {
	return &FileManagerImpl{gitignoreFactory: factory}
}

func (fm *FileManagerImpl) CreateFile(ctx context.Context, fs afero.Fs, path, content string) (*model.File, error) {
	exists, err := fileExists(fs, path)
	if err != nil {
		return nil, fmt.Errorf("Failed to check if file %s exists: %w", path, err)
	}

	if exists {
		return nil, NewFileAlreadyExistsError(path)
	}

	dir := filepath.Dir(path)
	if err := fs.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("Failed to create directory %s: %w", dir, err)
	}

	if err := afero.WriteFile(fs, path, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("Failed to write file %s: %w", path, err)
	}

	return model.NewFile(path, content), nil
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

	file := model.NewFile(path, content)

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

func (fm *FileManagerImpl) ListFiles(ctx context.Context, fs afero.Fs, opts ...ListFileOption) ([]*model.File, error) {
	var files []*model.File

	opt := &ListFilesOptions{}
	for _, o := range opts {
		o(opt)
	}

	m, err := fm.gitignoreFactory.NewMatcher(fs)
	if err != nil {
		return nil, fmt.Errorf("failed to create gitignore matcher: %w", err)
	}

	err = afero.Walk(fs, "/", func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return errors.New("context cancelled")
		default:
		}

		if err != nil {
			return fmt.Errorf("Error walking file tree on path %s: %w", path, err)
		}

		// check gitignore
		match, err := m.Match(path, info.IsDir())
		if err != nil {
			return fmt.Errorf("failed to match path %s: %w", path, err)
		}
		if match {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !opt.ShowHidden && isHidden(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		ok, err := opt.Filter.keep(path, info)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		if !info.IsDir() {
			if !opt.WithContent {
				path, err = filepath.Rel("/", path)
				if err != nil {
					return err
				}

				files = append(files, model.EmptyFile(path))
				return nil
			}

			file, err := readFile(fs, path)
			if err != nil {
				return fmt.Errorf("Error reading file %s: %w", path, err)
			}

			file.Path, err = filepath.Rel("/", file.Path)
			if err != nil {
				return err
			}

			files = append(files, file)
			return nil
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

	return model.NewFile(path, string(content)), nil
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
