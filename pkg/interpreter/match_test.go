package interpreter

import (
        "strings"
        "testing"
)

// === Match Expression - Literal Patterns (Int) ===

func TestMatchExprIntLiteralFirst(t *testing.T) {
        src := `module test
fn check(x: Int) -> String:
    let result = match x:
        1 -> "one"
        2 -> "two"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: 1}})
        expectString(t, result, "one")
}

func TestMatchExprIntLiteralSecond(t *testing.T) {
        src := `module test
fn check(x: Int) -> String:
    let result = match x:
        1 -> "one"
        2 -> "two"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: 2}})
        expectString(t, result, "two")
}

func TestMatchExprIntLiteralWildcard(t *testing.T) {
        src := `module test
fn check(x: Int) -> String:
    let result = match x:
        1 -> "one"
        2 -> "two"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: 99}})
        expectString(t, result, "other")
}

func TestMatchExprIntLiteralZero(t *testing.T) {
        src := `module test
fn check(x: Int) -> String:
    let result = match x:
        0 -> "zero"
        _ -> "nonzero"
    return result
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: 0}})
        expectString(t, result, "zero")
}

func TestMatchExprIntNegative(t *testing.T) {
        // Negative literal as expression: -1 is a unary expression, not a literal pattern
        // Use variable binding for now
        src := `module test
fn check(x: Int) -> String:
    let result = match x:
        0 -> "zero"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: -1}})
        expectString(t, result, "other")
}

// === Match Expression - Float Literal Patterns ===

func TestMatchExprFloatLiteral(t *testing.T) {
        src := `module test
fn check(x: Float) -> String:
    let result = match x:
        3.14 -> "pi"
        2.71 -> "e"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", []Value{&FloatVal{Val: 3.14}})
        expectString(t, result, "pi")
}

func TestMatchExprFloatLiteralSecond(t *testing.T) {
        src := `module test
fn check(x: Float) -> String:
    let result = match x:
        3.14 -> "pi"
        2.71 -> "e"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", []Value{&FloatVal{Val: 2.71}})
        expectString(t, result, "e")
}

// === Match Expression - String Literal Patterns ===

func TestMatchExprStringLiteral(t *testing.T) {
        src := `module test
fn check(s: String) -> Int:
    let result = match s:
        "hello" -> 1
        "world" -> 2
        _ -> 0
    return result
`
        result := runFunc(t, src, "check", []Value{&StringVal{Val: "hello"}})
        expectInt(t, result, 1)
}

func TestMatchExprStringLiteralNoMatch(t *testing.T) {
        src := `module test
fn check(s: String) -> Int:
    let result = match s:
        "hello" -> 1
        "world" -> 2
        _ -> 0
    return result
`
        result := runFunc(t, src, "check", []Value{&StringVal{Val: "foo"}})
        expectInt(t, result, 0)
}

// === Match Expression - Bool Literal Patterns ===

func TestMatchExprBoolLiteralTrue(t *testing.T) {
        src := `module test
fn check(b: Bool) -> String:
    let result = match b:
        true -> "yes"
        false -> "no"
    return result
`
        result := runFunc(t, src, "check", []Value{&BoolVal{Val: true}})
        expectString(t, result, "yes")
}

func TestMatchExprBoolLiteralFalse(t *testing.T) {
        src := `module test
fn check(b: Bool) -> String:
    let result = match b:
        true -> "yes"
        false -> "no"
    return result
`
        result := runFunc(t, src, "check", []Value{&BoolVal{Val: false}})
        expectString(t, result, "no")
}

// === Match Expression - Wildcard Pattern ===

func TestMatchExprWildcardAlwaysMatches(t *testing.T) {
        src := `module test
fn check(x: Int) -> String:
    let result = match x:
        _ -> "always"
    return result
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: 42}})
        expectString(t, result, "always")
}

// === Match Expression - Variable Binding Patterns ===

func TestMatchExprVariableBinding(t *testing.T) {
        src := `module test
fn check(x: Int) -> Int:
    let result = match x:
        n -> n + 1
    return result
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: 10}})
        expectInt(t, result, 11)
}

func TestMatchExprVariableBindingAfterLiteral(t *testing.T) {
        src := `module test
fn check(x: Int) -> String:
    let result = match x:
        0 -> "zero"
        n -> "nonzero"
    return result
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: 5}})
        expectString(t, result, "nonzero")
}

