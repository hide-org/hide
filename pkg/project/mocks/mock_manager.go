package mocks

import "github.com/artmoskvin/hide/pkg/project"

// MockProjectManager is a mock of the project.Manager interface for testing
type MockProjectManager struct {
	CreateProjectFunc    func(request project.CreateProjectRequest) (project.Project, error)
	GetProjectFunc       func(projectId string) (project.Project, error)
	ResolveTaskAliasFunc func(projectId string, alias string) (project.Task, error)
	CreateTaskFunc       func(projectId string, command string) (project.TaskResult, error)
}

func (m *MockProjectManager) CreateProject(request project.CreateProjectRequest) (project.Project, error) {
	return m.CreateProjectFunc(request)
}

func (m *MockProjectManager) GetProject(projectId string) (project.Project, error) {
	return m.GetProjectFunc(projectId)
}

func (m *MockProjectManager) ResolveTaskAlias(projectId string, alias string) (project.Task, error) {
	return m.ResolveTaskAliasFunc(projectId, alias)
}

func (m *MockProjectManager) CreateTask(projectId string, command string) (project.TaskResult, error) {
	return m.CreateTaskFunc(projectId, command)
}
