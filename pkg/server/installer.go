package server

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/hide-org/hide/pkg/workspaces"
)

type Installer interface {
	Install(ctx context.Context, workspace workspaces.Workspace) (int, error)
}

type InstallerImpl struct {
	binPath         string
	releaseProvider ReleaseProvider
	sshOptions      []workspaces.SshOption
}

func NewInstaller(binaryPath string, releaseProvider ReleaseProvider, sshOptions []workspaces.SshOption) Installer {
	return &InstallerImpl{
		binPath:         binaryPath,
		releaseProvider: releaseProvider,
		sshOptions:      sshOptions,
	}
}

func (i *InstallerImpl) Install(ctx context.Context, workspace workspaces.Workspace) (int, error) {
	arch, err := workspaces.GetArchitecture(ctx, workspace)
	if err != nil {
		return 0, fmt.Errorf("failed to determine container architecture: %w", err)
	}

	// Get the specific asset URL
	downloadURL, err := i.releaseProvider.GetDownloadURL(ctx, arch)
	if err != nil {
		return 0, fmt.Errorf("failed to get download URL: %w", err)
	}

	// Create temp directory and write script
	tempDir := "/tmp/hide-installer"
	if _, err := workspace.Ssh(ctx, "mkdir -p "+tempDir); err != nil {
		return 0, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		// Clean up temp directory
		if _, err := workspace.Ssh(ctx, "rm -rf "+tempDir); err != nil {
			fmt.Printf("Warning: failed to remove temp directory: %v\n", err)
		}
	}()

	scriptPath := filepath.Join(tempDir, "install.sh")
	// Use the installScript constant
	cmd := fmt.Sprintf(`cat > %s << 'EOL'
%s
EOL`, scriptPath, installScript)

	if _, err := workspace.Ssh(ctx, cmd); err != nil {
		return 0, fmt.Errorf("failed to write install script: %w", err)
	}

	// Make the script executable
	if _, err := workspace.Ssh(ctx, "chmod +x "+scriptPath); err != nil {
		return 0, fmt.Errorf("failed to make install script executable: %w", err)
	}

	// Run the install script
	output, err := workspace.Ssh(ctx, fmt.Sprintf(
		"%s -b %s -u %s",
		scriptPath,
		i.binPath,
		downloadURL,
	), i.sshOptions...)
	if err != nil {
		return 0, fmt.Errorf("installation failed: %w", err)
	}

	// Extract port from output
	port, err := extractPort(output)
	if err != nil {
		return 0, fmt.Errorf("failed to extract port from server output: %w", err)
	}

	return port, nil
}

func extractPort(output string) (int, error) {
	// Split output into lines and find the most relevant one
	lines := strings.Split(output, "\n")
	var portLine string
	// First try "Server started successfully" as it's our final confirmation
	for _, line := range lines {
		if strings.Contains(line, "Server started successfully on port") {
			portLine = line
			break
		}
	}
	// Fallback to "Using port" if server start confirmation not found
	if portLine == "" {
		for _, line := range lines {
			if strings.Contains(line, "Using port:") {
				portLine = line
				break
			}
		}
	}

	if portLine == "" {
		return 0, fmt.Errorf("could not find port information in output:\n%s", output)
	}

	re := regexp.MustCompile(`\d+`)
	portStr := re.FindString(portLine)
	if portStr == "" {
		return 0, fmt.Errorf("could not extract port number from line: %s", portLine)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse port number '%s': %w", portStr, err)
	}

	return port, nil
}
