package lang

type LanguageID = string

// Language IDs based on https://github.com/go-enry/go-enry
// For reference see https://github.com/go-enry/go-enry/blob/master/data/languageInfo.go
const (
	Go         LanguageID = "Go"
	JavaScript LanguageID = "JavaScript"
	Python     LanguageID = "Python"
	TypeScript LanguageID = "TypeScript"
	// TODO: check that const is consistenet with go-entry
	TSX LanguageID = "TSX"
	JSX LanguageID = "JSX"
)

var Adapters = []Adapter{
	new(gopls), new(tsserver),
}
