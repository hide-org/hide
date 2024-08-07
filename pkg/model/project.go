package model

import (
	"errors"

	"github.com/artmoskvin/hide/pkg/devcontainer"
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
