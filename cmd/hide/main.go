package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/files"
	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/artmoskvin/hide/pkg/util"
	"github.com/docker/docker/client"
)

const ProjectsDir = "hide-projects"
const DefaultDotEnvPath = ".env"

func main() {
	envPath := flag.String("env", DefaultDotEnvPath, "path to the .env file")
	flag.Parse()

	_, err := os.Stat(*envPath)

	if os.IsNotExist(err) {
		log.Printf("Debug: Environment file %s does not exist.", *envPath)
	}

	if err == nil {
		dotEnvDir := filepath.Dir(*envPath)
		dotEnvFile := filepath.Base(*envPath)

		err := util.LoadEnv(os.DirFS(dotEnvDir), dotEnvFile)
		if err != nil {
			log.Printf("Warning: Cannot load environment variables from %s: %s", *envPath, err)
		}
	}

	dockerUser := os.Getenv("DOCKER_USER")
	dockerToken := os.Getenv("DOCKER_TOKEN")

	if dockerUser == "" || dockerToken == "" {
		log.Println("Warning: DOCKER_USER and DOCKER_TOKEN environment variables are empty. This might cause problems when pulling images from Docker Hub.")
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
	fileManager := files.NewFileManager()
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
