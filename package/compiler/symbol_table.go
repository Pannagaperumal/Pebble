package compiler

import (
	"pebble/package/symbol"
)

// SymbolTable is an alias for the symbol.SymbolTable type
type SymbolTable = symbol.SymbolTable

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	return symbol.NewSymbolTable()
}

// NewEnclosedSymbolTable creates a new symbol table with an outer scope
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	return symbol.NewEnclosedSymbolTable(outer)
}
