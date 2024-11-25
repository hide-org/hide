package lsp

import (
	"fmt"

	lang "github.com/hide-org/hide/pkg/lsp/v2/languages"
)

type LanguageNotSupportedError struct {
	languageId lang.LanguageID
}

func (e LanguageNotSupportedError) Error() string {
	return fmt.Sprintf("Language %s is not supported", e.languageId)
}

func NewLanguageNotSupportedError(languageId lang.LanguageID) *LanguageNotSupportedError {
	return &LanguageNotSupportedError{languageId: languageId}
}

type LanguageServerAlreadyExistsError struct {
	LanguageID lang.LanguageID
}

func (e LanguageServerAlreadyExistsError) Error() string {
	return fmt.Sprintf("language server already exists for language %s", e.LanguageID)
}

func NewLanguageServerAlreadyExistsError(languageId lang.LanguageID) *LanguageServerAlreadyExistsError {
	return &LanguageServerAlreadyExistsError{LanguageID: languageId}
}

type LanguageServerNotFoundError struct {
	LanguageID lang.LanguageID
}

func (e LanguageServerNotFoundError) Error() string {
	return fmt.Sprintf("language server not found for language %s", e.LanguageID)
}

func NewLanguageServerNotFoundError(languageId lang.LanguageID) *LanguageServerNotFoundError {
	return &LanguageServerNotFoundError{LanguageID: languageId}
}
