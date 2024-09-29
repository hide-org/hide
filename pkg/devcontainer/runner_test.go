package devcontainer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/devcontainer/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDockerRunner_Run(t *testing.T) {
	tests := []struct {
		name       string
		config     devcontainer.Config
		setupMocks func(*mocks.MockExecutor, *mocks.MockImageManager, *mocks.MockContainerManager)
		wantResult string
		wantError  string
	}{
		{
			name: "Successful run with local image",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Image: "test-image",
				},
			},
			setupMocks: func(me *mocks.MockExecutor, mim *mocks.MockImageManager, mcm *mocks.MockContainerManager) {
				mim.On("LocalImageExists", mock.Anything, "test-image").Return(true, nil)
				mcm.On("CreateContainer", mock.Anything, "test-image", mock.Anything, mock.Anything).Return("container-id", nil)
				mcm.On("StartContainer", mock.Anything, "container-id").Return(nil)
			},
			wantResult: "container-id",
		},
		{
			name: "Successful run with pulled image",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Image: "test-image",
				},
			},
			setupMocks: func(me *mocks.MockExecutor, mim *mocks.MockImageManager, mcm *mocks.MockContainerManager) {
				mim.On("LocalImageExists", mock.Anything, "test-image").Return(false, nil)
				mim.On("PullImage", mock.Anything, "test-image").Return(nil)
				mcm.On("CreateContainer", mock.Anything, "test-image", mock.Anything, mock.Anything).Return("container-id", nil)
				mcm.On("StartContainer", mock.Anything, "container-id").Return(nil)
			},
			wantResult: "container-id",
		},
		{
			name: "Successful run with Dockerfile",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Dockerfile: "Dockerfile",
				},
			},
			setupMocks: func(me *mocks.MockExecutor, mim *mocks.MockImageManager, mcm *mocks.MockContainerManager) {
				mim.On("BuildImage", mock.Anything, mock.Anything, mock.Anything).Return("built-image-id", nil)
				mcm.On("CreateContainer", mock.Anything, "built-image-id", mock.Anything, mock.Anything).Return("container-id", nil)
				mcm.On("StartContainer", mock.Anything, "container-id").Return(nil)
			},
			wantResult: "container-id",
		},
		{
			name: "Failed with docker compose",
			config: devcontainer.Config{
				DockerComposeProps: devcontainer.DockerComposeProps{
					DockerComposeFile: []string{"docker-compose.yml"},
					Service:           "test-service",
				},
			},
			setupMocks: func(me *mocks.MockExecutor, mim *mocks.MockImageManager, mcm *mocks.MockContainerManager) {},
			wantError:  "Docker Compose is not supported yet",
		},
		{
			name:       "Failed with invalid devcontainer",
			config:     devcontainer.Config{},
			setupMocks: func(me *mocks.MockExecutor, mim *mocks.MockImageManager, mcm *mocks.MockContainerManager) {},
			wantError:  "Invalid devcontainer configuration",
		},
		{
			name: "Failed local image check",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Image: "test-image",
				},
			},
			setupMocks: func(me *mocks.MockExecutor, mim *mocks.MockImageManager, mcm *mocks.MockContainerManager) {
				mim.On("LocalImageExists", mock.Anything, "test-image").Return(false, errors.New("check error"))
			},
			wantError: "Failed to check if image test-image exists: check error",
		},
		{
			name: "Failed image pull",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Image: "test-image",
				},
			},
			setupMocks: func(me *mocks.MockExecutor, mim *mocks.MockImageManager, mcm *mocks.MockContainerManager) {
				mim.On("LocalImageExists", mock.Anything, "test-image").Return(false, nil)
				mim.On("PullImage", mock.Anything, "test-image").Return(errors.New("pull error"))
			},
			wantError: "Failed to pull image test-image: pull error",
		},
		{
			name: "Failed image build",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Dockerfile: "Dockerfile",
				},
			},
			setupMocks: func(me *mocks.MockExecutor, mim *mocks.MockImageManager, mcm *mocks.MockContainerManager) {
				mim.On("BuildImage", mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("build error"))
			},
			wantError: "Failed to build image: build error",
		},
		{
			name: "Failed container creation",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Image: "test-image",
				},
			},
			setupMocks: func(me *mocks.MockExecutor, mim *mocks.MockImageManager, mcm *mocks.MockContainerManager) {
				mim.On("LocalImageExists", mock.Anything, "test-image").Return(true, nil)
				mcm.On("CreateContainer", mock.Anything, "test-image", mock.Anything, mock.Anything).Return("", errors.New("create error"))
			},
			wantError: "Failed to create container: create error",
		},
		{
			name: "Failed container start",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Image: "test-image",
				},
			},
			setupMocks: func(me *mocks.MockExecutor, mim *mocks.MockImageManager, mcm *mocks.MockContainerManager) {
				mim.On("LocalImageExists", mock.Anything, "test-image").Return(true, nil)
				mcm.On("CreateContainer", mock.Anything, "test-image", mock.Anything, mock.Anything).Return("container-id", nil)
				mcm.On("StartContainer", mock.Anything, "container-id").Return(errors.New("start error"))
			},
			wantError: "Failed to start container: start error",
		},
		{
			name: "Successful run with lifecycle commands",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Image: "test-image",
				},
				LifecycleProps: devcontainer.LifecycleProps{
					InitializeCommand:    devcontainer.LifecycleCommand{"command": []string{"initialize"}},
					OnCreateCommand:      devcontainer.LifecycleCommand{"command": []string{"onCreate"}},
					UpdateContentCommand: devcontainer.LifecycleCommand{"command": []string{"updateContent"}},
					PostCreateCommand:    devcontainer.LifecycleCommand{"command": []string{"postCreate"}},
					PostStartCommand:     devcontainer.LifecycleCommand{"command": []string{"postStart"}},
					PostAttachCommand:    devcontainer.LifecycleCommand{"command": []string{"postAttach"}},
				},
			},
			setupMocks: func(me *mocks.MockExecutor, mim *mocks.MockImageManager, mcm *mocks.MockContainerManager) {
				mim.On("LocalImageExists", mock.Anything, "test-image").Return(true, nil)
				mcm.On("CreateContainer", mock.Anything, "test-image", mock.Anything, mock.Anything).Return("container-id", nil)
				mcm.On("StartContainer", mock.Anything, "container-id").Return(nil)
				me.On("Run", []string{"initialize"}, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mcm.On("Exec", mock.Anything, "container-id", []string{"onCreate"}).Return(devcontainer.ExecResult{ExitCode: 0}, nil)
				mcm.On("Exec", mock.Anything, "container-id", []string{"updateContent"}).Return(devcontainer.ExecResult{ExitCode: 0}, nil)
				mcm.On("Exec", mock.Anything, "container-id", []string{"postCreate"}).Return(devcontainer.ExecResult{ExitCode: 0}, nil)
				me.On("Run", []string{"postStart"}, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				me.On("Run", []string{"postAttach"}, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			wantResult: "container-id",
		},
		{
			name: "Failed lifecycle command",
			config: devcontainer.Config{
				DockerImageProps: devcontainer.DockerImageProps{
					Image: "test-image",
				},
				LifecycleProps: devcontainer.LifecycleProps{
					OnCreateCommand: devcontainer.LifecycleCommand{"test": []string{"test", "command"}},
				},
			},
			setupMocks: func(me *mocks.MockExecutor, mim *mocks.MockImageManager, mcm *mocks.MockContainerManager) {
				mim.On("LocalImageExists", mock.Anything, "test-image").Return(true, nil)
				mcm.On("CreateContainer", mock.Anything, "test-image", mock.Anything, mock.Anything).Return("container-id", nil)
				mcm.On("StartContainer", mock.Anything, "container-id").Return(nil)
				mcm.On("Exec", mock.Anything, "container-id", []string{"test", "command"}).Return(devcontainer.ExecResult{ExitCode: 1, StdErr: "command failed"}, nil)
			},
			wantError: "Failed to run command test [test command] in container container-id",
		},
	}

	for _, tt := range tests {
		mockExecutor := &mocks.MockExecutor{}
		mockImageManager := &mocks.MockImageManager{}
		mockContainerManager := &mocks.MockContainerManager{}

		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(mockExecutor, mockImageManager, mockContainerManager)
			runner := devcontainer.NewDockerRunner(mockExecutor, mockImageManager, mockContainerManager)
			result, err := runner.Run(context.Background(), "/test/project", tt.config)

			if tt.wantError != "" {
				assert.Contains(t, err.Error(), tt.wantError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}

			mockExecutor.AssertExpectations(t)
			mockImageManager.AssertExpectations(t)
			mockContainerManager.AssertExpectations(t)
		})
	}
}

