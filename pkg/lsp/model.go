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
