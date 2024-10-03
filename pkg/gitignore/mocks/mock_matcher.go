package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/spf13/afero"

	"github.com/artmoskvin/hide/pkg/gitignore"
)

// MockMatcher is a mock implementation of the Matcher interface
type MockMatcher struct {
	mock.Mock
}

// Match mocks the Match method of the Matcher interface
func (m *MockMatcher) Match(path string, isDir bool) (bool, error) {
	args := m.Called(path, isDir)
	return args.Bool(0), args.Error(1)
}

func NewMockMatcher() *MockMatcher {
	return new(MockMatcher)
}

// MockGitignoreMatcher is a mock implementation of the gitignore.Matcher interface
type MockGitignoreMatcher struct {
	mock.Mock
}

// Match mocks the Match method of the gitignore.Matcher interface
func (m *MockGitignoreMatcher) Match(path []string, isDir bool) bool {
	args := m.Called(path, isDir)
	return args.Bool(0)
}

func NewMockGitignoreMatcher() *MockGitignoreMatcher {
	return new(MockGitignoreMatcher)
}

// MockMatcherFactory is a mock implementation of the MatcherFactory interface
type MockMatcherFactory struct {
	mock.Mock
}

// NewMatcher mocks the NewMatcher method of the MatcherFactory interface
func (m *MockMatcherFactory) NewMatcher(fs afero.Fs) (gitignore.Matcher, error) {
	args := m.Called(fs)
	return args.Get(0).(gitignore.Matcher), args.Error(1)
}

func NewMockMatcherFactory() *MockMatcherFactory {
	return new(MockMatcherFactory)
}
