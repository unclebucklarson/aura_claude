package goemit

import (
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	auralex "github.com/unclebucklarson/aura/pkg/lexer"
	auraparser "github.com/unclebucklarson/aura/pkg/parser"
)

// compile parses Aura source and emits Go, then verifies the Go parses cleanly.
func compile(t *testing.T, src string) string {
	t.Helper()
	l := auralex.New(src, "<test>")
	tokens, lexErrs := l.Tokenize()
	if len(lexErrs) > 0 {
		t.Fatalf("lex: %v", lexErrs[0])
	}
	p := auraparser.New(tokens, "<test>")
	mod, parseErrs := p.Parse()
	if len(parseErrs) > 0 {
		t.Fatalf("parse: %v", parseErrs[0])
	}
	em := New()
	out, warns := em.Emit(mod)
	for _, w := range warns {
		t.Logf("warn: %s", w)
	}
	// Verify the emitted Go is syntactically valid.
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "<gen>", out, 0); err != nil {
		t.Fatalf("emitted Go is not valid:\n%s\n\nerror: %v", out, err)
	}
	return out
}

// TestHelloWorld checks that a simple main function compiles to valid Go.
func TestHelloWorld(t *testing.T) {
	src := `module main

fn main():
    println("Hello, World!")
`
	out := compile(t, src)
	if !strings.Contains(out, "fmt.Println") {
		t.Errorf("expected fmt.Println in output:\n%s", out)
	}
}

// TestArithmetic checks basic integer arithmetic.
func TestArithmetic(t *testing.T) {
	src := `module math

fn add(a: Int, b: Int) -> Int:
    return a + b

fn multiply(a: Int, b: Int) -> Int:
    let result = a * b
    return result
`
	out := compile(t, src)
	if !strings.Contains(out, "func Add") && !strings.Contains(out, "func add") {
		t.Errorf("expected add function in output:\n%s", out)
	}
}

// TestStruct checks struct definition and construction.
func TestStruct(t *testing.T) {
	src := `module shapes

pub struct Point:
    x: Float
    y: Float

fn make_point(x: Float, y: Float) -> Point:
    return Point(x: x, y: y)
`
	out := compile(t, src)
	if !strings.Contains(out, "type Point struct") {
		t.Errorf("expected struct definition:\n%s", out)
	}
	if !strings.Contains(out, "Point{") {
		t.Errorf("expected struct literal:\n%s", out)
	}
}

// TestSimpleEnum checks a unit enum compiles to int consts.
func TestSimpleEnum(t *testing.T) {
	src := `module colors

enum Color:
    Red
    Green
    Blue
`
	out := compile(t, src)
	// Private enum → lowercase in Go (unexported)
	if !strings.Contains(out, "type color int") && !strings.Contains(out, "type Color int") {
		t.Errorf("expected color type:\n%s", out)
	}
	if !strings.Contains(out, "iota") {
		t.Errorf("expected iota:\n%s", out)
	}
}

// TestIfElse checks if/elif/else emission.
func TestIfElse(t *testing.T) {
	src := `module test

fn classify(n: Int) -> String:
    if n < 0:
        return "negative"
    elif n == 0:
        return "zero"
    else:
        return "positive"
`
	out := compile(t, src)
	if !strings.Contains(out, "} else if") {
		t.Errorf("expected else if:\n%s", out)
	}
}

// TestForLoop checks for-in loop emission.
func TestForLoop(t *testing.T) {
	src := `module test

fn sum(nums: [Int]) -> Int:
    let mut total = 0
    for n in nums:
        total = total + n
    return total
`
	out := compile(t, src)
	if !strings.Contains(out, "range") {
		t.Errorf("expected range in output:\n%s", out)
	}
}

// TestWhileLoop checks while loop emission.
func TestWhileLoop(t *testing.T) {
	src := `module test

fn countdown(n: Int) -> Int:
    let mut x = n
    while x > 0:
        x = x - 1
    return x
`
	out := compile(t, src)
	if !strings.Contains(out, "for ") {
		t.Errorf("expected for (while) in output:\n%s", out)
	}
}

// TestOptionType checks Option[T] maps to auraOption.
func TestOptionType(t *testing.T) {
	src := `module test

fn maybe_double(n: Int) -> Option[Int]:
    if n > 0:
        return Some(n * 2)
    return None
`
	out := compile(t, src)
	if !strings.Contains(out, "auraOption") {
		t.Errorf("expected auraOption in output:\n%s", out)
	}
	if !strings.Contains(out, "auraSome") {
		t.Errorf("expected auraSome in output:\n%s", out)
	}
}

// TestStringInterpolation checks that interpolated strings use fmt.Sprintf.
func TestStringInterpolation(t *testing.T) {
	src := `module test

fn greet(name: String) -> String:
    return "Hello, {name}!"
`
	out := compile(t, src)
	if !strings.Contains(out, "fmt.Sprintf") {
		t.Errorf("expected fmt.Sprintf for interpolation:\n%s", out)
	}
}

// TestListExpr checks list literal emission.
func TestListExpr(t *testing.T) {
	src := `module test

fn make_list() -> [Int]:
    return [1, 2, 3]
`
	out := compile(t, src)
	if !strings.Contains(out, "[]any{") {
		t.Errorf("expected list literal in output:\n%s", out)
	}
}

// TestTypeDef checks type alias emission.
func TestTypeDef(t *testing.T) {
	src := `module test

type Score = Int
`
	out := compile(t, src)
	// Private type → lowercase in Go
	if !strings.Contains(out, "type score = int64") && !strings.Contains(out, "type Score = int64") {
		t.Errorf("expected score type alias in output:\n%s", out)
	}
}

// TestFormattedOutput checks that the emitted Go can be gofmt'd without error.
func TestFormattedOutput(t *testing.T) {
	src := `module main

fn main():
    let x = 42
    println(x)
`
	out := compile(t, src)
	_, err := format.Source([]byte(out))
	if err != nil {
		t.Errorf("gofmt failed: %v\n%s", err, out)
	}
}

// TestPackageName checks that a module with a main() gets package main.
func TestPackageName(t *testing.T) {
	src := `module myapp

fn main():
    println("hi")
`
	out := compile(t, src)
	if !strings.HasPrefix(strings.TrimSpace(out), "package main") {
		t.Errorf("expected package main:\n%s", out)
	}
}

// TestNonMainPackage checks that a library module gets the module name.
func TestNonMainPackage(t *testing.T) {
	src := `module utils

pub fn double(n: Int) -> Int:
    return n * 2
`
	out := compile(t, src)
	if !strings.HasPrefix(strings.TrimSpace(out), "package utils") {
		t.Errorf("expected package utils:\n%s", out)
	}
}
