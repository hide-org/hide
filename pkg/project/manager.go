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

type TaskRequest struct {
	Command *string `json:"command,omitempty"`
	Alias   *string `json:"alias,omitempty"`
}

type Task struct {
	Alias   string `json:"alias"`
	Command string `json:"command"`
}

type DevContainerConfig struct {
	Tasks []Task `json:"tasks"`
}

type Project struct {
	Id   string `json:"id"`
	Path string `json:"path"`
	// TODO: container id is ephemeral, it should not be here and should not be exposed to the client
	ContainerId string `json:"containerId"`
	Tasks       []Task `json:"tasks"`
}

func (project *Project) findTaskByAlias(alias string) (Task, error) {
	for _, task := range project.Tasks {
		if task.Alias == alias {
			return task, nil
		}
	}
	return Task{}, errors.New("task not found")
}

type CmdResult struct {
	StdOut   string `json:"stdOut"`
	StdErr   string `json:"stdErr"`
	ExitCode int    `json:"exitCode"`
}

type Manager interface {
	CreateProject(request CreateProjectRequest) (Project, error)
	GetProject(projectId string) (Project, error)
	CreateTask(projectId string, request TaskRequest) (CmdResult, error)
}

type ManagerImpl struct {
	DevContainerManager devcontainer.Manager
	ProjectStore        map[string]Project
	ProjectsRoot        string
}

func NewProjectManager(devContainerManager devcontainer.Manager, projectStore map[string]Project, projectsRoot string) Manager {
	return ManagerImpl{DevContainerManager: devContainerManager, ProjectStore: projectStore, ProjectsRoot: projectsRoot}
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

	// TODO: who should be responsible for parsing devcontainer.json?
	devContainerConfig := pm.devContainerConfigFromProject(projectPath)

	devContainer, err := pm.DevContainerManager.StartContainer(projectPath)

	if err != nil {
		removeProjectDir(projectPath)
		return Project{}, fmt.Errorf("Failed to launch devcontainer: %w", err)
	}

	// TODO: save devcontainer commands to the project, maybe the whole config?
	project := Project{Id: projectId, Path: projectPath, ContainerId: devContainer.Id, Tasks: devContainerConfig.Tasks}
	pm.ProjectStore[devContainer.Id] = project

	return project, nil
}

func (pm ManagerImpl) GetProject(projectId string) (Project, error) {
	project, ok := pm.ProjectStore[projectId]

	if !ok {
		return Project{}, fmt.Errorf("Project with id %s not found", projectId)
	}

	return project, nil
}

func (pm ManagerImpl) CreateTask(projectId string, request TaskRequest) (CmdResult, error) {
	project, ok := pm.ProjectStore[projectId]

	if !ok {
		return CmdResult{}, fmt.Errorf("Project with id %s not found", projectId)
	}

	var command string

	if request.Alias != nil {
		task, err := project.findTaskByAlias(*request.Alias)

		if err != nil {
			return CmdResult{}, fmt.Errorf("Task with alias %s not found", *request.Alias)
		}

		command = task.Command
	}

	if request.Command != nil {
		command = *request.Command
	}

	// TODO: can both command and alias be empty?

	execResult, err := pm.DevContainerManager.Exec(project.Id, project.Path, command)

	if err != nil {
		return CmdResult{}, fmt.Errorf("Failed to execute command: %w", err)
	}

	return CmdResult{StdOut: execResult.StdOut, StdErr: execResult.StdErr, ExitCode: execResult.ExitCode}, nil
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
