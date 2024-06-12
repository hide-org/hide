package project

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/result"
	"github.com/artmoskvin/hide/pkg/util"
)

type Repository struct {
	Url    string  `json:"url"`
	Commit *string `json:"commit,omitempty"`
}

type CreateProjectRequest struct {
	Repository   Repository           `json:"repository"`
	DevContainer *devcontainer.Config `json:"devcontainer,omitempty"`
}

type Config struct {
	DevContainerConfig devcontainer.Config `json:"devContainerConfig"`
}

type ProjectId = string

type Project struct {
	Id          ProjectId `json:"id"`
	Path        string    `json:"path"`
	Config      Config    `json:"config"`
	containerId string
}

func NewProject(id ProjectId, path string, config Config, containerId string) Project {
	return Project{Id: id, Path: path, Config: config, containerId: containerId}
}

func (project *Project) FindTaskByAlias(alias string) (devcontainer.Task, error) {
	if project.Config.DevContainerConfig.Customizations.Hide == nil {
		return devcontainer.Task{}, errors.New("task not found")
	}

	for _, task := range project.Config.DevContainerConfig.Customizations.Hide.Tasks {
		if task.Alias == alias {
			return task, nil
		}
	}
	return devcontainer.Task{}, errors.New("task not found")
}

func (project *Project) GetTasks() []devcontainer.Task {
	if project.Config.DevContainerConfig.Customizations.Hide == nil {
		return []devcontainer.Task{}
	}

	return project.Config.DevContainerConfig.Customizations.Hide.Tasks
}

type TaskResult struct {
	StdOut   string `json:"stdOut"`
	StdErr   string `json:"stdErr"`
	ExitCode int    `json:"exitCode"`
}

type Manager interface {
	CreateProject(request CreateProjectRequest) <-chan result.Result[Project]
	GetProject(projectId ProjectId) (Project, error)
	GetProjects() ([]*Project, error)
	ResolveTaskAlias(projectId ProjectId, alias string) (devcontainer.Task, error)
	CreateTask(projectId ProjectId, command string) (TaskResult, error)
	Cleanup() error
}

type ManagerImpl struct {
	DevContainerRunner devcontainer.Runner
	Store              Store
	ProjectsRoot       string
}

func NewProjectManager(devContainerRunner devcontainer.Runner, projectStore Store, projectsRoot string) Manager {
	return ManagerImpl{DevContainerRunner: devContainerRunner, Store: projectStore, ProjectsRoot: projectsRoot}
}

func (pm ManagerImpl) CreateProject(request CreateProjectRequest) <-chan result.Result[Project] {
	c := make(chan result.Result[Project])

	go func() {
		log.Printf("Creating project for repo %s", request.Repository.Url)

		projectId := util.RandomString(10)
		projectPath := path.Join(pm.ProjectsRoot, projectId)

		if err := pm.createProjectDir(projectPath); err != nil {
			log.Printf("Failed to create project directory: %s", err)
			c <- result.Failure[Project](fmt.Errorf("Failed to create project directory: %w", err))
			return
		}

		if r := <-cloneGitRepo(request.Repository, projectPath); r.IsFailure() {
			log.Printf("Failed to clone git repo: %s", r.Error)
			removeProjectDir(projectPath)
			c <- result.Failure[Project](fmt.Errorf("Failed to clone git repo: %w", r.Error))
			return
		}

		var devContainerConfig devcontainer.Config

		if request.DevContainer != nil {
			devContainerConfig = *request.DevContainer
		} else {
			config, err := pm.configFromProject(os.DirFS(projectPath))

			if err != nil {
				log.Printf("Failed to get devcontainer config from repository %s: %s", request.Repository.Url, err)
				removeProjectDir(projectPath)
				c <- result.Failure[Project](fmt.Errorf("Failed to read devcontainer.json: %w", err))
				return
			}

			devContainerConfig = config
		}

		containerId, err := pm.DevContainerRunner.Run(projectPath, devContainerConfig)

		if err != nil {
			log.Println("Failed to launch devcontainer:", err)
			removeProjectDir(projectPath)
			c <- result.Failure[Project](fmt.Errorf("Failed to launch devcontainer: %w", err))
			return
		}

		project := Project{Id: projectId, Path: projectPath, Config: Config{DevContainerConfig: devContainerConfig}, containerId: containerId}

		if err := pm.Store.CreateProject(&project); err != nil {
			log.Printf("Failed to save project: %s", err)
			removeProjectDir(projectPath)
			c <- result.Failure[Project](fmt.Errorf("Failed to save project: %w", err))
			return
		}

		log.Printf("Created project %s for repo %s", projectId, request.Repository.Url)

		c <- result.Success(project)
	}()

	return c
}

