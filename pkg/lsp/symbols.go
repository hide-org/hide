package lsp

import (
	"slices"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

type SymbolFilter struct {
	include []protocol.SymbolKind
	exclude []protocol.SymbolKind
}

func NewIncludeSymbolFilter(include ...protocol.SymbolKind) SymbolFilter {
	return SymbolFilter{include: include}
}

func NewExcludeSymbolFilter(exclude ...protocol.SymbolKind) SymbolFilter {
	return SymbolFilter{exclude: exclude}
}
func (f *SymbolFilter) shouldExcludeSymbol(symbol protocol.SymbolInformation) bool {
	return slices.Contains(f.exclude, symbol.Kind)
}

func (f *SymbolFilter) shouldIncludeSymbol(symbol protocol.SymbolInformation) bool {
	if len(f.include) == 0 {
		return true
	}
	return slices.Contains(f.include, symbol.Kind)
}
