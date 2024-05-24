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
	"github.com/artmoskvin/hide/pkg/util"
)

type CreateProjectRequest struct {
	RepoUrl string `json:"repoUrl"`
}

type Task struct {
	Alias   string `json:"alias"`
	Command string `json:"command"`
}

type Config struct {
	DevContainerConfig *devcontainer.Config `json:"devContainerConfig"`
	Tasks              []Task               `json:"tasks"`
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

func (project *Project) FindTaskByAlias(alias string) (Task, error) {
	for _, task := range project.Config.Tasks {
		if task.Alias == alias {
			return task, nil
		}
	}
	return Task{}, errors.New("task not found")
}

type TaskResult struct {
	StdOut   string `json:"stdOut"`
	StdErr   string `json:"stdErr"`
	ExitCode int    `json:"exitCode"`
}

type Manager interface {
	CreateProject(request CreateProjectRequest) (Project, error)
	GetProject(projectId ProjectId) (Project, error)
	GetProjects() ([]*Project, error)
	ResolveTaskAlias(projectId ProjectId, alias string) (Task, error)
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

func (pm ManagerImpl) CreateProject(request CreateProjectRequest) (Project, error) {
	projectId := util.RandomString(10)
	projectPath := path.Join(pm.ProjectsRoot, projectId)

	if err := pm.createProjectDir(projectPath); err != nil {
		return Project{}, fmt.Errorf("Failed to create project directory: %w", err)
	}

	if err := cloneGitRepo(request.RepoUrl, projectPath); err != nil {
		removeProjectDir(projectPath)
		return Project{}, fmt.Errorf("Failed to clone git repo: %w", err)
	}

	config, err := pm.configFromProject(os.DirFS(projectPath))

	if err != nil {
		removeProjectDir(projectPath)
		return Project{}, fmt.Errorf("Failed to read devcontainer.json: %w", err)
	}

	containerId, err := pm.DevContainerRunner.Run(projectPath, config.DevContainerConfig)

	if err != nil {
		log.Println("Failed to launch devcontainer:", err)
		removeProjectDir(projectPath)
		return Project{}, fmt.Errorf("Failed to launch devcontainer: %w", err)
	}

	project := Project{Id: projectId, Path: projectPath, Config: config, containerId: containerId}

	if err := pm.Store.CreateProject(&project); err != nil {
		removeProjectDir(projectPath)
		return Project{}, fmt.Errorf("Failed to save project: %w", err)
	}

	return project, nil
}

func (pm ManagerImpl) GetProject(projectId string) (Project, error) {
	project, err := pm.Store.GetProject(projectId)

	if err != nil {
		return Project{}, fmt.Errorf("Project with id %s not found", projectId)
	}

	return *project, nil
}

func (pm ManagerImpl) GetProjects() ([]*Project, error) {
	projects, err := pm.Store.GetProjects()

	if err != nil {
		return nil, fmt.Errorf("Failed to get projects: %w", err)
	}

	return projects, nil
}

func (pm ManagerImpl) ResolveTaskAlias(projectId string, alias string) (Task, error) {
	project, err := pm.GetProject(projectId)

	if err != nil {
		return Task{}, fmt.Errorf("Project with id %s not found", projectId)
	}

	task, err := project.FindTaskByAlias(alias)

	if err != nil {
		return Task{}, fmt.Errorf("Task with alias %s not found", alias)
	}

	return task, nil
}

func (pm ManagerImpl) CreateTask(projectId string, command string) (TaskResult, error) {
	project, err := pm.GetProject(projectId)

	if err != nil {
		return TaskResult{}, fmt.Errorf("Project with id %s not found", projectId)
	}

	execResult, err := pm.DevContainerRunner.Exec(project.containerId, strings.Split(command, " "))

	if err != nil {
		return TaskResult{}, fmt.Errorf("Failed to execute command: %w", err)
	}

	return TaskResult{StdOut: execResult.StdOut, StdErr: execResult.StdErr, ExitCode: execResult.ExitCode}, nil
}

func (pm ManagerImpl) Cleanup() error {
	projects, err := pm.GetProjects()

	if err != nil {
		return fmt.Errorf("Failed to get projects: %w", err)
	}

	for _, project := range projects {
		pm.DevContainerRunner.Stop(project.containerId)
	}

	return nil
}

func (pm ManagerImpl) createProjectDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("Failed to create project directory: %w", err)
	}

	fmt.Println("Created project directory: ", path)

	return nil
}

func (pm ManagerImpl) configFromProject(fileSystem fs.FS) (Config, error) {
	configFile, err := devcontainer.FindConfig(fileSystem)

	if err != nil {
		return Config{}, fmt.Errorf("Failed to find devcontainer.json: %w", err)
	}

	config, err := devcontainer.ParseConfig(configFile)

	if err != nil {
		return Config{}, fmt.Errorf("Failed to parse devcontainer.json: %w", err)
	}

	// TODO: parse tasks from customizations
	var tasks []Task

	return Config{DevContainerConfig: config, Tasks: tasks}, nil
}

func removeProjectDir(projectPath string) {
	if err := os.RemoveAll(projectPath); err != nil {
		fmt.Printf("Failed to remove project directory %s: %s", projectPath, err)
		return
	}

	fmt.Println("Removed project directory: ", projectPath)

	return
}

func cloneGitRepo(githubUrl string, projectPath string) error {
	cmd := exec.Command("git", "clone", githubUrl, projectPath)
	cmdOut, err := cmd.Output()

	if err != nil {
		return fmt.Errorf("Failed to clone git repo: %w", err)
	}

	fmt.Println("> ", cmd.String())
	fmt.Println(string(cmdOut))

	return nil
}
