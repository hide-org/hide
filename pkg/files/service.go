package files

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/hide-org/hide/pkg/gitignore"
	"github.com/hide-org/hide/pkg/lsp/v2"
	"github.com/hide-org/hide/pkg/model"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

const MaxDiagnosticsDelay = time.Second * 1

type Service interface {
	CreateFile(ctx context.Context, path, content string) (*model.File, error)
	ReadFile(ctx context.Context, path string) (*model.File, error)
	UpdateFile(ctx context.Context, path, content string) (*model.File, error)
	DeleteFile(ctx context.Context, path string) error
	ListFiles(ctx context.Context, opts ...ListFileOption) (model.Files, error)
	ApplyPatch(ctx context.Context, path, patch string) (*model.File, error)
	UpdateLines(ctx context.Context, path string, lineDiff LineDiffChunk) (*model.File, error)
}

type ServiceImpl struct {
	gitignoreFactory gitignore.MatcherFactory
	lspService       lsp.Service
	fs               afero.Fs
}

func NewService(factory gitignore.MatcherFactory, lspService lsp.Service, fs afero.Fs) Service {
	return &ServiceImpl{gitignoreFactory: factory, lspService: lspService, fs: fs}
}

func (s *ServiceImpl) CreateFile(ctx context.Context, path, content string) (*model.File, error) {
	exists, err := fileExists(s.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to check if file %s exists: %w", path, err)
	}

	if exists {
		return nil, NewFileAlreadyExistsError(path)
	}

	dir := filepath.Dir(path)
	if err := s.fs.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := afero.WriteFile(s.fs, path, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("failed to write file %s: %w", path, err)
	}

	file, err := readFile(s.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s after creating it: %w", path, err)
	}

	// TODO: check if should fetch diagnostics here
	diagnostics, err := s.getDiagnostics(ctx, *file, MaxDiagnosticsDelay)
	if err != nil {
		log.Warn().Err(err).Str("path", path).Msg("Failed to get diagnostics but ignoring it.")

		return file, nil
	}

	return model.NewFile(path, content).WithDiagnostics(diagnostics), nil
}

