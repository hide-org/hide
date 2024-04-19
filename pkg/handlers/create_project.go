package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/artmoskvin/hide/pkg/project"
)

func CreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request project.LaunchDevContainerRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed parsing request body", http.StatusBadRequest)
		return
	}

	// TODO: is there a way to avoid creating a new instance of DevContainerCli every time?
	devContainerCli := project.DevContainerCli{}
	projectManager := project.NewProjectManager()

	projectPath, err := projectManager.CreateProjectDir()

	if err != nil {
		http.Error(w, "Failed to create project directory", http.StatusInternalServerError)
		return
	}

	devContainer, err := devContainerCli.Create(request, projectPath)

	if err != nil {
		projectManager.RemoveProjectDir()
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devContainer)
}

func ExecCmd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request project.ExecCmdRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed parsing request body", http.StatusBadRequest)
		return
	}

	devContainerCli := project.DevContainerCli{}

	execOut, err := devContainerCli.Exec(request)

	if err != nil {
		http.Error(w, "Failed to execute command", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(execOut)
}
