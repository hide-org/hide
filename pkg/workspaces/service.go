package workspaces

import "context"

type Service interface {
	Create(ctx context.Context, gitURL string) (Workspace, error)
	Get(ctx context.Context, ID string) (Workspace, error)
}
