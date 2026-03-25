package docgen

import (
	"strings"
	"testing"

	"github.com/unclebucklarson/aura/pkg/lexer"
	"github.com/unclebucklarson/aura/pkg/parser"
)

// parse is a test helper that lexes and parses Aura source into an ast.Module.
func parse(t *testing.T, src string) *DocPage {
	t.Helper()
	l := lexer.New(src, "<test>")
	tokens, _ := l.Tokenize()
	p := parser.New(tokens, "<test>")
	mod, errs := p.Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	return Generate(mod)
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("expected output to contain:\n  %q\ngot:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Errorf("expected output NOT to contain %q", want)
	}
}

// --- Module header ---

func TestDocModuleName(t *testing.T) {
	page := parse(t, `module myapp`)
	if page.Module != "myapp" {
		t.Errorf("expected module 'myapp', got %q", page.Module)
	}
	assertContains(t, page.Markdown(), "# myapp")
}

// --- Functions ---

func TestDocPubFunction(t *testing.T) {
	page := parse(t, `module test

## Add two integers.
pub fn add(a: Int, b: Int) -> Int:
    return a + b
`)
	if len(page.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(page.Functions))
	}
	f := page.Functions[0]
	if f.Signature != "fn add(a: Int, b: Int) -> Int" {
		t.Errorf("unexpected signature: %q", f.Signature)
	}
	if f.Doc != "Add two integers." {
		t.Errorf("unexpected doc: %q", f.Doc)
	}
	md := page.Markdown()
	assertContains(t, md, "### `fn add(a: Int, b: Int) -> Int`")
	assertContains(t, md, "Add two integers.")
}

func TestDocPrivateFunctionExcluded(t *testing.T) {
	page := parse(t, `module test

fn internal() -> Int:
    return 42
`)
	if len(page.Functions) != 0 {
		t.Errorf("expected private function to be excluded")
	}
}

func TestDocFunctionEffectsAndConstraints(t *testing.T) {
	page := parse(t, `module test

trait Printable:
    fn display(self) -> String

pub fn save[T](item: T) -> None with db where T: Printable:
    return None
`)
	if len(page.Functions) != 1 {
		t.Fatalf("expected 1 function")
	}
	f := page.Functions[0]
	if f.Signature != "fn save[T](item: T) -> None" {
		t.Errorf("unexpected signature: %q", f.Signature)
	}
	if len(f.Effects) != 1 || f.Effects[0] != "db" {
		t.Errorf("expected effects [db], got %v", f.Effects)
	}
	if len(f.Constraints) != 1 || f.Constraints[0] != "T: Printable" {
		t.Errorf("expected constraints [T: Printable], got %v", f.Constraints)
	}
	md := page.Markdown()
	assertContains(t, md, "**Effects:** db")
	assertContains(t, md, "**Constraints:** T: Printable")
}

// --- Types ---

func TestDocTypeAlias(t *testing.T) {
	page := parse(t, `module test

## A user identifier.
pub type UserId = String
`)
	if len(page.Types) != 1 {
		t.Fatalf("expected 1 type")
	}
	ty := page.Types[0]
	if ty.Name != "UserId" || ty.Body != "String" {
		t.Errorf("unexpected type: name=%q body=%q", ty.Name, ty.Body)
	}
	assertContains(t, page.Markdown(), "### `type UserId = String`")
	assertContains(t, page.Markdown(), "A user identifier.")
}

func TestDocGenericTypeAlias(t *testing.T) {
	page := parse(t, `module test

pub type Maybe[T] = Option[T]
`)
	if len(page.Types) != 1 {
		t.Fatalf("expected 1 type")
	}
	assertContains(t, page.Markdown(), "type Maybe[T] = Option[T]")
}

// --- Structs ---

func TestDocStruct(t *testing.T) {
	page := parse(t, `module test

## Represents a task.
pub struct Task:
    pub id: String
    pub title: String
    pub priority: Int
`)
	if len(page.Structs) != 1 {
		t.Fatalf("expected 1 struct")
	}
	s := page.Structs[0]
	if s.Name != "Task" {
		t.Errorf("expected struct Task")
	}
	if len(s.Fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(s.Fields))
	}
	md := page.Markdown()
	assertContains(t, md, "### `struct Task`")
	assertContains(t, md, "Represents a task.")
	assertContains(t, md, "`id: String`")
	assertContains(t, md, "`priority: Int`")
}

func TestDocStructWithImplMethods(t *testing.T) {
	page := parse(t, `module test

pub struct Point:
    pub x: Int
    pub y: Int

impl Point:
    pub fn distance(self) -> Float:
        return 0.0
`)
	if len(page.Structs) != 1 {
		t.Fatalf("expected 1 struct")
	}
	s := page.Structs[0]
	if len(s.Methods) != 1 {
		t.Errorf("expected 1 method, got %d", len(s.Methods))
	}
	assertContains(t, page.Markdown(), "**Methods:**")
	assertContains(t, page.Markdown(), "`fn distance(self) -> Float`")
}

// --- Enums ---

func TestDocEnum(t *testing.T) {
	page := parse(t, `module test

## Task status values.
pub enum Status:
    Pending
    InProgress
    Done(String)
`)
	if len(page.Enums) != 1 {
		t.Fatalf("expected 1 enum")
	}
	e := page.Enums[0]
	if e.Name != "Status" {
		t.Errorf("expected enum Status")
	}
	md := page.Markdown()
	assertContains(t, md, "### `enum Status`")
	assertContains(t, md, "Task status values.")
	assertContains(t, md, "`Pending`")
	assertContains(t, md, "`Done(String)`")
}

// --- Traits ---

func TestDocTrait(t *testing.T) {
	page := parse(t, `module test

## Types that can be displayed.
pub trait Printable:
    fn display(self) -> String
`)
	if len(page.Traits) != 1 {
		t.Fatalf("expected 1 trait")
	}
	md := page.Markdown()
	assertContains(t, md, "### `trait Printable`")
	assertContains(t, md, "Types that can be displayed.")
	assertContains(t, md, "`fn display(self) -> String`")
}

// --- Specs ---

func TestDocSpec(t *testing.T) {
	page := parse(t, `module test

spec CreateTask:
    doc: "Creates a new task."
    inputs:
        title: String - "The task title"
    guarantees:
        - "Returns a Task with status pending"
    effects: db
    errors:
        InvalidTitle(String) - "When title is empty"
`)
	if len(page.Specs) != 1 {
		t.Fatalf("expected 1 spec")
	}
	md := page.Markdown()
	assertContains(t, md, "### `spec CreateTask`")
	assertContains(t, md, "Creates a new task.")
	assertContains(t, md, "`title: String`")
	assertContains(t, md, "Returns a Task with status pending")
	assertContains(t, md, "**Effects:** db")
	assertContains(t, md, "`InvalidTitle`")
}

// --- JSON output ---

func TestDocJSON(t *testing.T) {
	page := parse(t, `module test

pub fn hello() -> String:
    return "hi"
`)
	j := page.JSON()
	assertContains(t, j, `"module": "test"`)
	assertContains(t, j, `"fn hello() -> String"`)
}
