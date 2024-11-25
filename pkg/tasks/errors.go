package tasks

import "fmt"

type TaskNotFoundError struct {
	alias string
}

func (e TaskNotFoundError) Error() string {
	return fmt.Sprintf("task with alias %s not found", e.alias)
}

func NewTaskNotFoundError(alias string) *TaskNotFoundError {
	return &TaskNotFoundError{alias: alias}
}
