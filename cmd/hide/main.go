package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/artmoskvin/hide/pkg/devcontainer"
	"github.com/artmoskvin/hide/pkg/files"
	"github.com/artmoskvin/hide/pkg/handlers"
	"github.com/artmoskvin/hide/pkg/languageserver"
	"github.com/artmoskvin/hide/pkg/model"
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/artmoskvin/hide/pkg/util"
	"github.com/docker/docker/client"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

const (
	ProjectsDir       = "hide-projects"
	DefaultDotEnvPath = ".env"
)

func main() {
	fmt.Print(Splash)

	envPath := flag.String("env", DefaultDotEnvPath, "path to the .env file")
	debug := flag.Bool("debug", false, "run service in a debug mode")
	flag.Parse()

	setupLogger(*debug)

	_, err := os.Stat(*envPath)

	if os.IsNotExist(err) {
		log.Debug().Msg(fmt.Sprintf("Environment file %s does not exist.", *envPath))
	}

	if err == nil {
		// NOTE: can use filepath.Split()
		dotEnvDir := filepath.Dir(*envPath)
		dotEnvFile := filepath.Base(*envPath)

		err := util.LoadEnv(os.DirFS(dotEnvDir), dotEnvFile)
		if err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("Cannot load environment variables from %s", *envPath))
		}
	}

	dockerUser := os.Getenv("DOCKER_USER")
	dockerToken := os.Getenv("DOCKER_TOKEN")

	if dockerUser == "" || dockerToken == "" {
		log.Warn().Msg("DOCKER_USER or DOCKER_TOKEN environment variables are empty. This might cause problems when pulling images from Docker Hub.")
	}

	mux := http.NewServeMux()
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize docker client")
	}

	context := context.Background()
	containerRunner := devcontainer.NewDockerRunner(dockerClient, util.NewExecutorImpl(), context, devcontainer.DockerRunnerConfig{Username: dockerUser, Password: dockerToken})
	projectStore := project.NewInMemoryStore(make(map[string]*model.Project))
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Err(err).Msg("User's home directory is not set")
	}

	projectsDir := filepath.Join(home, ProjectsDir)

	languageDetector := languageserver.NewLanguageDetector()
	lspService := languageserver.NewService(lspClientFactoryMethod, languageDetector)
	fileManager := files.NewLanguageServerAwareFileManager(files.NewFileManager(), lspService)
	projectManager := project.NewProjectManager(containerRunner, projectStore, projectsDir, fileManager, lspService)
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

	// TODO: make configurable
	port := ":8080"

	log.Info().Msg(fmt.Sprintf("Server started on %s\n", port))

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal().Err(err).Str("port", port).Msg("Failed to start server")
	}
}

func setupLogger(debug bool) {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: zerolog.TimeFormatUnix, NoColor: false}
	log.Logger = log.Output(output)

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func lspClientFactoryMethod(languageId, projectRoot string, diagnosticsChannel chan protocol.PublishDiagnosticsParams) languageserver.Client {
	ctx := context.Background()

	// Define the lsp server executable based on the languageId (currently only "go" is supported)
	lspServerExecutable := "gopls"

	// Start the language server
	process, err := languageserver.NewProcess(lspServerExecutable)
	if err != nil {
		log.Fatalf("Failed to create language server process: %s", err)
	}

	if err := process.Start(); err != nil {
		log.Fatalf("Failed to start language server: %s", err)
	}

	// TODO: fix me
	// defer process.Stop()

	// Create a client for the language server
	client := languageserver.NewClient(ctx, process.ReadWriteCloser(), diagnosticsChannel)

	// Initialize the language server
	root := languageserver.PathToURI(projectRoot)
	_, err = client.Initialize(context.Background(), protocol.InitializeParams{
		RootURI: &root,
		Capabilities: protocol.ClientCapabilities{
			TextDocument: &protocol.TextDocumentClientCapabilities{
				Synchronization: &protocol.TextDocumentSyncClientCapabilities{
					DynamicRegistration: boolPointer(true),
				},
			},
		},
		// WorkspaceFolders: []protocol.WorkspaceFolder{
		// 	{
		// 		URI:  root,
		// 		Name: "hide",
		// 	},
		// },
	})

	if err != nil {
		log.Fatalf("Failed to initialize language server: %s", err)
	}

	// Debug
	log.Printf("Initialized language server for project ??? (languageId: %s)", languageId)

	// if opt, ok := initResult.Capabilities.TextDocumentSync.(protocol.TextDocumentSyncOptions); ok {
	// 	log.Printf("Support open/close file: %t", *opt.OpenClose)
	// 	log.Printf("Support change notifications: %v", *opt.Change)
	// }

	// Notify that initialized
	if err := client.NotifyInitialized(ctx); err != nil {
		log.Fatalf("Failed to notify initialized: %s", err)
	}

	return client
}

func boolPointer(b bool) *bool {
	return &b
}
