package devcontainer

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/docker/docker/pkg/stringid"
	"github.com/hide-org/hide/pkg/daytona"
	"github.com/hide-org/hide/pkg/git"
	"github.com/rs/zerolog/log"
)

type DevContainer interface {
	ID() string
	Stop(ctx context.Context) error
	Exec(ctx context.Context, command []string) (ExecResult, error)
	ExecDetached(ctx context.Context, command []string) (string, error)
}

type DockerDevContainer struct {
	id     string
	runner Runner
}

func NewDockerDevContainer(id string, runner Runner) DevContainer {
	return &DockerDevContainer{id: id, runner: runner}
}

func (d *DockerDevContainer) ID() string {
	return d.id
}

func (d *DockerDevContainer) Stop(ctx context.Context) error {
	return d.runner.Stop(ctx, d.id)
}

func (d *DockerDevContainer) Exec(ctx context.Context, command []string) (ExecResult, error) {
	return d.runner.Exec(ctx, d.id, command)
}

func (d *DockerDevContainer) ExecDetached(ctx context.Context, command []string) (string, error) {
	return d.runner.ExecDetached(ctx, d.id, command)
}

func (d *DockerDevContainer) CopyFile(ctx context.Context, hostPath, containerPath string) error {
	return d.runner.CopyFile(ctx, d.id, hostPath, containerPath)
}

type Service interface {
	Create(ctx context.Context, gitURL string) (DevContainer, error)
	Get(ctx context.Context, containerId string) (DevContainer, error)
}

type DockerService struct {
	git          git.Client
	projectsRoot string
	randomFn     func(int) string
	runner       Runner
}

func (s *DockerService) Create(ctx context.Context, gitURL string, commit *string, config *Config) (DevContainer, error) {
	projectId := s.randomFn(10)
	projectPath := path.Join(s.projectsRoot, projectId)

	// Clone git repo
	if err := createProjectDir(projectPath); err != nil {
		log.Error().Err(err).Msg("Failed to create project directory")
		return nil, fmt.Errorf("Failed to create project directory: %w", err)
	}

	r, err := s.git.Clone(gitURL, projectPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to clone git repo")
		removeProjectDir(projectPath)
		return nil, fmt.Errorf("Failed to clone git repo: %w", err)
	}

	if commit != nil {
		if err := s.git.Checkout(*r, *commit); err != nil {
			log.Error().Err(err).Msg("Failed to checkout commit")
			removeProjectDir(projectPath)
			return nil, fmt.Errorf("Failed to checkout commit %s: %w", *commit, err)
		}
	}

	// Start devcontainer
	if config == nil {
		config, err = configFromProject(os.DirFS(projectPath))
		if err != nil {
			log.Error().Err(err).Msgf("Failed to get devcontainer config from repository %s", gitURL)
			removeProjectDir(projectPath)
			return nil, fmt.Errorf("Failed to read devcontainer.json: %w", err)
		}
	}

	containerID, err := s.runner.Run(ctx, projectPath, *config)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start devcontainer")
		removeProjectDir(projectPath)
		return nil, fmt.Errorf("Failed to start devcontainer: %w", err)
	}

	return NewDockerDevContainer(containerID, s.runner), nil
}

func configFromProject(fs fs.FS) (*Config, error) {
	configFile, err := FindConfig(fs)
	if err != nil {
		return nil, fmt.Errorf("Failed to find devcontainer.json: %w", err)
	}

	config, err := ParseConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse devcontainer.json: %w", err)
	}

	return config, nil
}

func createProjectDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("Failed to create project directory: %w", err)
	}

	log.Debug().Msgf("Created project directory: %s", path)

	return nil
}

func removeProjectDir(projectPath string) {
	if err := os.RemoveAll(projectPath); err != nil {
		log.Error().Err(err).Msgf("Failed to remove project directory %s", projectPath)
		return
	}

	log.Debug().Msgf("Removed project directory: %s", projectPath)

	return
}

type DaytonaContainer struct {
	workspace *daytona.Workspace
	daytona   *daytona.APIClient
}

func (c *DaytonaContainer) ID() string {
	return c.workspace.GetId()
}

func (c *DaytonaContainer) Stop(ctx context.Context) error {
	return nil
}

func (c *DaytonaContainer) Exec(ctx context.Context, command []string) (ExecResult, error) {
	return ExecResult{}, nil
}

func (c *DaytonaContainer) ExecDetached(ctx context.Context, command []string) (string, error) {
	return "", nil
}

func NewDaytonaContainer(workspace *daytona.Workspace, daytona *daytona.APIClient) DevContainer {
	return &DaytonaContainer{workspace: workspace, daytona: daytona}
}

type DaytonaService struct {
	daytona *daytona.APIClient
}

func NewDaytonaService(daytona *daytona.APIClient) Service {
	return &DaytonaService{daytona: daytona}
}

func (s *DaytonaService) Create(ctx context.Context, gitURL string) (DevContainer, error) {
	repositoryContext := *daytona.NewGetRepositoryContext(gitURL)
	repository, _, err := s.daytona.GitProviderAPI.GetGitContext(ctx).Repository(repositoryContext).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository context: %w", err)
	}

	wsID := workspaceID()
	wsName := workspaceName(repository, wsID)

	project := *daytona.NewCreateProjectDTO(
		make(map[string]string),
		repository.Name,
		*daytona.NewCreateProjectSourceDTO(*repository),
	)

	project.SetImage("daytonaio/workspace-project:latest")
	project.SetUser("daytona")

	payload := *daytona.NewCreateWorkspaceDTO(
		wsID,
		wsName,
		[]daytona.CreateProjectDTO{project},
		"local",
	)

	workspace, _, err := s.daytona.WorkspaceAPI.CreateWorkspace(ctx).Workspace(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	log.Debug().Msgf("Created workspace %v", workspace)
	return NewDaytonaContainer(workspace, s.daytona), nil
}

func (s *DaytonaService) Get(ctx context.Context, containerId string) (DevContainer, error) {
	return nil, nil
}

func workspaceID() string {
	return stringid.TruncateID(stringid.GenerateRandomID())
}

func workspaceName(repository *daytona.GitRepository, ID string) string {
	return fmt.Sprintf("%s-%s", repository.Name, ID)
}

func GetArchitecture(ctx context.Context, container DevContainer) (string, error) {
	result, err := container.Exec(ctx, []string{"uname", "-m"})
	if err != nil {
		return "", fmt.Errorf("failed to get container architecture: %w", err)
	}

	arch := strings.TrimSpace(result.StdOut)
	switch arch {
	case "x86_64":
		return "amd64", nil
	case "aarch64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("unsupported architecture: %s", arch)
	}
}