func (pm ManagerImpl) GetProject(projectId string) (Project, error) {
	log.Printf("Getting project %s", projectId)

	project, err := pm.Store.GetProject(projectId)

	if err != nil {
		log.Printf("Project with id %s not found", projectId)
		return Project{}, fmt.Errorf("Project with id %s not found", projectId)
	}

	log.Printf("Got project %+v", project)

	return *project, nil
}

func (pm ManagerImpl) GetProjects() ([]*Project, error) {
	log.Printf("Getting projects")

	projects, err := pm.Store.GetProjects()

	if err != nil {
		log.Printf("Failed to get projects: %s", err)
		return nil, fmt.Errorf("Failed to get projects: %w", err)
	}

	log.Printf("Got projects %+v", projects)

	return projects, nil
}

func (pm ManagerImpl) ResolveTaskAlias(projectId string, alias string) (devcontainer.Task, error) {
	log.Printf("Resolving task alias %s for project %s", alias, projectId)

	project, err := pm.GetProject(projectId)

	if err != nil {
		log.Printf("Project with id %s not found", projectId)
		return devcontainer.Task{}, fmt.Errorf("Project with id %s not found", projectId)
	}

	task, err := project.FindTaskByAlias(alias)

	if err != nil {
		log.Printf("Task with alias %s for project %s not found", alias, projectId)
		return devcontainer.Task{}, fmt.Errorf("Task with alias %s not found", alias)
	}

	log.Printf("Resolved task alias %s for project %s: %+v", alias, projectId, task)

	return task, nil
}

func (pm ManagerImpl) CreateTask(projectId string, command string) (TaskResult, error) {
	log.Printf("Creating task for project %s. Command: %s", projectId, command)

	project, err := pm.GetProject(projectId)

	if err != nil {
		log.Printf("Project with id %s not found", projectId)
		return TaskResult{}, fmt.Errorf("Project with id %s not found", projectId)
	}

	execResult, err := pm.DevContainerRunner.Exec(project.containerId, strings.Split(command, " "))

	if err != nil {
		log.Printf("Failed to execute command '%s' in container %s: %s", command, project.containerId, err)
		return TaskResult{}, fmt.Errorf("Failed to execute command: %w", err)
	}

	log.Printf("Task '%s' for project %s executed successfully", command, projectId)

	return TaskResult{StdOut: execResult.StdOut, StdErr: execResult.StdErr, ExitCode: execResult.ExitCode}, nil
}

func (pm ManagerImpl) Cleanup() error {
	log.Printf("Cleaning up projects")

	projects, err := pm.GetProjects()

	if err != nil {
		log.Printf("Failed to get projects: %s", err)
		return fmt.Errorf("Failed to get projects: %w", err)
	}

	for _, project := range projects {
		log.Printf("Cleaning up project %s", project.Id)
		pm.DevContainerRunner.Stop(project.containerId)
	}

	log.Printf("Cleaned up projects")

	return nil
}

func (pm ManagerImpl) createProjectDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("Failed to create project directory: %w", err)
	}

	log.Println("Created project directory: ", path)

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
		log.Printf("Failed to remove project directory %s: %s", projectPath, err)
		return
	}

	log.Println("Removed project directory: ", projectPath)

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

		log.Printf("Cloned git repo %s to %s", repository.Url, projectPath)
		log.Println(string(cmdOut))

		if repository.Commit != nil {
			cmd = exec.Command("git", "checkout", *repository.Commit)
			cmdOut, err = cmd.Output()

			if err != nil {
				c <- result.EmptyFailure(fmt.Errorf("Failed to checkout commit %s: %w", *repository.Commit, err))
				return
			}

			log.Printf("Checked out commit %s", *repository.Commit)
			log.Println(string(cmdOut))
		}

		c <- result.EmptySuccess()
	}()

	return c
}