func (s *ServiceImpl) ReadFile(ctx context.Context, path string) (*model.File, error) {
	exists, err := fileExists(s.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to check if file %s exists: %w", path, err)
	}

	if !exists {
		return nil, NewFileNotFoundError(path)
	}

	file, err := readFile(s.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	// TODO: check if should fetch diagnostics here
	diagnostics, err := s.getDiagnostics(ctx, *file, MaxDiagnosticsDelay)
	if err != nil {
		log.Warn().Err(err).Str("path", path).Msg("Failed to get diagnostics but ignoring it.")

		return file, nil
	}

	return file.WithDiagnostics(diagnostics), nil
}

func (s *ServiceImpl) UpdateFile(ctx context.Context, path, content string) (*model.File, error) {
	exists, err := fileExists(s.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to check if file %s exists: %w", path, err)
	}

	if !exists {
		return nil, NewFileNotFoundError(path)
	}

	file := model.NewFile(path, content)

	if err := writeFile(s.fs, file); err != nil {
		return nil, fmt.Errorf("failed to write file %s: %w", path, err)
	}

	file, err = readFile(s.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s after updating it: %w", path, err)
	}

	// TODO: check if should fetch diagnostics here
	diagnostics, err := s.getDiagnostics(ctx, *file, MaxDiagnosticsDelay)
	if err != nil {
		log.Warn().Err(err).Str("path", path).Msg("Failed to get diagnostics but ignoring it.")

		return file, nil
	}

	return file.WithDiagnostics(diagnostics), nil
}

func (s *ServiceImpl) DeleteFile(ctx context.Context, path string) error {
	exists, err := fileExists(s.fs, path)
	if err != nil {
		return fmt.Errorf("failed to check if file %s exists: %w", path, err)
	}

	if !exists {
		return NewFileNotFoundError(path)
	}
	return s.fs.Remove(path)
}

func (s *ServiceImpl) ListFiles(ctx context.Context, opts ...ListFileOption) (model.Files, error) {
	var files []*model.File

	opt := &ListFilesOptions{}
	for _, o := range opts {
		o(opt)
	}

	m, err := s.gitignoreFactory.NewMatcher(s.fs)
	if err != nil {
		return nil, fmt.Errorf("failed to create gitignore matcher: %w", err)
	}

	err = afero.Walk(s.fs, "/", func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return errors.New("context cancelled")
		default:
		}

		if err != nil {
			return fmt.Errorf("error walking file tree on path %s: %w", path, err)
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

			file, err := readFile(s.fs, path)
			if err != nil {
				return fmt.Errorf("error reading file %s: %w", path, err)
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

func (s *ServiceImpl) ApplyPatch(ctx context.Context, path, patch string) (*model.File, error) {
	exists, err := fileExists(s.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to check if file %s exists: %w", path, err)
	}

	if !exists {
		return nil, NewFileNotFoundError(path)
	}

	file, err := readFile(s.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	files, _, err := gitdiff.Parse(strings.NewReader(patch))
	if err != nil {
		return nil, fmt.Errorf("failed to parse patch: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files changed in patch")
	}

	if len(files) > 1 {
		return nil, fmt.Errorf("patch cannot contain multiple files")
	}

	var output bytes.Buffer

	if err := gitdiff.Apply(&output, strings.NewReader(file.GetContent()), files[0]); err != nil {
		return nil, fmt.Errorf("failed to apply patch to %s: %w\n%s", path, err, patch)
	}

	if err := afero.WriteFile(s.fs, path, output.Bytes(), 0o644); err != nil {
		return nil, fmt.Errorf("failed to write file %s after applying patch: %w", path, err)
	}

	file, err = readFile(s.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s after applying patch: %w", path, err)
	}

	// TODO: check if should fetch diagnostics here
	diagnostics, err := s.getDiagnostics(ctx, *file, MaxDiagnosticsDelay)
	if err != nil {
		log.Warn().Err(err).Str("path", path).Msg("Failed to get diagnostics but ignoring it.")

		return file, nil
	}

	return file.WithDiagnostics(diagnostics), nil
}

func (s *ServiceImpl) UpdateLines(ctx context.Context, path string, lineDiff LineDiffChunk) (*model.File, error) {
	exists, err := fileExists(s.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to check if file %s exists: %w", path, err)
	}

	if !exists {
		return nil, NewFileNotFoundError(path)
	}

	file, err := readFile(s.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	numLines := len(file.Lines)

	if lineDiff.StartLine == lineDiff.EndLine {
		return nil, fmt.Errorf("start line must be less than end line")
	}

	if lineDiff.StartLine > numLines {
		return nil, fmt.Errorf("start line must be less than or equal to %d", numLines)
	}

	if lineDiff.EndLine > numLines+1 {
		return nil, fmt.Errorf("end line must be less than or equal to %d", numLines+1)
	}

	file, err = file.ReplaceLineRange(lineDiff.StartLine, lineDiff.EndLine, lineDiff.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to replace lines: %w", err)
	}

	if err := writeFile(s.fs, file); err != nil {
		return nil, fmt.Errorf("failed to write file %s: %w", path, err)
	}

	file, err = readFile(s.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s after updating lines: %w", path, err)
	}

	// TODO: check if should fetch diagnostics here
	diagnostics, err := s.getDiagnostics(ctx, *file, MaxDiagnosticsDelay)
	if err != nil {
		log.Warn().Err(err).Str("path", path).Msg("Failed to get diagnostics but ignoring it.")

		return file, nil
	}

	return file.WithDiagnostics(diagnostics), nil
}

func (s *ServiceImpl) getDiagnostics(ctx context.Context, file model.File, waitFor time.Duration) ([]protocol.Diagnostic, error) {
	if err := s.lspService.NotifyDidOpen(ctx, file); err != nil {
		var lspLanguageServerNotFoundError *lsp.LanguageServerNotFoundError
		if errors.As(err, &lspLanguageServerNotFoundError) {
			return nil, nil
		}

		return nil, fmt.Errorf("Failed to notify didOpen while reading file %s: %w", file.Path, err)
	}

	// wait for diagnostics
	time.Sleep(waitFor)

	diagnostics, err := s.lspService.GetDiagnostics(ctx, file)
	if err != nil {
		var lspLanguageServerNotFoundError *lsp.LanguageServerNotFoundError
		if errors.As(err, &lspLanguageServerNotFoundError) {
			return nil, nil
		}

		return nil, fmt.Errorf("Failed to get diagnostics for file %s: %w", file.Path, err)
	}

	if err := s.lspService.NotifyDidClose(ctx, file); err != nil {
		var lspLanguageServerNotFoundError *lsp.LanguageServerNotFoundError
		if errors.As(err, &lspLanguageServerNotFoundError) {
			return nil, nil
		}

		return nil, fmt.Errorf("Failed to notify didClose while reading file %s: %w", file.Path, err)
	}

	return diagnostics, nil
}