func TestDockerRunnerStop(t *testing.T) {
	tests := []struct {
		name        string
		containerID string
		setupMocks  func(*mocks.MockContainerManager)
		wantError   string
	}{
		{
			name:        "Successful stop",
			containerID: "test-container-id",
			setupMocks: func(mcm *mocks.MockContainerManager) {
				mcm.On("StopContainer", mock.Anything, "test-container-id").Return(nil)
			},
			wantError: "",
		},
		{
			name:        "Failed stop",
			containerID: "failed-container-id",
			setupMocks: func(mcm *mocks.MockContainerManager) {
				mcm.On("StopContainer", mock.Anything, "failed-container-id").Return(errors.New("failed to stop container"))
			},
			wantError: "failed to stop container",
		},
	}

	for _, tt := range tests {
		mockExecutor := &mocks.MockExecutor{}
		mockImageManager := &mocks.MockImageManager{}
		mockContainerManager := &mocks.MockContainerManager{}

		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(mockContainerManager)
			runner := devcontainer.NewDockerRunner(mockExecutor, mockImageManager, mockContainerManager)
			err := runner.Stop(context.Background(), tt.containerID)

			if tt.wantError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
			} else {
				assert.NoError(t, err)
			}

			mockContainerManager.AssertExpectations(t)
		})
	}
}

