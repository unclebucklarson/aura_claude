package formatter

import (
	"os"
	"strings"
	"testing"

	"github.com/unclebucklarson/aura/pkg/lexer"
	"github.com/unclebucklarson/aura/pkg/parser"
)

func formatSource(t *testing.T, src string) string {
	t.Helper()
	l := lexer.New(src, "test.aura")
	tokens, lexErrors := l.Tokenize()
	if len(lexErrors) > 0 {
		t.Fatalf("lex errors: %v", lexErrors)
	}
	p := parser.New(tokens, "test.aura")
	module, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}
	f := New()
	return f.Format(module)
}

func TestFormatSimple(t *testing.T) {
	src := `module simple

type Name = String

pub struct Point:
    pub x: Float
    pub y: Float

pub fn add(a: Int, b: Int) -> Int:
    return a + b
`
	output := formatSource(t, src)
	if !strings.Contains(output, "module simple") {
		t.Error("expected 'module simple' in output")
	}
	if !strings.Contains(output, "pub struct Point:") {
		t.Error("expected 'pub struct Point:' in output")
	}
	if !strings.Contains(output, "pub fn add(") {
		t.Error("expected 'pub fn add(' in output")
	}
	if !strings.Contains(output, "return a + b") {
		t.Error("expected 'return a + b' in output")
	}
}

func TestFormatEnum(t *testing.T) {
	src := `module test

pub enum Color:
    Red
    Green
    Blue
`
	output := formatSource(t, src)
	if !strings.Contains(output, "pub enum Color:") {
		t.Error("expected enum definition")
	}
	if !strings.Contains(output, "    Red") {
		t.Error("expected indented variant 'Red'")
	}
}

func TestFormatIfStmt(t *testing.T) {
	src := `module test

pub fn f(x: Int) -> String:
    if x > 0:
        return "pos"
    else:
        return "neg"
`
	output := formatSource(t, src)
	if !strings.Contains(output, "if x > 0:") {
		t.Error("expected if statement")
	}
	if !strings.Contains(output, "else:") {
		t.Error("expected else clause")
	}
}

func TestFormatSpec(t *testing.T) {
	src := `module test

spec CreateTask:
    doc: "Creates a task."

    inputs:
        title: String - "The title"

    guarantees:
        - "Returns a task"

    effects: db, time

    errors:
        InvalidTitle(String) - "Bad title"
`
	output := formatSource(t, src)
	if !strings.Contains(output, "spec CreateTask:") {
		t.Error("expected spec block")
	}
	if !strings.Contains(output, "effects: db, time") {
		t.Error("expected effects line")
	}
}

func TestRoundTrip(t *testing.T) {
	// Parse → format → parse again and check both produce same formatting
	src := `module test

pub struct Point:
    pub x: Float
    pub y: Float

pub fn add(a: Int, b: Int) -> Int:
    return a + b
`
	// First pass
	output1 := formatSource(t, src)

	// Second pass (format the formatted output)
	output2 := formatSource(t, output1)

	if output1 != output2 {
		t.Errorf("round-trip failed:\nFirst:\n%s\nSecond:\n%s", output1, output2)
	}
}

func TestRoundTripModels(t *testing.T) {
	data, err := os.ReadFile("../../testdata/models.aura")
	if err != nil {
		t.Skip("testdata/models.aura not found")
	}
	src := string(data)
	output1 := formatSource(t, src)
	output2 := formatSource(t, output1)
	if output1 != output2 {
		t.Errorf("round-trip failed for models.aura:\nFirst:\n%s\nSecond:\n%s", output1, output2)
	}
}

func TestRoundTripControlFlow(t *testing.T) {
	data, err := os.ReadFile("../../testdata/control_flow.aura")
	if err != nil {
		t.Skip("testdata/control_flow.aura not found")
	}
	src := string(data)
	output1 := formatSource(t, src)
	output2 := formatSource(t, output1)
	if output1 != output2 {
		t.Errorf("round-trip failed for control_flow.aura:\nFirst:\n%s\nSecond:\n%s", output1, output2)
	}
}

func TestRoundTripService(t *testing.T) {
	data, err := os.ReadFile("../../testdata/service.aura")
	if err != nil {
		t.Skip("testdata/service.aura not found")
	}
	src := string(data)
	output1 := formatSource(t, src)
	output2 := formatSource(t, output1)
	if output1 != output2 {
		t.Errorf("round-trip failed for service.aura:\nFirst:\n%s\nSecond:\n%s", output1, output2)
	}
}

func TestFormatDeterministic(t *testing.T) {
	src := `module test

pub fn add(a: Int, b: Int) -> Int:
    return a + b
`
	// Format 10 times, should always produce same result
	first := formatSource(t, src)
	for i := 0; i < 10; i++ {
		output := formatSource(t, src)
		if output != first {
			t.Errorf("non-deterministic output on iteration %d", i)
		}
	}
}
