package devcontainer_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/devcontainer/mocks"
	"github.com/artmoskvin/hide/pkg/random"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/assert"
)

func TestDockerImageManager_PullOrBuildImage(t *testing.T) {
	tests := []struct {
		name           string
		config         devcontainer.Config
		credentials    devcontainer.RegistryCredentials
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
			expectedResult: "test:latest",
		},
		{
			name: "Build image from Dockerfile with name",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Build: &devcontainer.BuildProps{
						Dockerfile: "Dockerfile",
						Context:    ".",
					},
				},
				GeneralProperties: devcontainer.GeneralProperties{
					Name: "test-container",
				},
			},
			mockSetup: func(m *mocks.MockDockerClient) {
				m.ImageBuildFunc = func(ctx context.Context, dockerfile io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
					return types.ImageBuildResponse{Body: io.NopCloser(bytes.NewReader([]byte{}))}, nil
				}
			},
			expectedResult: "test-container-test:latest",
		},
		{
			name:   "Error when Dockerfile not found",
			config: devcontainer.Config{},
			mockSetup: func(m *mocks.MockDockerClient) {
			},
			expectedError: "Dockerfile not found",
		},
		{
			name: "Error building image",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Build: &devcontainer.BuildProps{
						Dockerfile: "Dockerfile",
						Context:    ".",
					},
				},
			},
			mockSetup: func(m *mocks.MockDockerClient) {
				m.ImageBuildFunc = func(ctx context.Context, dockerfile io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
					return types.ImageBuildResponse{}, errors.New("error building image")
				}
			},
			expectedError: "error building image",
		},
		{
			name: "Error pulling image",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Image: "test-image",
				},
			},
			mockSetup: func(m *mocks.MockDockerClient) {
				m.ImagePullFunc = func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
					return nil, errors.New("error pulling image")
				}
			},
			expectedError: "error pulling image",
		},
		{
			name: "Error pulling image because of invalid credentials",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Image: "test-image",
				},
			},
			mockSetup: func(m *mocks.MockDockerClient) {
				m.ImagePullFunc = func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
					return nil, errors.New("error pulling image")
				}
			},
			credentials: &mocks.MockRegistryCredentials{
				GetCredentialsFunc: func() (string, error) {
					return "", errors.New("error getting credentials")
				},
			},
			expectedError: "error getting credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mocks.MockDockerClient{}
			tt.mockSetup(mockClient)

			credentials := devcontainer.NewDockerHubRegistryCredentials("test-username", "test-password")
			if tt.credentials != nil {
				credentials = tt.credentials
			}

			imageManager := devcontainer.NewImageManager(mockClient, random.FixedString, credentials, devcontainer.NopLogger())

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
