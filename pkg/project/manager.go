package project

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/artmoskvin/hide/pkg/devcontainer"
)

type CreateProjectRequest struct {
	RepoUrl string `json:"repoUrl"`
}

type Task struct {
	Alias   string `json:"alias"`
	Command string `json:"command"`
}

type DevContainerConfig struct {
	Tasks []Task `json:"tasks"`
}

type Project struct {
	Id     string             `json:"id"`
	Path   string             `json:"path"`
	Config DevContainerConfig `json:"config"`
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
	GetProject(projectId string) (Project, error)
	ResolveTaskAlias(projectId string, alias string) (Task, error)
	CreateTask(projectId string, command string) (TaskResult, error)
}

type ManagerImpl struct {
	DevContainerManager devcontainer.Manager
	Store               Store
	ProjectsRoot        string
}

func NewProjectManager(devContainerManager devcontainer.Manager, projectStore Store, projectsRoot string) Manager {
	return ManagerImpl{DevContainerManager: devContainerManager, Store: projectStore, ProjectsRoot: projectsRoot}
}

func (pm ManagerImpl) CreateProject(request CreateProjectRequest) (Project, error) {
	projectId := randomString(10)
	projectPath := path.Join(pm.ProjectsRoot, projectId)

	if err := pm.createProjectDir(projectPath); err != nil {
		return Project{}, fmt.Errorf("Failed to create project directory: %w", err)
	}

	if err := cloneGitRepo(request.RepoUrl, projectPath); err != nil {
		removeProjectDir(projectPath)
		return Project{}, fmt.Errorf("Failed to clone git repo: %w", err)
	}

	devContainerConfig := pm.devContainerConfigFromProject(projectPath)

	if _, err := pm.DevContainerManager.StartContainer(projectPath); err != nil {
		removeProjectDir(projectPath)
		return Project{}, fmt.Errorf("Failed to launch devcontainer: %w", err)
	}

	project := Project{Id: projectId, Path: projectPath, Config: devContainerConfig}

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

	container, err := pm.DevContainerManager.FindContainerByProject(projectId)

	if err != nil {
		return TaskResult{}, fmt.Errorf("Failed to find container for project %s: %w", projectId, err)
	}

	execResult, err := pm.DevContainerManager.Exec(container.Id, project.Path, command)

	if err != nil {
		return TaskResult{}, fmt.Errorf("Failed to execute command: %w", err)
	}

	return TaskResult{StdOut: execResult.StdOut, StdErr: execResult.StdErr, ExitCode: execResult.ExitCode}, nil
}

func (pm ManagerImpl) createProjectDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("Failed to create project directory: %w", err)
	}

	fmt.Println("Created project directory: ", path)

	return nil
}

func (pm ManagerImpl) devContainerConfigFromProject(projectPath string) DevContainerConfig {
	// TODO: find devcontainer.json in the project and parse it into a Config struct
	return DevContainerConfig{Tasks: []Task{}}
}

func removeProjectDir(projectPath string) {
	if err := os.RemoveAll(projectPath); err != nil {
		fmt.Printf("Failed to remove project directory %s: %s", projectPath, err)
		return
	}

	fmt.Println("Removed project directory: ", projectPath)

	return
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
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
