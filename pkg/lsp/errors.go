package lsp

import "fmt"

type LanguageNotSupportedError struct {
	languageId LanguageId
}

func (e LanguageNotSupportedError) Error() string {
	return fmt.Sprintf("Language %s is not supported", e.languageId)
}

func NewLanguageNotSupportedError(languageId LanguageId) *LanguageNotSupportedError {
	return &LanguageNotSupportedError{languageId: languageId}
}

type LanguageServerAlreadyExistsError struct {
	ProjectId  ProjectId
	LanguageId LanguageId
}

func (e LanguageServerAlreadyExistsError) Error() string {
	return fmt.Sprintf("Language server already exists for project %s and language %s", e.ProjectId, e.LanguageId)
}

func NewLanguageServerAlreadyExistsError(projectId ProjectId, languageId LanguageId) *LanguageServerAlreadyExistsError {
	return &LanguageServerAlreadyExistsError{ProjectId: projectId, LanguageId: languageId}
}

type LanguageServerNotFoundError struct {
	ProjectId  ProjectId
	LanguageId LanguageId
}

func (e LanguageServerNotFoundError) Error() string {
	return fmt.Sprintf("Language server not found for project %s and language %s", e.ProjectId, e.LanguageId)
}

func NewLanguageServerNotFoundError(projectId ProjectId, languageId LanguageId) *LanguageServerNotFoundError {
	return &LanguageServerNotFoundError{ProjectId: projectId, LanguageId: languageId}
}
