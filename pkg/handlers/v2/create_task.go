package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hide-org/hide/pkg/tasks"
)

type TaskRequest struct {
	Command *string `json:"command,omitempty"`
	Alias   *string `json:"alias,omitempty"`
}

type CreateTaskHandler struct {
	Tasks tasks.Service
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
	var request TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "failed parsing request body", http.StatusBadRequest)
		return
	}

	if request.Alias != nil {
		// check for context cancellation error
		result, err := h.Tasks.Run(ctx, *request.Alias)
		if err != nil {
			var taskNotFoundError *tasks.TaskNotFoundError
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
				return
			}

			http.Error(w, fmt.Sprintf("failed to run task '%s': %s", *request.Alias, err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
		return
	}

	if request.Command != nil {
		result, err := h.Tasks.RunCommand(ctx, *request.Command)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				// do not write any response since it can only be cancelled by client
				return
			}

			if errors.Is(err, context.DeadlineExceeded) {
				http.Error(w, "", http.StatusRequestTimeout)
				return
			}

			http.Error(w, fmt.Sprintf("failed to run command '%s': %s", *request.Command, err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
		return
	}

	http.Error(w, "invalid request: either 'command' or 'alias' must be provided", http.StatusBadRequest)
	return 
}
