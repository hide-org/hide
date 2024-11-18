package lang

type LanguageID string

// Language IDs based on https://github.com/go-enry/go-enry
// For reference see https://github.com/go-enry/go-enry/blob/master/data/languageInfo.go
const (
	Go         LanguageID = "Go"
	JavaScript            = "JavaScript"
	Python                = "Python"
	TypeScript            = "TypeScript"
)

type LanguageAdapterRegistry map[LanguageID]Adapter

var AdapterRegistry = LanguageAdapterRegistry{
	Go: gopls{},
}
