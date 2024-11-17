package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/docker/docker/client"
	"github.com/go-playground/validator/v10"
	"github.com/hide-org/hide/pkg/devcontainer/v2"
	"github.com/hide-org/hide/pkg/files"
	"github.com/hide-org/hide/pkg/git"
	"github.com/hide-org/hide/pkg/gitignore"
	"github.com/hide-org/hide/pkg/handlers/v2"
	"github.com/hide-org/hide/pkg/lsp/v2"
	"github.com/hide-org/hide/pkg/model"
	"github.com/hide-org/hide/pkg/project/v2"
	"github.com/hide-org/hide/pkg/random"
	"github.com/hide-org/hide/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	pf := pmRunCmd.PersistentFlags()
	pf.StringVar(&envPath, "env", DefaultDotEnvPath, "path to the .env file")
	pf.BoolVar(&debug, "debug", false, "run service in a debug mode")
	pf.IntVar(&port, "port", 8080, "service port")

	rootCmd.AddCommand(pmCmd)
	pmCmd.AddCommand(pmRunCmd)
}

var pmCmd = &cobra.Command{
	Use:   "pm",
	Short: "Project related commands",
	Long:  "Commands for managing projects in Hide.",
}

var pmRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs Hide project manager",
	PreRun: func(cmd *cobra.Command, args []string) {
		setupLogger(debug)
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(splash)

		_, err := os.Stat(envPath)

		if os.IsNotExist(err) {
			log.Debug().Msgf("Environment file %s does not exist.", envPath)
		}

		if err == nil {
			dir, file := filepath.Split(envPath)

			if dir == "" {
				dir = "."
			}

			err := util.LoadEnv(os.DirFS(dir), file)
			if err != nil {
				log.Error().Err(err).Msgf("Cannot load environment variables from %s", envPath)
			}
		}

		dockerUser := os.Getenv("DOCKER_USER")
		dockerToken := os.Getenv("DOCKER_TOKEN")

		if dockerUser == "" || dockerToken == "" {
			log.Warn().Msg("DOCKER_USER or DOCKER_TOKEN environment variables are empty. This might cause problems when pulling images from Docker Hub.")
		}

		dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Fatal().Err(err).Msg("Cannot initialize docker client")
		}

		containerRunner := devcontainer.NewDockerRunner(devcontainer.NewExecutorImpl(), devcontainer.NewImageManager(dockerClient, random.String, devcontainer.NewDockerHubRegistryCredentials(dockerUser, dockerToken)), devcontainer.NewDockerContainerManager(dockerClient))
		projectStore := project.NewInMemoryStore(make(map[string]*model.Project))
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal().Err(err).Msg("User's home directory is not set")
		}

		projectsDir := filepath.Join(home, HidePath, ProjectsDir)

		languageDetector := lsp.NewLanguageDetector()
		diagnosticsStore := lsp.NewDiagnosticsStore()
		clientPool := lsp.NewClientPool()
		lspService := lsp.NewService(languageDetector, lsp.LspServerExecutables, diagnosticsStore, clientPool)
		fileManager := files.NewService(gitignore.NewMatcherFactory(), lspService)
		projectManager := project.NewProjectManager(containerRunner, projectStore, projectsDir, fileManager, lspService, languageDetector, random.String, git.NewClient())
		validator := validator.New(validator.WithRequiredStructEnabled())

		router := handlers.
			NewRouter().
			WithCreateProjectHandler(handlers.CreateProjectHandler{Manager: projectManager, Validator: validator}).
			WithDeleteProjectHandler(handlers.DeleteProjectHandler{Manager: projectManager}).
			Build()

		addr := fmt.Sprintf("127.0.0.1:%d", port)

		server := &http.Server{
			Handler: router,
			Addr:    addr,
		}

		log.Info().Msgf("Server started on %s\n", addr)

		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			log.Info().Msg("Server shutting down ...")
			if err := projectManager.Cleanup(ctx); err != nil {
				log.Warn().Err(err).Msgf("Failed to cleanup projects")
			}

			if err := server.Shutdown(ctx); err != nil {
				log.Warn().Err(err).Msgf("HTTP shutdown error: %v", err)
			}
		}()

		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msgf("HTTP server error: %v", err)
		}

		fmt.Println("👋 Goodbye!")
	},
}
