package interpreter

import "testing"

// --- Generic Functions ---

func TestGenericIdentityInt(t *testing.T) {
	result := runFunc(t, `module test

fn identity[T](x: T) -> T:
    return x

fn main() -> Int:
    return identity(42)
`, "main", nil)
	expectInt(t, result, 42)
}

func TestGenericIdentityString(t *testing.T) {
	result := runFunc(t, `module test

fn identity[T](x: T) -> T:
    return x

fn main() -> String:
    return identity("hello")
`, "main", nil)
	expectString(t, result, "hello")
}

func TestGenericIdentityBool(t *testing.T) {
	result := runFunc(t, `module test

fn identity[T](x: T) -> T:
    return x

fn main() -> Bool:
    return identity(true)
`, "main", nil)
	expectBool(t, result, true)
}

func TestGenericSwap(t *testing.T) {
	result := runFunc(t, `module test

fn swap[A, B](a: A, b: B) -> (B, A):
    return (b, a)
`, "swap", []Value{&StringVal{Val: "hi"}, &IntVal{Val: 99}})
	tv, ok := result.(*TupleVal)
	if !ok {
		t.Fatalf("expected TupleVal, got %T", result)
	}
	if len(tv.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(tv.Elements))
	}
	expectInt(t, tv.Elements[0], 99)
	expectString(t, tv.Elements[1], "hi")
}

// --- Generic Structs ---

func TestGenericStructConstruction(t *testing.T) {
	result := runFunc(t, `module test

struct Pair[A, B]:
    first: A
    second: B

fn main() -> Int:
    let p = Pair(first: 10, second: "hello")
    return p.first
`, "main", nil)
	expectInt(t, result, 10)
}

func TestGenericStructSecondField(t *testing.T) {
	result := runFunc(t, `module test

struct Pair[A, B]:
    first: A
    second: B

fn main() -> String:
    let p = Pair(first: 10, second: "world")
    return p.second
`, "main", nil)
	expectString(t, result, "world")
}

func TestGenericStructBoolFloat(t *testing.T) {
	result := runFunc(t, `module test

struct Pair[A, B]:
    first: A
    second: B

fn main() -> Bool:
    let p = Pair(first: true, second: 3.14)
    return p.first
`, "main", nil)
	expectBool(t, result, true)
}

func TestGenericStructInFunction(t *testing.T) {
	result := runFunc(t, `module test

struct Box[T]:
    value: T

fn unbox[T](b: Box[T]) -> T:
    return b.value

fn main() -> Int:
    let b = Box(value: 55)
    return unbox(b)
`, "main", nil)
	expectInt(t, result, 55)
}

// --- Generic Enums ---

func TestGenericEnumSomeVariant(t *testing.T) {
	result := runFunc(t, `module test

enum Maybe[T]:
    Just(T)
    Nothing

fn main() -> Int:
    let m = Maybe.Just(42)
    match m:
        case Maybe.Just(v):
            return v
        case Maybe.Nothing():
            return 0
`, "main", nil)
	expectInt(t, result, 42)
}

func TestGenericEnumNothingVariant(t *testing.T) {
	result := runFunc(t, `module test

enum Maybe[T]:
    Just(T)
    Nothing

fn main() -> Int:
    let m = Maybe.Nothing()
    match m:
        case Maybe.Just(v):
            return v
        case Maybe.Nothing():
            return -1
`, "main", nil)
	expectInt(t, result, -1)
}

func TestGenericEnumStringPayload(t *testing.T) {
	result := runFunc(t, `module test

enum Wrapper[T]:
    Wrap(T)

fn main() -> String:
    let w = Wrapper.Wrap("test")
    match w:
        case Wrapper.Wrap(v):
            return v
`, "main", nil)
	expectString(t, result, "test")
}

// --- Generic with Collections ---

func TestGenericFirstElement(t *testing.T) {
	result := runFunc(t, `module test

fn first[T](items: [T]) -> T:
    return items[0]

fn main() -> Int:
    return first([10, 20, 30])
`, "main", nil)
	expectInt(t, result, 10)
}

func TestGenericWrapInList(t *testing.T) {
	result := runFunc(t, `module test

fn wrap_list[T](x: T) -> [T]:
    return [x]

fn main() -> Int:
    let xs = wrap_list(7)
    return xs[0]
`, "main", nil)
	expectInt(t, result, 7)
}

// --- Polymorphic Use ---

func TestGenericCalledWithMultipleTypes(t *testing.T) {
	src := `module test

fn identity[T](x: T) -> T:
    return x

fn main():
    let n = identity(42)
    let s = identity("hello")
    let b = identity(true)
`
	interp := runModule(t, src)
	_ = interp
}

func TestGenericStructMultipleInstantiations(t *testing.T) {
	src := `module test

struct Pair[A, B]:
    first: A
    second: B

fn main():
    let p1 = Pair(first: 1, second: "one")
    let p2 = Pair(first: true, second: 3.14)
`
	interp := runModule(t, src)
	_ = interp
}
