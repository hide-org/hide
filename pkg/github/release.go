package github

import (
	"fmt"
	"strings"
)

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func (r *Release) GetAssetURL(arch string) (string, error) {
	assetName := fmt.Sprintf("hide_%s", arch)
	for _, asset := range r.Assets {
		if strings.Contains(asset.Name, assetName) {
			return asset.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("no matching binary found for architecture: %s", arch)
}
