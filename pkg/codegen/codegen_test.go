package codegen

import (
	"strings"
	"testing"

	"github.com/unclebucklarson/aura/pkg/ast"
	"github.com/unclebucklarson/aura/pkg/lexer"
	"github.com/unclebucklarson/aura/pkg/parser"
)

// parseModule is a helper that parses an Aura source string into a Module.
func parseModule(t *testing.T, src string) *ast.Module {
	t.Helper()
	l := lexer.New(src, "<test>")
	tokens, lexErrs := l.Tokenize()
	if len(lexErrs) > 0 {
		t.Fatalf("lex error: %v", lexErrs[0])
	}
	p := parser.New(tokens, "<test>")
	mod, parseErrs := p.Parse()
	if len(parseErrs) > 0 {
		t.Fatalf("parse error: %v", parseErrs[0])
	}
	return mod
}

// --- stripFences ---

func TestStripFences_NoFences(t *testing.T) {
	in := "fn foo():\n    return 1"
	got := stripFences(in)
	if got != in {
		t.Errorf("expected unchanged, got %q", got)
	}
}

func TestStripFences_WithFences(t *testing.T) {
	in := "```aura\nfn foo():\n    return 1\n```"
	got := stripFences(in)
	if strings.Contains(got, "```") {
		t.Errorf("fences not stripped: %q", got)
	}
	if !strings.Contains(got, "fn foo()") {
		t.Errorf("content missing: %q", got)
	}
}

func TestStripFences_GenericFence(t *testing.T) {
	in := "```\nfn bar():\n    return 2\n```"
	got := stripFences(in)
	if strings.Contains(got, "```") {
		t.Errorf("fences not stripped: %q", got)
	}
}

// --- FindUnimplementedSpecs ---

func TestFindUnimplementedSpecs_AllUnimplemented(t *testing.T) {
	src := `module test

spec Add:
    doc: "add two numbers"
    inputs:
        a: Int
        b: Int
`
	mod := parseModule(t, src)
	specs := FindUnimplementedSpecs(mod)
	if len(specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(specs))
	}
	if specs[0].Name != "Add" {
		t.Errorf("expected spec Add, got %s", specs[0].Name)
	}
}

func TestFindUnimplementedSpecs_Satisfied(t *testing.T) {
	src := `module test

spec Add:
    doc: "add two numbers"
    inputs:
        a: Int
        b: Int

fn add(a: Int, b: Int) -> Int satisfies Add:
    return a + b
`
	mod := parseModule(t, src)
	specs := FindUnimplementedSpecs(mod)
	if len(specs) != 0 {
		t.Errorf("expected 0 unimplemented specs, got %d", len(specs))
	}
}

func TestFindUnimplementedSpecs_Mixed(t *testing.T) {
	src := `module test

spec Add:
    doc: "add"
    inputs:
        a: Int
        b: Int

spec Subtract:
    doc: "subtract"
    inputs:
        a: Int
        b: Int

fn add(a: Int, b: Int) -> Int satisfies Add:
    return a + b
`
	mod := parseModule(t, src)
	specs := FindUnimplementedSpecs(mod)
	if len(specs) != 1 {
		t.Fatalf("expected 1 unimplemented spec, got %d", len(specs))
	}
	if specs[0].Name != "Subtract" {
		t.Errorf("expected Subtract, got %s", specs[0].Name)
	}
}

// --- ExtractContext ---

