package mocks

// MockRegistryCredentials is a mock of the RegistryCredentials interface for testing
type MockRegistryCredentials struct {
	GetCredentialsFunc func() (string, error)
}

func (m *MockRegistryCredentials) GetCredentials() (string, error) {
	return m.GetCredentialsFunc()
}
