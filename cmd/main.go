package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/filemanager"
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
	fileManager := filemanager.NewFileManager()
	createProjectHandler := handlers.CreateProjectHandler{Manager: projectManager}
	createTaskHandler := handlers.CreateTaskHandler{Manager: projectManager}
	createFileHandler := handlers.CreateFileHandler{Manager: projectManager, FileManager: fileManager}
	readFileHandler := handlers.ReadFileHandler{Manager: projectManager, FileManager: fileManager}
	updateFileHandler := handlers.UpdateFileHandler{Manager: projectManager, FileManager: fileManager}
	deleteFileHandler := handlers.DeleteFileHandler{Manager: projectManager, FileManager: fileManager}

	mux.Handle("POST /projects", createProjectHandler)
	mux.Handle("POST /projects/{id}/tasks", createTaskHandler)
	mux.Handle("POST /projects/{id}/files", createFileHandler)
	mux.Handle("GET /projects/{id}/files/{path...}", readFileHandler)
	mux.Handle("PUT /projects/{id}/files/{path...}", updateFileHandler)
	mux.Handle("DELETE /projects/{id}/files/{path...}", deleteFileHandler)

	port := ":8080"

	if err := http.ListenAndServe(port, mux); err != nil {
		fmt.Println("Error starting server")
		panic(err)
	}
}
