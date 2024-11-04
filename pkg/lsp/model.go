package lsp

import (
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type SymbolInfo struct {
	Name     string   `json:"name"`
	Kind     string   `json:"kind"`
	Location Location `json:"location"`
}

type Location struct {
	Path  string `json:"path"`
	Range Range  `json:"range"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type DocumentOutline struct {
	Path            string           `json:"path"`
	DocumentSymbols []DocumentSymbol `json:"document_symbols"`
}

type DocumentSymbol struct {
	Name     string           `json:"name"`
	Detail   string           `json:"detail"`
	Kind     string           `json:"kind"`
	Range    Range            `json:"range"`
	Children []DocumentSymbol `json:"children,omitempty"`
}

func documentOutlineFrom(symbols []protocol.DocumentSymbol, path string) DocumentOutline {
	out := DocumentOutline{
		Path:            path,
		DocumentSymbols: make([]DocumentSymbol, 0, len(symbols)),
	}

	for _, symbol := range symbols {
		out.DocumentSymbols = append(out.DocumentSymbols, parseDocumentSymbol(symbol))
	}

	return out
}

func parseDocumentSymbol(src protocol.DocumentSymbol) DocumentSymbol {
	out := DocumentSymbol{
		Name: src.Name,
		Kind: symbolKindToString(src.Kind),
		Range: Range{
			Start: Position{
				Line:      int(src.Range.Start.Line) + 1, // src is zero indexed
				Character: int(src.Range.Start.Character),
			},
			End: Position{
				Line:      int(src.Range.End.Line) + 1, // src is zero indexed
				Character: int(src.Range.End.Character),
			},
		},
		Children: make([]DocumentSymbol, 0, len(src.Children)),
	}

	if src.Detail != nil {
		out.Detail = *src.Detail
	}

	for _, child := range src.Children {
		out.Children = append(out.Children, parseDocumentSymbol(child))
	}

	return out
}

func symbolKindToString(kind protocol.SymbolKind) string {
	switch kind {
	case protocol.SymbolKindFile:
		return "File"
	case protocol.SymbolKindModule:
		return "Module"
	case protocol.SymbolKindNamespace:
		return "Namespace"
	case protocol.SymbolKindPackage:
		return "Package"
	case protocol.SymbolKindClass:
		return "Class"
	case protocol.SymbolKindMethod:
		return "Method"
	case protocol.SymbolKindProperty:
		return "Property"
	case protocol.SymbolKindField:
		return "Field"
	case protocol.SymbolKindConstructor:
		return "Constructor"
	case protocol.SymbolKindEnum:
		return "Enum"
	case protocol.SymbolKindInterface:
		return "Interface"
	case protocol.SymbolKindFunction:
		return "Function"
	case protocol.SymbolKindVariable:
		return "Variable"
	case protocol.SymbolKindConstant:
		return "Constant"
	case protocol.SymbolKindString:
		return "String"
	case protocol.SymbolKindNumber:
		return "Number"
	case protocol.SymbolKindBoolean:
		return "Boolean"
	case protocol.SymbolKindArray:
		return "Array"
	case protocol.SymbolKindObject:
		return "Object"
	case protocol.SymbolKindKey:
		return "Key"
	case protocol.SymbolKindNull:
		return "Null"
	case protocol.SymbolKindEnumMember:
		return "EnumMember"
	case protocol.SymbolKindStruct:
		return "Struct"
	case protocol.SymbolKindEvent:
		return "Event"
	case protocol.SymbolKindOperator:
		return "Operator"
	case protocol.SymbolKindTypeParameter:
		return "TypeParameter"
	default:
		return "Unknown"
	}
}