func TestExtractContext_Basic(t *testing.T) {
	src := `module math

struct Point:
    x: Float
    y: Float

enum Color:
    Red
    Green
    Blue

type Scalar = Float

fn distance(p: Point) -> Float:
    return 0.0
`
	mod := parseModule(t, src)
	ctx := ExtractContext(mod)
	if ctx.ModuleName != "math" {
		t.Errorf("expected module name math, got %q", ctx.ModuleName)
	}
	if len(ctx.Structs) != 1 || ctx.Structs[0].Name != "Point" {
		t.Errorf("expected 1 struct Point, got %v", ctx.Structs)
	}
	if len(ctx.Enums) != 1 || ctx.Enums[0].Name != "Color" {
		t.Errorf("expected 1 enum Color, got %v", ctx.Enums)
	}
	if len(ctx.Types) != 1 || ctx.Types[0].Name != "Scalar" {
		t.Errorf("expected 1 type Scalar, got %v", ctx.Types)
	}
	if len(ctx.Functions) != 1 || ctx.Functions[0].Name != "distance" {
		t.Errorf("expected 1 function distance, got %v", ctx.Functions)
	}
}

// --- BuildPrompt ---

func TestBuildPrompt_ContainsSpecName(t *testing.T) {
	src := `module test

spec Greet:
    doc: "Return a greeting string"
    inputs:
        name: String - "person's name"
`
	mod := parseModule(t, src)
	specs := FindUnimplementedSpecs(mod)
	if len(specs) == 0 {
		t.Fatal("no specs found")
	}
	ctx := ExtractContext(mod)
	prompt := BuildPrompt(specs[0], ctx)

	if !strings.Contains(prompt, "Greet") {
		t.Error("prompt missing spec name")
	}
	if !strings.Contains(prompt, "Return a greeting string") {
		t.Error("prompt missing doc description")
	}
	if !strings.Contains(prompt, "name: String") {
		t.Error("prompt missing input")
	}
	if !strings.Contains(prompt, "satisfies Greet") {
		t.Error("prompt missing satisfies instruction")
	}
}

func TestBuildPrompt_IncludesTypeContext(t *testing.T) {
	src := `module test

struct User:
    id: Int
    name: String

spec GetUser:
    doc: "Look up a user by id"
    inputs:
        id: Int
`
	mod := parseModule(t, src)
	specs := FindUnimplementedSpecs(mod)
	ctx := ExtractContext(mod)
	prompt := BuildPrompt(specs[0], ctx)

	if !strings.Contains(prompt, "struct User") {
		t.Error("prompt missing struct context")
	}
	if !strings.Contains(prompt, "id: Int") {
		t.Error("prompt missing struct field")
	}
}

func TestBuildPrompt_IncludesEffectsAndErrors(t *testing.T) {
	src := `module test

spec FetchData:
    doc: "Fetch remote data"
    inputs:
        url: String - "the URL to fetch"
    effects: net
    errors:
        NetworkError - "connection failed"
`
	mod := parseModule(t, src)
	specs := FindUnimplementedSpecs(mod)
	ctx := ExtractContext(mod)
	prompt := BuildPrompt(specs[0], ctx)

	if !strings.Contains(prompt, "net") {
		t.Error("prompt missing effect")
	}
	if !strings.Contains(prompt, "NetworkError") {
		t.Error("prompt missing error")
	}
}

func TestBuildPrompt_ContainsSyntaxGuide(t *testing.T) {
	src := `module test

spec Noop:
    doc: "does nothing"
`
	mod := parseModule(t, src)
	specs := FindUnimplementedSpecs(mod)
	ctx := ExtractContext(mod)
	prompt := BuildPrompt(specs[0], ctx)

	// Aura syntax cheat sheet should always be present
	if !strings.Contains(prompt, "fn name") {
		t.Error("prompt missing syntax guide")
	}
}

// --- Validate ---

func TestValidate_ValidCode(t *testing.T) {
	original := `module test

spec Add:
    inputs:
        a: Int
        b: Int
`
	generated := "fn add(a: Int, b: Int) -> Int satisfies Add:\n    return a + b\n"
	errs := Validate(original, generated, "<test>")
	if len(errs) > 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidate_InvalidCode(t *testing.T) {
	original := "module test\n"
	generated := "this is not valid aura code !!!\n"
	errs := Validate(original, generated, "<test>")
	if len(errs) == 0 {
		t.Error("expected parse/lex errors for invalid code")
	}
}
