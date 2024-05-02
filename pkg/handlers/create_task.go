package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/artmoskvin/hide/pkg/project"
)

type TaskRequest struct {
	Command *string `json:"command,omitempty"`
	Alias   *string `json:"alias,omitempty"`
}

func (t *TaskRequest) validate() error {
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
	projectId := r.PathValue("id")
	var request TaskRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed parsing request body", http.StatusBadRequest)
		return
	}

	if err := request.validate(); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err), http.StatusBadRequest)
		return
	}

	var command string

	if request.Alias != nil {
		task, err := h.Manager.ResolveTaskAlias(projectId, *request.Alias)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to resolve task alias: %s", err), http.StatusBadRequest)
			return
		}
		command = task.Command
	} else {
		command = *request.Command
	}

	taskResult, err := h.Manager.CreateTask(projectId, command)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to run task %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(taskResult)
}
