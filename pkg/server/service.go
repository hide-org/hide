package server

import "context"
import "github.com/hide-org/hide/pkg/devcontainer/v2"

type Service interface {
	Start(ctx context.Context, container devcontainer.DevContainer) error
	Stop(ctx context.Context, containerId string) error
}
