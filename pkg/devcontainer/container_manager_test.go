package devcontainer_test

import (
	"bufio"
	"bytes"
	"context"
	"slices"
	"testing"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/devcontainer/mocks"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDockerContainerManager_CreateContainer(t *testing.T) {
	tests := []struct {
		name           string
		image          string
		projectPath    string
		config         devcontainer.Config
		mockSetup      func(*mocks.MockDockerContainerClient)
		expectedResult string
		expectedError  string
	}{
		{
			name:        "Create container with default configuration",
			image:       "test-image",
			projectPath: "/test/project",
			config:      devcontainer.Config{},
			mockSetup: func(m *mocks.MockDockerContainerClient) {
				m.On("ContainerCreate", mock.Anything, mock.MatchedBy(func(config *container.Config) bool {
					return config.Image == "test-image" &&
						slices.Equal(config.Cmd, devcontainer.DefaultContainerCommand) &&
						config.WorkingDir == "/workspace"
				}), mock.MatchedBy(func(hostConfig *container.HostConfig) bool {
					return slices.Equal(hostConfig.Mounts, []mount.Mount{
						{
							Type:   mount.TypeBind,
							Source: "/test/project",
							Target: "/workspace",
						},
					})
				}), mock.Anything, mock.Anything, "").
					Return(container.CreateResponse{ID: "test-container-id"}, nil)
			},
			expectedResult: "test-container-id",
		},
		{
			name:        "Create container with custom configuration",
			image:       "custom-image",
			projectPath: "/custom/project",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					WorkspaceMount: &devcontainer.Mount{
						Source:      "/custom/source",
						Destination: "/custom/destination",
					},
					WorkspaceFolder: "/custom/folder",
					AppPort:         []int{8080},
				},
				GeneralProperties: devcontainer.GeneralProperties{
					ContainerEnv:  map[string]string{"ENV1": "value1"},
					ContainerUser: "custom-user",
					Init:          true,
					Privileged:    true,
					CapAdd:        []string{"SYS_PTRACE"},
					SecurityOpt:   []string{"seccomp=unconfined"},
				},
			},
			mockSetup: func(m *mocks.MockDockerContainerClient) {
				m.On("ContainerCreate", mock.Anything, mock.MatchedBy(func(config *container.Config) bool {
					return config.Image == "custom-image" &&
						slices.Equal(config.Env, []string{"ENV1=value1"}) &&
						config.User == "custom-user" &&
						config.WorkingDir == "/custom/folder"
				}), mock.MatchedBy(func(hostConfig *container.HostConfig) bool {
					return slices.Equal(hostConfig.Mounts, []mount.Mount{
						{
							Type:   mount.TypeBind,
							Source: "/custom/source",
							Target: "/custom/destination",
						},
					}) &&
						*hostConfig.Init &&
						hostConfig.Privileged &&
						slices.Equal(hostConfig.CapAdd, []string{"SYS_PTRACE"}) &&
						slices.Equal(hostConfig.SecurityOpt, []string{"seccomp=unconfined"}) &&
						slices.Equal(hostConfig.PortBindings[nat.Port("8080/tcp")], []nat.PortBinding{
							{
								HostIP:   "127.0.0.1",
								HostPort: "8080",
							},
						})
				}), mock.Anything, mock.Anything, "").
					Return(container.CreateResponse{ID: "custom-container-id"}, nil)
			},
			expectedResult: "custom-container-id",
		},
		{
			name:        "Error creating container",
			image:       "error-image",
			projectPath: "/error/project",
			config:      devcontainer.Config{},
			mockSetup: func(m *mocks.MockDockerContainerClient) {
				m.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, "").
					Return(container.CreateResponse{}, assert.AnError)
			},
			expectedError: "assert.AnError general error for testing",
		},
		{
			name: "Unsupported mount type",
			config: devcontainer.Config{
				GeneralProperties: devcontainer.GeneralProperties{
					Mounts: []devcontainer.Mount{
						{
							Type:        "unsupported-type",
							Source:      "/test/source",
							Destination: "/test/destination",
						},
					},
				},
			},
			mockSetup:     func(m *mocks.MockDockerContainerClient) {},
			expectedError: "Failed to convert mount type unsupported-type to mount.Type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mocks.MockDockerContainerClient{}
			tt.mockSetup(mockClient)

			containerManager := devcontainer.NewDockerContainerManager(mockClient)

			result, err := containerManager.CreateContainer(context.Background(), tt.image, tt.projectPath, tt.config)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestDockerContainerManager_StartContainer(t *testing.T) {
	tests := []struct {
		name          string
		containerId   string
		mockSetup     func(*mocks.MockDockerContainerClient)
		expectedError string
	}{
		{
			name:        "Start container",
			containerId: "test-container-id",
			mockSetup: func(m *mocks.MockDockerContainerClient) {
				m.On("ContainerStart", mock.Anything, "test-container-id", mock.AnythingOfType("container.StartOptions")).Return(nil)
			},
		},
		{
			name:        "Error starting container",
			containerId: "error-container-id",
			mockSetup: func(m *mocks.MockDockerContainerClient) {
				m.On("ContainerStart", mock.Anything, "error-container-id", mock.AnythingOfType("container.StartOptions")).Return(assert.AnError)
			},
			expectedError: "assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mocks.MockDockerContainerClient{}
			tt.mockSetup(mockClient)

			containerManager := devcontainer.NewDockerContainerManager(mockClient)

			err := containerManager.StartContainer(context.Background(), tt.containerId)

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

func TestDockerContainerManager_StopContainer(t *testing.T) {
	tests := []struct {
		name          string
		containerId   string
		mockSetup     func(*mocks.MockDockerContainerClient)
		expectedError string
	}{
		{
			name:        "Stop container",
			containerId: "test-container-id",
			mockSetup: func(m *mocks.MockDockerContainerClient) {
				m.On("ContainerStop", mock.Anything, "test-container-id", mock.AnythingOfType("container.StopOptions")).Return(nil)
			},
		},
		{
			name:        "Error stopping container",
			containerId: "error-container-id",
			mockSetup: func(m *mocks.MockDockerContainerClient) {
				m.On("ContainerStop", mock.Anything, "error-container-id", mock.AnythingOfType("container.StopOptions")).Return(assert.AnError)
			},
			expectedError: "assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mocks.MockDockerContainerClient{}
			tt.mockSetup(mockClient)

			containerManager := devcontainer.NewDockerContainerManager(mockClient)

			err := containerManager.StopContainer(context.Background(), tt.containerId)

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

func TestDockerContainerManager_Exec(t *testing.T) {
	tests := []struct {
		name          string
		containerId   string
		command       []string
		mockSetup     func(*mocks.MockDockerContainerClient)
		expectedError string
	}{
		{
			name:        "Exec command",
			containerId: "test-container-id",
			command:     []string{"/bin/sh", "-c", "echo test"},
			mockSetup: func(m *mocks.MockDockerContainerClient) {
				m.On("ContainerExecCreate", mock.Anything, "test-container-id", mock.MatchedBy(func(config types.ExecConfig) bool {
					return slices.Equal(config.Cmd, []string{"/bin/sh", "-c", "echo test"}) &&
						config.AttachStdout &&
						config.AttachStderr
				})).Return(types.IDResponse{ID: "test-exec-id"}, nil)

				m.On("ContainerExecAttach", mock.Anything, "test-exec-id", mock.AnythingOfType("types.ExecStartCheck")).Return(types.HijackedResponse{Reader: bufio.NewReader(bytes.NewReader([]byte("test-stdout\ntest-stderr\n")))}, nil)

				m.On("ContainerExecInspect", mock.Anything, "test-exec-id").Return(types.ContainerExecInspect{ExitCode: 0}, nil)
			},
		},
		{
			name:        "Error creating exec configuration",
			containerId: "error-container-id",
			command:     []string{"/bin/sh", "-c", "echo test"},
			mockSetup: func(m *mocks.MockDockerContainerClient) {
				m.On("ContainerExecCreate", mock.Anything, "error-container-id", mock.MatchedBy(func(config types.ExecConfig) bool {
					return slices.Equal(config.Cmd, []string{"/bin/sh", "-c", "echo test"}) &&
						config.AttachStdout &&
						config.AttachStderr
				})).Return(types.IDResponse{}, assert.AnError)
			},
			expectedError: "assert.AnError general error for testing",
		},
		{
			name:        "Error attaching to exec process",
			containerId: "test-container-id",
			command:     []string{"/bin/sh", "-c", "echo test"},
			mockSetup: func(m *mocks.MockDockerContainerClient) {
				m.On("ContainerExecCreate", mock.Anything, "test-container-id", mock.MatchedBy(func(config types.ExecConfig) bool {
					return slices.Equal(config.Cmd, []string{"/bin/sh", "-c", "echo test"}) &&
						config.AttachStdout &&
						config.AttachStderr
				})).Return(types.IDResponse{ID: "test-exec-id"}, nil)
				m.On("ContainerExecAttach", mock.Anything, "test-exec-id", mock.AnythingOfType("types.ExecStartCheck")).Return(types.HijackedResponse{}, assert.AnError)
			},
			expectedError: "assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mocks.MockDockerContainerClient{}
			tt.mockSetup(mockClient)

			containerManager := devcontainer.NewDockerContainerManager(mockClient)

			_, err := containerManager.Exec(context.Background(), tt.containerId, tt.command)

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
