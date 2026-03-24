package interpreter

import (
        "strings"
        "testing"
)

// === Tuple Literal Creation ===

func TestTupleLiteralTwoElements(t *testing.T) {
        src := `module test
fn make_tuple() -> (Int, Int):
    return (1, 2)
`
        result := runFunc(t, src, "make_tuple", nil)
        tv, ok := result.(*TupleVal)
        if !ok {
                t.Fatalf("expected TupleVal, got %T (%v)", result, result)
        }
        if len(tv.Elements) != 2 {
                t.Fatalf("expected 2 elements, got %d", len(tv.Elements))
        }
        expectInt(t, tv.Elements[0], 1)
        expectInt(t, tv.Elements[1], 2)
}

func TestTupleLiteralThreeElements(t *testing.T) {
        src := `module test
fn make_tuple() -> (Int, String, Bool):
    return (42, "hello", true)
`
        result := runFunc(t, src, "make_tuple", nil)
        tv, ok := result.(*TupleVal)
        if !ok {
                t.Fatalf("expected TupleVal, got %T (%v)", result, result)
        }
        if len(tv.Elements) != 3 {
                t.Fatalf("expected 3 elements, got %d", len(tv.Elements))
        }
        expectInt(t, tv.Elements[0], 42)
        expectString(t, tv.Elements[1], "hello")
        expectBool(t, tv.Elements[2], true)
}

func TestTupleLiteralSingleElement(t *testing.T) {
        src := `module test
fn make_tuple():
    return (42,)
`
        result := runFunc(t, src, "make_tuple", nil)
        tv, ok := result.(*TupleVal)
        if !ok {
                t.Fatalf("expected TupleVal, got %T (%v)", result, result)
        }
        if len(tv.Elements) != 1 {
                t.Fatalf("expected 1 element, got %d", len(tv.Elements))
        }
        expectInt(t, tv.Elements[0], 42)
}

func TestTupleLiteralEmpty(t *testing.T) {
        src := `module test
fn make_tuple():
    return ()
`
        result := runFunc(t, src, "make_tuple", nil)
        tv, ok := result.(*TupleVal)
        if !ok {
                t.Fatalf("expected TupleVal, got %T (%v)", result, result)
        }
        if len(tv.Elements) != 0 {
                t.Fatalf("expected 0 elements, got %d", len(tv.Elements))
        }
}

func TestTupleNestedTuples(t *testing.T) {
        src := `module test
fn nested():
    let t = ((1, 2), (3, 4))
    return t
`
        result := runFunc(t, src, "nested", nil)
        tv, ok := result.(*TupleVal)
        if !ok {
                t.Fatalf("expected TupleVal, got %T (%v)", result, result)
        }
        if len(tv.Elements) != 2 {
                t.Fatalf("expected 2 elements, got %d", len(tv.Elements))
        }
        inner1, ok := tv.Elements[0].(*TupleVal)
        if !ok {
                t.Fatalf("expected inner TupleVal, got %T", tv.Elements[0])
        }
        expectInt(t, inner1.Elements[0], 1)
        expectInt(t, inner1.Elements[1], 2)
}

// === Tuple Destructuring ===

func TestTupleDestructuringBasic(t *testing.T) {
        src := `module test
fn destruct() -> Int:
    let (x, y) = (10, 20)
    return x + y
`
        result := runFunc(t, src, "destruct", nil)
        expectInt(t, result, 30)
}

func TestTupleDestructuringThreeElements(t *testing.T) {
        src := `module test
fn destruct() -> String:
    let (a, b, c) = ("hello", " ", "world")
    return a + b + c
`
        result := runFunc(t, src, "destruct", nil)
        expectString(t, result, "hello world")
}

func TestTupleDestructuringWithWildcard(t *testing.T) {
        src := `module test
fn destruct() -> Int:
    let (x, _, z) = (1, 2, 3)
    return x + z
`
        result := runFunc(t, src, "destruct", nil)
        expectInt(t, result, 4)
}

func TestTupleDestructuringMutable(t *testing.T) {
        src := `module test
fn destruct() -> Int:
    let mut (x, y) = (10, 20)
    x = 100
    return x + y
`
        result := runFunc(t, src, "destruct", nil)
        expectInt(t, result, 120)
}

