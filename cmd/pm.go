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

	"github.com/go-playground/validator/v10"
	"github.com/hide-org/hide/pkg/daytona"
	"github.com/hide-org/hide/pkg/github"
	"github.com/hide-org/hide/pkg/handlers/v2"
	"github.com/hide-org/hide/pkg/model"
	"github.com/hide-org/hide/pkg/project/v2"
	"github.com/hide-org/hide/pkg/server"
	"github.com/hide-org/hide/pkg/util"
	"github.com/hide-org/hide/pkg/workspaces"
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

		var releaseProvider server.ReleaseProvider
		var sshOpts []workspaces.SshOption
		if os.Getenv("RELEASE_PROVIDER") == "local" {
			releaseProvider = server.NewStaticReleaseProvider("127.0.0.1:8000/bin/hide_amd64")
			sshOpts = append(sshOpts, workspaces.NewSshOption("RemoteForward", fmt.Sprintf("8000 %s", "127.0.0.1:8000")))
		} else {
			releaseProvider = server.NewGithubReleaseProvider(github.NewClient("hide-org", "hide"))
		}

		serverInstaller := server.NewInstaller("~/.local/bin/hide", releaseProvider, sshOpts)
		projectStore := project.NewInMemoryStore(make(map[string]*model.Project))
		daytonaAPI := daytona.NewAPIClient(daytona.NewConfiguration())
		workspacesService := workspaces.NewDaytonaService(daytonaAPI, os.Getenv("DAYTONA_API_KEY"))

		projectManager := project.NewProjectManager(serverInstaller, projectStore, workspacesService)
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
