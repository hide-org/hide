package main

import "os/exec"
import "math/rand"
import "time"
import "os"
import "fmt"
import "encoding/json"
import "strings"

const ProjectsDir = "hide-projects"

type LaunchDevContainerRequest struct {
	githubUrl string
}

type ExecCmdRequest struct {
	devContainer DevContainer
	cmd          string
}

type DevContainer struct {
	id   string
	path string
}

type DevContainerManager interface {
	Launch() DevContainer
	Exec()
}

type DevContainerCli struct{}

func (pm DevContainerCli) Create(request LaunchDevContainerRequest) (DevContainer, error) {
	projectPath, err := createProjectDir()

	if err != nil {
		return DevContainer{}, fmt.Errorf("Failed to create project directory: %w", err)
	}

	if err = cloneGitRepo(request.githubUrl, projectPath); err != nil {
		return DevContainer{}, fmt.Errorf("Failed to clone git repo: %w", err)
	}

	devContainer, err := launchDevContainer(projectPath)

	if err != nil {
		return devContainer, fmt.Errorf("Failed to launch devcontainer: %w", err)
	}

	return devContainer, nil
}

func (pm DevContainerCli) Exec(request ExecCmdRequest) (string, error) {
	allArgs := append([]string{"exec", "--workspace-folder", request.devContainer.path}, strings.Split(request.cmd, " ")...)
	cmd := exec.Command("devcontainer", allArgs...)
	cmdOut, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("Failed to exec command %s in devcontainer %s: %w", request.devContainer.id, request.cmd, err)
	}

	fmt.Println(">", cmd.String())
	fmt.Println(string(cmdOut))

	return string(cmdOut), nil
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

	fmt.Println(">", cmd.String())
	fmt.Println(string(cmdOut))

	return nil
}

func launchDevContainer(projectPath string) (DevContainer, error) {
	cmd := exec.Command("devcontainer", "up", "--log-format", "json", "--workspace-folder", projectPath)
	cmdOut, err := cmd.Output()

	if err != nil {
		return DevContainer{}, fmt.Errorf("Failed to launch devcontainer: %w", err)
	}

	fmt.Println(">", cmd.String())
	fmt.Println(string(cmdOut))

	jsonOutput := string(cmdOut)
	// lines := strings.Split(jsonOutput, "\n")
	// lastLine := lines[len(lines)-1]

	var dat map[string]interface{}

	fmt.Println("Trying to parse json: ", jsonOutput)

	if err := json.Unmarshal([]byte(jsonOutput), &dat); err != nil {
		return DevContainer{}, fmt.Errorf("Failed to parse devcontainer output: %w", err)
	}

	containerId := dat["containerId"].(string)
	return DevContainer{id: containerId, path: projectPath}, nil
}
