package mocks

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/mock"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var _ client.ContainerAPIClient = (*MockDockerContainerClient)(nil)

type MockDockerContainerClient struct {
	mock.Mock
}

func (m *MockDockerContainerClient) ContainerAttach(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
	args := m.Called(ctx, container, options)
	return args.Get(0).(types.HijackedResponse), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerCommit(ctx context.Context, container string, options container.CommitOptions) (types.IDResponse, error) {
	args := m.Called(ctx, container, options)
	return args.Get(0).(types.IDResponse), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	args := m.Called(ctx, config, hostConfig, networkingConfig, platform, containerName)
	return args.Get(0).(container.CreateResponse), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerDiff(ctx context.Context, cntnr string) ([]container.FilesystemChange, error) {
	args := m.Called(ctx, cntnr)
	return args.Get(0).([]container.FilesystemChange), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error) {
	args := m.Called(ctx, execID, config)
	return args.Get(0).(types.HijackedResponse), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error) {
	args := m.Called(ctx, container, config)
	return args.Get(0).(types.IDResponse), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error) {
	args := m.Called(ctx, execID)
	return args.Get(0).(types.ContainerExecInspect), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerExecResize(ctx context.Context, execID string, options container.ResizeOptions) error {
	args := m.Called(ctx, execID, options)
	return args.Error(0)
}

func (m *MockDockerContainerClient) ContainerExecStart(ctx context.Context, execID string, config types.ExecStartCheck) error {
	args := m.Called(ctx, execID, config)
	return args.Error(0)
}

func (m *MockDockerContainerClient) ContainerExport(ctx context.Context, container string) (io.ReadCloser, error) {
	args := m.Called(ctx, container)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error) {
	args := m.Called(ctx, container)
	return args.Get(0).(types.ContainerJSON), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerInspectWithRaw(ctx context.Context, container string, getSize bool) (types.ContainerJSON, []byte, error) {
	args := m.Called(ctx, container, getSize)
	return args.Get(0).(types.ContainerJSON), args.Get(1).([]byte), args.Error(2)
}

func (m *MockDockerContainerClient) ContainerKill(ctx context.Context, container string, signal string) error {
	args := m.Called(ctx, container, signal)
	return args.Error(0)
}

func (m *MockDockerContainerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]types.Container), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, container, options)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerPause(ctx context.Context, container string) error {
	args := m.Called(ctx, container)
	return args.Error(0)
}

func (m *MockDockerContainerClient) ContainerRemove(ctx context.Context, container string, options container.RemoveOptions) error {
	args := m.Called(ctx, container, options)
	return args.Error(0)
}

func (m *MockDockerContainerClient) ContainerRename(ctx context.Context, container string, newContainerName string) error {
	args := m.Called(ctx, container, newContainerName)
	return args.Error(0)
}

func (m *MockDockerContainerClient) ContainerResize(ctx context.Context, container string, options container.ResizeOptions) error {
	args := m.Called(ctx, container, options)
	return args.Error(0)
}

func (m *MockDockerContainerClient) ContainerRestart(ctx context.Context, container string, options container.StopOptions) error {
	args := m.Called(ctx, container, options)
	return args.Error(0)
}

func (m *MockDockerContainerClient) ContainerStatPath(ctx context.Context, container string, path string) (types.ContainerPathStat, error) {
	args := m.Called(ctx, container, path)
	return args.Get(0).(types.ContainerPathStat), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerStats(ctx context.Context, container string, stream bool) (types.ContainerStats, error) {
	args := m.Called(ctx, container, stream)
	return args.Get(0).(types.ContainerStats), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerStatsOneShot(ctx context.Context, container string) (types.ContainerStats, error) {
	args := m.Called(ctx, container)
	return args.Get(0).(types.ContainerStats), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerStart(ctx context.Context, container string, options container.StartOptions) error {
	args := m.Called(ctx, container, options)
	return args.Error(0)
}

func (m *MockDockerContainerClient) ContainerStop(ctx context.Context, container string, options container.StopOptions) error {
	args := m.Called(ctx, container, options)
	return args.Error(0)
}

func (m *MockDockerContainerClient) ContainerTop(ctx context.Context, cntnr string, arguments []string) (container.ContainerTopOKBody, error) {
	args := m.Called(ctx, cntnr, arguments)
	return args.Get(0).(container.ContainerTopOKBody), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerUnpause(ctx context.Context, container string) error {
	args := m.Called(ctx, container)
	return args.Error(0)
}

func (m *MockDockerContainerClient) ContainerUpdate(ctx context.Context, cntnr string, updateConfig container.UpdateConfig) (container.ContainerUpdateOKBody, error) {
	args := m.Called(ctx, cntnr, updateConfig)
	return args.Get(0).(container.ContainerUpdateOKBody), args.Error(1)
}

func (m *MockDockerContainerClient) ContainerWait(ctx context.Context, cntnr string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	args := m.Called(ctx, cntnr, condition)
	return args.Get(0).(<-chan container.WaitResponse), args.Get(1).(<-chan error)
}

func (m *MockDockerContainerClient) CopyFromContainer(ctx context.Context, container string, srcPath string) (io.ReadCloser, types.ContainerPathStat, error) {
	args := m.Called(ctx, container, srcPath)
	return args.Get(0).(io.ReadCloser), args.Get(1).(types.ContainerPathStat), args.Error(2)
}

func (m *MockDockerContainerClient) CopyToContainer(ctx context.Context, container string, path string, content io.Reader, options types.CopyToContainerOptions) error {
	args := m.Called(ctx, container, path, content, options)
	return args.Error(0)
}

func (m *MockDockerContainerClient) ContainersPrune(ctx context.Context, pruneFilters filters.Args) (types.ContainersPruneReport, error) {
	args := m.Called(ctx, pruneFilters)
	return args.Get(0).(types.ContainersPruneReport), args.Error(1)
}
