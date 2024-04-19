package project

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type LaunchDevContainerRequest struct {
	GithubUrl string `json:"githubUrl"`
}

type ExecCmdRequest struct {
	DevContainer DevContainer `json:"devContainer"`
	Cmd          string       `json:"cmd"`
}

type DevContainer struct {
	Id   string `json:"id"`
	Path string `json:"path"`
}

type CmdOutput struct {
	Output string `json:"output"`
}

type DevContainerManager interface {
	Launch() DevContainer
	Exec()
}

type DevContainerCli struct{}

func NewDevContainerCli() *DevContainerCli {
	return &DevContainerCli{}
}

func (cl DevContainerCli) Create(request LaunchDevContainerRequest, projectPath string) (DevContainer, error) {
	if err := cloneGitRepo(request.GithubUrl, projectPath); err != nil {
		return DevContainer{}, fmt.Errorf("Failed to clone git repo: %w", err)
	}

	devContainer, err := launchDevContainer(projectPath)

	if err != nil {
		return devContainer, fmt.Errorf("Failed to launch devcontainer: %w", err)
	}

	return devContainer, nil
}

func (cl DevContainerCli) Exec(request ExecCmdRequest) (CmdOutput, error) {
	// TODO: use container id instead of path
	allArgs := append([]string{"exec", "--workspace-folder", request.DevContainer.Path}, strings.Split(request.Cmd, " ")...)
	cmd := exec.Command("devcontainer", allArgs...)
	cmdOut, err := cmd.Output()

	if err != nil {
		return CmdOutput{}, fmt.Errorf("Failed to exec command %s in devcontainer %s: %w", request.DevContainer.Id, request.Cmd, err)
	}

	fmt.Println(">", cmd.String())
	fmt.Println(string(cmdOut))

	return CmdOutput{Output: string(cmdOut)}, nil
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

	var dat map[string]interface{}

	fmt.Println("Trying to parse json: ", jsonOutput)

	if err := json.Unmarshal([]byte(jsonOutput), &dat); err != nil {
		return DevContainer{}, fmt.Errorf("Failed to parse devcontainer output: %w", err)
	}

	containerId := dat["containerId"].(string)
	return DevContainer{Id: containerId, Path: projectPath}, nil
}
