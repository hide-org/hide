package project

type Store interface {
	GetProject(id string) (*Project, error)
	GetProjects() ([]*Project, error)
	CreateProject(project *Project) error
	UpdateProject(project *Project) error
	DeleteProject(id string) error
}
