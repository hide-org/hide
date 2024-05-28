package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/artmoskvin/hide/pkg/project"
)

type ListTasksHandler struct {
	Manager project.Manager
}

func (h ListTasksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectId := r.PathValue("id")
	project, err := h.Manager.GetProject(projectId)

	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	tasks := project.GetTasks()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}
