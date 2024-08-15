package project

import "fmt"

type ProjectNotFoundError struct {
	projectId string
}

func (e ProjectNotFoundError) Error() string {
	return fmt.Sprintf("project %s not found", e.projectId)
}

func NewProjectNotFoundError(projectId string) *ProjectNotFoundError {
	return &ProjectNotFoundError{projectId: projectId}
}

type ProjectAlreadyExistsError struct {
	ProjectId string
}

func (e ProjectAlreadyExistsError) Error() string {
	return fmt.Sprintf("project %s already exists", e.ProjectId)
}

func NewProjectAlreadyExistsError(projectId string) *ProjectAlreadyExistsError {
	return &ProjectAlreadyExistsError{ProjectId: projectId}
}
