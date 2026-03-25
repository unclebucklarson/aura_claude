package interpreter

import "testing"

// Constraints are type-erased at runtime. These tests verify that
// constrained generic functions execute correctly.

func TestConstraintGenericFnRunsWithConstraint(t *testing.T) {
	val := runFunc(t, `module test

trait Printable:
    fn display() -> String

struct Label:
    text: String

impl Printable for Label:
    fn display(self: Label) -> String:
        return self.text

fn show[T](x: T) -> Int where T: Printable:
    return 1

fn test() -> Int:
    let l = Label(text: "hi")
    return show(l)
`, "test", nil)
	expectInt(t, val, 1)
}

func TestConstraintGenericFnReturnsValue(t *testing.T) {
	val := runFunc(t, `module test

trait Sizeable:
    fn size() -> Int

struct Vec:
    len: Int

impl Sizeable for Vec:
    fn size(self: Vec) -> Int:
        return self.len

fn get_size[T](x: T) -> Int where T: Sizeable:
    return 42

fn test() -> Int:
    let v = Vec(len: 10)
    return get_size(v)
`, "test", nil)
	expectInt(t, val, 42)
}

func TestConstraintMultipleConstraintsFnRuns(t *testing.T) {
	val := runFunc(t, `module test

trait Printable:
    fn display() -> String

trait Countable:
    fn count() -> Int

struct Items:
    n: Int

impl Printable for Items:
    fn display(self: Items) -> String:
        return "items"

impl Countable for Items:
    fn count(self: Items) -> Int:
        return self.n

fn process[T](x: T) -> Int where T: Printable, T: Countable:
    return 99

fn test() -> Int:
    let items = Items(n: 5)
    return process(items)
`, "test", nil)
	expectInt(t, val, 99)
}

func TestConstraintTwoTypeParamsBothRun(t *testing.T) {
	val := runFunc(t, `module test

trait Printable:
    fn display() -> String

struct A:
    v: Int

struct B:
    v: Int

impl Printable for A:
    fn display(self: A) -> String:
        return "a"

impl Printable for B:
    fn display(self: B) -> String:
        return "b"

fn both[X, Y](x: X, y: Y) -> Int where X: Printable, Y: Printable:
    return 7

fn test() -> Int:
    let a = A(v: 1)
    let b = B(v: 2)
    return both(a, b)
`, "test", nil)
	expectInt(t, val, 7)
}

func TestConstraintTwoTypeParamsReturnFirst(t *testing.T) {
	val := runFunc(t, `module test

trait Printable:
    fn display() -> String

struct Num:
    val: Int

impl Printable for Num:
    fn display(self: Num) -> String:
        return "num"

fn first[X, Y](x: X, y: Y) -> Int where X: Printable, Y: Printable:
    return 5

fn test() -> Int:
    let a = Num(val: 1)
    let b = Num(val: 2)
    return first(a, b)
`, "test", nil)
	expectInt(t, val, 5)
}

func TestConstraintImplMethodCallableInBody(t *testing.T) {
	val := runFunc(t, `module test

trait Printable:
    fn display() -> String

struct Widget:
    name: String

impl Printable for Widget:
    fn display(self: Widget) -> String:
        return self.name

fn show_it[T](x: T) -> String where T: Printable:
    return x.display()

fn test() -> String:
    let w = Widget(name: "button")
    return show_it(w)
`, "test", nil)
	expectString(t, val, "button")
}

func TestConstraintInherentImplSatisfiesConstraint(t *testing.T) {
	val := runFunc(t, `module test

trait Greetable:
    fn greet() -> String

struct Person:
    name: String

impl Greetable for Person:
    fn greet(self: Person) -> String:
        return self.name

fn welcome[T](x: T) -> String where T: Greetable:
    return x.greet()

fn test() -> String:
    let p = Person(name: "Alice")
    return welcome(p)
`, "test", nil)
	expectString(t, val, "Alice")
}

func TestConstraintRefinementTypeStillEnforced(t *testing.T) {
	// Verify refinement runtime enforcement wasn't broken by the WHERE guard
	interp := runModule(t, `module test

type Score = Int where self >= 0

fn main():
    let x: Score = 10
`)
	_ = interp // just verify it runs without panic
}

func TestConstraintRefinementTypeStillEnforcedViolation(t *testing.T) {
	// Verify that a refinement violation still panics at runtime
	module := parseModule(t, `module test

type Score = Int where self >= 0

fn run_violation() -> Int:
    let x: Score = -1
    return x
`)
	interp := New(module)
	_, err := interp.Run()
	if err != nil {
		t.Fatalf("module init error: %v", err)
	}
	_, err = interp.RunFunction("run_violation", nil)
	if err == nil {
		t.Fatal("expected refinement violation error, got nil")
	}
}

func TestConstraintGenericIdentityWithConstraint(t *testing.T) {
	val := runFunc(t, `module test

trait Printable:
    fn display() -> String

struct Tag:
    id: Int

impl Printable for Tag:
    fn display(self: Tag) -> String:
        return "tag"

fn identity[T](x: T) -> T where T: Printable:
    return x

fn test() -> Int:
    let t = Tag(id: 99)
    let result = identity(t)
    return result.id
`, "test", nil)
	expectInt(t, val, 99)
}
