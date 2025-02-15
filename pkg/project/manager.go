package project

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sync"
	"time"

	"github.com/hide-org/hide/pkg/devcontainer"
	"github.com/hide-org/hide/pkg/files"
	"github.com/hide-org/hide/pkg/git"
	"github.com/hide-org/hide/pkg/lsp"
	"github.com/hide-org/hide/pkg/model"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/rs/zerolog/log"

	"github.com/spf13/afero"
)

const MaxDiagnosticsDelay = time.Second * 1

type CreateProjectRequest struct {
	Repository   model.Repository     `json:"repository" validate:"required"`
	DevContainer *devcontainer.Config `json:"devcontainer,omitempty"`
	Languages    []lsp.LanguageId     `json:"languages,omitempty" validate:"dive,oneof=Go JavaScript Python TypeScript"`
}

type TaskResult struct {
	StdOut   string `json:"stdout"`
	StdErr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

type Manager interface {
	// Deprecated: use files.Service instead
	ApplyPatch(ctx context.Context, projectId, path, patch string) (*model.File, error)
	Cleanup(ctx context.Context) error
	// Deprecated: use files.Service instead
	CreateFile(ctx context.Context, projectId, path, content string) (*model.File, error)
	CreateProject(ctx context.Context, request CreateProjectRequest) (*model.Project, error)
	// Deprecated: use tasks.Service instead
	CreateTask(ctx context.Context, projectId model.ProjectId, command string) (TaskResult, error)
	// Deprecated: use files.Service instead
	DeleteFile(ctx context.Context, projectId, path string) error
	DeleteProject(ctx context.Context, projectId model.ProjectId) error
	GetProject(ctx context.Context, projectId model.ProjectId) (model.Project, error)
	GetProjects(ctx context.Context) ([]*model.Project, error)
	// Deprecated: use files.Service instead
	ListFiles(ctx context.Context, projectId string, opts ...files.ListFileOption) (model.Files, error)
	// Deprecated: use files.Service instead
	ReadFile(ctx context.Context, projectId, path string) (*model.File, error)
	// Deprecated: use tasks.Service instead
	ResolveTaskAlias(ctx context.Context, projectId model.ProjectId, alias string) (devcontainer.Task, error)
	// Deprecated: use symbol.Service instead
	SearchSymbols(ctx context.Context, projectId model.ProjectId, query string, symbolFilter lsp.SymbolFilter) ([]lsp.SymbolInfo, error)
	// Deprecated: use files.Service instead
	UpdateFile(ctx context.Context, projectId, path, content string) (*model.File, error)
	// Deprecated: use files.Service instead
	UpdateLines(ctx context.Context, projectId, path string, lineDiff files.LineDiffChunk) (*model.File, error)
	// Deprecated: use outline.Service instead
	DocumentOutline(ctx context.Context, projectId, path string) (lsp.DocumentOutline, error)
}

type ManagerImpl struct {
	devContainerRunner devcontainer.Runner
	store              Store
	projectsRoot       string
	fileManager        files.FileManager
	lspService         lsp.Service
	languageDetector   lsp.LanguageDetector
	randomString       func(int) string
	git                git.Client
}

func NewProjectManager(
	devContainerRunner devcontainer.Runner,
	projectStore Store,
	projectsRoot string,
	fileManager files.FileManager,
	lspService lsp.Service,
	languageDetector lsp.LanguageDetector,
	randomString func(int) string,
	git git.Client,
) Manager {
	return ManagerImpl{
		devContainerRunner: devContainerRunner,
		store:              projectStore,
		projectsRoot:       projectsRoot,
		fileManager:        fileManager,
		lspService:         lspService,
		languageDetector:   languageDetector,
		randomString:       randomString,
		git:                git,
	}
}

func (pm ManagerImpl) CreateProject(ctx context.Context, request CreateProjectRequest) (*model.Project, error) {
	log.Debug().Msgf("Creating project for repo %s", request.Repository.Url)

	projectId := pm.randomString(10)
	projectPath := path.Join(pm.projectsRoot, projectId)

	// Clone git repo
	if err := pm.createProjectDir(projectPath); err != nil {
		log.Error().Err(err).Msg("Failed to create project directory")
		return nil, fmt.Errorf("Failed to create project directory: %w", err)
	}

	r, err := pm.git.Clone(request.Repository.Url, projectPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to clone git repo")
		removeProjectDir(projectPath)
		return nil, fmt.Errorf("Failed to clone git repo: %w", err)
	}

	if request.Repository.Commit != nil {
		if err := pm.git.Checkout(*r, *request.Repository.Commit); err != nil {
			log.Error().Err(err).Msg("Failed to checkout commit")
			removeProjectDir(projectPath)
			return nil, fmt.Errorf("Failed to checkout commit %s: %w", *request.Repository.Commit, err)
		}
	}

	// Start devcontainer
	var devContainerConfig devcontainer.Config

	if request.DevContainer != nil {
		devContainerConfig = *request.DevContainer
	} else {
		config, err := pm.configFromProject(os.DirFS(projectPath))
		if err != nil {
			log.Error().Err(err).Msgf("Failed to get devcontainer config from repository %s", request.Repository.Url)
			removeProjectDir(projectPath)
			return nil, fmt.Errorf("Failed to read devcontainer.json: %w", err)
		}

		devContainerConfig = config
	}

	containerId, err := pm.devContainerRunner.Run(ctx, projectPath, devContainerConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to launch devcontainer")
		removeProjectDir(projectPath)
		return nil, fmt.Errorf("Failed to launch devcontainer: %w", err)
	}

	project := model.NewProject(projectId, projectPath, model.Config{DevContainerConfig: devContainerConfig}, containerId, request.Repository)

	languages := request.Languages
	if len(languages) == 0 {
		languages, err = pm.detectLanguages(project)
		if err != nil {
			log.Error().Err(err).Msg("Failed to detect project languages")
			removeProjectDir(projectPath)
			return nil, fmt.Errorf("Failed to detect project languages: %w", err)
		}
	}

	for _, language := range languages {
		if err := pm.lspService.StartServer(model.NewContextWithProject(context.Background(), &project), language); err != nil {
			log.Warn().Err(err).Msg("Failed to start LSP server. Diagnostics will not be available.")
		}
	}

	// Save project in store
	if err := pm.store.CreateProject(&project); err != nil {
		log.Error().Err(err).Msg("Failed to save project")
		removeProjectDir(projectPath)
		return nil, fmt.Errorf("Failed to save project: %w", err)
	}

	log.Debug().Msgf("Created project %s for repo %s", projectId, request.Repository.Url)

	return &project, nil
}

func (pm ManagerImpl) GetProject(ctx context.Context, projectId string) (model.Project, error) {
	project, err := pm.store.GetProject(projectId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get project with id %s", projectId)
		return model.Project{}, err
	}

	return *project, nil
}

func (pm ManagerImpl) GetProjects(ctx context.Context) ([]*model.Project, error) {
	projects, err := pm.store.GetProjects()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get projects")
		return nil, fmt.Errorf("Failed to get projects: %w", err)
	}

	return projects, nil
}

