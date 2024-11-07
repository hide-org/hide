package devcontainer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	DefaultShell      = "/bin/sh"
	DefaultWorkingDir = "/workspace"
)

var DefaultContainerCommand = []string{DefaultShell, "-c", "while sleep 1000; do :; done"}

type ContainerManager interface {
	CreateContainer(ctx context.Context, image string, projectPath string, config Config) (string, error)
	StartContainer(ctx context.Context, containerId string) error
	StopContainer(ctx context.Context, containerId string) error
	Exec(ctx context.Context, containerId string, command []string) (ExecResult, error)
}

type DockerContainerManager struct {
	client.ContainerAPIClient
}

func NewDockerContainerManager(dockerContainerCli client.ContainerAPIClient) ContainerManager {
	return &DockerContainerManager{ContainerAPIClient: dockerContainerCli}
}

func (cm *DockerContainerManager) CreateContainer(ctx context.Context, image string, projectPath string, config Config) (string, error) {
	containerConfig := &container.Config{Image: image, Cmd: DefaultContainerCommand}

	if len(config.ContainerEnv) > 0 {
		env := []string{}

		for envKey, envValue := range config.ContainerEnv {
			env = append(env, fmt.Sprintf("%s=%s", envKey, envValue))
		}

		containerConfig.Env = env
	}

	if config.ContainerUser != "" {
		containerConfig.User = config.ContainerUser
	}

	hostConfig := &container.HostConfig{
		Init:        &config.Init,
		Privileged:  config.Privileged,
		CapAdd:      config.CapAdd,
		SecurityOpt: config.SecurityOpt,
	}

	if len(config.AppPort) > 0 {
		portBindings := make(nat.PortMap)

		for _, port := range config.AppPort {
			port_str := strconv.Itoa(port)
			port, err := nat.NewPort("tcp", port_str)
			if err != nil {
				return "", fmt.Errorf("Failed to create new TCP port from port %s: %w", port_str, err)
			}

			portBindings[port] = []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: port_str}}
		}

		hostConfig.PortBindings = portBindings
	}

	mounts := []mount.Mount{}
	workspaceSource := projectPath
	workspaceTarget := DefaultWorkingDir
	containerConfig.WorkingDir = DefaultWorkingDir

	if config.WorkspaceMount != nil && config.WorkspaceFolder != "" {
		workspaceSource = config.WorkspaceMount.Source
		workspaceTarget = config.WorkspaceMount.Destination
		containerConfig.WorkingDir = config.WorkspaceFolder
	}

	mounts = append(mounts, mount.Mount{
		Type:   mount.TypeBind,
		Source: workspaceSource,
		Target: workspaceTarget,
	})

	if len(config.Mounts) > 0 {
		for _, m := range config.Mounts {
			mountType, err := stringToType(m.Type)
			if err != nil {
				return "", fmt.Errorf("Failed to convert mount type %s to mount.Type: %w", m.Type, err)
			}

			mounts = append(mounts, mount.Mount{
				Type:   mountType,
				Source: m.Source,
				Target: m.Destination,
			})
		}
	}

	hostConfig.Mounts = mounts
	createResponse, err := cm.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", err
	}

	return createResponse.ID, nil
}

func (cm *DockerContainerManager) StartContainer(ctx context.Context, containerId string) error {
	return cm.ContainerStart(ctx, containerId, container.StartOptions{})
}

func (cm *DockerContainerManager) StopContainer(ctx context.Context, containerId string) error {
	return cm.ContainerStop(ctx, containerId, container.StopOptions{})
}

func (cm *DockerContainerManager) Exec(ctx context.Context, containerId string, command []string) (ExecResult, error) {
	execConfig := types.ExecConfig{
		Cmd:          command,
		AttachStdout: true,
		AttachStderr: true,
	}

	execIDResp, err := cm.ContainerExecCreate(ctx, containerId, execConfig)
	if err != nil {
		return ExecResult{}, fmt.Errorf("Failed to create exec configuration for command %s in container %s: %w", command, containerId, err)
	}

	execID := execIDResp.ID
	resp, err := cm.ContainerExecAttach(ctx, execID, types.ExecStartCheck{})
	if err != nil {
		return ExecResult{}, fmt.Errorf("Failed to attach to exec process %s in container %s: %w", execID, containerId, err)
	}
	defer resp.Close()

	var stdOut, stdErr bytes.Buffer
	logPipe := &logPipe{}

	if err := readOutputFromContainer(ctx, resp.Reader, io.MultiWriter(&stdOut, logPipe), io.MultiWriter(&stdErr, logPipe)); err != nil {
		return ExecResult{}, fmt.Errorf("Failed reading output from container %s: %w", containerId, err)
	}

	inspectResp, err := cm.ContainerExecInspect(ctx, execID)
	if err != nil {
		return ExecResult{}, fmt.Errorf("Failed to inspect exec process %s in container %s: %w", execID, containerId, err)
	}

	return ExecResult{StdOut: stdOut.String(), StdErr: stdErr.String(), ExitCode: inspectResp.ExitCode}, nil
}

func stringToType(s string) (mount.Type, error) {
	switch s {
	case string(mount.TypeBind):
		return mount.TypeBind, nil
	case string(mount.TypeVolume):
		return mount.TypeVolume, nil
	case string(mount.TypeTmpfs):
		return mount.TypeTmpfs, nil
	case string(mount.TypeNamedPipe):
		return mount.TypeNamedPipe, nil
	case string(mount.TypeCluster):
		return mount.TypeCluster, nil
	default:
		return "", fmt.Errorf("Unsupported mount type: %s", s)
	}
}
