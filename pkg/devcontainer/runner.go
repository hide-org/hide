package devcontainer

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/artmoskvin/hide/pkg/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

type Runner interface {
	Run(projectPath string, config Config) (string, error)
	Stop(containerId string) error
	Exec(containerId string, command string) (string, error)
}

type RunnerImpl struct {
	dockerClient    client.Client
	commandExecutor util.Executor
}

func NewRunnerImpl(client client.Client) Runner {
	return &RunnerImpl{
		dockerClient: client,
	}
}

func (r *RunnerImpl) Run(projectPath string, config Config) (string, error) {
	// Run initialize commands
	if config.LifecycleProps.InitializeCommand != nil {
		for _, command := range config.LifecycleProps.InitializeCommand {
			if err := r.commandExecutor.Run(command, projectPath, os.Stdout, os.Stderr); err != nil {
				return "", fmt.Errorf("Failed to run initialize command %s: %w", command, err)
			}
		}
	}

	// Pull image
	if config.Image != "" {
		log.Println("Pulling image", config.Image)
		output, err := r.dockerClient.ImagePull(context.Background(), config.Image, image.PullOptions{})
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
	if config.Build != nil || config.Dockerfile != "" {
		log.Println("Building image")
		buildContext, err := archive.TarWithOptions(projectPath, &archive.TarOptions{})

		if err != nil {
			return "", fmt.Errorf("Failed to create tar archive for Docker build context: %w", err)
		}

		imageBuildResponse, err := r.dockerClient.ImageBuild(context.Background(), buildContext, types.ImageBuildOptions{
			// TODO: add build args
			Tags:       []string{config.Image},
			Dockerfile: config.Build.Dockerfile,
			// Context:    config.Build.Context,
			// BuildArgs:  config.Build.Args,
			// Platform:   config.Build.Platform,
			// NoCache:    config.Build.NoCache,
		})

		if err != nil {
			return "", fmt.Errorf("Failed to build Docker image: %w", err)
		}

		defer imageBuildResponse.Body.Close()

		if err := util.StreamOutput(imageBuildResponse.Body, os.Stdout); err != nil {
			log.Printf("Error streaming output: %v\n", err)
		}

		log.Println("Built image", config.Image)
	}

	return "", nil
}

func (r *RunnerImpl) Stop(containerId string) error {
	return nil
}

func (r *RunnerImpl) Exec(containerId string, command string) (string, error) {
	return "", nil
}
