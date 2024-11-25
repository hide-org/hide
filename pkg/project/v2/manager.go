package project

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hide-org/hide/pkg/devcontainer/v2"
	lang "github.com/hide-org/hide/pkg/lsp/v2/languages"
	"github.com/hide-org/hide/pkg/model"
	"github.com/hide-org/hide/pkg/server"
	"github.com/hide-org/hide/pkg/workspaces"

	"github.com/rs/zerolog/log"
)

const HideBin = "/go/bin/hide_linux_arm64"

type Repository struct {
	Url    string  `json:"url" validate:"required,url"`
	Commit *string `json:"commit,omitempty"`
}

type CreateProjectRequest struct {
	Repository   Repository           `json:"repository" validate:"required"`
	DevContainer *devcontainer.Config `json:"devcontainer,omitempty"`
	Languages    []lang.LanguageID    `json:"languages,omitempty" validate:"dive,oneof=Go JavaScript Python TypeScript"`
}

type Manager interface {
	Cleanup(ctx context.Context) error
	CreateProject(ctx context.Context, request CreateProjectRequest) (*model.Project, error)
	DeleteProject(ctx context.Context, projectId model.ProjectId) error
	GetProject(ctx context.Context, projectId model.ProjectId) (model.Project, error)
	GetProjects(ctx context.Context) ([]*model.Project, error)
}

type ManagerImpl struct {
	servers    server.Installer
	store      Store
	workspaces workspaces.Service
}

func NewProjectManager(
	servers server.Installer,
	store Store,
	workspaces workspaces.Service,
) Manager {
	return &ManagerImpl{
		servers:    servers,
		store:      store,
		workspaces: workspaces,
	}
}

func (pm ManagerImpl) CreateProject(ctx context.Context, request CreateProjectRequest) (*model.Project, error) {
	log.Debug().Msgf("Creating project for repo %s", request.Repository.Url)

	// Create workspace
	workspace, err := pm.workspaces.Create(ctx, request.Repository.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// Wait until SSH server is ready
	log.Debug().Msg("Waiting for SSH server to be ready")
	time.Sleep(10 * time.Second)
	count := 0
	for {
		ready, err := workspace.SSHIsReady(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to check if SSH server is ready: %w", err)
		}
		if ready {
			log.Debug().Msg("SSH server is ready")
			break
		}
		log.Debug().Msg("SSH server is not ready yet")
		count++
		if count > 5 {
			return nil, fmt.Errorf("SSH server is not ready after 5 attempts")
		}
		log.Debug().Msgf("Waiting for SSH server to be ready (%d/%d)", count, 5)
		time.Sleep(1 * time.Second)
	}

	// Start dev server
	log.Debug().Msg("Installing dev server")
	port, err := pm.servers.Install(ctx, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to start dev server: %w", err)
	}

	log.Debug().Str("port", fmt.Sprintf("%d", port)).Msg("Installed dev server")

	project := model.Project{
		ID:          workspace.ID(),
		ContainerId: workspace.ID(),
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
		return fmt.Errorf("failed to get project with id %s: %w", projectId, err)
	}

	workspace, err := pm.workspaces.Get(ctx, project.ContainerId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get workspace %s", project.ContainerId)
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	if err := workspace.Stop(ctx); err != nil {
		log.Error().Err(err).Msgf("Failed to stop workspace %s", project.ContainerId)
		return fmt.Errorf("failed to stop workspace: %w", err)
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
