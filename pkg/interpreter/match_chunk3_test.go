package interpreter

import (
	"testing"
)

// =============================================================================
// Phase 3.2 Chunk 3: Guard Clauses, Or-patterns, Binding Patterns (as)
// =============================================================================

// ---------------------------------------------------------------------------
// Guard clauses on match expressions
// ---------------------------------------------------------------------------

func TestMatchExprGuardPositive(t *testing.T) {
	src := `module test
fn check(n: Int) -> String:
    return match n:
        x if x > 0 -> "positive"
        x if x < 0 -> "negative"
        _ -> "zero"
`
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 5}}), "positive")
}

func TestMatchExprGuardNegative(t *testing.T) {
	src := `module test
fn check(n: Int) -> String:
    return match n:
        x if x > 0 -> "positive"
        x if x < 0 -> "negative"
        _ -> "zero"
`
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: -3}}), "negative")
}

func TestMatchExprGuardZero(t *testing.T) {
	src := `module test
fn check(n: Int) -> String:
    return match n:
        x if x > 0 -> "positive"
        x if x < 0 -> "negative"
        _ -> "zero"
`
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 0}}), "zero")
}

func TestMatchExprGuardFallthroughToNextArm(t *testing.T) {
	// Guard fails → next arm tried
	src := `module test
fn check(n: Int) -> String:
    return match n:
        1 if false -> "never"
        1 -> "one"
        _ -> "other"
`
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 1}}), "one")
}

func TestMatchExprGuardBindingUsedInBody(t *testing.T) {
	src := `module test
fn check(n: Int) -> Int:
    return match n:
        x if x > 10 -> x * 2
        x -> x + 1
`
	expectInt(t, runFunc(t, src, "check", []Value{&IntVal{Val: 20}}), 40)
	expectInt(t, runFunc(t, src, "check", []Value{&IntVal{Val: 5}}), 6)
}

func TestMatchExprGuardTuplePattern(t *testing.T) {
	src := `module test
fn check(x: Int, y: Int) -> String:
    let p = (x, y)
    return match p:
        (a, b) if a == b -> "equal"
        (a, b) if a > b  -> "first bigger"
        _ -> "second bigger"
`
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 3}, &IntVal{Val: 3}}), "equal")
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 5}, &IntVal{Val: 2}}), "first bigger")
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 1}, &IntVal{Val: 9}}), "second bigger")
}

func TestMatchExprGuardConstructorSome(t *testing.T) {
	src := `module test
fn check(n: Int) -> String:
    let v = Some(n)
    return match v:
        Some(x) if x > 100 -> "big"
        Some(x) -> "small"
        None -> "none"
`
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 200}}), "big")
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 5}}), "small")
}

func TestMatchExprGuardMultipleGuardsSamePattern(t *testing.T) {
	src := `module test
fn check(n: Int) -> String:
    return match n:
        x if x > 100 -> "huge"
        x if x > 50  -> "large"
        x if x > 10  -> "medium"
        _ -> "small"
`
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 200}}), "huge")
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 75}}), "large")
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 15}}), "medium")
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 3}}), "small")
}

// ---------------------------------------------------------------------------
// Or-patterns
// ---------------------------------------------------------------------------

func TestOrPatternIntLiterals(t *testing.T) {
	src := `module test
fn check(n: Int) -> String:
    return match n:
        1 | 2 | 3 -> "low"
        4 | 5 | 6 -> "mid"
        _ -> "high"
`
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 2}}), "low")
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 5}}), "mid")
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 9}}), "high")
}

func TestOrPatternStringLiterals(t *testing.T) {
	src := `module test
fn check(s: String) -> String:
    return match s:
        "yes" | "y" | "true"  -> "affirmative"
        "no"  | "n" | "false" -> "negative"
        _ -> "unknown"
`
	expectString(t, runFunc(t, src, "check", []Value{&StringVal{Val: "yes"}}), "affirmative")
	expectString(t, runFunc(t, src, "check", []Value{&StringVal{Val: "n"}}), "negative")
	expectString(t, runFunc(t, src, "check", []Value{&StringVal{Val: "maybe"}}), "unknown")
}

func TestOrPatternTwoAlternatives(t *testing.T) {
	src := `module test
fn check(n: Int) -> String:
    return match n:
        0 | 1 -> "zero or one"
        _ -> "other"
`
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 0}}), "zero or one")
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 1}}), "zero or one")
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 2}}), "other")
}

func TestOrPatternWithGuard(t *testing.T) {
	src := `module test
fn check(n: Int) -> String:
    return match n:
        0 | 1 if true -> "zero or one"
        _ -> "other"
`
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 0}}), "zero or one")
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 1}}), "zero or one")
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 2}}), "other")
}

func TestOrPatternEnumVariants(t *testing.T) {
	src := `module test
enum Direction:
    North
    South
    East
    West

fn check(d: Direction) -> String:
    return match d:
        Direction.North | Direction.South -> "vertical"
        Direction.East  | Direction.West  -> "horizontal"
`
	resultV := runFunc(t, src, "check", []Value{&EnumVal{TypeName: "Direction", VariantName: "North"}})
	expectString(t, resultV, "vertical")
	resultH := runFunc(t, src, "check", []Value{&EnumVal{TypeName: "Direction", VariantName: "East"}})
	expectString(t, resultH, "horizontal")
}

// ---------------------------------------------------------------------------
// Binding patterns (pattern as name)
// ---------------------------------------------------------------------------

func TestAsPatternBindsWhole(t *testing.T) {
	src := `module test
fn check(n: Int) -> Int:
    return match n:
        _ as x -> x + 10
`
	expectInt(t, runFunc(t, src, "check", []Value{&IntVal{Val: 5}}), 15)
}

func TestAsPatternOnTuple(t *testing.T) {
	src := `module test
fn check(x: Int, y: Int) -> Int:
    let t = (x, y)
    return match t:
        (a, b) as pair -> a + b
`
	expectInt(t, runFunc(t, src, "check", []Value{&IntVal{Val: 3}, &IntVal{Val: 4}}), 7)
}

func TestAsPatternOnLiteral(t *testing.T) {
	src := `module test
fn check(n: Int) -> Int:
    return match n:
        42 as the_answer -> the_answer * 2
        x -> x
`
	expectInt(t, runFunc(t, src, "check", []Value{&IntVal{Val: 42}}), 84)
	expectInt(t, runFunc(t, src, "check", []Value{&IntVal{Val: 7}}), 7)
}

func TestAsPatternOnConstructor(t *testing.T) {
	src := `module test
fn check(n: Int) -> String:
    let v = Some(n)
    return match v:
        Some(x) as s -> "got some"
        None -> "nothing"
`
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 42}}), "got some")
}

func TestAsPatternWithGuard(t *testing.T) {
	src := `module test
fn check(n: Int) -> String:
    return match n:
        _ as x if x > 0 -> "positive"
        _ -> "non-positive"
`
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: 7}}), "positive")
	expectString(t, runFunc(t, src, "check", []Value{&IntVal{Val: -1}}), "non-positive")
}

func TestAsPatternOnListSpread(t *testing.T) {
	src := `module test
fn check() -> Int:
    let xs = [10, 20, 30]
    return match xs:
        [first, ...rest] as whole -> first
        _ -> 0
`
	expectInt(t, runFunc(t, src, "check", []Value{}), 10)
}
