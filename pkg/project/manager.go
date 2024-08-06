package project

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/files"
	"github.com/artmoskvin/hide/pkg/languageserver"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/artmoskvin/hide/pkg/result"
	"github.com/artmoskvin/hide/pkg/util"

	"github.com/rs/zerolog/log"

	"github.com/spf13/afero"
)

const MaxDiagnosticsDelay = time.Second * 1

type Repository struct {
	Url    string  `json:"url"`
	Commit *string `json:"commit,omitempty"`
}

type CreateProjectRequest struct {
	Repository   Repository           `json:"repository"`
	DevContainer *devcontainer.Config `json:"devcontainer,omitempty"`
}

type TaskResult struct {
	StdOut   string `json:"stdOut"`
	StdErr   string `json:"stdErr"`
	ExitCode int    `json:"exitCode"`
}

type Manager interface {
	CreateProject(request CreateProjectRequest) <-chan result.Result[model.Project]
	GetProject(projectId model.ProjectId) (model.Project, error)
	GetProjects() ([]*model.Project, error)
	DeleteProject(projectId model.ProjectId) <-chan result.Empty
	ResolveTaskAlias(projectId model.ProjectId, alias string) (devcontainer.Task, error)
	CreateTask(projectId model.ProjectId, command string) (TaskResult, error)
	Cleanup() error
	CreateFile(ctx context.Context, projectId, path, content string) (model.File, error)
	ReadFile(ctx context.Context, projectId, path string, props files.ReadProps) (model.File, error)
	UpdateFile(ctx context.Context, projectId, path, content string) (model.File, error)
}

type ManagerImpl struct {
	DevContainerRunner devcontainer.Runner
	Store              Store
	ProjectsRoot       string
	fileManager        files.FileManager
	lspService         languageserver.Service
}

func NewProjectManager(devContainerRunner devcontainer.Runner, projectStore Store, projectsRoot string, fileManager files.FileManager, lspService languageserver.Service) Manager {
	return ManagerImpl{DevContainerRunner: devContainerRunner, Store: projectStore, ProjectsRoot: projectsRoot, fileManager: fileManager, lspService: lspService}
}

func (pm ManagerImpl) CreateProject(request CreateProjectRequest) <-chan result.Result[model.Project] {
	c := make(chan result.Result[model.Project])

	go func() {
		log.Debug().Msgf("Creating project for repo %s", request.Repository.Url)

		projectId := util.RandomString(10)
		projectPath := path.Join(pm.ProjectsRoot, projectId)

		if err := pm.createProjectDir(projectPath); err != nil {
			log.Error().Err(err).Msg("Failed to create project directory")
			c <- result.Failure[model.Project](fmt.Errorf("Failed to create project directory: %w", err))
			return
		}

		if r := <-cloneGitRepo(request.Repository, projectPath); r.IsFailure() {
			log.Error().Err(r.Error).Msg("Failed to clone git repo")
			removeProjectDir(projectPath)
			c <- result.Failure[model.Project](fmt.Errorf("Failed to clone git repo: %w", r.Error))
			return
		}

		var devContainerConfig devcontainer.Config

		if request.DevContainer != nil {
			devContainerConfig = *request.DevContainer
		} else {
			config, err := pm.configFromProject(os.DirFS(projectPath))

			if err != nil {
				log.Error().Err(err).Msgf("Failed to get devcontainer config from repository %s", request.Repository.Url)
				removeProjectDir(projectPath)
				c <- result.Failure[model.Project](fmt.Errorf("Failed to read devcontainer.json: %w", err))
				return
			}

			devContainerConfig = config
		}

		containerId, err := pm.DevContainerRunner.Run(projectPath, devContainerConfig)

		if err != nil {
			log.Error().Err(err).Msg("Failed to launch devcontainer")
			removeProjectDir(projectPath)
			c <- result.Failure[model.Project](fmt.Errorf("Failed to launch devcontainer: %w", err))
			return
		}

		project := model.Project{Id: projectId, Path: projectPath, Config: model.Config{DevContainerConfig: devContainerConfig}, ContainerId: containerId}

		if err := pm.Store.CreateProject(&project); err != nil {
			log.Error().Err(err).Msg("Failed to save project")
			removeProjectDir(projectPath)
			c <- result.Failure[model.Project](fmt.Errorf("Failed to save project: %w", err))
			return
		}

		log.Debug().Msgf("Created project %s for repo %s", projectId, request.Repository.Url)

		c <- result.Success(project)
	}()

	return c
}

func (pm ManagerImpl) GetProject(projectId string) (model.Project, error) {
	project, err := pm.Store.GetProject(projectId)

	if err != nil {
		log.Error().Err(err).Msgf("Project with id %s not found", projectId)
		return model.Project{}, fmt.Errorf("Project with id %s not found", projectId)
	}

	return *project, nil
}

func (pm ManagerImpl) GetProjects() ([]*model.Project, error) {
	projects, err := pm.Store.GetProjects()

	if err != nil {
		log.Error().Err(err).Msg("Failed to get projects")
		return nil, fmt.Errorf("Failed to get projects: %w", err)
	}

	return projects, nil
}

func (pm ManagerImpl) DeleteProject(projectId string) <-chan result.Empty {
	c := make(chan result.Empty)

	go func() {
		log.Debug().Msgf("Deleting project %s", projectId)

		project, err := pm.GetProject(projectId)

		if err != nil {
			log.Error().Err(err).Msgf("Project with id %s not found", projectId)
			c <- result.EmptyFailure(fmt.Errorf("Project with id %s not found", projectId))
			return
		}

		if err := pm.DevContainerRunner.Stop(project.ContainerId); err != nil {
			log.Error().Err(err).Msgf("Failed to stop container %s", project.ContainerId)
			c <- result.EmptyFailure(fmt.Errorf("Failed to stop container: %w", err))
			return
		}

		if err := pm.Store.DeleteProject(projectId); err != nil {
			log.Error().Err(err).Msgf("Failed to delete project %s", projectId)
			c <- result.EmptyFailure(fmt.Errorf("Failed to delete project: %w", err))
			return
		}

		log.Debug().Msgf("Deleted project %s", projectId)

		c <- result.EmptySuccess()
	}()

	return c
}

