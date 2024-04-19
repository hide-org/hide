package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/artmoskvin/hide/pkg/project"
)

type CreateProjectHandler struct {
	Manager project.Manager
}

func (h CreateProjectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request project.CreateProjectRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed parsing request body", http.StatusBadRequest)
		return
	}

	project, err := h.Manager.CreateProject(request)

	if err != nil {
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

type ExecCmdHandler struct {
	Manager project.Manager
}

func (h ExecCmdHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectId := r.PathValue("id")
	var request project.ExecCmdRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed parsing request body", http.StatusBadRequest)
		return
	}

	execOut, err := h.Manager.ExecCmd(projectId, request)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute command: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(execOut)
}
