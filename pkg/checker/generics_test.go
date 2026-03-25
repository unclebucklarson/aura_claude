package checker

import "testing"

// --- Generic Declarations ---

func TestGenericFunctionDeclaration(t *testing.T) {
	errs := checkSource(t, `module test

fn identity[T](x: T) -> T:
    return x
`)
	expectNoErrors(t, errs)
}

func TestGenericStructDeclaration(t *testing.T) {
	errs := checkSource(t, `module test

struct Pair[A, B]:
    first: A
    second: B
`)
	expectNoErrors(t, errs)
}

func TestGenericEnumDeclaration(t *testing.T) {
	errs := checkSource(t, `module test

enum Tree[T]:
    Leaf
    Node(T)
`)
	expectNoErrors(t, errs)
}

func TestGenericTypeAlias(t *testing.T) {
	errs := checkSource(t, `module test

type Wrapper[T] = T
`)
	expectNoErrors(t, errs)
}

func TestGenericMultiParam(t *testing.T) {
	errs := checkSource(t, `module test

fn swap[A, B](a: A, b: B) -> (B, A):
    return (b, a)
`)
	expectNoErrors(t, errs)
}

// --- Body Type Checking ---

func TestGenericFunctionBodyReturnTypeParam(t *testing.T) {
	errs := checkSource(t, `module test

fn identity[T](x: T) -> T:
    return x
`)
	expectNoErrors(t, errs)
}

func TestGenericSwapBody(t *testing.T) {
	errs := checkSource(t, `module test

fn swap[A, B](a: A, b: B) -> (B, A):
    return (b, a)
`)
	expectNoErrors(t, errs)
}

func TestGenericFunctionMultipleTypeParams(t *testing.T) {
	errs := checkSource(t, `module test

fn map_fn[A, B](items: [A], f: fn(A) -> B) -> [B]:
    return []
`)
	expectNoErrors(t, errs)
}

// --- Call-Site Inference ---

func TestGenericCallIntArg(t *testing.T) {
	errs := checkSource(t, `module test

fn identity[T](x: T) -> T:
    return x

fn main():
    let n = identity(42)
`)
	expectNoErrors(t, errs)
}

func TestGenericCallStringArg(t *testing.T) {
	errs := checkSource(t, `module test

fn identity[T](x: T) -> T:
    return x

fn main():
    let s = identity("hello")
`)
	expectNoErrors(t, errs)
}

func TestGenericCallReturnAssignableToInferred(t *testing.T) {
	errs := checkSource(t, `module test

fn identity[T](x: T) -> T:
    return x

fn main():
    let n: Int = identity(42)
`)
	expectNoErrors(t, errs)
}

func TestGenericCallReturnMismatch(t *testing.T) {
	errs := checkSource(t, `module test

fn identity[T](x: T) -> T:
    return x

fn main():
    let s: String = identity(42)
`)
	expectErrorCode(t, errs, ErrTypeMismatch)
}

func TestGenericCallSwapNoErrors(t *testing.T) {
	errs := checkSource(t, `module test

fn swap[A, B](a: A, b: B) -> (B, A):
    return (b, a)

fn main():
    let result = swap(1, "hi")
`)
	expectNoErrors(t, errs)
}

func TestGenericCallBoolArg(t *testing.T) {
	errs := checkSource(t, `module test

fn identity[T](x: T) -> T:
    return x

fn main():
    let b = identity(true)
`)
	expectNoErrors(t, errs)
}

// --- Generic Struct Instantiation ---

func TestGenericStructInstantiation(t *testing.T) {
	errs := checkSource(t, `module test

struct Pair[A, B]:
    first: A
    second: B

fn main():
    let p: Pair[Int, String] = Pair(first: 1, second: "hi")
`)
	expectNoErrors(t, errs)
}

func TestGenericStructUsedInFunction(t *testing.T) {
	errs := checkSource(t, `module test

struct Pair[A, B]:
    first: A
    second: B

fn make_pair[A, B](a: A, b: B) -> Pair[A, B]:
    return Pair(first: a, second: b)
`)
	expectNoErrors(t, errs)
}

func TestGenericStructWrongTypeArgCount(t *testing.T) {
	errs := checkSource(t, `module test

struct Pair[A, B]:
    first: A
    second: B

fn main():
    let p: Pair[Int] = Pair(first: 1, second: "hi")
`)
	expectErrorCode(t, errs, ErrTypeParamCount)
}

func TestOptionAndResultRegression(t *testing.T) {
	errs := checkSource(t, `module test

fn safe_div(a: Int, b: Int) -> Option[Int]:
    if b == 0:
        return None
    return Some(a / b)
`)
	expectNoErrors(t, errs)
}

func TestNestedGenericType(t *testing.T) {
	errs := checkSource(t, `module test

struct Pair[A, B]:
    first: A
    second: B

fn main():
    let pairs: [Pair[Int, String]] = []
`)
	expectNoErrors(t, errs)
}

// --- Generic Enum Instantiation ---

func TestGenericEnumInstantiation(t *testing.T) {
	errs := checkSource(t, `module test

enum Option2[T]:
    Some(T)
    None

fn main():
    let x: Option2[Int] = Option2.Some(42)
`)
	expectNoErrors(t, errs)
}

func TestGenericEnumInFunctionReturn(t *testing.T) {
	errs := checkSource(t, `module test

enum Maybe[T]:
    Just(T)
    Nothing

fn wrap[T](val: T) -> Maybe[T]:
    return Maybe.Just(val)
`)
	expectNoErrors(t, errs)
}

func TestGenericEnumNested(t *testing.T) {
	errs := checkSource(t, `module test

enum Tree[T]:
    Leaf
    Node(T)

fn make_leaf[T]() -> Tree[T]:
    return Tree.Leaf()
`)
	expectNoErrors(t, errs)
}

// --- Error Cases ---

func TestGenericFunctionWrongArgCount(t *testing.T) {
	errs := checkSource(t, `module test

fn identity[T](x: T) -> T:
    return x

fn main():
    let n = identity(1, 2)
`)
	expectErrorCode(t, errs, ErrArgCount)
}

func TestUndefinedTypeStillCaught(t *testing.T) {
	errs := checkSource(t, `module test

fn foo(x: DoesNotExist) -> Int:
    return 0
`)
	expectErrorCode(t, errs, ErrUndefinedType)
}

func TestGenericFunctionDeclarationMultiUse(t *testing.T) {
	errs := checkSource(t, `module test

fn identity[T](x: T) -> T:
    return x

fn main():
    let n = identity(42)
    let s = identity("hello")
    let b = identity(true)
`)
	expectNoErrors(t, errs)
}
