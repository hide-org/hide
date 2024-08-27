package devcontainer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/artmoskvin/hide/pkg/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
)

const DefaultShell = "/bin/sh"

var DefaultContainerCommand = []string{DefaultShell, "-c", "while sleep 1000; do :; done"}

type DockerRunnerConfig struct {
	Username string
	Password string
}

type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

type Runner interface {
	Run(ctx context.Context, projectPath string, config Config) (string, error)
	Stop(ctx context.Context, containerId string) error
	Exec(ctx context.Context, containerId string, command []string) (ExecResult, error)
}

type DockerRunner struct {
	dockerClient    DockerClient
	commandExecutor util.Executor
	config          DockerRunnerConfig
	imageManager    ImageManager
}

func NewDockerRunner(client DockerClient, commandExecutor util.Executor, config DockerRunnerConfig, imageManager ImageManager) Runner {
	return &DockerRunner{
		dockerClient:    client,
		commandExecutor: commandExecutor,
		config:          config,
		imageManager:    imageManager,
	}
}

func (r *DockerRunner) Run(ctx context.Context, projectPath string, config Config) (string, error) {
	// Run initialize commands
	if command := config.LifecycleProps.InitializeCommand; command != nil {
		if err := r.executeLifecycleCommand(command, projectPath); err != nil {
			return "", fmt.Errorf("Failed to run initialize command %s: %w", command, err)
		}
	}

	// Build docker compose
	if len(config.DockerComposeFile) > 0 {
		// TODO: build docker-compose file
		return "", fmt.Errorf("Docker Compose is not supported yet")
	}

	// Pull or build image
	imageId, err := r.imageManager.PullOrBuildImage(ctx, projectPath, config)

	if err != nil {
		return "", fmt.Errorf("Failed to pull or build image: %w", err)
	}

	// Create container
	containerId, err := r.createContainer(ctx, imageId, projectPath, config)

	if err != nil {
		return "", fmt.Errorf("Failed to create container: %w", err)
	}

	// Start container
	if err := r.startContainer(ctx, containerId); err != nil {
		return "", fmt.Errorf("Failed to start container: %w", err)
	}

	// Run onCreate commands
	if command := config.LifecycleProps.OnCreateCommand; command != nil {
		if err := r.executeLifecycleCommandInContainer(ctx, command, containerId); err != nil {
			return "", fmt.Errorf("Failed to run onCreate command %s: %w", command, err)
		}
	}

	// Run updateContent commands
	if command := config.LifecycleProps.UpdateContentCommand; command != nil {
		if err := r.executeLifecycleCommandInContainer(ctx, command, containerId); err != nil {
			return "", fmt.Errorf("Failed to run updateContent command %s: %w", command, err)
		}
	}

	// Run postCreate commands
	if command := config.LifecycleProps.PostCreateCommand; command != nil {
		if err := r.executeLifecycleCommandInContainer(ctx, command, containerId); err != nil {
			return "", fmt.Errorf("Failed to run postCreate command %s: %w", command, err)
		}
	}

	// Run postStart commands
	if command := config.LifecycleProps.PostStartCommand; command != nil {
		if err := r.executeLifecycleCommand(command, projectPath); err != nil {
			return "", fmt.Errorf("Failed to run postStart command %s: %w", command, err)
		}
	}

	// Run postAttach commands
	if command := config.LifecycleProps.PostAttachCommand; command != nil {
		if err := r.executeLifecycleCommand(command, projectPath); err != nil {
			return "", fmt.Errorf("Failed to run postAttach command %s: %w", command, err)
		}
	}

	return containerId, nil
}

func (r *DockerRunner) Stop(ctx context.Context, containerId string) error {
	if err := r.dockerClient.ContainerStop(ctx, containerId, container.StopOptions{}); err != nil {
		return fmt.Errorf("Failed to stop container %s: %w", containerId, err)
	}

	return nil
}

