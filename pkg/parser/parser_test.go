package parser

import (
	"testing"

	"github.com/unclebucklarson/aura/pkg/ast"
	"github.com/unclebucklarson/aura/pkg/lexer"
)

func parseSource(t *testing.T, src string) *ast.Module {
	t.Helper()
	l := lexer.New(src, "test.aura")
	tokens, lexErrors := l.Tokenize()
	if len(lexErrors) > 0 {
		t.Fatalf("lex errors: %v", lexErrors)
	}
	p := New(tokens, "test.aura")
	module, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}
	return module
}

func TestParseModuleDecl(t *testing.T) {
	src := "module myapp.models\n"
	m := parseSource(t, src)
	if m.Name == nil {
		t.Fatal("expected module name")
	}
	if m.Name.String() != "myapp.models" {
		t.Errorf("expected 'myapp.models', got %q", m.Name.String())
	}
}

func TestParseImports(t *testing.T) {
	src := "module test\n\nimport std.time as time\nfrom std.collections import List, Map\n"
	m := parseSource(t, src)
	if len(m.Imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(m.Imports))
	}
	if m.Imports[0].Alias != "time" {
		t.Errorf("expected alias 'time', got %q", m.Imports[0].Alias)
	}
	if len(m.Imports[1].Names) != 2 {
		t.Errorf("expected 2 imported names, got %d", len(m.Imports[1].Names))
	}
}

func TestParseTypeDef(t *testing.T) {
	src := "module test\n\ntype TaskId = String where len >= 1 and len <= 64\n"
	m := parseSource(t, src)
	if len(m.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(m.Items))
	}
	td, ok := m.Items[0].(*ast.TypeDef)
	if !ok {
		t.Fatalf("expected TypeDef, got %T", m.Items[0])
	}
	if td.Name != "TaskId" {
		t.Errorf("expected name 'TaskId', got %q", td.Name)
	}
	_, isRefinement := td.Body.(*ast.RefinementType)
	if !isRefinement {
		t.Error("expected RefinementType body")
	}
}

func TestParseUnionType(t *testing.T) {
	src := "module test\n\ntype TaskStatus = \"pending\" | \"done\"\n"
	m := parseSource(t, src)
	if len(m.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(m.Items))
	}
	td := m.Items[0].(*ast.TypeDef)
	_, isUnion := td.Body.(*ast.UnionType)
	if !isUnion {
		t.Errorf("expected UnionType, got %T", td.Body)
	}
}

func TestParseStructDef(t *testing.T) {
	src := `module test

pub struct Task:
    pub id: String
    pub title: String = ""
    pub priority: Int = 3
`
	m := parseSource(t, src)
	if len(m.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(m.Items))
	}
	sd, ok := m.Items[0].(*ast.StructDef)
	if !ok {
		t.Fatalf("expected StructDef, got %T", m.Items[0])
	}
	if sd.Name != "Task" {
		t.Errorf("expected name 'Task', got %q", sd.Name)
	}
	if sd.Visibility != ast.Public {
		t.Error("expected public visibility")
	}
	if len(sd.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(sd.Fields))
	}
	if sd.Fields[1].Default == nil {
		t.Error("expected default value for title field")
	}
}

func TestParseEnumDef(t *testing.T) {
	src := `module test

pub enum TaskError:
    NotFound(String)
    InvalidTitle(String)
    Generic
`
	m := parseSource(t, src)
	ed, ok := m.Items[0].(*ast.EnumDef)
	if !ok {
		t.Fatalf("expected EnumDef, got %T", m.Items[0])
	}
	if len(ed.Variants) != 3 {
		t.Fatalf("expected 3 variants, got %d", len(ed.Variants))
	}
	if ed.Variants[0].Name != "NotFound" {
		t.Errorf("expected 'NotFound', got %q", ed.Variants[0].Name)
	}
	if len(ed.Variants[0].Fields) != 1 {
		t.Errorf("expected 1 field for NotFound, got %d", len(ed.Variants[0].Fields))
	}
	if len(ed.Variants[2].Fields) != 0 {
		t.Errorf("expected 0 fields for Generic, got %d", len(ed.Variants[2].Fields))
	}
}

func TestParseFnDef(t *testing.T) {
	src := `module test

pub fn add(a: Int, b: Int) -> Int:
    return a + b
`
	m := parseSource(t, src)
	fd, ok := m.Items[0].(*ast.FnDef)
	if !ok {
		t.Fatalf("expected FnDef, got %T", m.Items[0])
	}
	if fd.Name != "add" {
		t.Errorf("expected 'add', got %q", fd.Name)
	}
	if len(fd.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(fd.Params))
	}
	if fd.ReturnType == nil {
		t.Error("expected return type")
	}
	if len(fd.Body) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(fd.Body))
	}
}

func TestParseFnWithEffects(t *testing.T) {
	src := `module test

pub fn create(title: String) -> Result[Task, Error] with db, time satisfies CreateSpec:
    return Ok(Task())
`
	m := parseSource(t, src)
	fd := m.Items[0].(*ast.FnDef)
	if len(fd.Effects) != 2 {
		t.Fatalf("expected 2 effects, got %d: %v", len(fd.Effects), fd.Effects)
	}
	if fd.Satisfies != "CreateSpec" {
		t.Errorf("expected 'CreateSpec', got %q", fd.Satisfies)
	}
}

