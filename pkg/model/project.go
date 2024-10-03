package model

import (
	"context"
	"fmt"

	"github.com/hide-org/hide/pkg/devcontainer"
)

type Config struct {
	DevContainerConfig devcontainer.Config `json:"devContainerConfig"`
}

type ProjectId = string

type Project struct {
	Id          ProjectId `json:"id"`
	Path        string    `json:"path"`
	Config      Config    `json:"config"`
	ContainerId string
}

func NewProject(id ProjectId, path string, config Config, containerId string) Project {
	return Project{Id: id, Path: path, Config: config, ContainerId: containerId}
}

func (project *Project) FindTaskByAlias(alias string) (devcontainer.Task, error) {
	if project.Config.DevContainerConfig.Customizations.Hide == nil {
		return devcontainer.Task{}, NewTaskNotFoundError(alias)
	}

	for _, task := range project.Config.DevContainerConfig.Customizations.Hide.Tasks {
		if task.Alias == alias {
			return task, nil
		}
	}
	return devcontainer.Task{}, NewTaskNotFoundError(alias)
}

func (project *Project) GetTasks() []devcontainer.Task {
	if project.Config.DevContainerConfig.Customizations.Hide == nil {
		return []devcontainer.Task{}
	}

	return project.Config.DevContainerConfig.Customizations.Hide.Tasks
}

// unexported key type for Project; prevents collisions with keys defined in other packages
type key int

// allocated key for Project
var projectKey key

// NewContextWithProject returns a new context with the project set
func NewContextWithProject(ctx context.Context, project *Project) context.Context {
	return context.WithValue(ctx, projectKey, project)
}

// ProjectFromContext returns the project from the context
func ProjectFromContext(ctx context.Context) (*Project, bool) {
	project, ok := ctx.Value(projectKey).(*Project)
	return project, ok
}

type TaskNotFoundError struct {
	taskId string
}

func (e TaskNotFoundError) Error() string {
	return fmt.Sprintf("task %s not found", e.taskId)
}

func NewTaskNotFoundError(taskId string) *TaskNotFoundError {
	return &TaskNotFoundError{taskId: taskId}
}