func (pm ManagerImpl) DeleteProject(ctx context.Context, projectId string) error {
	log.Debug().Msgf("Deleting project %s", projectId)

	project, err := pm.GetProject(ctx, projectId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get project with id %s", projectId)
		return fmt.Errorf("Failed to get project with id %s: %w", projectId, err)
	}

	if err := pm.devContainerRunner.Stop(ctx, project.ContainerId); err != nil {
		log.Error().Err(err).Msgf("Failed to stop container %s", project.ContainerId)
		return fmt.Errorf("Failed to stop container: %w", err)
	}

	if err := pm.lspService.CleanupProject(ctx, projectId); err != nil {
		log.Error().Err(err).Str("projectId", projectId).Msg("Failed to stop LSP server(s)")
		return fmt.Errorf("Failed to stop LSP server(s): %w", err)
	}

	if err := pm.store.DeleteProject(projectId); err != nil {
		log.Error().Err(err).Msgf("Failed to delete project %s", projectId)
		return fmt.Errorf("Failed to delete project: %w", err)
	}

	log.Debug().Msgf("Deleted project %s", projectId)

	return nil
}

func (pm ManagerImpl) ResolveTaskAlias(ctx context.Context, projectId string, alias string) (devcontainer.Task, error) {
	log.Debug().Msgf("Resolving task alias %s for project %s", alias, projectId)

	project, err := pm.GetProject(ctx, projectId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get project with id %s", projectId)
		return devcontainer.Task{}, fmt.Errorf("Failed to get project with id %s: %w", projectId, err)
	}

	task, err := project.FindTaskByAlias(alias)
	if err != nil {
		log.Error().Err(err).Msgf("Task with alias %s for project %s not found", alias, projectId)
		return devcontainer.Task{}, err
	}

	log.Debug().Msgf("Resolved task alias %s for project %s: %+v", alias, projectId, task)

	return task, nil
}

