package project

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"sync"

	"github.com/hide-org/hide/pkg/devcontainer/v2"
	"github.com/hide-org/hide/pkg/git"
	"github.com/hide-org/hide/pkg/lsp/v2"
	"github.com/hide-org/hide/pkg/model"
	"github.com/hide-org/hide/pkg/server"

	"github.com/rs/zerolog/log"
	// "github.com/spf13/afero"
)

const HideBin = "/go/bin/hide_linux_arm64"

type Repository struct {
	Url    string  `json:"url" validate:"required,url"`
	Commit *string `json:"commit,omitempty"`
}

type CreateProjectRequest struct {
	Repository   Repository           `json:"repository" validate:"required"`
	DevContainer *devcontainer.Config `json:"devcontainer,omitempty"`
	Languages    []lsp.LanguageId     `json:"languages,omitempty" validate:"dive,oneof=Go JavaScript Python TypeScript"`
}

type Manager interface {
	Cleanup(ctx context.Context) error
	CreateProject(ctx context.Context, request CreateProjectRequest) (*model.Project, error)
	DeleteProject(ctx context.Context, projectId model.ProjectId) error
	GetProject(ctx context.Context, projectId model.ProjectId) (model.Project, error)
	GetProjects(ctx context.Context) ([]*model.Project, error)
}

type ManagerImpl struct {
	devcontainers devcontainer.Service
	servers       server.Service
	store         Store
	git           git.Client
}

func NewProjectManager(
	devcontainers devcontainer.Service,
	servers server.Service,
	store Store,
	git git.Client,
) Manager {
	return &ManagerImpl{
		devcontainers: devcontainers,
		servers:       servers,
		store:         store,
		git:           git,
	}
}

func (pm ManagerImpl) CreateProject(ctx context.Context, request CreateProjectRequest) (*model.Project, error) {
	log.Debug().Msgf("Creating project for repo %s", request.Repository.Url)

	// Create and start container
	container, err := pm.devcontainers.Create(ctx, request.Repository.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to create dev container: %w", err)
	}

	// Start dev server
	if err := pm.servers.Start(ctx, container); err != nil {
		return nil, fmt.Errorf("failed to start dev server: %w", err)
	}

	project := model.Project{
		ID:          container.ID(),
		ContainerId: container.ID(),
	}

	// Save project in store
	if err := pm.store.CreateProject(&project); err != nil {
		return nil, fmt.Errorf("failed to save project: %w", err)
	}

	log.Debug().Msgf("Created project %s for repo %s", project.ID, request.Repository.Url)

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

	container, err := pm.devcontainers.Get(ctx, project.ContainerId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get container %s", project.ContainerId)
		return fmt.Errorf("Failed to get container: %w", err)
	}

	if err := container.Stop(ctx); err != nil {
		log.Error().Err(err).Msgf("Failed to stop container %s", project.ContainerId)
		return fmt.Errorf("Failed to stop container: %w", err)
	}

	if err := pm.store.DeleteProject(projectId); err != nil {
		log.Error().Err(err).Msgf("Failed to delete project %s", projectId)
		return fmt.Errorf("Failed to delete project: %w", err)
	}

	log.Debug().Msgf("Deleted project %s", projectId)

	return nil
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

// func (pm ManagerImpl) detectLanguages(project model.Project) ([]lsp.LanguageId, error) {
// 	files, err := pm.fileManager.ListFiles(model.NewContextWithProject(context.Background(), &project), afero.NewBasePathFs(afero.NewOsFs(), project.Path), files.ListFilesWithContent())
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to list files: %w", err)
// 	}
//
// 	// TODO: handle multiple main language
// 	language := pm.languageDetector.DetectMainLanguage(files)
// 	log.Debug().Msgf("Detected main language %s for project %s", language, project.Id)
// 	return []lsp.LanguageId{language}, nil
// }

func removeProjectDir(projectPath string) {
	if err := os.RemoveAll(projectPath); err != nil {
		log.Error().Err(err).Msgf("Failed to remove project directory %s", projectPath)
		return
	}

	log.Debug().Msgf("Removed project directory: %s", projectPath)

	return
}
