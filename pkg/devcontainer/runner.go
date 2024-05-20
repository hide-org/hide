package devcontainer

import (
	"context"
	"fmt"

	// "io"
	"log"
	"os"

	// "os/exec"

	"github.com/artmoskvin/hide/pkg/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	// "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	// v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type Runner interface {
	Run(projectPath string, config *Config) (string, error)
	Stop(containerId string) error
	Exec(containerId string, command string) (string, error)
}

type RunnerImpl struct {
	dockerClient    *client.Client
	commandExecutor util.Executor
}

func NewRunnerImpl(client *client.Client, commandExecutor util.Executor) Runner {
	return &RunnerImpl{
		dockerClient:    client,
		commandExecutor: commandExecutor,
	}
}

func (r *RunnerImpl) Run(projectPath string, config *Config) (string, error) {
	// Run initialize commands
	if config.LifecycleProps.InitializeCommand != nil {
		for _, command := range config.LifecycleProps.InitializeCommand {
			if err := r.commandExecutor.Run(command, projectPath, os.Stdout, os.Stderr); err != nil {
				return "", fmt.Errorf("Failed to run initialize command %s: %w", command, err)
			}
		}
	}

	ctx := context.Background()

	// Pull image
	if config.Image != "" {
		log.Println("Pulling image...", config.Image)
		output, err := r.dockerClient.ImagePull(ctx, config.Image, image.PullOptions{})
		if err != nil {
			return "", fmt.Errorf("Failed to pull image %s: %w", config.Image, err)
		}
		defer output.Close()

		if err := util.StreamOutput(output, os.Stdout); err != nil {
			log.Printf("Error streaming output: %v\n", err)
		}

		log.Println("Pulled image", config.Image)
	}

	// Build image
	if config.Dockerfile != "" || (config.Build != nil && config.Build.Dockerfile != "") {
		log.Println("Building image...")
		buildContext, err := archive.TarWithOptions(projectPath, &archive.TarOptions{})

		if err != nil {
			return "", fmt.Errorf("Failed to create tar archive for Docker build context: %w", err)
		}

		var dockerFile string

		if config.Dockerfile != "" {
			dockerFile = config.Dockerfile
		} else if config.Build != nil && config.Build.Dockerfile != "" {
			dockerFile = config.Build.Dockerfile
		}

		imageBuildResponse, err := r.dockerClient.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
			// TODO: add build args
			Dockerfile: dockerFile,
			BuildArgs:  config.Build.Args,
			Context:    buildContext,
			CacheFrom:  config.Build.CacheFrom,
			Target:     config.Build.Target,
		})

		if err != nil {
			return "", fmt.Errorf("Failed to build Docker image: %w", err)
		}

		defer imageBuildResponse.Body.Close()

		if err := util.StreamOutput(imageBuildResponse.Body, os.Stdout); err != nil {
			log.Printf("Error streaming output: %v\n", err)
		}

		log.Println("Built image")
	}

	// Build docker compose
	if len(config.DockerComposeFile) > 0 {
		log.Println("Building docker-compose...")
		// TODO: build docker-compose file
	}

	containerConfig := &container.Config{Image: config.Image, Cmd: []string{"/bin/sh", "-c", "while sleep 1000; do :; done"}}
	createResponse, err := r.dockerClient.ContainerCreate(ctx, containerConfig, nil, nil, nil, "")

	if err != nil {
		return "", fmt.Errorf("Failed to create container: %w", err)
	}

	containerId := createResponse.ID

	log.Println("Created container", containerId)

	// start container

	if err := r.dockerClient.ContainerStart(ctx, containerId, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("Failed to start container: %w", err)
	}

	log.Println("Started container", containerId)

	return containerId, nil
}

func (r *RunnerImpl) Stop(containerId string) error {
	return nil
}

func (r *RunnerImpl) Exec(containerId string, command string) (string, error) {
	return "", nil
}