func (pm ManagerImpl) CreateTask(ctx context.Context, projectId string, command string) (TaskResult, error) {
	log.Debug().Msgf("Creating task for project %s. Command: %s", projectId, command)

	project, err := pm.GetProject(ctx, projectId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get project with id %s", projectId)
		return TaskResult{}, fmt.Errorf("Failed to get project with id %s: %w", projectId, err)
	}

	execResult, err := pm.devContainerRunner.Exec(ctx, project.ContainerId, cmdMaybeWithTimeout(ctx, command))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to execute command '%s' in container %s", command, project.ContainerId)
		return TaskResult{}, fmt.Errorf("Failed to execute command: %w", err)
	}

	log.Debug().Msgf("Task '%s' for project %s completed", command, projectId)

	return TaskResult{StdOut: execResult.StdOut, StdErr: execResult.StdErr, ExitCode: execResult.ExitCode}, nil
}

// cmdMaybeWithTimeout prepends timeout command to the command.
//
// Note: this is a workaround to ensure that the process is actually stopped after the timeout duration exceeded.
func cmdMaybeWithTimeout(ctx context.Context, cmd string) (command []string) {
	deadline, ok := ctx.Deadline()
	if !ok {
		return []string{"/bin/bash", "-c", cmd}
	}

	duration := time.Until(deadline)
	return []string{"timeout", "--kill-after=1s", "--verbose", fmt.Sprintf("%fs", duration.Seconds()), "/bin/bash", "-c", cmd}
}

func (pm ManagerImpl) Cleanup(ctx context.Context) error {
	log.Info().Msg("Cleaning up projects")

	projects, err := pm.GetProjects(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get projects")
		return fmt.Errorf("Failed to get projects: %w", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(projects))

	for _, project := range projects {
		wg.Add(1)
		go func(p *model.Project) {
			defer wg.Done()
			if err := pm.DeleteProject(ctx, p.ID); err != nil {
				errChan <- err
			}
			return
		}(project)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("Errors occurred during cleanup: %v", errs)
	}

	log.Info().Msg("Cleaned up projects")
	return nil
}

func (pm ManagerImpl) CreateFile(ctx context.Context, projectId, path, content string) (*model.File, error) {
	log.Debug().Str("projectId", projectId).Str("path", path).Msg("Creating file")

	project, err := pm.GetProject(ctx, projectId)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Msg("Failed to get project")
		return nil, fmt.Errorf("Failed to get project with id %s: %w", projectId, err)
	}

	ctx = model.NewContextWithProject(ctx, &project)

	file, err := pm.fileManager.CreateFile(ctx, afero.NewBasePathFs(afero.NewOsFs(), project.Path), path, content)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Str("path", path).Msg("Failed to create file")
		return file, err
	}

	if diagnostics, err := pm.getDiagnostics(ctx, *file, MaxDiagnosticsDelay); err != nil {
		log.Warn().Err(err).Str("projectId", projectId).Str("path", path).Msg("Failed to get diagnostics")
	} else {
		file.Diagnostics = diagnostics
	}

	return file, nil
}

func (pm ManagerImpl) ReadFile(ctx context.Context, projectId, path string) (*model.File, error) {
	log.Debug().Str("projectId", projectId).Str("path", path).Msg("Reading file")

	project, err := pm.GetProject(ctx, projectId)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Msg("Failed to get project")
		return nil, fmt.Errorf("Failed to get project with id %s: %w", projectId, err)
	}

	ctx = model.NewContextWithProject(ctx, &project)
	file, err := pm.fileManager.ReadFile(ctx, afero.NewBasePathFs(afero.NewOsFs(), project.Path), path)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Str("path", path).Msg("Failed to read file")
		return file, err
	}

	if diagnostics, err := pm.getDiagnostics(ctx, *file, MaxDiagnosticsDelay); err != nil {
		log.Warn().Err(err).Str("projectId", projectId).Str("path", path).Msg("Failed to get diagnostics")
	} else {
		file.Diagnostics = diagnostics
	}

	return file, nil
}

