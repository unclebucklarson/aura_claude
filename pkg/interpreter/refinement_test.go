package interpreter

import (
	"strings"
	"testing"
)

// =============================================================================
// Issue #8: Refinement Type Runtime Enforcement
// =============================================================================

// ---------------------------------------------------------------------------
// Inline refinement on let bindings
// ---------------------------------------------------------------------------

func TestRefinementLetPassesWhenSatisfied(t *testing.T) {
	src := `module test
fn check() -> Int:
    let n: Int where self >= 0 = 5
    return n
`
	expectInt(t, runFunc(t, src, "check", nil), 5)
}

func TestRefinementLetPanicsWhenViolated(t *testing.T) {
	src := `module test
fn check():
    let n: Int where self >= 0 = -1
    return n
`
	expectRuntimeError(t, src, "check", "refinement type constraint violated")
}

func TestRefinementLetCompoundPredicate(t *testing.T) {
	src := `module test
fn check() -> Int:
    let n: Int where self >= 1 and self <= 10 = 5
    return n
`
	expectInt(t, runFunc(t, src, "check", nil), 5)
}

func TestRefinementLetCompoundPredicateViolatesLower(t *testing.T) {
	src := `module test
fn check():
    let n: Int where self >= 1 and self <= 10 = 0
    return n
`
	expectRuntimeError(t, src, "check", "refinement type constraint violated")
}

func TestRefinementLetCompoundPredicateViolatesUpper(t *testing.T) {
	src := `module test
fn check():
    let n: Int where self >= 1 and self <= 10 = 11
    return n
`
	expectRuntimeError(t, src, "check", "refinement type constraint violated")
}

// ---------------------------------------------------------------------------
// Named refinement types (type alias)
// ---------------------------------------------------------------------------

func TestRefinementNamedTypePassesWhenSatisfied(t *testing.T) {
	src := `module test
type Priority = Int where self >= 1 and self <= 5

fn check() -> Int:
    let p: Priority = 3
    return p
`
	expectInt(t, runFunc(t, src, "check", nil), 3)
}

func TestRefinementNamedTypePanicsWhenViolated(t *testing.T) {
	src := `module test
type Priority = Int where self >= 1 and self <= 5

fn check():
    let p: Priority = 0
    return p
`
	expectRuntimeError(t, src, "check", "refinement type constraint violated")
}

// ---------------------------------------------------------------------------
// Refinement on function parameters
// ---------------------------------------------------------------------------

func TestRefinementParamPassesWhenSatisfied(t *testing.T) {
	src := `module test
fn double(n: Int where self > 0) -> Int:
    return n * 2
`
	expectInt(t, runFunc(t, src, "double", []Value{&IntVal{Val: 5}}), 10)
}

func TestRefinementParamPanicsWhenViolated(t *testing.T) {
	src := `module test
fn double(n: Int where self > 0) -> Int:
    return n * 2
`
	module := parseModule(t, src)
	interp := New(module)
	if _, err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	_, err := interp.RunFunction("double", []Value{&IntVal{Val: -3}})
	if err == nil {
		t.Fatal("expected runtime error for refinement violation, got nil")
	}
	if !strings.Contains(err.Error(), "refinement type constraint violated") {
		t.Fatalf("expected refinement error, got: %v", err)
	}
}

func TestRefinementNamedParamPanicsWhenViolated(t *testing.T) {
	src := `module test
type Score = Int where self >= 0 and self <= 100

fn grade(s: Score) -> String:
    return "ok"
`
	module := parseModule(t, src)
	interp := New(module)
	if _, err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	_, err := interp.RunFunction("grade", []Value{&IntVal{Val: 150}})
	if err == nil {
		t.Fatal("expected runtime error for refinement violation, got nil")
	}
	if !strings.Contains(err.Error(), "refinement type constraint violated") {
		t.Fatalf("expected refinement error, got: %v", err)
	}
}
