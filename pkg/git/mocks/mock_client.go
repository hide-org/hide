package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/hide-org/hide/pkg/git"
)

var _ git.Client = &MockClient{}

// MockClient is a mock implementation of the Client interface
type MockClient struct {
	mock.Mock
}

// Checkout mocks the Checkout method of the Client interface
func (m *MockClient) Checkout(repo git.Repository, commit string) error {
	args := m.Called(repo, commit)
	return args.Error(0)
}

// Clone mocks the Clone method of the Client interface
func (m *MockClient) Clone(url, dst string) (*git.Repository, error) {
	args := m.Called(url, dst)
	return args.Get(0).(*git.Repository), args.Error(1)
}

// NewMockClient creates and returns a new instance of MockClient
func NewMockClient() *MockClient {
	return new(MockClient)
}
