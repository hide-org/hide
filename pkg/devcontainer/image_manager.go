package devcontainer

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/rs/zerolog/log"
)

type ImageManager interface {
	PullImage(ctx context.Context, name string) error
	BuildImage(ctx context.Context, workingDir string, config Config) (string, error)
}

type DockerImageManager struct {
	client.ImageAPIClient
	randomString func(int) string
	credentials  RegistryCredentials
	// logger for docker build and pull logs
	logger Logger
}

func NewImageManager(dockerImageCli client.ImageAPIClient, randomString func(int) string, credentials RegistryCredentials, logger Logger) ImageManager {
	return &DockerImageManager{ImageAPIClient: dockerImageCli, randomString: randomString, credentials: credentials, logger: logger}
}

func (im *DockerImageManager) PullImage(ctx context.Context, name string) error {
	log.Debug().Str("image", name).Msg("Pulling image")

	authStr, err := im.credentials.GetCredentials()
	if err != nil {
		log.Error().Err(err).Msg("Failed to encode registry auth")
		return fmt.Errorf("Failed to encode registry auth: %w", err)
	}

	output, err := im.ImagePull(ctx, name, image.PullOptions{RegistryAuth: authStr})
	if err != nil {
		return err
	}
	defer output.Close()

	im.logger.Log(output)

	log.Debug().Str("image", name).Msg("Pulled image")
	return nil
}

func (im *DockerImageManager) BuildImage(ctx context.Context, workingDir string, config Config) (string, error) {
	var dockerFile, context string

	if config.Dockerfile != "" {
		dockerFile = config.Dockerfile
	} 

	if config.Build != nil && config.Build.Dockerfile != "" {
		dockerFile = config.Build.Dockerfile
	}

	if dockerFile == "" {
		return "", fmt.Errorf("Dockerfile not found")
	}

	if config.Context != "" {
		context = config.Context
	}

	if config.Build != nil && config.Build.Context != "" {
		context = config.Build.Context
	}

	if context == "" {
		context = "."
	}

	dockerFilePath := filepath.Join(workingDir, config.Path, dockerFile)
	contextPath := filepath.Join(workingDir, config.Path, context)
	dockerFileRelativePath, err := filepath.Rel(contextPath, dockerFilePath)
	if err != nil {
		return "", fmt.Errorf("Failed to get relative path of Dockerfile: %w", err)
	}

	log.Debug().Str("buildContextPath", contextPath).Msg("Building image")

	buildContext, err := archive.TarWithOptions(contextPath, &archive.TarOptions{})
	if err != nil {
		return "", fmt.Errorf("Failed to create tar archive from %s for Docker build context: %w", contextPath, err)
	}

	var tag string
	if config.Name != "" {
		containerName := sanitizeContainerName(config.Name)
		tag = fmt.Sprintf("%s-%s:%s", containerName, im.randomString(6), "latest")
	} else {
		tag = fmt.Sprintf("%s:%s", im.randomString(6), "latest")
	}
	tag = strings.ToLower(tag)

	options := types.ImageBuildOptions{
		Tags:       []string{tag},
		Dockerfile: dockerFileRelativePath,
		Context:    buildContext,
	}

	if config.Build != nil {
		options.BuildArgs = config.Build.Args
		options.CacheFrom = config.Build.CacheFrom
		options.Target = config.Build.Target
		// NOTE: config.Build.RunArgs are ignored because they are defined as []string and it's not obvious how to map them into types.ImageBuildOptions{}
	}

	imageBuildResponse, err := im.ImageBuild(ctx, buildContext, options)
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
