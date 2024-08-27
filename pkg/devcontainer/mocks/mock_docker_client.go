package mocks

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// MockDockerClient is a mock implementation of the DockerClient interface for testing
type MockDockerClient struct {
	ContainerCreateFunc         func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
	ContainerStartFunc          func(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStopFunc           func(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerExecCreateFunc     func(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error)
	ContainerExecAttachFunc     func(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error)
	ContainerExecInspectFunc    func(ctx context.Context, execID string) (types.ContainerExecInspect, error)
	ImagePullFunc               func(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
	ImageBuildFunc              func(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
}

func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	return m.ContainerCreateFunc(ctx, config, hostConfig, networkingConfig, platform, containerName)
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	return m.ContainerStartFunc(ctx, containerID, options)
}

func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	return m.ContainerStopFunc(ctx, containerID, options)
}

func (m *MockDockerClient) ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error) {
	return m.ContainerExecCreateFunc(ctx, container, config)
}

func (m *MockDockerClient) ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error) {
	return m.ContainerExecAttachFunc(ctx, execID, config)
}

func (m *MockDockerClient) ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error) {
	return m.ContainerExecInspectFunc(ctx, execID)
}

func (m *MockDockerClient) ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
	return m.ImagePullFunc(ctx, refStr, options)
}

func (m *MockDockerClient) ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	return m.ImageBuildFunc(ctx, buildContext, options)
}
