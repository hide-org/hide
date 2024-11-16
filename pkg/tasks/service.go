package task

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

type Result struct {
	StdOut   string `json:"stdout"`
	StdErr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

type Task struct {
	Alias   string `json:"alias"`
	Command string `json:"command"`
}

type Service interface {
	Get(ctx context.Context, alias string) (Task, error)
	List(ctx context.Context) ([]Task, error)
	Run(ctx context.Context, alias string) (Result, error)
	RunCommand(ctx context.Context, command string) (Result, error)
	Upsert(ctx context.Context, alias, command string) (Task, error)
}

type ServiceImpl struct {
	executor Executor
	store map[string]Task
}

func NewService(executor Executor, store map[string]Task) Service {
	return ServiceImpl{
		executor: executor,
		store: store,
	}
}

func (s ServiceImpl) Get(ctx context.Context, alias string) (Task, error) {
	task, ok := s.store[alias]
	if !ok {
		return Task{}, NewTaskNotFoundError(alias)
	}
	return task, nil
}

func (s ServiceImpl) List(ctx context.Context) ([]Task, error) {
	tasks := make([]Task, 0, len(s.store))
	for _, task := range s.store {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s ServiceImpl) Run(ctx context.Context, alias string) (Result, error) {
	task, err := s.Get(ctx, alias)
	if err != nil {
		return Result{}, err
	}

	return s.RunCommand(ctx, task.Command)
}

func (s ServiceImpl) RunCommand(ctx context.Context, command string) (Result, error) {
	return Result{}, nil
}

func (s ServiceImpl) Upsert(ctx context.Context, alias, command string) (Task, error) {
	s.store[alias] = Task{Alias: alias, Command: command}
	return s.Get(ctx, alias)
}

func (s ServiceImpl) CreateTask(ctx context.Context, command string) (Result, error) {
	log.Debug().Msgf("Creating task for command: %s", command)

	// TODO: implement running commands, i.e. handling output, dirs, etc.
	stdout, stderr := os.Stdout, os.Stderr
	err := s.executor.Run(cmdMaybeWithTimeout(ctx, command), "/workspace", stdout, stderr)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to execute command '%s'", command)
		return Result{}, fmt.Errorf("Failed to execute command: %w", err)
	}

	log.Debug().Msgf("Task for command %s completed", command, )

	return Result{StdOut: "", StdErr: "", ExitCode: 0}, nil
}

func (s ServiceImpl) ListTasks(ctx context.Context) ([]Task, error) {
	// TODO: implement listing tasks
	return []Task{}, nil
}

func (s ServiceImpl) ResolveTaskAlias(ctx context.Context, alias string) (Task, error) {
	// TODO: implement resolving task aliases
	return Task{}, nil
}

// cmdMaybeWithTimeout prepends timeout command to the command.
//
// Note: this is a workaround to ensure that the process is actually stopped after the timeout duration exceeded.
func cmdMaybeWithTimeout(ctx context.Context, cmd string) (command []string) {
	deadline, ok := ctx.Deadline()
	if !ok {
		return []string{"/bin/bash", "-c", cmd}
	}

	duration := time.Until(deadline)
	return []string{"timeout", "--kill-after=1s", "--verbose", fmt.Sprintf("%fs", duration.Seconds()), "/bin/bash", "-c", cmd}
}
