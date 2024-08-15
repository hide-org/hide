package project

import (
	"github.com/artmoskvin/hide/pkg/model"
)

type Store interface {
	GetProject(id string) (*model.Project, error)
	GetProjects() ([]*model.Project, error)
	CreateProject(project *model.Project) error
	UpdateProject(project *model.Project) error
	DeleteProject(id string) error
}

type InMemoryStore struct {
	projects map[string]*model.Project
}

func NewInMemoryStore(store map[string]*model.Project) *InMemoryStore {
	return &InMemoryStore{projects: store}
}

func (s *InMemoryStore) GetProject(id string) (*model.Project, error) {
	project, ok := s.projects[id]

	if !ok {
		return nil, NewProjectNotFoundError(id)
	}

	return project, nil
}

func (s *InMemoryStore) GetProjects() ([]*model.Project, error) {
	projects := make([]*model.Project, 0, len(s.projects))

	for _, project := range s.projects {
		projects = append(projects, project)
	}

	return projects, nil
}

func (s *InMemoryStore) CreateProject(project *model.Project) error {
	if _, ok := s.projects[project.Id]; ok {
		return NewProjectAlreadyExistsError(project.Id)
	}

	s.projects[project.Id] = project

	return nil
}

func (s *InMemoryStore) UpdateProject(project *model.Project) error {
	if _, ok := s.projects[project.Id]; !ok {
		return NewProjectNotFoundError(project.Id)
	}

	s.projects[project.Id] = project

	return nil
}

func (s *InMemoryStore) DeleteProject(id string) error {
	delete(s.projects, id)

	return nil
}
