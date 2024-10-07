package xrt

import "github.com/bosley/xla/xlist"

type SymbolEnvironment struct {
	Symbols map[string]xlist.Element
	Parent  *SymbolEnvironment
	Child   *SymbolEnvironment
}

// NewSymbolEnvironment creates a new SymbolEnvironment with an optional parent
func NewSymbolEnvironment(parent *SymbolEnvironment) *SymbolEnvironment {
	return &SymbolEnvironment{
		Symbols: make(map[string]xlist.Element),
		Parent:  parent,
	}
}

func OrphanSymbolEnvironmentFrom(mapIn map[string]xlist.Element) *SymbolEnvironment {
	return &SymbolEnvironment{
		Symbols: mapIn,
		Parent:  nil,
	}
}

// MergeSymbols merges the provided symbols into the current SymbolEnvironment
func (se *SymbolEnvironment) MergeSymbols(sym map[string]xlist.Element) {
	for k, v := range sym {
		se.Symbols[k] = v
	}
}

// PushSymbols creates a child SymbolEnvironment with the current as the parent
func (se *SymbolEnvironment) PushSymbols() *SymbolEnvironment {
	child := NewSymbolEnvironment(se)
	se.Child = child
	return child
}

// PopSymbols returns the parent SymbolEnvironment or the current if no parent exists
func (se *SymbolEnvironment) PopSymbols() *SymbolEnvironment {
	if se.Parent == nil {
		return se
	}
	return se.Parent
}

// SearchSymbol searches for a symbol in the current and parent environments
func (se *SymbolEnvironment) SearchSymbol(key string, permitParentSearch bool) (xlist.Element, bool) {
	if value, ok := se.Symbols[key]; ok {
		return value, true
	}

	if permitParentSearch && se.Parent != nil {
		return se.Parent.SearchSymbol(key, permitParentSearch)
	}

	return xlist.Element{}, false
}

// SetSymbol sets a symbol in the current environment's symbol map
func (se *SymbolEnvironment) SetSymbol(identifier string, result xlist.Element) {
	se.Symbols[identifier] = result
}
