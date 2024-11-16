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
	LanguageId LanguageId
}

func (e LanguageServerAlreadyExistsError) Error() string {
	return fmt.Sprintf("language server already exists for language %s", e.LanguageId)
}

func NewLanguageServerAlreadyExistsError(languageId LanguageId) *LanguageServerAlreadyExistsError {
	return &LanguageServerAlreadyExistsError{LanguageId: languageId}
}

type LanguageServerNotFoundError struct {
	LanguageId LanguageId
}

func (e LanguageServerNotFoundError) Error() string {
	return fmt.Sprintf("language server not found for language %s", e.LanguageId)
}

func NewLanguageServerNotFoundError(languageId LanguageId) *LanguageServerNotFoundError {
	return &LanguageServerNotFoundError{LanguageId: languageId}
}
