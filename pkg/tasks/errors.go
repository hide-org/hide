package task

import "fmt"

type TaskNotFoundError string

func (e TaskNotFoundError) Error() string {
	return fmt.Sprintf("task with alias %s not found", string(e))
}

func NewTaskNotFoundError(alias string) TaskNotFoundError {
	return TaskNotFoundError(alias)
}
