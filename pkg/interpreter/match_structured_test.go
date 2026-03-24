package interpreter

import (
        "strings"
        "testing"
)

// ============================================================
// Tuple Pattern Tests
// ============================================================

func TestMatchTupleBasic(t *testing.T) {
        src := `module test
fn check(x: Int, y: Int) -> String:
    let point = (x, y)
    let result = match point:
        (0, 0) -> "origin"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: 0}, &IntVal{Val: 0}})
        expectString(t, result, "origin")
}

func TestMatchTupleWithVariables(t *testing.T) {
        src := `module test
fn check(x: Int, y: Int) -> String:
    let point = (x, y)
    let result = match point:
        (0, 0) -> "origin"
        (a, 0) -> "x-axis"
        (0, b) -> "y-axis"
        (a, b) -> "other"
    return result
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: 5}, &IntVal{Val: 0}})
        expectString(t, result, "x-axis")
}

func TestMatchTupleYAxis(t *testing.T) {
        src := `module test
fn check(x: Int, y: Int) -> String:
    let point = (x, y)
    let result = match point:
        (0, 0) -> "origin"
        (a, 0) -> "x-axis"
        (0, b) -> "y-axis"
        (a, b) -> "other"
    return result
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: 0}, &IntVal{Val: 7}})
        expectString(t, result, "y-axis")
}

func TestMatchTupleGeneral(t *testing.T) {
        src := `module test
fn check(x: Int, y: Int) -> String:
    let point = (x, y)
    let result = match point:
        (0, 0) -> "origin"
        (a, 0) -> "x-axis"
        (0, b) -> "y-axis"
        (a, b) -> "other"
    return result
`
        result := runFunc(t, src, "check", []Value{&IntVal{Val: 3}, &IntVal{Val: 4}})
        expectString(t, result, "other")
}

func TestMatchTupleThreeElements(t *testing.T) {
        src := `module test
fn check() -> String:
    let t = (1, 2, 3)
    let result = match t:
        (1, 2, 3) -> "exact"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "exact")
}

func TestMatchTupleMismatchedSize(t *testing.T) {
        src := `module test
fn check() -> String:
    let t = (1, 2, 3)
    let result = match t:
        (1, 2) -> "two"
        _ -> "fallback"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "fallback")
}

func TestMatchTupleWithWildcard(t *testing.T) {
        src := `module test
fn check() -> String:
    let t = (1, 2)
    let result = match t:
        (_, 2) -> "second is two"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "second is two")
}

func TestMatchNestedTuples(t *testing.T) {
        src := `module test
fn check() -> String:
    let t = ((1, 2), 3)
    let result = match t:
        ((1, 2), 3) -> "nested match"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "nested match")
}

// ============================================================
// List Pattern Tests
// ============================================================

func TestMatchListEmpty(t *testing.T) {
        src := `module test
fn check() -> String:
    let lst = []
    let result = match lst:
        [] -> "empty"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "empty")
}

func TestMatchListSingleElement(t *testing.T) {
        src := `module test
fn check() -> String:
    let lst = [42]
    let result = match lst:
        [] -> "empty"
        [x] -> "single"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "single")
}

func TestMatchListTwoElements(t *testing.T) {
        src := `module test
fn check() -> String:
    let lst = [1, 2]
    let result = match lst:
        [] -> "empty"
        [x] -> "single"
        [x, y] -> "pair"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "pair")
}

func TestMatchListWithLiterals(t *testing.T) {
        src := `module test
fn check() -> String:
    let lst = [1, 2, 3]
    let result = match lst:
        [1, 2, 3] -> "exact"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "exact")
}

func TestMatchListMismatchedSize(t *testing.T) {
        src := `module test
fn check() -> String:
    let lst = [1, 2, 3]
    let result = match lst:
        [1, 2] -> "two"
        _ -> "fallback"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "fallback")
}

func TestMatchListSpreadBasic(t *testing.T) {
        src := `module test
fn check() -> String:
    let lst = [1, 2, 3, 4]
    let result = match lst:
        [first, ...rest] -> "first is " + str(first)
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "first is 1")
}

func TestMatchListSpreadEmpty(t *testing.T) {
        src := `module test
fn check() -> String:
    let lst = [1]
    let result = match lst:
        [first, ...rest] -> "rest len: " + str(len(rest))
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "rest len: 0")
}

func TestMatchListSpreadMultiple(t *testing.T) {
        src := `module test
fn check() -> String:
    let lst = [10, 20, 30, 40, 50]
    let result = match lst:
        [a, b, ...rest] -> "rest len: " + str(len(rest))
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "rest len: 3")
}

func TestMatchListSpreadNotEnough(t *testing.T) {
        src := `module test
fn check() -> String:
    let lst = []
    let result = match lst:
        [first, ...rest] -> "has elements"
        [] -> "empty"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "empty")
}