func TestTupleDestructuringFromList(t *testing.T) {
        src := `module test
fn destruct() -> Int:
    let (a, b) = [5, 10]
    return a + b
`
        result := runFunc(t, src, "destruct", nil)
        expectInt(t, result, 15)
}

func TestTupleDestructuringSizeMismatch(t *testing.T) {
        src := `module test
fn destruct():
    let (x, y, z) = (1, 2)
    return x
`
        module := parseModule(t, src)
        interp := New(module)
        _, err := interp.Run()
        if err != nil {
                t.Fatalf("run error: %v", err)
        }
        _, err = interp.RunFunction("destruct", nil)
        if err == nil {
                t.Fatal("expected error for size mismatch")
        }
        if !strings.Contains(err.Error(), "cannot destructure") {
                t.Fatalf("unexpected error: %s", err.Error())
        }
}

// === Tuple Indexing ===

func TestTupleIndexing(t *testing.T) {
        src := `module test
fn idx() -> Int:
    let t = (10, 20, 30)
    return t[1]
`
        result := runFunc(t, src, "idx", nil)
        expectInt(t, result, 20)
}

func TestTupleIndexOutOfBounds(t *testing.T) {
        src := `module test
fn idx():
    let t = (1, 2, 3)
    return t[5]
`
        module := parseModule(t, src)
        interp := New(module)
        _, err := interp.Run()
        if err != nil {
                t.Fatalf("run error: %v", err)
        }
        _, err = interp.RunFunction("idx", nil)
        if err == nil {
                t.Fatal("expected error for out of bounds")
        }
        if !strings.Contains(err.Error(), "out of bounds") {
                t.Fatalf("unexpected error: %s", err.Error())
        }
}

// === Tuple Methods ===

func TestTupleMethodLen(t *testing.T) {
        src := `module test
fn test_len() -> Int:
    let t = (1, 2, 3)
    return t.len()
`
        result := runFunc(t, src, "test_len", nil)
        expectInt(t, result, 3)
}

func TestTupleMethodLenEmpty(t *testing.T) {
        src := `module test
fn test_len() -> Int:
    let t = ()
    return t.len()
`
        result := runFunc(t, src, "test_len", nil)
        expectInt(t, result, 0)
}

func TestTupleMethodGet(t *testing.T) {
        src := `module test
fn test_get() -> Int:
    let t = (10, 20, 30)
    let val = t.get(1)
    return val.unwrap()
`
        result := runFunc(t, src, "test_get", nil)
        expectInt(t, result, 20)
}

func TestTupleMethodGetOutOfBounds(t *testing.T) {
        src := `module test
fn test_get() -> Bool:
    let t = (1, 2)
    let val = t.get(5)
    return val.is_none()
`
        result := runFunc(t, src, "test_get", nil)
        expectBool(t, result, true)
}

func TestTupleMethodToList(t *testing.T) {
        src := `module test
fn test_to_list() -> Int:
    let t = (1, 2, 3)
    let lst = t.to_list()
    return lst.len()
`
        result := runFunc(t, src, "test_to_list", nil)
        expectInt(t, result, 3)
}

func TestTupleMethodIsEmpty(t *testing.T) {
        src := `module test
fn test_empty() -> Bool:
    let t = ()
    return t.is_empty()
`
        result := runFunc(t, src, "test_empty", nil)
        expectBool(t, result, true)
}

func TestTupleMethodContains(t *testing.T) {
        src := `module test
fn test_contains() -> Bool:
    let t = (1, 2, 3)
    return t.contains(2)
`
        result := runFunc(t, src, "test_contains", nil)
        expectBool(t, result, true)
}

func TestTupleMethodContainsFalse(t *testing.T) {
        src := `module test
fn test_contains() -> Bool:
    let t = (1, 2, 3)
    return t.contains(99)
`
        result := runFunc(t, src, "test_contains", nil)
        expectBool(t, result, false)
}

func TestTupleMethodFirstLast(t *testing.T) {
        src := `module test
fn test_fl() -> Int:
    let t = (10, 20, 30)
    let f = t.first().unwrap()
    let l = t.last().unwrap()
    return f + l
`
        result := runFunc(t, src, "test_fl", nil)
        expectInt(t, result, 40)
}

