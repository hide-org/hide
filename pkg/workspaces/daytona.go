package workspaces

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/docker/docker/pkg/stringid"
	"github.com/hide-org/hide/pkg/daytona"
	"github.com/rs/zerolog/log"
)

type DaytonaWorkspace struct {
	workspace *daytona.Workspace
	daytona   *daytona.APIClient
}

func (c *DaytonaWorkspace) ID() string {
	return c.workspace.GetId()
}

// Runs the command in the workspace container using daytona ssh CLI. 
// Example: `daytona ssh hide-8a2c5bfdccea hide "curl -fsSL -o ~/.local/bin/hide 127.0.0.1:8000/bin/hide_amd64" -o "RemoteForward=8000 127.0.0.1:8000"`
func (c *DaytonaWorkspace) Ssh(ctx context.Context, command string, opts ...SshOption) (string, error) {
	projects, ok := c.workspace.GetProjectsOk()
	if !ok || len(projects) == 0 {
		return "", fmt.Errorf("workspace has no projects")
	}
	if len(projects) > 1 {
		log.Warn().Msgf("workspace has more than one project, using the first one")
	}

	project := projects[0]
	cmdArgs := []string{"ssh", c.workspace.GetId(), project.GetName(), command}
	for _, opt := range opts {
		cmdArgs = append(cmdArgs, fmt.Sprintf("-o %s=%s", opt.key, opt.value))
	}

	log.Debug().Msgf("Running command: daytona %s", strings.Join(cmdArgs, " "))

	cmd := exec.CommandContext(ctx, "daytona", cmdArgs...)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Read stdout and stderr
	stdoutBytes, err := io.ReadAll(stdout)
	if err != nil {
		return "", fmt.Errorf("failed to read stdout: %w", err)
	}
	stderrBytes, _ := io.ReadAll(stderr) // We can ignore stderr read errors

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("failed to run command: %w\nstderr: %s", err, stderrBytes)
	}

	log.Debug().Msgf("Command output:\n%s", stdoutBytes)
	return string(stdoutBytes), nil
}

func (c *DaytonaWorkspace) Stop(ctx context.Context) error {
	_, err := c.daytona.WorkspaceAPI.StopWorkspace(ctx, c.workspace.GetId()).Execute()
	return err
}

func NewDaytonaWorkspace(workspace *daytona.Workspace, daytona *daytona.APIClient) Workspace {
	return &DaytonaWorkspace{workspace: workspace, daytona: daytona}
}

type DaytonaService struct {
	daytona *daytona.APIClient
}

func NewDaytonaService(daytona *daytona.APIClient) Service {
	return &DaytonaService{daytona: daytona}
}

func (s *DaytonaService) Create(ctx context.Context, gitURL string) (Workspace, error) {
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
		"docker-remote",
	)

	workspace, _, err := s.daytona.WorkspaceAPI.CreateWorkspace(ctx).Workspace(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	log.Debug().Msgf("Created workspace %v", workspace)
	return NewDaytonaWorkspace(workspace, s.daytona), nil
}

func (s *DaytonaService) Get(ctx context.Context, ID string) (Workspace, error) {
	workspace, _, err := s.daytona.WorkspaceAPI.GetWorkspace(ctx, ID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	return NewDaytonaWorkspace(daytona.NewWorkspace(workspace.GetId(), workspace.GetName(), workspace.GetProjects(), workspace.GetTarget()), s.daytona), nil
}

func workspaceID() string {
	return stringid.TruncateID(stringid.GenerateRandomID())
}

func workspaceName(repository *daytona.GitRepository, ID string) string {
	return fmt.Sprintf("%s-%s", repository.Name, ID)
}
