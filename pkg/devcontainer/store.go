package devcontainer

import "fmt"

type Store interface {
	GetContainer(id string) (*Container, error)
	GetContainerByProject(projectID string) ([]*Container, error)
	CreateContainer(container *Container) error
	UpdateContainer(container *Container) error
	DeleteContainer(id string) error
}

type InMemoryStore struct {
	containers map[string]*Container
}

func NewInMemoryStore(store map[string]*Container) *InMemoryStore {
	return &InMemoryStore{containers: store}
}

func (s *InMemoryStore) GetContainer(id string) (*Container, error) {
	container, ok := s.containers[id]

	if !ok {
		return nil, fmt.Errorf("Container with id %s not found", id)
	}

	return container, nil
}

func (s *InMemoryStore) GetContainerByProject(projectID string) ([]*Container, error) {
	var containers []*Container

	for _, container := range s.containers {
		if container.ProjectID == projectID {
			containers = append(containers, container)
		}
	}

	return containers, nil
}

func (s *InMemoryStore) CreateContainer(container *Container) error {
	if _, ok := s.containers[container.Id]; ok {
		return fmt.Errorf("Container with id %s already exists", container.Id)
	}

	s.containers[container.Id] = container
	return nil
}

func (s *InMemoryStore) UpdateContainer(container *Container) error {
	if _, ok := s.containers[container.Id]; !ok {
		return fmt.Errorf("Container with id %s not found", container.Id)
	}

	s.containers[container.Id] = container
	return nil
}

func (s *InMemoryStore) DeleteContainer(id string) error {
	if _, ok := s.containers[id]; !ok {
		return fmt.Errorf("Container with id %s not found", id)
	}

	delete(s.containers, id)
	return nil
}