func TestTupleMethodFirstLastEmpty(t *testing.T) {
        src := `module test
fn test_fl() -> Bool:
    let t = ()
    return t.first().is_none() and t.last().is_none()
`
        result := runFunc(t, src, "test_fl", nil)
        expectBool(t, result, true)
}

func TestTupleMethodReverse(t *testing.T) {
        src := `module test
fn test_rev() -> Int:
    let t = (1, 2, 3)
    let r = t.reverse()
    return r[0]
`
        result := runFunc(t, src, "test_rev", nil)
        expectInt(t, result, 3)
}

func TestTupleMethodMap(t *testing.T) {
        src := `module test
fn test_map() -> Int:
    let t = (1, 2, 3)
    let doubled = t.map(|x| -> x * 2)
    return doubled[0] + doubled[1] + doubled[2]
`
        result := runFunc(t, src, "test_map", nil)
        expectInt(t, result, 12)
}

func TestTupleMethodEnumerate(t *testing.T) {
        src := `module test
fn test_enum() -> Int:
    let t = (10, 20, 30)
    let pairs = t.enumerate()
    let first_pair = pairs[0]
    return first_pair[0] + first_pair[1]
`
        result := runFunc(t, src, "test_enum", nil)
        expectInt(t, result, 10) // 0 + 10
}

// === Tuple Iteration ===

func TestTupleForLoop(t *testing.T) {
        src := `module test
fn test_iter() -> Int:
    let t = (1, 2, 3, 4)
    let mut sum = 0
    for x in t:
        sum = sum + x
    return sum
`
        result := runFunc(t, src, "test_iter", nil)
        expectInt(t, result, 10)
}

// === Tuple Equality ===

func TestTupleEquality(t *testing.T) {
        src := `module test
fn test_eq() -> Bool:
    let a = (1, 2, 3)
    let b = (1, 2, 3)
    return a == b
`
        result := runFunc(t, src, "test_eq", nil)
        expectBool(t, result, true)
}

func TestTupleInequality(t *testing.T) {
        src := `module test
fn test_neq() -> Bool:
    let a = (1, 2, 3)
    let b = (1, 2, 4)
    return a != b
`
        result := runFunc(t, src, "test_neq", nil)
        expectBool(t, result, true)
}

// === Tuple String Representation ===

func TestTupleString(t *testing.T) {
        tv := &TupleVal{Elements: []Value{
                &IntVal{Val: 1},
                &StringVal{Val: "hello"},
                &BoolVal{Val: true},
        }}
        expected := `(1, "hello", true)`
        if tv.String() != expected {
                t.Fatalf("expected %q, got %q", expected, tv.String())
        }
}

func TestTupleEmptyString(t *testing.T) {
        tv := &TupleVal{Elements: nil}
        expected := `()`
        if tv.String() != expected {
                t.Fatalf("expected %q, got %q", expected, tv.String())
        }
}

// === Tuple Zip ===

func TestTupleMethodZip(t *testing.T) {
        src := `module test
fn test_zip() -> Int:
    let a = (1, 2, 3)
    let b = (10, 20, 30)
    let zipped = a.zip(b)
    let pair = zipped[1]
    return pair[0] + pair[1]
`
        result := runFunc(t, src, "test_zip", nil)
        expectInt(t, result, 22) // 2 + 20
}

// === Tuple as Function Return Value ===

func TestTupleAsReturnValue(t *testing.T) {
        src := `module test
fn divide(a: Int, b: Int):
    if b == 0:
        return (0, "division by zero")
    return (a / b, "ok")

fn test_return() -> Int:
    let (result, msg) = divide(10, 2)
    return result
`
        result := runFunc(t, src, "test_return", nil)
        expectInt(t, result, 5)
}

// === Grouped Expression Still Works ===

func TestGroupedExpressionNotTuple(t *testing.T) {
        src := `module test
fn test_group() -> Int:
    return (1 + 2) * 3
`
        result := runFunc(t, src, "test_group", nil)
        expectInt(t, result, 9)
}

// === Tuple with Expressions ===

func TestTupleWithExpressions(t *testing.T) {
        src := `module test
fn test_expr() -> Int:
    let t = (1 + 2, 3 * 4, 10 - 5)
    return t[0] + t[1] + t[2]
`
        result := runFunc(t, src, "test_expr", nil)
        expectInt(t, result, 20) // 3 + 12 + 5
}