func (r *DockerRunner) Exec(ctx context.Context, containerID string, command []string) (ExecResult, error) {
	execConfig := types.ExecConfig{
		Cmd:          command,
		AttachStdout: true,
		AttachStderr: true,
	}
	execIDResp, err := r.dockerClient.ContainerExecCreate(ctx, containerID, execConfig)

	if err != nil {
		return ExecResult{}, fmt.Errorf("Failed to create execute configuration for command %s in container %s: %w", command, containerID, err)
	}

	execID := execIDResp.ID

	resp, err := r.dockerClient.ContainerExecAttach(ctx, execID, types.ExecStartCheck{})

	if err != nil {
		return ExecResult{}, fmt.Errorf("Failed to attach to exec process %s in container %s: %w", execID, containerID, err)
	}

	defer resp.Close()

	var stdOut, stdErr bytes.Buffer

	stdOutWriter := io.MultiWriter(os.Stdout, &stdOut)
	stdErrWriter := io.MultiWriter(os.Stderr, &stdErr)

	if err := ReadOutputFromContainer(resp.Reader, stdOutWriter, stdErrWriter); err != nil {
		return ExecResult{}, fmt.Errorf("Error reading output from container: %w", err)
	}

	inspectResp, err := r.dockerClient.ContainerExecInspect(ctx, execID)

	if err != nil {
		return ExecResult{}, fmt.Errorf("Failed to inspect exec process %s in container %s: %w", execID, containerID, err)
	}

	return ExecResult{StdOut: stdOut.String(), StdErr: stdErr.String(), ExitCode: inspectResp.ExitCode}, nil
}

func (r *DockerRunner) executeLifecycleCommand(lifecycleCommand LifecycleCommand, workingDir string) error {
	for _, command := range lifecycleCommand {
		log.Debug().Str("command", fmt.Sprintf("%s", command)).Msg("Running command")

		if err := r.commandExecutor.Run(command, workingDir, os.Stdout, os.Stderr); err != nil {
			return err
		}
	}

	return nil
}

func (r *DockerRunner) executeLifecycleCommandInContainer(ctx context.Context, lifecycleCommand LifecycleCommand, containerId string) error {
	for _, command := range lifecycleCommand {
		log.Debug().Str("command", fmt.Sprintf("%s", command)).Msg("Running command")

		result, err := r.Exec(ctx, containerId, command)

		if err != nil {
			return err
		}

		if result.ExitCode != 0 {
			return fmt.Errorf("Exit code %d. Stdout: %s, Stderr: %s", result.ExitCode, result.StdOut, result.StdErr)
		}

	}

	return nil
}

func (r *DockerRunner) createContainer(ctx context.Context, image string, projectPath string, config Config) (string, error) {
	log.Debug().Msg("Creating container")

	env := []string{}

	for envKey, envValue := range config.ContainerEnv {
		env = append(env, fmt.Sprintf("%s=%s", envKey, envValue))
	}

	containerConfig := &container.Config{Image: image, Cmd: DefaultContainerCommand, Env: env}

	if config.ContainerUser != "" {
		containerConfig.User = config.ContainerUser
	}

	portBindings := make(nat.PortMap)

	for _, port := range config.AppPort {
		port_str := strconv.Itoa(port)
		port, err := nat.NewPort("tcp", port_str)

		if err != nil {
			return "", fmt.Errorf("Failed to create new TCP port from port %s: %w", port_str, err)
		}

		portBindings[port] = []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: port_str}}
	}

	mounts := []mount.Mount{}

	if config.WorkspaceMount != nil && config.WorkspaceFolder != "" {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: config.WorkspaceMount.Source,
			Target: config.WorkspaceMount.Destination,
		})

		containerConfig.WorkingDir = config.WorkspaceFolder
	} else {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: projectPath,
			Target: "/workspace",
		})

		containerConfig.WorkingDir = "/workspace"
	}

	if len(config.Mounts) > 0 {
		for _, _mount := range config.Mounts {
			mountType, err := stringToType(_mount.Type)

			if err != nil {
				return "", fmt.Errorf("Failed to convert mount type %s to type.Type: %w", _mount.Type, err)
			}

			mounts = append(mounts, mount.Mount{
				Type:   mountType,
				Source: _mount.Source,
				Target: _mount.Destination,
			})
		}
	}

	hostConfig := container.HostConfig{
		PortBindings: portBindings,
		Mounts:       mounts,
		Init:         &config.Init,
		Privileged:   config.Privileged,
		CapAdd:       config.CapAdd,
		SecurityOpt:  config.SecurityOpt,
	}

	createResponse, err := r.dockerClient.ContainerCreate(ctx, containerConfig, &hostConfig, nil, nil, "")

	if err != nil {
		return "", err
	}

	containerId := createResponse.ID

	log.Debug().Str("containerId", containerId).Msg("Created container")

	return containerId, nil
}

func (r *DockerRunner) startContainer(ctx context.Context, containerId string) error {
	log.Debug().Msg("Starting container")

	err := r.dockerClient.ContainerStart(ctx, containerId, container.StartOptions{})

	if err != nil {
		return err
	}

	log.Debug().Str("containerId", containerId).Msg("Started container")

	return nil
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