func TestDockerRunnerExec(t *testing.T) {
	tests := []struct {
		name        string
		containerID string
		command     []string
		setupMocks  func(*mocks.MockContainerManager)
		wantResult  devcontainer.ExecResult
		wantError   string
	}{
		{
			name:        "Successful exec",
			containerID: "test-container-id",
			command:     []string{"echo", "hello"},
			setupMocks: func(mcm *mocks.MockContainerManager) {
				mcm.On("Exec", mock.Anything, "test-container-id", []string{"echo", "hello"}).Return(devcontainer.ExecResult{ExitCode: 0, StdOut: "hello\n"}, nil)
			},
			wantResult: devcontainer.ExecResult{ExitCode: 0, StdOut: "hello\n"},
			wantError:  "",
		},
		{
			name:        "Failed exec",
			containerID: "failed-container-id",
			command:     []string{"non-existent-command"},
			setupMocks: func(mcm *mocks.MockContainerManager) {
				mcm.On("Exec", mock.Anything, "failed-container-id", []string{"non-existent-command"}).Return(devcontainer.ExecResult{}, errors.New("command not found"))
			},
			wantResult: devcontainer.ExecResult{},
			wantError:  "command not found",
		},
		{
			name:        "Exec with non-zero exit code",
			containerID: "exit-code-container-id",
			command:     []string{"exit", "1"},
			setupMocks: func(mcm *mocks.MockContainerManager) {
				mcm.On("Exec", mock.Anything, "exit-code-container-id", []string{"exit", "1"}).Return(devcontainer.ExecResult{ExitCode: 1, StdErr: "Exit status 1"}, nil)
			},
			wantResult: devcontainer.ExecResult{ExitCode: 1, StdErr: "Exit status 1"},
			wantError:  "",
		},
	}

	for _, tt := range tests {
		mockExecutor := &mocks.MockExecutor{}
		mockImageManager := &mocks.MockImageManager{}
		mockContainerManager := &mocks.MockContainerManager{}

		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(mockContainerManager)
			runner := devcontainer.NewDockerRunner(mockExecutor, mockImageManager, mockContainerManager)
			result, err := runner.Exec(context.Background(), tt.containerID, tt.command)

			if tt.wantError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}

			mockContainerManager.AssertExpectations(t)
		})
	}
}
