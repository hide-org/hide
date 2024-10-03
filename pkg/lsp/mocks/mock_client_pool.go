package mocks

import (
	"github.com/hide-org/hide/pkg/lsp"
	"github.com/stretchr/testify/mock"
)

var _ lsp.ClientPool = (*MockClientPool)(nil)

type MockClientPool struct {
	mock.Mock
}

func (m *MockClientPool) Get(projectId lsp.ProjectId, languageId lsp.LanguageId) (lsp.Client, bool) {
	args := m.Called(projectId, languageId)
	return args.Get(0).(lsp.Client), args.Bool(1)
}

func (m *MockClientPool) GetAllForProject(projectId lsp.ProjectId) (map[lsp.LanguageId]lsp.Client, bool) {
	args := m.Called(projectId)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(map[lsp.LanguageId]lsp.Client), args.Bool(1)
}

func (m *MockClientPool) Set(projectId lsp.ProjectId, languageId lsp.LanguageId, client lsp.Client) {
	m.Called(projectId, languageId, client)
}

func (m *MockClientPool) Delete(projectId lsp.ProjectId, languageId lsp.LanguageId) {
	m.Called(projectId, languageId)
}

func (m *MockClientPool) DeleteAllForProject(projectId lsp.ProjectId) {
	m.Called(projectId)
}
