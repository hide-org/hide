package server

import (
	"context"
	"fmt"

	"github.com/hide-org/hide/pkg/github"
)

type ReleaseProvider interface {
	GetDownloadURL(ctx context.Context, arch string) (string, error)
}

type GithubReleaseProvider struct {
	githubClient github.Client
}

func NewGithubReleaseProvider(client github.Client) ReleaseProvider {
	return &GithubReleaseProvider{githubClient: client}
}

func (g *GithubReleaseProvider) GetDownloadURL(ctx context.Context, arch string) (string, error) {
	release, err := g.githubClient.GetLatestRelease()
	if err != nil {
		return "", fmt.Errorf("failed to get latest release: %w", err)
	}
	
	return release.GetAssetURL(arch)
}

type StaticReleaseProvider struct {
	downloadURL string
}

func NewStaticReleaseProvider(url string) ReleaseProvider {
	return &StaticReleaseProvider{downloadURL: url}
}

func (s *StaticReleaseProvider) GetDownloadURL(ctx context.Context, arch string) (string, error) {
	return s.downloadURL, nil
} 