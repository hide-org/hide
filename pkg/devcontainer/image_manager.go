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
}

func NewImageManager(dockerImageCli client.ImageAPIClient, randomString func(int) string, credentials RegistryCredentials) ImageManager {
	return &DockerImageManager{ImageAPIClient: dockerImageCli, randomString: randomString, credentials: credentials}
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

	if err := logResponse(output); err != nil {
		log.Error().Err(err)
	}

	log.Debug().Str("image", name).Msg("Pulled image")
	return nil
}

func (im *DockerImageManager) BuildImage(ctx context.Context, workingDir string, config Config) (string, error) {
	var dockerFile, context string

	dockerFile = config.Dockerfile
	if config.Build != nil {
		dockerFile = config.Build.Dockerfile
	}

	if dockerFile == "" {
		return "", fmt.Errorf("Dockerfile not found")
	}

	context = config.Context
	if config.Build != nil {
		context = config.Build.Context
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

	imageBuildResponse, err := im.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
		Tags:       []string{tag},
		Dockerfile: dockerFileRelativePath,
		BuildArgs:  config.Build.Args,
		Context:    buildContext,
		CacheFrom:  config.Build.CacheFrom,
		Target:     config.Build.Target,
		// NOTE: other options are ignored because in the devcontainer spec they are defined as []string and it's not obvious how to parse them into types.ImageBuildOptions{}
	})
	if err != nil {
		return "", fmt.Errorf("Failed to build Docker image: %w", err)
	}
	defer imageBuildResponse.Body.Close()

	if err := logResponse(imageBuildResponse.Body); err != nil {
		log.Error().Err(err)
	}

	log.Debug().Str("tag", tag).Msg("Built image")
	return tag, nil
}

func sanitizeContainerName(containerName string) string {
	containerName = strings.ReplaceAll(containerName, " ", "-")
	return containerName
}