func (pm ManagerImpl) ResolveTaskAlias(projectId string, alias string) (devcontainer.Task, error) {
	log.Debug().Msgf("Resolving task alias %s for project %s", alias, projectId)

	project, err := pm.GetProject(projectId)

	if err != nil {
		log.Error().Err(err).Msgf("Project with id %s not found", projectId)
		return devcontainer.Task{}, fmt.Errorf("Project with id %s not found", projectId)
	}

	task, err := project.FindTaskByAlias(alias)

	if err != nil {
		log.Error().Err(err).Msgf("Task with alias %s for project %s not found", alias, projectId)
		return devcontainer.Task{}, fmt.Errorf("Task with alias %s not found", alias)
	}

	log.Debug().Msgf("Resolved task alias %s for project %s: %+v", alias, projectId, task)

	return task, nil
}

func (pm ManagerImpl) CreateTask(projectId string, command string) (TaskResult, error) {
	log.Debug().Msgf("Creating task for project %s. Command: %s", projectId, command)

	project, err := pm.GetProject(projectId)

	if err != nil {
		log.Error().Err(err).Msgf("Project with id %s not found", projectId)
		return TaskResult{}, fmt.Errorf("Project with id %s not found", projectId)
	}

	execResult, err := pm.DevContainerRunner.Exec(project.ContainerId, []string{"/bin/bash", "-c", command})

	if err != nil {
		log.Error().Err(err).Msgf("Failed to execute command '%s' in container %s", command, project.ContainerId)
		return TaskResult{}, fmt.Errorf("Failed to execute command: %w", err)
	}

	log.Debug().Msgf("Task '%s' for project %s completed", command, projectId)

	return TaskResult{StdOut: execResult.StdOut, StdErr: execResult.StdErr, ExitCode: execResult.ExitCode}, nil
}

func (pm ManagerImpl) Cleanup() error {
	log.Info().Msg("Cleaning up projects")

	projects, err := pm.GetProjects()

	if err != nil {
		log.Error().Err(err).Msg("Failed to get projects")
		return fmt.Errorf("Failed to get projects: %w", err)
	}

	for _, project := range projects {
		log.Debug().Msgf("Cleaning up project %s", project.Id)
		pm.DevContainerRunner.Stop(project.ContainerId)
	}

	log.Info().Msg("Cleaned up projects")

	return nil
}

func (pm ManagerImpl) CreateFile(ctx context.Context, projectId, path, content string) (model.File, error) {
	log.Debug().Msgf("Creating file %s in project %s", path, projectId)

	project, err := pm.GetProject(projectId)

	if err != nil {
		log.Error().Err(err).Msgf("Project with id %s not found", projectId)
		return model.File{}, fmt.Errorf("Project with id %s not found", projectId)
	}

	return pm.fileManager.CreateFile(model.NewContextWithProject(ctx, &project), afero.NewBasePathFs(afero.NewOsFs(), project.Path), path, content)
}

func (pm ManagerImpl) ReadFile(ctx context.Context, projectId, path string, props files.ReadProps) (model.File, error) {
	project, err := pm.GetProject(projectId)

	if err != nil {
		return model.File{}, fmt.Errorf("Project with id %s not found", projectId)
	}

	return pm.fileManager.ReadFile(model.NewContextWithProject(ctx, &project), afero.NewBasePathFs(afero.NewOsFs(), project.Path), path, props)
}

func (pm ManagerImpl) UpdateFile(ctx context.Context, projectId, path, content string) (model.File, error) {
	project, err := pm.GetProject(projectId)

	if err != nil {
		return model.File{}, fmt.Errorf("Project with id %s not found", projectId)
	}

	return pm.fileManager.UpdateFile(model.NewContextWithProject(ctx, &project), afero.NewBasePathFs(afero.NewOsFs(), project.Path), path, content)
}

func (pm ManagerImpl) createProjectDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
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

func removeProjectDir(projectPath string) {
	if err := os.RemoveAll(projectPath); err != nil {
		log.Error().Err(err).Msgf("Failed to remove project directory %s", projectPath)
		return
	}

	log.Debug().Msgf("Removed project directory: %s", projectPath)

	return
}

func cloneGitRepo(repository Repository, projectPath string) <-chan result.Empty {
	c := make(chan result.Empty)

	go func() {
		cmd := exec.Command("git", "clone", repository.Url, projectPath)
		cmdOut, err := cmd.Output()

		if err != nil {
			c <- result.EmptyFailure(fmt.Errorf("Failed to clone git repo: %w", err))
			return
		}

		log.Debug().Msgf("Cloned git repo %s to %s", repository.Url, projectPath)
		log.Debug().Msg(string(cmdOut))

		if repository.Commit != nil {
			cmd = exec.Command("git", "checkout", *repository.Commit)
			cmd.Dir = projectPath
			cmdOut, err = cmd.Output()

			if err != nil {
				c <- result.EmptyFailure(fmt.Errorf("Failed to checkout commit %s: %w", *repository.Commit, err))
				return
			}

			log.Debug().Msgf("Checked out commit %s", *repository.Commit)
			log.Debug().Msg(string(cmdOut))
		}

		c <- result.EmptySuccess()
	}()

	return c
}
