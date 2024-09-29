package devcontainer

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

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
	commandExecutor  Executor
	imageManager     ImageManager
	containerManager ContainerManager
}

func NewDockerRunner(commandExecutor Executor, imageManager ImageManager, containerManager ContainerManager) Runner {
	return &DockerRunner{
		commandExecutor:  commandExecutor,
		imageManager:     imageManager,
		containerManager: containerManager,
	}
}

func (r *DockerRunner) Run(ctx context.Context, projectPath string, config Config) (string, error) {
	log.Debug().Any("config", config).Msg("Running container")
	// Run initialize commands
	if command := config.LifecycleProps.InitializeCommand; command != nil {
		if err := r.executeLifecycleCommand(command, projectPath); err != nil {
			return "", fmt.Errorf("Failed to run initialize commands: %w", err)
		}
	}

	// Get image
	imageId, err := r.getImage(ctx, config, projectPath)
	if err != nil {
		return "", fmt.Errorf("Failed to get image: %w", err)
	}

	// Create container
	containerId, err := r.containerManager.CreateContainer(ctx, imageId, projectPath, config)
	if err != nil {
		return "", fmt.Errorf("Failed to create container: %w", err)
	}

	// Start container
	if err := r.containerManager.StartContainer(ctx, containerId); err != nil {
		return "", fmt.Errorf("Failed to start container: %w", err)
	}

	// Run onCreate commands
	if command := config.LifecycleProps.OnCreateCommand; command != nil {
		if err := r.executeLifecycleCommandInContainer(ctx, command, containerId); err != nil {
			return "", fmt.Errorf("Failed to run onCreate commands: %w", err)
		}
	}

	// Run updateContent commands
	if command := config.LifecycleProps.UpdateContentCommand; command != nil {
		if err := r.executeLifecycleCommandInContainer(ctx, command, containerId); err != nil {
			return "", fmt.Errorf("Failed to run updateContent commands: %w", err)
		}
	}

	// Run postCreate commands
	if command := config.LifecycleProps.PostCreateCommand; command != nil {
		if err := r.executeLifecycleCommandInContainer(ctx, command, containerId); err != nil {
			return "", fmt.Errorf("Failed to run postCreate commands: %w", err)
		}
	}

	// Run postStart commands
	if command := config.LifecycleProps.PostStartCommand; command != nil {
		if err := r.executeLifecycleCommand(command, projectPath); err != nil {
			return "", fmt.Errorf("Failed to run postStart commands: %w", err)
		}
	}

	// Run postAttach commands
	if command := config.LifecycleProps.PostAttachCommand; command != nil {
		if err := r.executeLifecycleCommand(command, projectPath); err != nil {
			return "", fmt.Errorf("Failed to run postAttach commands: %w", err)
		}
	}

	return containerId, nil
}

func (r *DockerRunner) Stop(ctx context.Context, containerId string) error {
	return r.containerManager.StopContainer(ctx, containerId)
}

func (r *DockerRunner) Exec(ctx context.Context, containerID string, command []string) (ExecResult, error) {
	return r.containerManager.Exec(ctx, containerID, command)
}

func (r *DockerRunner) executeLifecycleCommand(lifecycleCommand LifecycleCommand, workingDir string) error {
	for name, command := range lifecycleCommand {
		log.Debug().Str("name", name).Str("command", fmt.Sprintf("%s", command)).Msg("Running command")

		if err := r.commandExecutor.Run(command, workingDir, os.Stdout, os.Stderr); err != nil {
			return fmt.Errorf("Failed to run command %s %s: %w", name, command, err)
		}
	}

	return nil
}

func (r *DockerRunner) executeLifecycleCommandInContainer(ctx context.Context, lifecycleCommand LifecycleCommand, containerId string) error {
	for name, command := range lifecycleCommand {
		log.Debug().Str("name", name).Str("command", fmt.Sprintf("%s", command)).Msg("Running command")

		result, err := r.Exec(ctx, containerId, command)
		if err != nil {
			return fmt.Errorf("Failed to run command %s %s in container %s: %w", name, command, containerId, err)
		}

		if result.ExitCode != 0 {
			return fmt.Errorf("Failed to run command %s %s in container %s: Exit code %d. Stdout: %s, Stderr: %s", name, command, containerId, result.ExitCode, result.StdOut, result.StdErr)
		}

	}

	return nil
}

func (r *DockerRunner) getImage(ctx context.Context, config Config, projectPath string) (string, error) {
	switch {
	case config.IsImageDevContainer():
		return r.getOrPullImage(ctx, config.DockerImageProps.Image)
	case config.IsDockerfileDevContainer():
		imageId, err := r.imageManager.BuildImage(ctx, projectPath, config)
		if err != nil {
			return "", fmt.Errorf("Failed to build image: %w", err)
		}
		return imageId, nil
	case config.IsComposeDevContainer():
		// TODO: build docker-compose file
		return "", fmt.Errorf("Docker Compose is not supported yet")
	default:
		return "", fmt.Errorf("Invalid devcontainer configuration")
	}
}

func (r *DockerRunner) getOrPullImage(ctx context.Context, imageId string) (string, error) {
	if imageId == "" {
		return "", fmt.Errorf("image id is empty")
	}

	exists, err := r.imageManager.LocalImageExists(ctx, imageId)
	if err != nil {
		return "", fmt.Errorf("Failed to check if image %s exists: %w", imageId, err)
	}

	if exists {
		return imageId, nil
	}

	if err := r.imageManager.PullImage(ctx, imageId); err != nil {
		return "", fmt.Errorf("Failed to pull image %s: %w", imageId, err)
	}

	return imageId, nil
}
