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
	"github.com/stretchr/testify/mock"
)

func TestDockerImageManager_PullImage(t *testing.T) {
	tests := []struct {
		name          string
		imageName     string
		credentials   devcontainer.RegistryCredentials
		mockSetup     func(*mocks.MockDockerImageClient)
		expectedError string
	}{
		{
			name:      "Pull existing image",
			imageName: "test-image",
			mockSetup: func(m *mocks.MockDockerImageClient) {
				m.On("ImagePull", mock.Anything, "test-image", mock.AnythingOfType("image.PullOptions")).
					Return(io.NopCloser(bytes.NewReader([]byte{})), nil)
			},
		},
		{
			name:      "Error pulling image",
			imageName: "test-image",
			mockSetup: func(m *mocks.MockDockerImageClient) {
				m.On("ImagePull", mock.Anything, "test-image", mock.AnythingOfType("image.PullOptions")).
					Return(nil, errors.New("error pulling image"))
			},
			expectedError: "error pulling image",
		},
		{
			name:      "Error pulling image because of invalid credentials",
			imageName: "test-image",
			mockSetup: func(m *mocks.MockDockerImageClient) {},
			credentials: &mocks.MockRegistryCredentials{
				GetCredentialsFunc: func() (string, error) {
					return "", errors.New("error getting credentials")
				},
			},
			expectedError: "Failed to encode registry auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mocks.MockDockerImageClient{}
			tt.mockSetup(mockClient)

			credentials := devcontainer.NewDockerHubRegistryCredentials("test-username", "test-password")
			if tt.credentials != nil {
				credentials = tt.credentials
			}

			imageManager := devcontainer.NewImageManager(mockClient, random.FixedString, credentials)

			err := imageManager.PullImage(context.Background(), tt.imageName)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestDockerImageManager_BuildImage(t *testing.T) {
	tests := []struct {
		name           string
		config         devcontainer.Config
		mockSetup      func(*mocks.MockDockerImageClient)
		expectedResult string
		expectedError  string
	}{
		{
			name: "Build image",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Dockerfile: "Dockerfile",
					Context:    ".",
				},
			},
			mockSetup: func(m *mocks.MockDockerImageClient) {
				m.On("ImageBuild", mock.Anything, mock.AnythingOfType("*io.PipeReader"), mock.MatchedBy(func(options types.ImageBuildOptions) bool {
					return options.Dockerfile == "Dockerfile" && len(options.Tags) > 0 && options.Tags[0] == "test:latest"
				})).Return(types.ImageBuildResponse{Body: io.NopCloser(bytes.NewReader([]byte{}))}, nil)
			},
			expectedResult: "test:latest",
		},
		{
			name: "Build image with name",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Dockerfile: "Dockerfile",
					Context:    ".",
				},
				GeneralProperties: devcontainer.GeneralProperties{
					Name: "test-container",
				},
			},
			mockSetup: func(m *mocks.MockDockerImageClient) {
				m.On("ImageBuild", mock.Anything, mock.AnythingOfType("*io.PipeReader"), mock.MatchedBy(func(options types.ImageBuildOptions) bool {
					return options.Dockerfile == "Dockerfile" && len(options.Tags) > 0 && options.Tags[0] == "test-container-test:latest"
				})).Return(types.ImageBuildResponse{Body: io.NopCloser(bytes.NewReader([]byte{}))}, nil)
			},
			expectedResult: "test-container-test:latest",
		},
		{
			name: "Build image from build props",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Build: &devcontainer.BuildProps{
						Dockerfile: "Dockerfile",
						Context:    ".",
					},
				},
			},
			mockSetup: func(m *mocks.MockDockerImageClient) {
				m.On("ImageBuild", mock.Anything, mock.AnythingOfType("*io.PipeReader"), mock.MatchedBy(func(options types.ImageBuildOptions) bool {
					return options.Dockerfile == "Dockerfile" && len(options.Tags) > 0 && options.Tags[0] == "test:latest"
				})).Return(types.ImageBuildResponse{Body: io.NopCloser(bytes.NewReader([]byte{}))}, nil)
			},
			expectedResult: "test:latest",
		},
		{
			name:          "Error when Dockerfile not found",
			config:        devcontainer.Config{},
			mockSetup:     func(m *mocks.MockDockerImageClient) {},
			expectedError: "Dockerfile not found",
		},
		{
			name: "Error building image",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Dockerfile: "Dockerfile",
					Context:    ".",
				},
			},
			mockSetup: func(m *mocks.MockDockerImageClient) {
				m.On("ImageBuild", mock.Anything, mock.AnythingOfType("*io.PipeReader"), mock.AnythingOfType("types.ImageBuildOptions")).
					Return(types.ImageBuildResponse{}, errors.New("error building image"))
			},
			expectedError: "Failed to build Docker image",
		},
		{
			name: "Build image with custom build args",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Dockerfile: "Dockerfile",
					Context:    ".",
					Build: &devcontainer.BuildProps{
						Args: map[string]*string{
							"ARG1": strPtr("value1"),
							"ARG2": strPtr("value2"),
						},
					},
				},
			},
			mockSetup: func(m *mocks.MockDockerImageClient) {
				m.On("ImageBuild", mock.Anything, mock.AnythingOfType("*io.PipeReader"), mock.MatchedBy(func(options types.ImageBuildOptions) bool {
					return options.Dockerfile == "Dockerfile" &&
						*options.BuildArgs["ARG1"] == "value1" &&
						*options.BuildArgs["ARG2"] == "value2"
				})).Return(types.ImageBuildResponse{Body: io.NopCloser(bytes.NewReader([]byte{}))}, nil)
			},
			expectedResult: "test:latest",
		},
		{
			name: "Build image with custom cacheFrom",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Dockerfile: "Dockerfile",
					Context:    ".",
					Build: &devcontainer.BuildProps{
						CacheFrom: []string{"cache1", "cache2"},
					},
				},
			},
			mockSetup: func(m *mocks.MockDockerImageClient) {
				m.On("ImageBuild", mock.Anything, mock.AnythingOfType("*io.PipeReader"), mock.MatchedBy(func(options types.ImageBuildOptions) bool {
					return options.Dockerfile == "Dockerfile" &&
						len(options.CacheFrom) == 2 &&
						options.CacheFrom[0] == "cache1" &&
						options.CacheFrom[1] == "cache2"
				})).Return(types.ImageBuildResponse{Body: io.NopCloser(bytes.NewReader([]byte{}))}, nil)
			},
			expectedResult: "test:latest",
		},
		{
			name: "Build image with custom target",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Dockerfile: "Dockerfile",
					Context:    ".",
					Build: &devcontainer.BuildProps{
						Target: "custom-target",
					},
				},
			},
			mockSetup: func(m *mocks.MockDockerImageClient) {
				m.On("ImageBuild", mock.Anything, mock.AnythingOfType("*io.PipeReader"), mock.MatchedBy(func(options types.ImageBuildOptions) bool {
					return options.Dockerfile == "Dockerfile" &&
						options.Target == "custom-target"
				})).Return(types.ImageBuildResponse{Body: io.NopCloser(bytes.NewReader([]byte{}))}, nil)
			},
			expectedResult: "test:latest",
		},
		{
			name: "Build image with non-default docker context",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Dockerfile: "Dockerfile",
					Context:    "..",
				},
			},
			mockSetup: func(m *mocks.MockDockerImageClient) {
				m.On("ImageBuild", mock.Anything, mock.AnythingOfType("*io.PipeReader"), mock.MatchedBy(func(options types.ImageBuildOptions) bool {
					return options.Dockerfile == "workdir/Dockerfile"
				})).Return(types.ImageBuildResponse{Body: io.NopCloser(bytes.NewReader([]byte{}))}, nil)
			},
			expectedResult: "test:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mocks.MockDockerImageClient{}
			tt.mockSetup(mockClient)

			imageManager := devcontainer.NewImageManager(mockClient, random.FixedString, nil)

			result, err := imageManager.BuildImage(context.Background(), "/workdir", tt.config)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, result, tt.expectedResult)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestDockerImageManager_CheckLocalImage(t *testing.T) {
	tests := []struct {
		name           string
		image          string
		mockSetup      func(m *mocks.MockDockerImageClient)
		expectedResult bool
		expectedError  string
	}{
		{
			name:  "image exists",
			image: "test-image",
			mockSetup: func(m *mocks.MockDockerImageClient) {
				m.On("ImageList", mock.Anything, mock.Anything).
					Return([]image.Summary{{ID: "test-image"}}, nil)
			},
			expectedResult: true,
		},
		{
			name:  "image doesn't exist",
			image: "test-image",
			mockSetup: func(m *mocks.MockDockerImageClient) {
				m.On("ImageList", mock.Anything, mock.Anything).
					Return([]image.Summary{}, nil)
			},
			expectedResult: false,
		},
		{
			name:  "error",
			image: "test-image",
			mockSetup: func(m *mocks.MockDockerImageClient) {
				m.On("ImageList", mock.Anything, mock.Anything).
					Return([]image.Summary{}, errors.New("test error"))
			},
			expectedResult: false,
			expectedError:  "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mocks.MockDockerImageClient{}
			tt.mockSetup(mockClient)

			imageManager := devcontainer.NewImageManager(mockClient, random.String, nil)

			result, err := imageManager.CheckLocalImage(context.Background(), tt.image)

			if tt.expectedResult {
				assert.True(t, result)
			} else {
				assert.False(t, result)
			}

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func strPtr(s string) *string {
	return &s
}
