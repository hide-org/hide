package devcontainer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/artmoskvin/hide/pkg/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/pkg/archive"
	"github.com/rs/zerolog/log"
)

type ImageManager interface {
	PullOrBuildImage(ctx context.Context, workingDir string, config Config) (string, error)
	PullImage(ctx context.Context, name string) error
	BuildImage(ctx context.Context, buildContextPath, dockerFilePath string, buildProps *BuildProps, containerName string) (string, error)
}

type DockerImageManager struct {
	dockerClient DockerClient
	config       DockerRunnerConfig
}

func NewImageManager(dockerClient DockerClient, config DockerRunnerConfig) ImageManager {
	return &DockerImageManager{dockerClient: dockerClient, config: config}
}

func (im *DockerImageManager) PullOrBuildImage(ctx context.Context, workingDir string, config Config) (string, error) {
	if config.Image != "" {
		if err := im.PullImage(ctx, config.Image); err != nil {
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
		// NOTE: this is bad; default values should be set during parsing
		// default value
		context = "."
	}

	dockerFilePath := filepath.Join(workingDir, config.Path, dockerFile)
	contextPath := filepath.Join(workingDir, config.Path, context)
	imageId, err := im.BuildImage(ctx, contextPath, dockerFilePath, config.Build, config.Name)

	if err != nil {
		return "", fmt.Errorf("Failed to build image: %w", err)
	}

	return imageId, nil
}

func (im *DockerImageManager) PullImage(ctx context.Context, name string) error {
	log.Debug().Str("image", name).Msg("Pulling image")

	authStr, err := im.encodeRegistryAuth(im.config.Username, im.config.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to encode registry auth")
		return fmt.Errorf("Failed to encode registry auth: %w", err)
	}

	output, err := im.dockerClient.ImagePull(ctx, name, image.PullOptions{RegistryAuth: authStr})
	if err != nil {
		return err
	}
	defer output.Close()

	if err := util.ReadOutput(output, os.Stdout); err != nil {
		log.Error().Err(err).Msg("Error streaming output")
	}

	log.Debug().Str("image", name).Msg("Pulled image")
	return nil
}

func (im *DockerImageManager) BuildImage(ctx context.Context, buildContextPath, dockerFilePath string, buildProps *BuildProps, containerName string) (string, error) {
	log.Debug().Str("buildContextPath", buildContextPath).Msg("Building image")
	log.Warn().Msg("Building images is not stable yet")

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

	imageBuildResponse, err := im.dockerClient.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
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

	if err := util.ReadOutput(imageBuildResponse.Body, os.Stdout); err != nil {
		log.Error().Err(err).Msg("Error streaming output")
	}

	log.Debug().Str("tag", tag).Msg("Built image")
	return tag, nil
}

func (im *DockerImageManager) encodeRegistryAuth(username, password string) (string, error) {
	authConfig := registry.AuthConfig{
		Username: username,
		Password: password,
	}

	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}

	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	return authStr, nil
}