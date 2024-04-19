package main

import (
	"fmt"
	"net/http"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/project"
)

func main() {
	mux := http.NewServeMux()
	projectManager := project.SimpleManager{DevContainerManager: devcontainer.CliManager{}, InMemoryProjects: make(map[string]project.Project)}
	createProjectHandler := handlers.CreateProjectHandler{Manager: projectManager}
	execCmdHandler := handlers.ExecCmdHandler{Manager: projectManager}

	mux.Handle("POST /projects", createProjectHandler)
	mux.Handle("POST /projects/{id}/exec", execCmdHandler)

	port := ":8080"
	err := http.ListenAndServe(port, mux)

	if err != nil {
		fmt.Println("Error starting server: ", err)
	}

	fmt.Println("Server started on port", port)
}
