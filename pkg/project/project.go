package project

import "os/exec"
import "math/rand"
import "time"
import "os"
import "fmt"
import "github.com/artmoskvin/hide/pkg/devcontainer"

const ProjectsDir = "hide-projects"

type CreateProjectRequest struct {
	RepoUrl string `json:"repoUrl"`
}

type ExecCmdRequest struct {
	Cmd string `json:"cmd"`
}

type Project struct {
	// Project id is a container id for now. It can change in the future.
	Id   string `json:"id"`
	Path string `json:"path"`
}

type CmdResult struct {
	StdOut   string `json:"stdOut"`
	StdErr   string `json:"stdErr"`
	ExitCode int    `json:"exitCode"`
}

type Manager interface {
	CreateProject(request CreateProjectRequest) (Project, error)
	ExecCmd(projectId string, request ExecCmdRequest) (CmdResult, error)
}

type SimpleManager struct {
	DevContainerManager devcontainer.Manager
	InMemoryProjects    map[string]Project
}

func (pm SimpleManager) CreateProject(request CreateProjectRequest) (Project, error) {
	projectPath, err := createProjectDir()

	if err != nil {
		return Project{}, fmt.Errorf("Failed to create project directory: %w", err)
	}

	if err = cloneGitRepo(request.RepoUrl, projectPath); err != nil {
		return Project{}, fmt.Errorf("Failed to clone git repo: %w", err)
	}

	devContainer, err := pm.DevContainerManager.StartContainer(projectPath)

	if err != nil {
		return Project{}, fmt.Errorf("Failed to launch devcontainer: %w", err)
	}

	project := Project{Id: devContainer.Id, Path: projectPath}
	pm.InMemoryProjects[devContainer.Id] = project

	return project, nil
}

func (pm SimpleManager) ExecCmd(projectId string, request ExecCmdRequest) (CmdResult, error) {
	project, ok := pm.InMemoryProjects[projectId]

	if !ok {
		return CmdResult{}, fmt.Errorf("Project with id %s not found", projectId)
	}

	execResult, err := pm.DevContainerManager.Exec(project.Id, project.Path, request.Cmd)

	if err != nil {
		return CmdResult{}, fmt.Errorf("Failed to execute command: %w", err)
	}

	return CmdResult{StdOut: execResult.StdOut, StdErr: execResult.StdErr, ExitCode: execResult.ExitCode}, nil
}

func createProjectDir() (string, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return "", fmt.Errorf("Failed to get user home directory: %w", err)
	}

	projectParentDir := fmt.Sprintf("%s/%s", home, ProjectsDir)
	dirName := randomString(10)
	projectPath := fmt.Sprintf("%s/%s", projectParentDir, dirName)

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return "", fmt.Errorf("Failed to create project directory: %w", err)
	}

	fmt.Println("Created project directory: ", projectPath)

	return projectPath, nil
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