func (pm ManagerImpl) UpdateFile(ctx context.Context, projectId, path, content string) (*model.File, error) {
	log.Debug().Msgf("Updating file %s in project %s", path, projectId)

	project, err := pm.GetProject(ctx, projectId)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Msg("Failed to get project")
		return nil, fmt.Errorf("Failed to get project with id %s: %w", projectId, err)
	}

	ctx = model.NewContextWithProject(ctx, &project)

	file, err := pm.fileManager.UpdateFile(ctx, afero.NewBasePathFs(afero.NewOsFs(), project.Path), path, content)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Str("path", path).Msg("Failed to update file")
		return file, err
	}

	if diagnostics, err := pm.getDiagnostics(ctx, *file, MaxDiagnosticsDelay); err != nil {
		log.Warn().Err(err).Str("projectId", projectId).Str("path", path).Msg("Failed to get diagnostics")
	} else {
		file.Diagnostics = diagnostics
	}

	return file, nil
}

func (pm ManagerImpl) DeleteFile(ctx context.Context, projectId, path string) error {
	log.Debug().Msgf("Deleting file %s in project %s", path, projectId)

	project, err := pm.GetProject(ctx, projectId)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Msg("Failed to get project")
		return fmt.Errorf("Failed to get project with id %s: %w", projectId, err)
	}

	return pm.fileManager.DeleteFile(model.NewContextWithProject(ctx, &project), afero.NewBasePathFs(afero.NewOsFs(), project.Path), path)
}

func (pm ManagerImpl) ListFiles(ctx context.Context, projectId string, opts ...files.ListFileOption) (model.Files, error) {
	log.Debug().Str("projectId", projectId).Msg("Listing files")

	project, err := pm.GetProject(ctx, projectId)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Msg("Failed to get project")
		return nil, fmt.Errorf("Failed to get project with id %s: %w", projectId, err)
	}

	ctx = model.NewContextWithProject(ctx, &project)

	files, err := pm.fileManager.ListFiles(ctx, afero.NewBasePathFs(afero.NewOsFs(), project.Path), opts...)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Msgf("Failed to list files")
		return nil, fmt.Errorf("Failed to list files in project %s: %w", projectId, err)
	}

	return files, nil
}

func (pm ManagerImpl) ApplyPatch(ctx context.Context, projectId, path, patch string) (*model.File, error) {
	log.Debug().Str("projectId", projectId).Str("path", path).Msg("Patching file")

	project, err := pm.GetProject(ctx, projectId)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Msg("Failed to get project")
		return nil, fmt.Errorf("Failed to get project with id %s: %w", projectId, err)
	}

	ctx = model.NewContextWithProject(ctx, &project)
	file, err := pm.fileManager.ApplyPatch(ctx, afero.NewBasePathFs(afero.NewOsFs(), project.Path), path, patch)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Str("path", path).Msg("Failed to patch file")
		return nil, fmt.Errorf("Failed to patch file %s: %w", path, err)
	}

	if diagnostics, err := pm.getDiagnostics(ctx, *file, MaxDiagnosticsDelay); err != nil {
		log.Warn().Err(err).Str("projectId", projectId).Str("path", path).Msg("Failed to get diagnostics")
	} else {
		file.Diagnostics = diagnostics
	}

	return file, nil
}

func (pm ManagerImpl) UpdateLines(ctx context.Context, projectId, path string, lineDiff files.LineDiffChunk) (*model.File, error) {
	log.Debug().Str("projectId", projectId).Str("path", path).Msg("Replacing lines in file")

	project, err := pm.GetProject(ctx, projectId)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Msg("Failed to get project")
		return nil, fmt.Errorf("Failed to get project with id %s: %w", projectId, err)
	}

	ctx = model.NewContextWithProject(ctx, &project)
	file, err := pm.fileManager.UpdateLines(ctx, afero.NewBasePathFs(afero.NewOsFs(), project.Path), path, lineDiff)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Str("path", path).Msg("Failed to replace lines in file")
		return nil, fmt.Errorf("Failed to replace lines in file %s: %w", path, err)
	}

	if diagnostics, err := pm.getDiagnostics(ctx, *file, MaxDiagnosticsDelay); err != nil {
		log.Warn().Err(err).Str("projectId", projectId).Str("path", path).Msg("Failed to get diagnostics")
	} else {
		file.Diagnostics = diagnostics
	}

	return file, nil
}

