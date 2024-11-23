package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/hide-org/hide/pkg/files"
	"github.com/hide-org/hide/pkg/gitignore"
	"github.com/hide-org/hide/pkg/handlers/v2"
	"github.com/hide-org/hide/pkg/lsp/v2"
	"github.com/hide-org/hide/pkg/middleware"
	"github.com/hide-org/hide/pkg/outline"
	"github.com/hide-org/hide/pkg/symbols"
	"github.com/hide-org/hide/pkg/tasks"
	"github.com/hide-org/hide/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var workspace string

func init() {
	pf := serverRunCmd.PersistentFlags()
	pf.StringVar(&envPath, "env", DefaultDotEnvPath, "path to the .env file")
	pf.BoolVar(&debug, "debug", false, "run service in a debug mode")
	pf.IntVar(&port, "port", 8080, "service port")
	cwd, _ := os.Getwd() // how can it fail?
	pf.StringVar(&workspace, "workspace", cwd, "path to the workspace directory")

	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(serverRunCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Server related commands",
	Long:  "Commands for managing the development server of Hide.",
}

var serverRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs Hide development server",
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

		languageDetector := lsp.NewLanguageDetector()
		diagnosticsStore := lsp.NewDiagnosticsStore()
		clientPool := lsp.NewClientPool()
		lspService := lsp.NewService(languageDetector, lsp.LspServerExecutables, diagnosticsStore, clientPool)

		taskService := tasks.NewService(tasks.NewExecutorImpl(), map[string]tasks.Task{})
		fileService := files.NewService(gitignore.NewMatcherFactory(), lspService, afero.NewBasePathFs(afero.NewOsFs(), workspace))
		symbolSearch := symbols.NewService(lspService)
		outlineService := outline.NewService(lspService)
		router := handlers.
			NewRouter().
			WithCreateTaskHandler(handlers.CreateTaskHandler{Tasks: taskService}).
			WithListTasksHandler(handlers.ListTasksHandler{Tasks: taskService}).
			WithCreateFileHandler(handlers.CreateFileHandler{Files: fileService}).
			WithListFilesHandler(handlers.ListFilesHandler{Files: fileService}).
			WithReadFileHandler(middleware.PathValidator(handlers.ReadFileHandler{Files: fileService})).
			WithUpdateFileHandler(middleware.PathValidator(handlers.UpdateFileHandler{Files: fileService})).
			WithDeleteFileHandler(middleware.PathValidator(handlers.DeleteFileHandler{Files: fileService})).
			WithSearchFileHandler(handlers.SearchFilesHandler{Files: fileService}).
			WithSearchSymbolsHandler(handlers.NewSearchSymbolsHandler(symbolSearch)).
			WithDocumentOutlineHandler(handlers.DocumentOutline{Outline: outlineService}).
			Build()

		addr := fmt.Sprintf("0.0.0.0:%d", port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create listener")
		}
		fmt.Printf("Server listening on %s\n", listener.Addr().String())

		server := &http.Server{
			Handler: router,
		}

		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			log.Info().Msg("Server shutting down ...")

			if err := server.Shutdown(ctx); err != nil {
				log.Warn().Err(err).Msgf("HTTP shutdown error: %v", err)
			}
		}()

		if err := server.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msgf("HTTP server error: %v", err)
		}

		fmt.Println("👋 Goodbye!")
	},
}
