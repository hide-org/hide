package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hide-org/hide/pkg/tasks"
)

type ListTasksHandler struct {
	Tasks tasks.Service
}

func (h ListTasksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.Tasks.List(r.Context())

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get project: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
	return
}
