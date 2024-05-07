package project_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/devcontainer/mocks"
	"github.com/artmoskvin/hide/pkg/project"
)

func TestProject_findTaskByAlias(t *testing.T) {
	project := project.Project{
		Config: project.DevContainerConfig{
			Tasks: []project.Task{
				{Alias: "test-task", Command: "echo test"},
			},
		},
	}

	task, err := project.FindTaskByAlias("test-task")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if task.Alias != "test-task" {
		t.Errorf("Expected task alias to be test-task, got %s", task.Alias)
	}
}

func TestProject_findTaskByAlias_notFound(t *testing.T) {
	project := project.Project{
		Config: project.DevContainerConfig{
			Tasks: []project.Task{
				{Alias: "test-task", Command: "echo test"},
			},
		},
	}

	_, err := project.FindTaskByAlias("missing-task")

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestManagerImpl_CreateProject(t *testing.T) {
	t.Skip("Skipping test because it depends on external shell command `git` and file system")
}

func TestManagerImpl_GetProject_Succeeds(t *testing.T) {
	_project := project.Project{Id: "test-project", Path: "/tmp/test-project", Config: project.DevContainerConfig{}}
	pm := project.NewProjectManager(nil, project.NewInMemoryStore(map[string]*project.Project{"test-project": &_project}), "/tmp")
	project, err := pm.GetProject("test-project")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(project, _project) {
		t.Errorf("Expected project id to be test-project, got %s", project.Id)
	}
}

func TestManagerImpl_GetProject_Fails(t *testing.T) {
	pm := project.NewProjectManager(nil, project.NewInMemoryStore(map[string]*project.Project{}), "/tmp")
	_, err := pm.GetProject("missing-project")

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestManagerImpl_ResolveTaskAlias_Succeeds(t *testing.T) {
	task := project.Task{Alias: "test-alias", Command: "echo test"}
	_project := project.Project{Id: "test-project", Path: "/tmp/test-project", Config: project.DevContainerConfig{Tasks: []project.Task{task}}}
	pm := project.NewProjectManager(nil, project.NewInMemoryStore(map[string]*project.Project{"test-project": &_project}), "/tmp")
	resolvedTask, err := pm.ResolveTaskAlias("test-project", "test-alias")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(resolvedTask, task) {
		t.Errorf("Expected task alias to be test-alias, got %s", resolvedTask.Alias)
	}
}

func TestManagerImpl_ResolveTaskAlias_ProjectNotFound(t *testing.T) {
	pm := project.NewProjectManager(nil, project.NewInMemoryStore(map[string]*project.Project{}), "/tmp")
	_, err := pm.ResolveTaskAlias("missing-project", "test-alias")

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestManagerImpl_ResolveTaskAlias_TaskNotFound(t *testing.T) {
	_project := project.Project{Id: "test-project", Path: "/tmp/test-project", Config: project.DevContainerConfig{}}
	pm := project.NewProjectManager(nil, project.NewInMemoryStore(map[string]*project.Project{"test-project": &_project}), "/tmp")
	_, err := pm.ResolveTaskAlias("test-project", "missing-alias")

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestManagerImpl_CreateTask(t *testing.T) {
	const projectId = "test-project"
	_project := project.Project{Id: projectId, Path: "/tmp/test-project", Config: project.DevContainerConfig{}}
	container := devcontainer.Container{Id: "test-container", ProjectId: projectId}
	devContainerManager := &mocks.MockDevContainerManager{
		FindContainerByProjectFunc: func(projectId string) (devcontainer.Container, error) {
			return container, nil
		},
		ExecFunc: func(containerId string, projectPath string, command string) (devcontainer.ExecResult, error) {
			return devcontainer.ExecResult{StdOut: "test-stdout", StdErr: "test-stderr", ExitCode: 1}, nil
		}}
	pm := project.NewProjectManager(devContainerManager, project.NewInMemoryStore(map[string]*project.Project{projectId: &_project}), "/tmp")

	taskResult, err := pm.CreateTask(projectId, "echo test")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedTaskResult := project.TaskResult{StdOut: "test-stdout", StdErr: "test-stderr", ExitCode: 1}

	if !reflect.DeepEqual(taskResult, expectedTaskResult) {
		t.Errorf("Expected empty stdout, got %s", taskResult.StdOut)
	}
}

func TestManagerImpl_CreateTask_ProjectNotFound(t *testing.T) {
	pm := project.NewProjectManager(nil, project.NewInMemoryStore(map[string]*project.Project{}), "/tmp")
	_, err := pm.CreateTask("missing-project", "echo test")

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestManagerImpl_CreateTask_ContainerNotFound(t *testing.T) {
	const projectId = "test-project"
	_project := project.Project{Id: projectId, Path: "/tmp/test-project", Config: project.DevContainerConfig{}}
	devContainerManager := &mocks.MockDevContainerManager{
		FindContainerByProjectFunc: func(projectId string) (devcontainer.Container, error) {
			return devcontainer.Container{}, errors.New("container not found")
		},
	}
	pm := project.NewProjectManager(devContainerManager, project.NewInMemoryStore(map[string]*project.Project{projectId: &_project}), "/tmp")

	_, err := pm.CreateTask(projectId, "echo test")

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestManagerImpl_CreateTask_ExecError(t *testing.T) {
	const projectId = "test-project"
	_project := project.Project{Id: projectId, Path: "/tmp/test-project", Config: project.DevContainerConfig{}}
	container := devcontainer.Container{Id: "test-container", ProjectId: projectId}
	devContainerManager := &mocks.MockDevContainerManager{
		FindContainerByProjectFunc: func(projectId string) (devcontainer.Container, error) {
			return container, nil
		},
		ExecFunc: func(containerId string, projectPath string, command string) (devcontainer.ExecResult, error) {
			return devcontainer.ExecResult{}, errors.New("exec error")
		},
	}
	pm := project.NewProjectManager(devContainerManager, project.NewInMemoryStore(map[string]*project.Project{projectId: &_project}), "/tmp")

	_, err := pm.CreateTask(projectId, "echo test")

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
