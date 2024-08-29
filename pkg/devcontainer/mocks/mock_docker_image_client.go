package mocks

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/mock"
)

var _ client.ImageAPIClient = (*MockDockerImageClient)(nil)

type MockDockerImageClient struct {
	mock.Mock
}

func (m *MockDockerImageClient) ImageBuild(ctx context.Context, context io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	args := m.Called(ctx, context, options)
	return args.Get(0).(types.ImageBuildResponse), args.Error(1)
}

func (m *MockDockerImageClient) BuildCachePrune(ctx context.Context, opts types.BuildCachePruneOptions) (*types.BuildCachePruneReport, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*types.BuildCachePruneReport), args.Error(1)
}

func (m *MockDockerImageClient) BuildCancel(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDockerImageClient) ImageCreate(ctx context.Context, parentReference string, options image.CreateOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, parentReference, options)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockDockerImageClient) ImageHistory(ctx context.Context, img string) ([]image.HistoryResponseItem, error) {
	args := m.Called(ctx, img)
	return args.Get(0).([]image.HistoryResponseItem), args.Error(1)
}

func (m *MockDockerImageClient) ImageImport(ctx context.Context, source types.ImageImportSource, ref string, options image.ImportOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, source, ref, options)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockDockerImageClient) ImageInspectWithRaw(ctx context.Context, image string) (types.ImageInspect, []byte, error) {
	args := m.Called(ctx, image)
	return args.Get(0).(types.ImageInspect), args.Get(1).([]byte), args.Error(2)
}

func (m *MockDockerImageClient) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]image.Summary), args.Error(1)
}

func (m *MockDockerImageClient) ImageLoad(ctx context.Context, input io.Reader, quiet bool) (types.ImageLoadResponse, error) {
	args := m.Called(ctx, input, quiet)
	return args.Get(0).(types.ImageLoadResponse), args.Error(1)
}

func (m *MockDockerImageClient) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, ref, options)
	var r io.ReadCloser
	if rf, ok := args.Get(0).(io.ReadCloser); ok {
		r = rf
	}
	return r, args.Error(1)
}

func (m *MockDockerImageClient) ImagePush(ctx context.Context, ref string, options image.PushOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, ref, options)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockDockerImageClient) ImageRemove(ctx context.Context, img string, options image.RemoveOptions) ([]image.DeleteResponse, error) {
	args := m.Called(ctx, img, options)
	return args.Get(0).([]image.DeleteResponse), args.Error(1)
}

func (m *MockDockerImageClient) ImageSearch(ctx context.Context, term string, options types.ImageSearchOptions) ([]registry.SearchResult, error) {
	args := m.Called(ctx, term, options)
	return args.Get(0).([]registry.SearchResult), args.Error(1)
}

func (m *MockDockerImageClient) ImageSave(ctx context.Context, images []string) (io.ReadCloser, error) {
	args := m.Called(ctx, images)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockDockerImageClient) ImageTag(ctx context.Context, image string, ref string) error {
	args := m.Called(ctx, image, ref)
	return args.Error(0)
}

func (m *MockDockerImageClient) ImagesPrune(ctx context.Context, pruneFilter filters.Args) (types.ImagesPruneReport, error) {
	args := m.Called(ctx, pruneFilter)
	return args.Get(0).(types.ImagesPruneReport), args.Error(1)
}
