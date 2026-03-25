package interpreter

import "testing"

// These tests verify that type-inferred code runs correctly at runtime.

// --- Empty Collection Runtime ---

func TestInferenceEmptyAnnotatedListLenZero(t *testing.T) {
	val := runFunc(t, `module test

fn test() -> Int:
    let xs: [Int] = []
    return len(xs)
`, "test", nil)
	expectInt(t, val, 0)
}

func TestInferenceEmptyAnnotatedMapLenZero(t *testing.T) {
	val := runFunc(t, `module test

fn test() -> Int:
    let m: {String: Int} = {}
    return len(m)
`, "test", nil)
	expectInt(t, val, 0)
}

func TestInferenceEmptyListThenAppend(t *testing.T) {
	val := runFunc(t, `module test

fn test() -> Int:
    let mut xs: [Int] = []
    xs = xs + [1, 2, 3]
    return len(xs)
`, "test", nil)
	expectInt(t, val, 3)
}

// --- Constructor Inference Runtime ---

func TestInferenceSomeAnnotatedMatch(t *testing.T) {
	val := runFunc(t, `module test

fn test() -> Int:
    let x: Option[Int] = Some(42)
    match x:
        case Some(v):
            return v
        case None:
            return 0
`, "test", nil)
	expectInt(t, val, 42)
}

func TestInferenceOkAnnotatedMatch(t *testing.T) {
	val := runFunc(t, `module test

fn test() -> Int:
    let x: Result[Int, String] = Ok(7)
    match x:
        case Ok(v):
            return v
        case Err(e):
            return 0
`, "test", nil)
	expectInt(t, val, 7)
}

func TestInferenceErrAnnotatedMatch(t *testing.T) {
	val := runFunc(t, `module test

fn test() -> String:
    let x: Result[Int, String] = Err("fail")
    match x:
        case Ok(v):
            return "ok"
        case Err(e):
            return e
`, "test", nil)
	expectString(t, val, "fail")
}

func TestInferenceNoneAnnotatedMatch(t *testing.T) {
	val := runFunc(t, `module test

fn test() -> Int:
    let x: Option[Int] = None
    match x:
        case Some(v):
            return v
        case None:
            return -1
`, "test", nil)
	expectInt(t, val, -1)
}

// --- Generic Alias Runtime ---

func TestInferenceGenericAliasMaybeMatch(t *testing.T) {
	val := runFunc(t, `module test

type Maybe[T] = Option[T]

fn test() -> Int:
    let x: Maybe[Int] = Some(99)
    match x:
        case Some(v):
            return v
        case None:
            return 0
`, "test", nil)
	expectInt(t, val, 99)
}

func TestInferenceGenericAliasListOps(t *testing.T) {
	val := runFunc(t, `module test

type Nums = [Int]

fn test() -> Int:
    let xs: Nums = [10, 20, 30]
    return len(xs)
`, "test", nil)
	expectInt(t, val, 3)
}

func TestInferenceGenericAliasReturnValue(t *testing.T) {
	val := runFunc(t, `module test

type MaybeInt = Option[Int]

fn get_val(n: Int) -> MaybeInt:
    if n > 0:
        return Some(n)
    return None

fn test() -> Int:
    let result = get_val(5)
    match result:
        case Some(v):
            return v
        case None:
            return 0
`, "test", nil)
	expectInt(t, val, 5)
}

func TestInferenceGenericAliasListWrapper(t *testing.T) {
	val := runFunc(t, `module test

type Wrapper[T] = [T]

fn test() -> Int:
    let xs: Wrapper[Int] = [1, 2, 3, 4]
    return len(xs)
`, "test", nil)
	expectInt(t, val, 4)
}
