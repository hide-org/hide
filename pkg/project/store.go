package project

import "fmt"

type Store interface {
	GetProject(id string) (*Project, error)
	GetProjects() ([]*Project, error)
	CreateProject(project *Project) error
	UpdateProject(project *Project) error
	DeleteProject(id string) error
}

type InMemoryStore struct {
	projects map[string]*Project
}

func NewInMemoryStore(store map[string]*Project) *InMemoryStore {
	return &InMemoryStore{projects: store}
}

func (s *InMemoryStore) GetProject(id string) (*Project, error) {
	project, ok := s.projects[id]

	if !ok {
		return nil, fmt.Errorf("project with id %s not found", id)
	}

	return project, nil
}

func (s *InMemoryStore) GetProjects() ([]*Project, error) {
	projects := make([]*Project, 0, len(s.projects))

	for _, project := range s.projects {
		projects = append(projects, project)
	}

	return projects, nil
}

func (s *InMemoryStore) CreateProject(project *Project) error {
	if _, ok := s.projects[project.Id]; ok {
		return fmt.Errorf("project with id %s already exists", project.Id)
	}

	s.projects[project.Id] = project

	return nil
}

func (s *InMemoryStore) UpdateProject(project *Project) error {
	if _, ok := s.projects[project.Id]; !ok {
		return fmt.Errorf("project with id %s not found", project.Id)
	}

	s.projects[project.Id] = project

	return nil
}

func (s *InMemoryStore) DeleteProject(id string) error {
	delete(s.projects, id)

	return nil
}
