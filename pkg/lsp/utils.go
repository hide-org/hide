package lsp

import (
	"path/filepath"

	"github.com/artmoskvin/hide/pkg/model"
	"github.com/go-enry/go-enry/v2"
	"github.com/rs/zerolog/log"
)

// Language IDs based on https://github.com/go-enry/go-enry
// For reference see https://github.com/go-enry/go-enry/blob/master/data/languageInfo.go
const (
	Go         = LanguageId("Go")
	JavaScript = LanguageId("JavaScript")
	Python     = LanguageId("Python")
	TypeScript = LanguageId("TypeScript")
)

type LanguageDetector interface {
	DetectLanguage(file *model.File) string
	DetectLanguages(files []*model.File) map[string]int
	DetectMainLanguage(files []*model.File) string
}

// LanguageDetectorImpl implements LanguageDetector using https://github.com/go-enry/go-enry
type LanguageDetectorImpl struct{}

func NewFileExtensionBasedLanguageDetector() LanguageDetector {
	return &LanguageDetectorImpl{}
}
func (ld LanguageDetectorImpl) DetectLanguage(file *model.File) LanguageId {
	return enry.GetLanguage(filepath.Base(file.Path), file.GetContentBytes())
}

func (ld LanguageDetectorImpl) DetectLanguages(files []*model.File) map[string]int {
	languages := make(map[string]int)
	for _, file := range files {
		language := ld.DetectLanguage(file)
		languages[language] += len(file.GetContentBytes())
	}
	return languages
}

func (ld LanguageDetectorImpl) DetectMainLanguage(files []*model.File) string {
	languages := ld.DetectLanguages(files)
	log.Debug().Any("languages", languages).Msg("Detected languages")
	var maxLanguage string
	maxCount := 0
	for language, count := range languages {
		if count > maxCount {
			maxLanguage = language
			maxCount = count
		}
	}
	return maxLanguage
}