func TestParseIfStmt(t *testing.T) {
	src := `module test

pub fn f(x: Int) -> String:
    if x > 0:
        return "pos"
    elif x < 0:
        return "neg"
    else:
        return "zero"
`
	m := parseSource(t, src)
	fd := m.Items[0].(*ast.FnDef)
	if len(fd.Body) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(fd.Body))
	}
	ifStmt, ok := fd.Body[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", fd.Body[0])
	}
	if len(ifStmt.ElifClauses) != 1 {
		t.Errorf("expected 1 elif clause, got %d", len(ifStmt.ElifClauses))
	}
	if len(ifStmt.ElseBody) != 1 {
		t.Errorf("expected 1 else statement, got %d", len(ifStmt.ElseBody))
	}
}

func TestParseMatchStmt(t *testing.T) {
	src := `module test

pub fn f(x: Int) -> String:
    match x:
        case 0:
            return "zero"
        case _:
            return "other"
`
	m := parseSource(t, src)
	fd := m.Items[0].(*ast.FnDef)
	ms, ok := fd.Body[0].(*ast.MatchStmt)
	if !ok {
		t.Fatalf("expected MatchStmt, got %T", fd.Body[0])
	}
	if len(ms.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(ms.Cases))
	}
}

func TestParseForStmt(t *testing.T) {
	src := `module test

pub fn f() -> Int:
    let mut sum = 0
    for i in range(10):
        sum = sum + i
    return sum
`
	m := parseSource(t, src)
	fd := m.Items[0].(*ast.FnDef)
	if len(fd.Body) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(fd.Body))
	}
	_, ok := fd.Body[1].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", fd.Body[1])
	}
}

func TestParseSpecBlock(t *testing.T) {
	src := `module test

spec CreateNewTask:
    doc: "Creates a new task."

    inputs:
        title: String - "The task title"
        priority: Int = 3 - "Priority"

    guarantees:
        - "Returns a Task"
        - "Task has status pending"

    effects: db, time

    errors:
        InvalidTitle(String) - "When title is empty"
`
	m := parseSource(t, src)
	sb, ok := m.Items[0].(*ast.SpecBlock)
	if !ok {
		t.Fatalf("expected SpecBlock, got %T", m.Items[0])
	}
	if sb.Name != "CreateNewTask" {
		t.Errorf("expected 'CreateNewTask', got %q", sb.Name)
	}
	if sb.Doc != "Creates a new task." {
		t.Errorf("expected doc, got %q", sb.Doc)
	}
	if len(sb.Inputs) != 2 {
		t.Fatalf("expected 2 inputs, got %d", len(sb.Inputs))
	}
	if len(sb.Guarantees) != 2 {
		t.Fatalf("expected 2 guarantees, got %d", len(sb.Guarantees))
	}
	if len(sb.Effects) != 2 {
		t.Fatalf("expected 2 effects, got %d", len(sb.Effects))
	}
	if len(sb.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(sb.Errors))
	}
}

func TestParseExpressions(t *testing.T) {
	src := `module test

pub fn f() -> Int:
    let a = 1 + 2 * 3
    let b = not true
    let c = [1, 2, 3]
    return a
`
	m := parseSource(t, src)
	fd := m.Items[0].(*ast.FnDef)
	if len(fd.Body) != 4 {
		t.Fatalf("expected 4 statements, got %d", len(fd.Body))
	}

	// First let: 1 + 2 * 3 should be BinaryOp
	ls := fd.Body[0].(*ast.LetStmt)
	_, ok := ls.Value.(*ast.BinaryOp)
	if !ok {
		t.Errorf("expected BinaryOp for '1 + 2 * 3', got %T", ls.Value)
	}
}

func TestParseListComp(t *testing.T) {
	src := `module test

pub fn f() -> [Int]:
    let xs = [x for x in items if x > 0]
    return xs
`
	m := parseSource(t, src)
	fd := m.Items[0].(*ast.FnDef)
	ls := fd.Body[0].(*ast.LetStmt)
	lc, ok := ls.Value.(*ast.ListComp)
	if !ok {
		t.Fatalf("expected ListComp, got %T", ls.Value)
	}
	if lc.Variable != "x" {
		t.Errorf("expected variable 'x', got %q", lc.Variable)
	}
	if lc.Filter == nil {
		t.Error("expected filter expression")
	}
}

func TestParseEmpty(t *testing.T) {
	src := ""
	m := parseSource(t, src)
	if m == nil {
		t.Fatal("expected non-nil module")
	}
}

func TestParseTestBlock(t *testing.T) {
	src := `module test

test "addition works":
    let r = add(1, 2)
    assert r == 3
`
	m := parseSource(t, src)
	tb, ok := m.Items[0].(*ast.TestBlock)
	if !ok {
		t.Fatalf("expected TestBlock, got %T", m.Items[0])
	}
	if tb.Name != "addition works" {
		t.Errorf("expected test name 'addition works', got %q", tb.Name)
	}
	if len(tb.Body) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(tb.Body))
	}
}
