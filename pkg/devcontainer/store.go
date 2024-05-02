package devcontainer

type Store interface {
	GetContainer(id string) (*Container, error)
	GetContainerByProject(projectID string) ([]*Container, error)
	CreateContainer(container *Container) error
	UpdateContainer(container *Container) error
	DeleteContainer(id string) error
}
