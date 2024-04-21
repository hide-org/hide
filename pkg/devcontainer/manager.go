package devcontainer

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Container struct {
	Id string
}

type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

type Manager interface {
	StartContainer(projectPath string) (Container, error)
	Exec(containerId string, projectPath string, command string) (ExecResult, error)
}

type CliManager struct{}

func NewDevContainerManager() Manager {
	return CliManager{}
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
	return Container{Id: containerId}, nil
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
