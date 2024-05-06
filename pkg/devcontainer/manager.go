package devcontainer

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Container struct {
	Id        string
	ProjectID string
}

type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

type Manager interface {
	StartContainer(projectPath string) (Container, error)
	FindContainerByProject(projectId string) (Container, error)
	StopContainer(containerId string) error
	Exec(containerId string, projectPath string, command string) (ExecResult, error)
}

type CliManager struct {
	Store Store
}

func NewDevContainerManager() Manager {
	return CliManager{Store: NewInMemoryStore(make(map[string]*Container))}
}

func (m CliManager) StartContainer(projectPath string) (Container, error) {
	cmd := exec.Command("devcontainer", "up", "--log-format", "json", "--workspace-folder", projectPath)
	cmdOut, err := cmd.Output()

	if err != nil {
		return Container{}, fmt.Errorf("Failed to launch devcontainer: %w", err)
	}

	fmt.Println(">", cmd.String())
	fmt.Println(string(cmdOut))

	jsonOutput := string(cmdOut)

	var response map[string]interface{}

	fmt.Println("Trying to parse json: ", jsonOutput)

	if err := json.Unmarshal([]byte(jsonOutput), &response); err != nil {
		return Container{}, fmt.Errorf("Failed to parse devcontainer output: %w", err)
	}

	containerId := response["containerId"].(string)
	container := Container{Id: containerId}
	m.Store.CreateContainer(&container)
	return container, nil
}

func (m CliManager) FindContainerByProject(projectId string) (Container, error) {
	containers, err := m.Store.GetContainerByProject(projectId)
	if err != nil {
		return Container{}, fmt.Errorf("Failed to find container for project %s: %w", projectId, err)
	}

	if len(containers) == 0 {
		return Container{}, fmt.Errorf("No container found for project %s", projectId)
	}

	if len(containers) > 1 {
		return Container{}, fmt.Errorf("Multiple containers found for project %s", projectId)
	}

	return *containers[0], nil
}

func (m CliManager) StopContainer(containerId string) error {
	cmd := exec.Command("docker", "stop", containerId)

	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("Failed to stop container %s: %w", containerId, err)
	}

	if err := m.Store.DeleteContainer(containerId); err != nil {
		return fmt.Errorf("Failed to delete container %s: %w", containerId, err)
	}

	return nil
}

func (m CliManager) Exec(containerId string, projectPath string, command string) (ExecResult, error) {
	allArgs := append([]string{"exec", "--workspace-folder", projectPath}, strings.Split(command, " ")...)
	cmd := exec.Command("devcontainer", allArgs...)
	cmdOut, err := cmd.Output()

	if err != nil {
		return ExecResult{}, fmt.Errorf("Failed to exec command %s in devcontainer %s: %w", command, containerId, err)
	}

	fmt.Println("> ", cmd.String())
	fmt.Println(string(cmdOut))

	// TODO: how to get exit code and stderr?
	return ExecResult{StdOut: string(cmdOut)}, nil
}
