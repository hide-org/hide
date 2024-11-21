package server

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hide-org/hide/pkg/devcontainer/v2"
	"github.com/hide-org/hide/pkg/github"
)

type Installer interface {
	Install(ctx context.Context, container devcontainer.DevContainer) (int, error)
}

type InstallerImpl struct {
	binPath string
	githubClient github.Client
}

func NewInstaller(binaryPath string, githubClient github.Client) Installer {
	return &InstallerImpl{
		binPath: binaryPath,
		githubClient: githubClient,
	}
}

func (i *InstallerImpl) Install(ctx context.Context, container devcontainer.DevContainer) (int, error) {
	arch, err := devcontainer.GetArchitecture(ctx, container)
	if err != nil {
		return 0, fmt.Errorf("failed to determine container architecture: %w", err)
	}

	// Get latest release
	release, err := i.githubClient.GetLatestRelease()
	if err != nil {
		return 0, fmt.Errorf("failed to get latest release: %w", err)
	}

	// Get the specific asset URL
	downloadURL, err := release.GetAssetURL(arch)
	if err != nil {
		return 0, fmt.Errorf("failed to get download URL: %w", err)
	}

	// Download the binary
	_, err = container.Exec(ctx, []string{"curl", "-fsSL", "-o", i.binPath, downloadURL})
	if err != nil {
		return 0, fmt.Errorf("failed to download server binary: %w", err)
	}

	// Make binary executable
	_, err = container.Exec(ctx, []string{"chmod", "+x", i.binPath})
	if err != nil {
		return 0, fmt.Errorf("failed to make server binary executable: %w", err)
	}

	// Verify binary works
	result, err := container.Exec(ctx, []string{i.binPath, "version"})
	if err != nil {
		return 0, fmt.Errorf("failed to verify server binary: %w", err)
	}
	if result.ExitCode != 0 {
		return 0, fmt.Errorf("server binary verification failed: %s", result.StdErr)
	}

	// Start server
	processID, err := container.ExecDetached(ctx, []string{i.binPath, "server", "run"})
	if err != nil {
		return 0, fmt.Errorf("failed to start server: %w", err)
	}

	// Get server port from logs
	result, err = container.Exec(ctx, []string{"grep", "-m", "1", "Server started on", fmt.Sprintf("/proc/%s/fd/1", processID)})
	if err != nil {
		return 0, fmt.Errorf("failed to get server port: %w", err)
	}

	port, err := extractPort(result.StdOut)
	if err != nil {
		return 0, fmt.Errorf("failed to extract port from server output: %w", err)
	}

	return port, nil
}

func extractPort(output string) (int, error) {
	re := regexp.MustCompile(`Server started on .+:(\d+)`)
	matches := re.FindStringSubmatch(strings.TrimSpace(output))
	if len(matches) != 2 {
		return 0, fmt.Errorf("unexpected server output format: %s", output)
	}

	var port int
	_, err := fmt.Sscanf(matches[1], "%d", &port)
	if err != nil {
		return 0, fmt.Errorf("failed to parse port number: %w", err)
	}

	return port, nil
}
