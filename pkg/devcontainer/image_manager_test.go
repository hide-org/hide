package devcontainer_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/devcontainer/mocks"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/assert"
)

func TestDockerImageManager_PullOrBuildImage(t *testing.T) {
	tests := []struct {
		name           string
		config         devcontainer.Config
		mockSetup      func(*mocks.MockDockerClient)
		expectedResult string
		expectedError  string
	}{
		{
			name: "Pull existing image",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Image: "test-image",
				},
			},
			mockSetup: func(m *mocks.MockDockerClient) {
				m.ImagePullFunc = func(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
					return io.NopCloser(bytes.NewReader([]byte{})), nil
				}
			},
			expectedResult: "test-image",
		},
		{
			name: "Build image from Dockerfile",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Build: &devcontainer.BuildProps{
						Dockerfile: "Dockerfile",
						Context:    ".",
					},
				},
			},
			mockSetup: func(m *mocks.MockDockerClient) {
				m.ImageBuildFunc = func(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
					return types.ImageBuildResponse{Body: io.NopCloser(bytes.NewReader([]byte{}))}, nil
				}
			},
			expectedResult: "mock-image-id",
		},
		{
			name:   "Error when Dockerfile not found",
			config: devcontainer.Config{},
			mockSetup: func(m *mocks.MockDockerClient) {
			},
			expectedError: "Dockerfile not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mocks.MockDockerClient{}
			tt.mockSetup(mockClient)

			imageManager := devcontainer.NewImageManager(mockClient, devcontainer.DockerRunnerConfig{})

			result, err := imageManager.PullOrBuildImage(context.Background(), "/workdir", tt.config)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestDockerImageManager_PullImage(t *testing.T) {
	tests := []struct {
		name          string
		imageName     string
		mockSetup     func(*mocks.MockDockerClient)
		expectedError string
	}{
		{
			name:      "Successfully pull image",
			imageName: "test-image:latest",
			mockSetup: func(m *mocks.MockDockerClient) {
				m.ImagePullFunc = func(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
					return io.NopCloser(bytes.NewReader([]byte{})), nil
				}
			},
		},
		{
			name:      "Error pulling image",
			imageName: "non-existent-image:latest",
			mockSetup: func(m *mocks.MockDockerClient) {
				m.ImagePullFunc = func(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
					return nil, errors.New("image not found")
				}
			},
			expectedError: "image not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mocks.MockDockerClient{}
			tt.mockSetup(mockClient)

			imageManager := devcontainer.NewImageManager(mockClient, devcontainer.DockerRunnerConfig{})

			err := imageManager.PullImage(context.Background(), tt.imageName)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDockerImageManager_BuildImage(t *testing.T) {
	tests := []struct {
		name             string
		buildContextPath string
		dockerFilePath   string
		buildProps       *devcontainer.BuildProps
		containerName    string
		mockSetup        func(*mocks.MockDockerClient)
		expectedImageID  string
		expectedError    string
	}{
		{
			name:             "Successfully build image",
			buildContextPath: "/test/context",
			dockerFilePath:   "/test/Dockerfile",
			buildProps:       &devcontainer.BuildProps{},
			containerName:    "test-container",
			mockSetup: func(m *mocks.MockDockerClient) {
				m.ImageBuildFunc = func(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
					return types.ImageBuildResponse{Body: io.NopCloser(bytes.NewReader([]byte{}))}, nil
				}
			},
			expectedImageID: "test-container-",
		},
		{
			name:             "Error building image",
			buildContextPath: "/test/context",
			dockerFilePath:   "/test/Dockerfile",
			buildProps:       &devcontainer.BuildProps{},
			containerName:    "test-container",
			mockSetup: func(m *mocks.MockDockerClient) {
				m.ImageBuildFunc = func(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
					return types.ImageBuildResponse{}, errors.New("build failed")
				}
			},
			expectedError: "Failed to build Docker image: build failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mocks.MockDockerClient{}
			tt.mockSetup(mockClient)

			imageManager := devcontainer.NewImageManager(mockClient, devcontainer.DockerRunnerConfig{})

			imageID, err := imageManager.BuildImage(context.Background(), tt.buildContextPath, tt.dockerFilePath, tt.buildProps, tt.containerName)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, imageID, tt.expectedImageID)
			}
		})
	}
}
