package devcontainer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"

	"log"
	"os"

	"github.com/artmoskvin/hide/pkg/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"

	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

var DefaultContainerCommand = []string{"/bin/sh", "-c", "while sleep 1000; do :; done"}

type Runner interface {
	Run(projectPath string, config *Config) (string, error)
	Stop(containerId string) error
	Exec(containerId string, command []string) (ExecResult, error)
}

type DockerRunner struct {
	dockerClient    *client.Client
	commandExecutor util.Executor
}

func NewRunnerImpl(client *client.Client, commandExecutor util.Executor) Runner {
	return &DockerRunner{
		dockerClient:    client,
		commandExecutor: commandExecutor,
	}
}

func (r *DockerRunner) Run(projectPath string, config *Config) (string, error) {
	// Run initialize commands
	if command := config.LifecycleProps.InitializeCommand; command != nil {
		if err := r.executeLifecycleCommand(command, projectPath); err != nil {
			return "", fmt.Errorf("Failed to run initialize command %s: %w", command, err)
		}
	}

	ctx := context.Background()

	// Build docker compose
	if len(config.DockerComposeFile) > 0 {
		// TODO: build docker-compose file
		return "", fmt.Errorf("Docker Compose is not supported yet")
	}

	// Pull or build image
	imageId, err := r.pullOrBuildImage(ctx, projectPath, config)

	if err != nil {
		return "", fmt.Errorf("Failed to pull or build image: %w", err)
	}

	// Create container
	containerId, err := r.createContainer(ctx, imageId)

	if err != nil {
		return "", fmt.Errorf("Failed to create container: %w", err)
	}

	// Start container
	if err := r.startContainer(ctx, containerId); err != nil {
		return "", fmt.Errorf("Failed to start container: %w", err)
	}

	// Run onCreate commands
	// TODO: run inside of container
	if command := config.LifecycleProps.OnCreateCommand; command != nil {
		if err := r.executeLifecycleCommand(command, projectPath); err != nil {
			return "", fmt.Errorf("Failed to run onCreate command %s: %w", command, err)
		}
	}

	// Run updateContent commands
	// TODO: run inside of container
	if command := config.LifecycleProps.UpdateContentCommand; command != nil {
		if err := r.executeLifecycleCommand(command, projectPath); err != nil {
			return "", fmt.Errorf("Failed to run updateContent command %s: %w", command, err)
		}
	}

	// Run postCreate commands
	// TODO: run inside of container
	if command := config.LifecycleProps.PostCreateCommand; command != nil {
		if err := r.executeLifecycleCommand(command, projectPath); err != nil {
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

func (r *DockerRunner) Stop(containerId string) error {
	if err := r.dockerClient.ContainerStop(context.Background(), containerId, container.StopOptions{}); err != nil {
		return fmt.Errorf("Failed to stop container %s: %w", containerId, err)
	}

	return nil
}

func (r *DockerRunner) Exec(containerID string, command []string) (ExecResult, error) {
	execConfig := types.ExecConfig{
		Cmd:          command,
		AttachStdout: true,
		AttachStderr: true,
	}
	execIDResp, err := r.dockerClient.ContainerExecCreate(context.Background(), containerID, execConfig)

	if err != nil {
		return ExecResult{}, fmt.Errorf("Failed to create execute configuration for command %s in container %s: %w", command, containerID, err)
	}

	execID := execIDResp.ID

	resp, err := r.dockerClient.ContainerExecAttach(context.Background(), execID, types.ExecStartCheck{})

	if err != nil {
		return ExecResult{}, fmt.Errorf("Failed to attach to exec process %s in container %s: %w", execID, containerID, err)
	}

	defer resp.Close()

	var buf bytes.Buffer
	multiWriter := io.MultiWriter(os.Stdout, &buf)

	if err := util.StreamOutput(resp.Reader, multiWriter); err != nil {
		return ExecResult{}, fmt.Errorf("Error streaming output: %w", err)
	}

	inspectResp, err := r.dockerClient.ContainerExecInspect(context.Background(), execID)

	if err != nil {
		return ExecResult{}, fmt.Errorf("Failed to inspect exec process %s in container %s: %w", execID, containerID, err)
	}

	return ExecResult{StdOut: buf.String(), StdErr: "", ExitCode: inspectResp.ExitCode}, nil
}

// TODO: execute commands in a separate goroutine
func (r *DockerRunner) executeLifecycleCommand(lifecycleCommand LifecycleCommand, workingDir string) error {
	for name, command := range lifecycleCommand {
		if name != "" {
			log.Println("Running command: ", name)
		}

		if err := r.commandExecutor.Run(command, workingDir, os.Stdout, os.Stderr); err != nil {
			return err
		}
	}

	return nil
}

func (r *DockerRunner) pullOrBuildImage(ctx context.Context, workingDir string, config *Config) (string, error) {
	if config.Image != "" {
		if err := r.pullImage(ctx, config.Image); err != nil {
			return "", fmt.Errorf("Failed to pull image %s: %w", config.Image, err)
		}

		return config.Image, nil
	}

	var dockerFile, context string

	if config.Dockerfile != "" {
		dockerFile = config.Dockerfile
	} else if config.Build != nil && config.Build.Dockerfile != "" {
		dockerFile = config.Build.Dockerfile
	} else {
		return "", fmt.Errorf("Dockerfile not found")
	}

	if config.Context != "" {
		context = config.Context
	} else if config.Build != nil && config.Build.Context != "" {
		context = config.Build.Context
	} else {
		// default value
		// NOTE: this is bad; default values should be set during parsing
		context = "."
	}

	dockerFilePath := filepath.Join(workingDir, config.Path, dockerFile)
	contextPath := filepath.Join(workingDir, config.Path, context)
	imageId, err := r.buildImage(ctx, contextPath, dockerFilePath, config.Build, config.Name)

	if err != nil {
		return "", fmt.Errorf("Failed to build image: %w", err)
	}

	return imageId, nil

}

func (r *DockerRunner) pullImage(ctx context.Context, _image string) error {
	log.Println("Pulling image...", _image)

	output, err := r.dockerClient.ImagePull(ctx, _image, image.PullOptions{})

	if err != nil {
		return err
	}

	defer output.Close()

	if err := util.StreamOutput(output, os.Stdout); err != nil {
		log.Printf("Error streaming output: %v\n", err)
	}

	log.Println("Pulled image", _image)

	return nil
}

func (r *DockerRunner) buildImage(ctx context.Context, buildContextPath string, dockerFilePath string, buildProps *BuildProps, containerName string) (string, error) {
	log.Println("Building image from", buildContextPath)

	buildContext, err := archive.TarWithOptions(buildContextPath, &archive.TarOptions{})

	if err != nil {
		return "", fmt.Errorf("Failed to create tar archive from %s for Docker build context: %w", buildContextPath, err)
	}

	var tag string

	if containerName != "" {
		tag = fmt.Sprintf("%s-%s:%s", containerName, util.RandomString(6), "latest")
	} else {
		tag = fmt.Sprintf("%s:%s", util.RandomString(6), "latest")
	}

	imageBuildResponse, err := r.dockerClient.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
		Tags:       []string{tag},
		Dockerfile: dockerFilePath,
		BuildArgs:  buildProps.Args,
		Context:    buildContext,
		CacheFrom:  buildProps.CacheFrom,
		Target:     buildProps.Target,
		// NOTE: other options are ignored because in the devcontainer spec they are defined as []string and it's too cumbersome to parse them into types.ImageBuildOptions{}
	})

	if err != nil {
		return "", fmt.Errorf("Failed to build Docker image: %w", err)
	}

	defer imageBuildResponse.Body.Close()

	if err := util.StreamOutput(imageBuildResponse.Body, os.Stdout); err != nil {
		log.Printf("Error streaming output: %v\n", err)
	}

	log.Println("Built image with tag", tag)

	return tag, nil
}

func (r *DockerRunner) createContainer(ctx context.Context, image string) (string, error) {
	log.Println("Creating container...")

	// TODO: add other container parameters
	containerConfig := &container.Config{Image: image, Cmd: DefaultContainerCommand}
	createResponse, err := r.dockerClient.ContainerCreate(ctx, containerConfig, nil, nil, nil, "")

	if err != nil {
		return "", err
	}

	containerId := createResponse.ID

	log.Println("Created container", containerId)

	return containerId, nil
}

func (r *DockerRunner) startContainer(ctx context.Context, containerId string) error {
	log.Println("Starting container...")

	err := r.dockerClient.ContainerStart(ctx, containerId, container.StartOptions{})

	if err != nil {
		return err
	}

	log.Println("Started container", containerId)

	return nil
}
