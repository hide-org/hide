package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/project"
)

const ProjectsDir = "hide-projects"

func main() {
	mux := http.NewServeMux()
	devContainerManager := devcontainer.NewDevContainerManager()
	projectStore := make(map[string]project.Project)
	home, err := os.UserHomeDir()

	if err != nil {
		panic(err)
	}

	projectsDir := filepath.Join(home, ProjectsDir)

	projectManager := project.NewProjectManager(devContainerManager, projectStore, projectsDir)
	createProjectHandler := handlers.CreateProjectHandler{Manager: projectManager}
	execCmdHandler := handlers.ExecCmdHandler{Manager: projectManager}

	mux.Handle("POST /projects", createProjectHandler)
	mux.Handle("POST /projects/{id}/exec", execCmdHandler)

	port := ":8080"

	if err := http.ListenAndServe(port, mux); err != nil {
		fmt.Println("Error starting server")
		panic(err)
	}
}
