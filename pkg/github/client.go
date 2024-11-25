package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client interface {
	GetLatestRelease() (*Release, error)
}

type ClientImpl struct {
	owner string
	repo  string
}

func (c *ClientImpl) GetLatestRelease() (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", c.owner, c.repo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	return &release, nil
}

func NewClient(owner, repo string) Client {
	return &ClientImpl{
		owner: owner,
		repo:  repo,
	}
}