func (pm ManagerImpl) SearchSymbols(ctx context.Context, projectId model.ProjectId, query string, symbolFilter lsp.SymbolFilter) ([]lsp.SymbolInfo, error) {
	log.Debug().Str("projectId", projectId).Str("query", query).Msg("Searching symbols")

	project, err := pm.GetProject(ctx, projectId)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Msg("Failed to get project")
		return nil, fmt.Errorf("failed to get project with id %s: %w", projectId, err)
	}

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled")
	default:
	}

	symbols, err := pm.lspService.GetWorkspaceSymbols(model.NewContextWithProject(ctx, &project), query, symbolFilter)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Msg("failed to get workspace symbols")
		return nil, fmt.Errorf("failed to get workspace symbols: %w", err)
	}

	log.Debug().Str("projectId", projectId).Str("query", query).Msgf("Found %d symbols", len(symbols))
	return symbols, nil
}

func (pm ManagerImpl) DocumentOutline(ctx context.Context, projectId, path string) (lsp.DocumentOutline, error) {
	log.Debug().Str("projectId", projectId).Str("path", path).Msg("Creating outline")

	project, err := pm.GetProject(ctx, projectId)
	if err != nil {
		log.Error().Err(err).Str("projectId", projectId).Msg("Failed to get project")
		return lsp.DocumentOutline{}, fmt.Errorf("failed to get project with id %s: %w", projectId, err)
	}

	return pm.lspService.GetDocumentOutline(model.NewContextWithProject(ctx, &project), model.File{Path: path})
}

func (pm ManagerImpl) createProjectDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("Failed to create project directory: %w", err)
	}

	log.Debug().Msgf("Created project directory: %s", path)

	return nil
}

func (pm ManagerImpl) configFromProject(fileSystem fs.FS) (devcontainer.Config, error) {
	configFile, err := devcontainer.FindConfig(fileSystem)
	if err != nil {
		return devcontainer.Config{}, fmt.Errorf("Failed to find devcontainer.json: %w", err)
	}

	config, err := devcontainer.ParseConfig(configFile)
	if err != nil {
		return devcontainer.Config{}, fmt.Errorf("Failed to parse devcontainer.json: %w", err)
	}

	return *config, nil
}

func (pm ManagerImpl) getDiagnostics(ctx context.Context, file model.File, waitFor time.Duration) ([]protocol.Diagnostic, error) {
	if err := pm.lspService.NotifyDidOpen(ctx, file); err != nil {
		var lspLanguageServerNotFoundError *lsp.LanguageServerNotFoundError
		if errors.As(err, &lspLanguageServerNotFoundError) {
			return nil, nil
		}

		return nil, fmt.Errorf("Failed to notify didOpen while reading file %s: %w", file.Path, err)
	}

	// wait for diagnostics
	time.Sleep(waitFor)

	diagnostics, err := pm.lspService.GetDiagnostics(ctx, file)
	if err != nil {
		var lspLanguageServerNotFoundError *lsp.LanguageServerNotFoundError
		if errors.As(err, &lspLanguageServerNotFoundError) {
			return nil, nil
		}

		return nil, fmt.Errorf("Failed to get diagnostics for file %s: %w", file.Path, err)
	}

	if err := pm.lspService.NotifyDidClose(ctx, file); err != nil {
		var lspLanguageServerNotFoundError *lsp.LanguageServerNotFoundError
		if errors.As(err, &lspLanguageServerNotFoundError) {
			return nil, nil
		}

		return nil, fmt.Errorf("Failed to notify didClose while reading file %s: %w", file.Path, err)
	}

	return diagnostics, nil
}

func (pm ManagerImpl) detectLanguages(project model.Project) ([]lsp.LanguageId, error) {
	files, err := pm.fileManager.ListFiles(model.NewContextWithProject(context.Background(), &project), afero.NewBasePathFs(afero.NewOsFs(), project.Path), files.ListFilesWithContent())
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// TODO: handle multiple main language
	language := pm.languageDetector.DetectMainLanguage(files)
	log.Debug().Msgf("Detected main language %s for project %s", language, project.ID)
	return []lsp.LanguageId{language}, nil
}

func removeProjectDir(projectPath string) {
	if err := os.RemoveAll(projectPath); err != nil {
		log.Error().Err(err).Msgf("Failed to remove project directory %s", projectPath)
		return
	}

	log.Debug().Msgf("Removed project directory: %s", projectPath)

	return
}
