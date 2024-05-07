package devcontainer_test

import (
	"github.com/artmoskvin/hide/pkg/devcontainer"
	"testing"
)

func TestCliManager_StartContainer(t *testing.T) {
	t.Skip("Skipping test because it depends on external shell command `devcontainer`")
}

func TestCliManager_FindContainerByProject(t *testing.T) {
	container := devcontainer.Container{Id: "test-container", ProjectId: "test-project"}
	cliManager := devcontainer.CliManager{Store: devcontainer.NewInMemoryStore(map[string]*devcontainer.Container{"test-container": &container})}
	foundContainer, err := cliManager.FindContainerByProject("test-project")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if foundContainer != container {
		t.Errorf("Expected container id to be %s, got %s", container.Id, foundContainer.Id)
	}
}

func TestCliManager_StopContainer(t *testing.T) {
	t.Skip("Skipping test because it depends on external shell command `docker`")
}

func TestCliManager_Exec(t *testing.T) {
	t.Skip("Skipping test because it depends on external shell command `devcontainer`")
}
