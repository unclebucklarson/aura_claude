// Package symbols implements the symbol table for the Aura type checker.
// It provides scope management with nested scopes for functions, blocks, and modules.
package symbols

import (
	"fmt"

	"github.com/unclebucklarson/aura/pkg/token"
)

// SymbolKind represents what kind of thing a symbol is.
type SymbolKind int

const (
	SymVariable SymbolKind = iota
	SymFunction
	SymType
	SymStruct
	SymEnum
	SymTrait
	SymSpec
	SymModule
	SymParam
	SymConst
	SymEnumVariant
	SymField
)

func (k SymbolKind) String() string {
	names := [...]string{
		"variable", "function", "type", "struct", "enum",
		"trait", "spec", "module", "parameter", "constant",
		"enum_variant", "field",
	}
	if int(k) < len(names) {
		return names[k]
	}
	return fmt.Sprintf("SymbolKind(%d)", int(k))
}

// Symbol represents a named entity in the program.
type Symbol struct {
	Name       string
	Kind       SymbolKind
	Span       token.Span
	Mutable    bool   // for variables: let mut
	Public     bool   // pub visibility
	TypeID     int    // index into the type registry (set by checker)
	SpecName   string // for functions: satisfies spec name
	Effects    []string // for functions: declared effects
	TypeParams []string // for generic types/functions
}

// ScopeKind indicates the type of scope.
type ScopeKind int

const (
	ScopeModule ScopeKind = iota
	ScopeFunction
	ScopeBlock
	ScopeLoop
	ScopeTest
)

func (k ScopeKind) String() string {
	names := [...]string{"module", "function", "block", "loop", "test"}
	if int(k) < len(names) {
		return names[k]
	}
	return fmt.Sprintf("ScopeKind(%d)", int(k))
}

// Scope represents a lexical scope containing symbol definitions.
type Scope struct {
	Kind     ScopeKind
	Name     string // e.g., function name, module name
	Parent   *Scope
	Children []*Scope
	Symbols  map[string]*Symbol
}

// NewScope creates a new scope with the given kind and parent.
func NewScope(kind ScopeKind, name string, parent *Scope) *Scope {
	s := &Scope{
		Kind:    kind,
		Name:    name,
		Parent:  parent,
		Symbols: make(map[string]*Symbol),
	}
	if parent != nil {
		parent.Children = append(parent.Children, s)
	}
	return s
}

// Define adds a symbol to this scope. Returns an error if already defined.
func (s *Scope) Define(sym *Symbol) error {
	if existing, ok := s.Symbols[sym.Name]; ok {
		return fmt.Errorf("symbol %q already defined as %s at %s",
			sym.Name, existing.Kind, existing.Span.Start)
	}
	s.Symbols[sym.Name] = sym
	return nil
}

// Lookup searches for a symbol in this scope and all parent scopes.
func (s *Scope) Lookup(name string) (*Symbol, bool) {
	if sym, ok := s.Symbols[name]; ok {
		return sym, true
	}
	if s.Parent != nil {
		return s.Parent.Lookup(name)
	}
	return nil, false
}

// LookupLocal searches for a symbol only in this scope.
func (s *Scope) LookupLocal(name string) (*Symbol, bool) {
	sym, ok := s.Symbols[name]
	return sym, ok
}

// IsInsideLoop checks if this scope or any ancestor is a loop scope.
func (s *Scope) IsInsideLoop() bool {
	for cur := s; cur != nil; cur = cur.Parent {
		if cur.Kind == ScopeLoop {
			return true
		}
	}
	return false
}

// EnclosingFunction returns the nearest enclosing function scope, or nil.
func (s *Scope) EnclosingFunction() *Scope {
	for cur := s; cur != nil; cur = cur.Parent {
		if cur.Kind == ScopeFunction {
			return cur
		}
	}
	return nil
}

// Table is the top-level symbol table manager.
type Table struct {
	Root    *Scope
	Current *Scope
}

// NewTable creates a new symbol table with a module-level root scope.
func NewTable(moduleName string) *Table {
	root := NewScope(ScopeModule, moduleName, nil)
	return &Table{
		Root:    root,
		Current: root,
	}
}

// PushScope enters a new child scope.
func (t *Table) PushScope(kind ScopeKind, name string) *Scope {
	s := NewScope(kind, name, t.Current)
	t.Current = s
	return s
}

// PopScope exits the current scope and returns to the parent.
func (t *Table) PopScope() *Scope {
	old := t.Current
	if t.Current.Parent != nil {
		t.Current = t.Current.Parent
	}
	return old
}

// Define adds a symbol to the current scope.
func (t *Table) Define(sym *Symbol) error {
	return t.Current.Define(sym)
}

// Lookup searches for a symbol starting from the current scope.
func (t *Table) Lookup(name string) (*Symbol, bool) {
	return t.Current.Lookup(name)
}

// LookupLocal searches only the current scope.
func (t *Table) LookupLocal(name string) (*Symbol, bool) {
	return t.Current.LookupLocal(name)
}