func TestMatchExprVariableBindingUsedInBody(t *testing.T) {
        src := `module test
fn double_or_zero(x: Int) -> Int:
    let result = match x:
        0 -> 0
        n -> n * 2
    return result
`
        result := runFunc(t, src, "double_or_zero", []Value{&IntVal{Val: 7}})
        expectInt(t, result, 14)
}

// === Match Expression - First-Match Semantics ===

func TestMatchExprFirstMatchWins(t *testing.T) {
        src := `module test
fn check(x: Int) -> String:
    let result = match x:
        1 -> "first"
        1 -> "second"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: 1}})
        expectString(t, result, "first")
}

func TestMatchExprVariableBeforeWildcard(t *testing.T) {
        // Variable binding matches before wildcard
        src := `module test
fn check(x: Int) -> Int:
    let result = match x:
        n -> n
        _ -> 0
    return result
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: 42}})
        expectInt(t, result, 42)
}

// === Match Expression - No Match Error ===

func TestMatchExprNoMatchPanics(t *testing.T) {
        src := `module test
fn check(x: Int) -> String:
    let result = match x:
        1 -> "one"
        2 -> "two"
    return result
`
        interp := runModule(t, src)
        _, err := interp.RunFunction("check", []Value{&IntVal{Val: 3}})
        if err == nil {
                t.Fatal("expected error for no matching pattern")
        }
        if !strings.Contains(err.Error(), "no pattern matched") {
                t.Fatalf("expected 'no pattern matched' error, got: %s", err.Error())
        }
}

// === Match Expression - Mixed Pattern Types ===

func TestMatchExprMixedLiteralAndWildcard(t *testing.T) {
        src := `module test
fn describe(s: String) -> String:
    let result = match s:
        "success" -> "ok"
        "error" -> "fail"
        _ -> "unknown"
    return result
`
        result := runFunc(t, src, "describe", []Value{&StringVal{Val: "error"}})
        expectString(t, result, "fail")
}

func TestMatchExprMixedLiteralAndVariable(t *testing.T) {
        src := `module test
fn describe(x: Int) -> String:
    let result = match x:
        0 -> "zero"
        1 -> "one"
        other -> "number"
    return result
`
        result := runFunc(t, src, "describe", []Value{&IntVal{Val: 100}})
        expectString(t, result, "number")
}

// === Match Expression - Used Inline ===

func TestMatchExprInReturn(t *testing.T) {
        src := `module test
fn check(x: Int) -> String:
    return match x:
        1 -> "one"
        _ -> "other"
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: 1}})
        expectString(t, result, "one")
}

// === Match Expression - Nested Match ===

func TestMatchExprNested(t *testing.T) {
        src := `module test
fn classify(x: Int, y: Int) -> String:
    let outer = match x:
        0 -> "x_zero"
        _ -> match y:
            0 -> "y_zero"
            _ -> "both_nonzero"
    return outer
`
        result := runFunc(t, src, "classify", []Value{&IntVal{Val: 1}, &IntVal{Val: 0}})
        expectString(t, result, "y_zero")
}

func TestMatchExprNestedXZero(t *testing.T) {
        src := `module test
fn classify(x: Int, y: Int) -> String:
    let outer = match x:
        0 -> "x_zero"
        _ -> match y:
            0 -> "y_zero"
            _ -> "both_nonzero"
    return outer
`
        result := runFunc(t, src, "classify", []Value{&IntVal{Val: 0}, &IntVal{Val: 5}})
        expectString(t, result, "x_zero")
}

// === Match Expression - With Different Value Types ===

func TestMatchExprNoneLiteral(t *testing.T) {
        src := `module test
fn check() -> String:
    let x = None
    let result = match x:
        None -> "nothing"
        _ -> "something"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "nothing")
}

// === Existing Match Statement Still Works ===

func TestMatchStmtStillWorks(t *testing.T) {
        src := `module test
fn check(x: Int) -> String:
    match x:
        case 1:
            return "one"
        case 2:
            return "two"
        case _:
            return "other"
    return "unreachable"
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: 2}})
        expectString(t, result, "two")
}
