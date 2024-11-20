package server

import (
	"fmt"
	"os"
	"path/filepath"
)

type Architecture string

const (
	AMD64 Architecture = "amd64"
	ARM64 Architecture = "arm64"
)

type Downloader interface {
	// Download downloads or retrieves cached binary for specific architecture
	Download(name string, arch Architecture) (string, error)
	// GetContainerPath returns the path where binary should be mounted in container
	GetContainerPath(name string) string
}

type BinaryManagerImpl struct {
	binaryCache map[string]string
	downloadFn  func(name string, arch Architecture) (string, error)
}

func NewBinaryManager(downloadFn func(string, Architecture) (string, error)) Downloader {
	return &BinaryManagerImpl{
		binaryCache: make(map[string]string),
		downloadFn:  downloadFn,
	}
}

func (m *BinaryManagerImpl) Download(name string, arch Architecture) (string, error) {
	cacheKey := fmt.Sprintf("%s_%s", name, arch)
	if path, ok := m.binaryCache[cacheKey]; ok {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	path, err := m.downloadFn(name, arch)
	if err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
	}

	m.binaryCache[cacheKey] = path
	return path, nil
}

func (m *BinaryManagerImpl) GetContainerPath(name string) string {
	return filepath.Join("/usr/local/bin", name)
} 
