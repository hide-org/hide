package mocks

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
)

type MockDockerImageClient struct{}

func (m *MockDockerImageClient) ImageBuild(ctx context.Context, context io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerImageClient) BuildCachePrune(ctx context.Context, opts types.BuildCachePruneOptions) (*types.BuildCachePruneReport, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerImageClient) BuildCancel(ctx context.Context, id string) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerImageClient) ImageCreate(ctx context.Context, parentReference string, options image.CreateOptions) (io.ReadCloser, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerImageClient) ImageHistory(ctx context.Context, image string) ([]image.HistoryResponseItem, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerImageClient) ImageImport(ctx context.Context, source types.ImageImportSource, ref string, options image.ImportOptions) (io.ReadCloser, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerImageClient) ImageInspectWithRaw(ctx context.Context, image string) (types.ImageInspect, []byte, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerImageClient) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerImageClient) ImageLoad(ctx context.Context, input io.Reader, quiet bool) (types.ImageLoadResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerImageClient) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerImageClient) ImagePush(ctx context.Context, ref string, options image.PushOptions) (io.ReadCloser, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerImageClient) ImageRemove(ctx context.Context, image string, options image.RemoveOptions) ([]image.DeleteResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerImageClient) ImageSearch(ctx context.Context, term string, options types.ImageSearchOptions) ([]registry.SearchResult, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerImageClient) ImageSave(ctx context.Context, images []string) (io.ReadCloser, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerImageClient) ImageTag(ctx context.Context, image string, ref string) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockDockerImageClient) ImagesPrune(ctx context.Context, pruneFilter filters.Args) (types.ImagesPruneReport, error) {
	panic("not implemented") // TODO: Implement
}
