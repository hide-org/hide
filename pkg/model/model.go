package model

import (
	"context"
	"errors"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type File struct {
	Path        string                `json:"path"`
	Content     string                `json:"content"`
	Diagnostics []protocol.Diagnostic `json:"diagnostics,omitempty"`
}

func (f *File) Equals(other *File) bool {
	if f == nil && other == nil {
		return true
	}

	if f == nil || other == nil {
		return false
	}

	// TODO: compare diagnostics
	return f.Path == other.Path && f.Content == other.Content
}

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
		return devcontainer.Task{}, errors.New("task not found")
	}

	for _, task := range project.Config.DevContainerConfig.Customizations.Hide.Tasks {
		if task.Alias == alias {
			return task, nil
		}
	}
	return devcontainer.Task{}, errors.New("task not found")
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
