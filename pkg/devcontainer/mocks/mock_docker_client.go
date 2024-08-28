package mocks

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/mock"
)

type MockDockerClient struct {
	mock.Mock
}

func (m *MockDockerClient) ContainerAttach(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerCommit(ctx context.Context, container string, options container.CommitOptions) (types.IDResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerDiff(ctx context.Context, container string) ([]container.FilesystemChange, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerExecResize(ctx context.Context, execID string, options container.ResizeOptions) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerExecStart(ctx context.Context, execID string, config types.ExecStartCheck) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerExport(ctx context.Context, container string) (io.ReadCloser, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerInspectWithRaw(ctx context.Context, container string, getSize bool) (types.ContainerJSON, []byte, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerKill(ctx context.Context, container string, signal string) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerPause(ctx context.Context, container string) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerRemove(ctx context.Context, container string, options container.RemoveOptions) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerRename(ctx context.Context, container string, newContainerName string) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerResize(ctx context.Context, container string, options container.ResizeOptions) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerRestart(ctx context.Context, container string, options container.StopOptions) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerStatPath(ctx context.Context, container string, path string) (types.ContainerPathStat, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerStats(ctx context.Context, container string, stream bool) (types.ContainerStats, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerStatsOneShot(ctx context.Context, container string) (types.ContainerStats, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, container string, options container.StartOptions) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerStop(ctx context.Context, container string, options container.StopOptions) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerTop(ctx context.Context, container string, arguments []string) (container.ContainerTopOKBody, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerUnpause(ctx context.Context, container string) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerUpdate(ctx context.Context, container string, updateConfig container.UpdateConfig) (container.ContainerUpdateOKBody, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainerWait(ctx context.Context, container string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) CopyFromContainer(ctx context.Context, container string, srcPath string) (io.ReadCloser, types.ContainerPathStat, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) CopyToContainer(ctx context.Context, container string, path string, content io.Reader, options types.CopyToContainerOptions) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerClient) ContainersPrune(ctx context.Context, pruneFilters filters.Args) (types.ContainersPruneReport, error) {
	panic("not implemented") // TODO: Implement
}

// // MockDockerClient is a mock implementation of the DockerClient interface for testing
// type MockDockerClient struct {
// 	ContainerCreateFunc         func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
// 	ContainerStartFunc          func(ctx context.Context, containerID string, options container.StartOptions) error
// 	ContainerStopFunc           func(ctx context.Context, containerID string, options container.StopOptions) error
// 	ContainerExecCreateFunc     func(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error)
// 	ContainerExecAttachFunc     func(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error)
// 	ContainerExecInspectFunc    func(ctx context.Context, execID string) (types.ContainerExecInspect, error)
// 	ImagePullFunc               func(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
// 	ImageBuildFunc              func(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
// }

// func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
// 	return m.ContainerCreateFunc(ctx, config, hostConfig, networkingConfig, platform, containerName)
// }

// func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
// 	return m.ContainerStartFunc(ctx, containerID, options)
// }

// func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
// 	return m.ContainerStopFunc(ctx, containerID, options)
// }

// func (m *MockDockerClient) ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error) {
// 	return m.ContainerExecCreateFunc(ctx, container, config)
// }

// func (m *MockDockerClient) ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error) {
// 	return m.ContainerExecAttachFunc(ctx, execID, config)
// }

// func (m *MockDockerClient) ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error) {
// 	return m.ContainerExecInspectFunc(ctx, execID)
// }

// func (m *MockDockerClient) ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
// 	return m.ImagePullFunc(ctx, refStr, options)
// }

// func (m *MockDockerClient) ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
// 	return m.ImageBuildFunc(ctx, buildContext, options)
// }