func TestMatchListNestedPatterns(t *testing.T) {
        src := `module test
fn check() -> String:
    let lst = [1, 2]
    let result = match lst:
        [1, _] -> "starts with one"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "starts with one")
}

// ============================================================
// Constructor Pattern Tests
// ============================================================

func TestMatchSomeBasic(t *testing.T) {
        src := `module test
fn check() -> String:
    let val = Some(42)
    let result = match val:
        Some(x) -> "has: " + str(x)
        None -> "empty"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "has: 42")
}

func TestMatchNoneBasic(t *testing.T) {
        src := `module test
fn check() -> String:
    let val = None
    let result = match val:
        Some(x) -> "has: " + str(x)
        None -> "empty"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "empty")
}

func TestMatchOkBasic(t *testing.T) {
        src := `module test
fn check() -> String:
    let val = Ok(100)
    let result = match val:
        Ok(v) -> "success: " + str(v)
        Err(e) -> "error: " + str(e)
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "success: 100")
}

func TestMatchErrBasic(t *testing.T) {
        src := `module test
fn check() -> String:
    let val = Err("oops")
    let result = match val:
        Ok(v) -> "success: " + str(v)
        Err(e) -> "error: " + e
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "error: oops")
}

func TestMatchSomeWithLiteral(t *testing.T) {
        src := `module test
fn check() -> String:
    let val = Some(42)
    let result = match val:
        Some(42) -> "exact"
        Some(x) -> "other some"
        None -> "empty"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "exact")
}

func TestMatchSomeWithWildcard(t *testing.T) {
        src := `module test
fn check() -> String:
    let val = Some(99)
    let result = match val:
        Some(_) -> "has something"
        None -> "empty"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "has something")
}

// ============================================================
// Nested Pattern Tests
// ============================================================

func TestMatchSomeWithTuple(t *testing.T) {
        src := `module test
fn check() -> String:
    let val = Some((1, 2))
    let result = match val:
        Some((x, y)) -> "point: " + str(x) + "," + str(y)
        None -> "empty"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "point: 1,2")
}

func TestMatchSomeWithList(t *testing.T) {
        src := `module test
fn check() -> String:
    let val = Some([1, 2, 3])
    let result = match val:
        Some([first, ...rest]) -> "first: " + str(first) + " rest: " + str(len(rest))
        None -> "empty"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "first: 1 rest: 2")
}

func TestMatchTupleWithConstructor(t *testing.T) {
        src := `module test
fn check() -> String:
    let t = (Some(1), None)
    let result = match t:
        (Some(x), None) -> "first: " + str(x)
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "first: 1")
}

func TestMatchListOfTuples(t *testing.T) {
        src := `module test
fn check() -> String:
    let lst = [(1, 2)]
    let result = match lst:
        [(a, b)] -> "pair: " + str(a) + "," + str(b)
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "pair: 1,2")
}

func TestMatchDeepNesting(t *testing.T) {
        src := `module test
fn check() -> String:
    let val = Ok(Some((1, 2)))
    let result = match val:
        Ok(Some((x, y))) -> str(x) + "+" + str(y)
        Ok(None) -> "ok-none"
        Err(e) -> "err"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "1+2")
}

// ============================================================
// Edge Cases and Error Handling
// ============================================================

func TestMatchTupleNotATuple(t *testing.T) {
        src := `module test
fn check() -> String:
    let val = 42
    let result = match val:
        (1, 2) -> "tuple"
        42 -> "number"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "number")
}

func TestMatchListNotAList(t *testing.T) {
        src := `module test
fn check() -> String:
    let val = 42
    let result = match val:
        [1, 2] -> "list"
        42 -> "number"
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "number")
}

func TestMatchNoPatternPanicStructured(t *testing.T) {
        src := `module test
fn check() -> String:
    let val = (1, 2, 3)
    let result = match val:
        (0, 0, 0) -> "origin"
    return result
`
        module := parseModule(t, src)
        interp := New(module)
        _, runErr := interp.Run()
        if runErr != nil {
                t.Fatalf("run error: %v", runErr)
        }
        _, err := interp.RunFunction("check", nil)
        if err == nil {
                t.Fatal("expected error for no matching pattern")
        }
        if !strings.Contains(err.Error(), "no pattern matched") {
                t.Fatalf("unexpected error: %s", err.Error())
        }
}

func TestMatchSpreadWithTrailingPattern(t *testing.T) {
        src := `module test
fn check() -> String:
    let lst = [1, 2, 3, 4, 5]
    let result = match lst:
        [first, ...middle, last] -> "first: " + str(first) + " last: " + str(last) + " mid: " + str(len(middle))
        _ -> "other"
    return result
`
        result := runFunc(t, src, "check", nil)
        expectString(t, result, "first: 1 last: 5 mid: 3")
}
