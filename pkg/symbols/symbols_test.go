package symbols

import (
	"testing"

	"github.com/unclebucklarson/aura/pkg/token"
)

func span() token.Span {
	return token.Span{File: "test.aura", Start: token.Position{Line: 1, Column: 1}}
}

func TestScopeDefineAndLookup(t *testing.T) {
	s := NewScope(ScopeModule, "test", nil)
	sym := &Symbol{Name: "x", Kind: SymVariable, Span: span()}
	if err := s.Define(sym); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, ok := s.Lookup("x")
	if !ok {
		t.Fatal("expected to find symbol x")
	}
	if found.Kind != SymVariable {
		t.Errorf("expected SymVariable, got %v", found.Kind)
	}
}

func TestScopeDuplicateDefine(t *testing.T) {
	s := NewScope(ScopeModule, "test", nil)
	sym1 := &Symbol{Name: "x", Kind: SymVariable, Span: span()}
	sym2 := &Symbol{Name: "x", Kind: SymFunction, Span: span()}

	if err := s.Define(sym1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Define(sym2); err == nil {
		t.Fatal("expected error for duplicate define")
	}
}

func TestNestedScopeLookup(t *testing.T) {
	parent := NewScope(ScopeModule, "mod", nil)
	parent.Define(&Symbol{Name: "x", Kind: SymVariable, Span: span()})

	child := NewScope(ScopeFunction, "fn", parent)
	child.Define(&Symbol{Name: "y", Kind: SymParam, Span: span()})

	// child can see parent symbols
	if _, ok := child.Lookup("x"); !ok {
		t.Error("child should see parent symbol x")
	}
	if _, ok := child.Lookup("y"); !ok {
		t.Error("child should see own symbol y")
	}

	// parent cannot see child symbols
	if _, ok := parent.Lookup("y"); ok {
		t.Error("parent should not see child symbol y")
	}
}

func TestLookupLocal(t *testing.T) {
	parent := NewScope(ScopeModule, "mod", nil)
	parent.Define(&Symbol{Name: "x", Kind: SymVariable, Span: span()})

	child := NewScope(ScopeFunction, "fn", parent)

	if _, ok := child.LookupLocal("x"); ok {
		t.Error("LookupLocal should not traverse parent")
	}
}

func TestShadowing(t *testing.T) {
	parent := NewScope(ScopeModule, "mod", nil)
	parent.Define(&Symbol{Name: "x", Kind: SymVariable, Span: span(), TypeID: 1})

	child := NewScope(ScopeBlock, "block", parent)
	child.Define(&Symbol{Name: "x", Kind: SymVariable, Span: span(), TypeID: 2})

	found, ok := child.Lookup("x")
	if !ok {
		t.Fatal("expected to find x")
	}
	if found.TypeID != 2 {
		t.Error("child x should shadow parent x")
	}
}

func TestTablePushPop(t *testing.T) {
	tbl := NewTable("test_module")

	if tbl.Current.Kind != ScopeModule {
		t.Error("root should be module scope")
	}

	tbl.Define(&Symbol{Name: "a", Kind: SymType, Span: span()})
	tbl.PushScope(ScopeFunction, "myFunc")
	tbl.Define(&Symbol{Name: "b", Kind: SymParam, Span: span()})

	if _, ok := tbl.Lookup("a"); !ok {
		t.Error("should see module symbol from function scope")
	}

	tbl.PopScope()

	if _, ok := tbl.Lookup("b"); ok {
		t.Error("should not see function symbol after pop")
	}
	if _, ok := tbl.Lookup("a"); !ok {
		t.Error("should still see module symbol after pop")
	}
}

func TestIsInsideLoop(t *testing.T) {
	tbl := NewTable("mod")
	tbl.PushScope(ScopeFunction, "fn")

	if tbl.Current.IsInsideLoop() {
		t.Error("should not be inside loop")
	}

	tbl.PushScope(ScopeLoop, "for")
	if !tbl.Current.IsInsideLoop() {
		t.Error("should be inside loop")
	}

	tbl.PushScope(ScopeBlock, "if")
	if !tbl.Current.IsInsideLoop() {
		t.Error("nested block inside loop should still report inside loop")
	}
}

func TestEnclosingFunction(t *testing.T) {
	tbl := NewTable("mod")
	if tbl.Current.EnclosingFunction() != nil {
		t.Error("module scope should have no enclosing function")
	}

	fnScope := tbl.PushScope(ScopeFunction, "myFunc")
	if tbl.Current.EnclosingFunction() != fnScope {
		t.Error("function scope should be its own enclosing function")
	}

	tbl.PushScope(ScopeBlock, "if")
	if tbl.Current.EnclosingFunction() != fnScope {
		t.Error("block should find enclosing function")
	}
}

func TestSymbolKindString(t *testing.T) {
	if SymVariable.String() != "variable" {
		t.Errorf("expected 'variable', got %q", SymVariable.String())
	}
	if SymFunction.String() != "function" {
		t.Errorf("expected 'function', got %q", SymFunction.String())
	}
}
