package languageserver

import "github.com/artmoskvin/hide/pkg/model"

type LanguageDetector interface {
	DetectLanguage(file model.File) string
}

type LanguageDetectorImpl struct{}

func NewLanguageDetector() LanguageDetector {
	return &LanguageDetectorImpl{}
}

func (ld LanguageDetectorImpl) DetectLanguage(file model.File) string {
	// TODO: implement
	return "go"
}
