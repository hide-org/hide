package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hide-org/hide/pkg/model"
	"github.com/hide-org/hide/pkg/project"
)

type TaskRequest struct {
	Command *string `json:"command,omitempty"`
	Alias   *string `json:"alias,omitempty"`
}

func (t *TaskRequest) Validate() error {
	if t.Command == nil && t.Alias == nil {
		return errors.New("either command or alias must be provided")
	}

	if t.Command != nil && t.Alias != nil {
		return errors.New("only one of command or alias must be provided")
	}

	return nil
}

type CreateTaskHandler struct {
	Manager project.Manager
}

func (h CreateTaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	timeOutSec := getTimeOutSeconds(r)
	if timeOutSec <= 0 {
		h.do(r.Context(), w, r)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*time.Duration(timeOutSec))
	defer cancel()

	h.do(ctx, w, r)
}

func (h CreateTaskHandler) do(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	projectID, err := getProjectID(r)
	if err != nil {
		http.Error(w, "invalid project ID", http.StatusBadRequest)
		return
	}

	var request TaskRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed parsing request body", http.StatusBadRequest)
		return
	}

	if err := request.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err), http.StatusBadRequest)
		return
	}

	var command string

	if request.Alias != nil {
		// check for context cancellation error
		task, err := h.Manager.ResolveTaskAlias(ctx, projectID, *request.Alias)
		if err != nil {
			var projectNotFoundError *project.ProjectNotFoundError
			if errors.As(err, &projectNotFoundError) {
				http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
				return
			}

			var taskNotFoundError *model.TaskNotFoundError
			if errors.As(err, &taskNotFoundError) {
				http.Error(w, taskNotFoundError.Error(), http.StatusNotFound)
				return
			}

			if errors.Is(err, context.Canceled) {
				// do not write any response since it can only be cancelled by client
				return
			}

			if errors.Is(err, context.DeadlineExceeded) {
				http.Error(w, "", http.StatusRequestTimeout)
			}

			http.Error(w, fmt.Sprintf("Failed to resolve task alias: %s", err), http.StatusInternalServerError)
			return
		}
		command = task.Command
	} else {
		command = *request.Command
	}

	taskResult, err := h.Manager.CreateTask(ctx, projectID, command)
	if err != nil {
		var projectNotFoundError *project.ProjectNotFoundError
		if errors.As(err, &projectNotFoundError) {
			http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, context.Canceled) {
			// do not write any response since it can only be cancelled by client
			return
		}

		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "", http.StatusRequestTimeout)
		}

		http.Error(w, fmt.Sprintf("Failed to run task %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(taskResult)

	return
}
