package project

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

const ProjectsDir = "hide-projects"

type ProjectManager struct {
	projectParentDir string
	dirName          string
	projectPath      string
}

func NewProjectManager() *ProjectManager {
	return &ProjectManager{}
}

func (pm *ProjectManager) CreateProjectDir() (string, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return "", fmt.Errorf("Failed to get user home directory: %w", err)
	}

	projectParentDir := fmt.Sprintf("%s/%s", home, ProjectsDir)
	dirName := randomString(10)
	pm.projectPath = fmt.Sprintf("%s/%s", projectParentDir, dirName)

	if err := os.MkdirAll(pm.projectPath, 0755); err != nil {
		return "", fmt.Errorf("Failed to create project directory: %w", err)
	}

	fmt.Println("Created project directory: ", pm.projectPath)

	return pm.projectPath, nil
}

func (pm *ProjectManager) RemoveProjectDir() error {
	if err := os.RemoveAll(pm.projectPath); err != nil {
		return fmt.Errorf("Failed to remove project directory: %w", err)
	}

	fmt.Println("Removed project directory: ", pm.projectPath)

	return nil
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
