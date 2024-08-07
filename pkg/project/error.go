package project

import "fmt"

type ProjectNotFoundError struct {
	ProjectId string
}

func (e ProjectNotFoundError) Error() string {
	return fmt.Sprintf("project %s not found", e.ProjectId)
}

type ProjectAlreadyExistsError struct {
	ProjectId string
}

func (e ProjectAlreadyExistsError) Error() string {
	return fmt.Sprintf("project %s already exists", e.ProjectId)
}
