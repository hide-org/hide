package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/filemanager"
	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/artmoskvin/hide/pkg/util"
	"github.com/docker/docker/client"
)

const ProjectsDir = "hide-projects"
const DotEnvPath = "."
const DotEnvFile = ".env"

func main() {
	err := util.LoadEnv(os.DirFS(DotEnvPath), DotEnvFile)
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	dockerUser := os.Getenv("DOCKER_USER")
	dockerToken := os.Getenv("DOCKER_TOKEN")

	if dockerUser == "" || dockerToken == "" {
		log.Fatal("Error: DOCKER_USER and DOCKER_TOKEN environment variables are required")
	}

	mux := http.NewServeMux()
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		panic(err)
	}

	context := context.Background()
	containerRunner := devcontainer.NewDockerRunner(dockerClient, util.NewExecutorImpl(), context, devcontainer.DockerRunnerConfig{Username: dockerUser, Password: dockerToken})
	projectStore := project.NewInMemoryStore(make(map[string]*project.Project))
	home, err := os.UserHomeDir()

	if err != nil {
		panic(err)
	}

	projectsDir := filepath.Join(home, ProjectsDir)

	projectManager := project.NewProjectManager(containerRunner, projectStore, projectsDir)
	fileManager := filemanager.NewFileManager()
	createProjectHandler := handlers.CreateProjectHandler{Manager: projectManager}
	deleteProjectHandler := handlers.DeleteProjectHandler{Manager: projectManager}
	createTaskHandler := handlers.CreateTaskHandler{Manager: projectManager}
	listTasksHandler := handlers.ListTasksHandler{Manager: projectManager}
	createFileHandler := handlers.CreateFileHandler{Manager: projectManager, FileManager: fileManager}
	readFileHandler := handlers.ReadFileHandler{Manager: projectManager, FileManager: fileManager}
	updateFileHandler := handlers.UpdateFileHandler{Manager: projectManager, FileManager: fileManager}
	deleteFileHandler := handlers.DeleteFileHandler{Manager: projectManager, FileManager: fileManager}
	listFilesHandler := handlers.ListFilesHandler{ProjectManager: projectManager, FileManager: fileManager}

	mux.Handle("POST /projects", createProjectHandler)
	mux.Handle("DELETE /projects/{id}", deleteProjectHandler)
	mux.Handle("POST /projects/{id}/tasks", createTaskHandler)
	mux.Handle("GET /projects/{id}/tasks", listTasksHandler)
	mux.Handle("POST /projects/{id}/files", createFileHandler)
	mux.Handle("GET /projects/{id}/files", listFilesHandler)
	mux.Handle("GET /projects/{id}/files/{path...}", readFileHandler)
	mux.Handle("PUT /projects/{id}/files/{path...}", updateFileHandler)
	mux.Handle("DELETE /projects/{id}/files/{path...}", deleteFileHandler)

	port := ":8080"

	fmt.Print(Splash)
	log.Printf("Server started on %s\n", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		fmt.Println("Error starting server")
		panic(err)
	}
}
