package devcontainer

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/archive"
	"github.com/rs/zerolog/log"
)

type ImageManager interface {
	PullOrBuildImage(ctx context.Context, workingDir string, config Config) (string, error)
}

type DockerImageManager struct {
	dockerClient DockerClient
	randomString func(int) string
	credentials  RegistryCredentials
	// logger for docker build and pull logs
	logger Logger
}

func NewImageManager(dockerClient DockerClient, randomString func(int) string, credentials RegistryCredentials, logger Logger) ImageManager {
	return &DockerImageManager{dockerClient: dockerClient, randomString: randomString, credentials: credentials, logger: logger}
}

func (im *DockerImageManager) PullOrBuildImage(ctx context.Context, workingDir string, config Config) (string, error) {
	if config.Image != "" {
		if err := im.pullImage(ctx, config.Image); err != nil {
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

	context = config.Context
	if config.Build != nil {
		context = config.Build.Context
	}

	dockerFilePath := filepath.Join(workingDir, config.Path, dockerFile)
	contextPath := filepath.Join(workingDir, config.Path, context)
	imageId, err := im.buildImage(ctx, contextPath, dockerFilePath, config.Build, config.Name)

	if err != nil {
		return "", fmt.Errorf("Failed to build image: %w", err)
	}

	return imageId, nil
}

func (im *DockerImageManager) pullImage(ctx context.Context, name string) error {
	log.Debug().Str("image", name).Msg("Pulling image")

	authStr, err := im.credentials.GetCredentials()
	if err != nil {
		log.Error().Err(err).Msg("Failed to encode registry auth")
		return fmt.Errorf("Failed to encode registry auth: %w", err)
	}

	output, err := im.dockerClient.ImagePull(ctx, name, image.PullOptions{RegistryAuth: authStr})
	if err != nil {
		return err
	}
	defer output.Close()

	im.logger.Log(output)

	log.Debug().Str("image", name).Msg("Pulled image")
	return nil
}

func (im *DockerImageManager) buildImage(ctx context.Context, buildContextPath, dockerFilePath string, buildProps *BuildProps, containerName string) (string, error) {
	log.Warn().Msg("Building images is not stable yet")
	log.Debug().Str("buildContextPath", buildContextPath).Msg("Building image")

	buildContext, err := archive.TarWithOptions(buildContextPath, &archive.TarOptions{})
	if err != nil {
		return "", fmt.Errorf("Failed to create tar archive from %s for Docker build context: %w", buildContextPath, err)
	}

	var tag string
	if containerName != "" {
		containerName = sanitizeContainerName(containerName)
		tag = fmt.Sprintf("%s-%s:%s", containerName, im.randomString(6), "latest")
	} else {
		tag = fmt.Sprintf("%s:%s", im.randomString(6), "latest")
	}
	tag = strings.ToLower(tag)

	imageBuildResponse, err := im.dockerClient.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
		Tags:       []string{tag},
		Dockerfile: dockerFilePath,
		BuildArgs:  buildProps.Args,
		Context:    buildContext,
		CacheFrom:  buildProps.CacheFrom,
		Target:     buildProps.Target,
		// NOTE: other options are ignored because in the devcontainer spec they are defined as []string and it's not obvious how to parse them into types.ImageBuildOptions{}
	})
	if err != nil {
		return "", fmt.Errorf("Failed to build Docker image: %w", err)
	}
	defer imageBuildResponse.Body.Close()

	im.logger.Log(imageBuildResponse.Body)

	log.Debug().Str("tag", tag).Msg("Built image")
	return tag, nil
}

func sanitizeContainerName(containerName string) string {
	containerName = strings.ReplaceAll(containerName, " ", "-")
	return containerName
}
