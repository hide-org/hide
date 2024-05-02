package mocks

import "github.com/artmoskvin/hide/pkg/project"

// MockProjectManager is a mock of the project.Manager interface for testing
type MockProjectManager struct {
	CreateProjectFunc func(request project.CreateProjectRequest) (project.Project, error)
	GetProjectFunc    func(projectId string) (project.Project, error)
	CreateTaskFunc    func(projectId string, request project.TaskRequest) (project.TaskResult, error)
}

func (m *MockProjectManager) CreateProject(request project.CreateProjectRequest) (project.Project, error) {
	return m.CreateProjectFunc(request)
}

func (m *MockProjectManager) GetProject(projectId string) (project.Project, error) {
	return m.GetProjectFunc(projectId)
}

func (m *MockProjectManager) CreateTask(projectId string, request project.TaskRequest) (project.TaskResult, error) {
	return m.CreateTaskFunc(projectId, request)
}
