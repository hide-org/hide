package project_test

import (
	"reflect"
	"testing"

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
