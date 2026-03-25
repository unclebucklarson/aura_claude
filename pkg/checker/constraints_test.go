package checker

import "testing"

// --- Constraint Declaration ---

func TestConstraintSingleDeclarationAccepted(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

fn show[T](x: T) -> String where T: Printable:
    return "ok"
`)
	expectNoErrors(t, errs)
}

func TestConstraintMultipleDeclarationsAccepted(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

trait Measurable:
    fn size() -> Int

fn describe[T](x: T) -> String where T: Printable, T: Measurable:
    return "ok"
`)
	expectNoErrors(t, errs)
}

// --- Trait Existence Validation ---

func TestConstraintUndefinedTrait(t *testing.T) {
	errs := checkSource(t, `module test

fn show[T](x: T) -> String where T: NonExistent:
    return "ok"
`)
	expectErrorCode(t, errs, ErrUndefinedType)
}

func TestConstraintNonTraitType(t *testing.T) {
	errs := checkSource(t, `module test

struct Point:
    x: Int

fn show[T](x: T) -> String where T: Point:
    return "ok"
`)
	expectErrorCode(t, errs, ErrTypeMismatch)
}

// --- Call-site Checking ---

func TestConstraintCallSiteSatisfied(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

struct Label:
    text: String

impl Printable for Label:
    fn display(self: Label) -> String:
        return self.text

fn show[T](x: T) -> String where T: Printable:
    return "ok"

fn main():
    let l = Label(text: "hi")
    show(l)
`)
	expectNoErrors(t, errs)
}

func TestConstraintCallSiteNotSatisfied(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

struct Rect:
    w: Int

fn show[T](x: T) -> String where T: Printable:
    return "ok"

fn main():
    let r = Rect(w: 5)
    show(r)
`)
	expectErrorCode(t, errs, ErrConstraintNotSatisfied)
}

func TestConstraintMultipleAllSatisfied(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

trait Measurable:
    fn size() -> Int

struct Box:
    w: Int

impl Printable for Box:
    fn display(self: Box) -> String:
        return "box"

impl Measurable for Box:
    fn size(self: Box) -> Int:
        return self.w

fn describe[T](x: T) -> String where T: Printable, T: Measurable:
    return "ok"

fn main():
    let b = Box(w: 3)
    describe(b)
`)
	expectNoErrors(t, errs)
}

func TestConstraintMultipleOneNotSatisfied(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

trait Measurable:
    fn size() -> Int

struct Box:
    w: Int

impl Printable for Box:
    fn display(self: Box) -> String:
        return "box"

fn describe[T](x: T) -> String where T: Printable, T: Measurable:
    return "ok"

fn main():
    let b = Box(w: 3)
    describe(b)
`)
	expectErrorCode(t, errs, ErrConstraintNotSatisfied)
}

// --- Constraint with Effects ---

func TestConstraintWithEffectsAccepted(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

fn show[T](x: T) -> String with log where T: Printable:
    return "ok"
`)
	expectNoErrors(t, errs)
}

// --- Multi Type Param ---

func TestConstraintTwoTypeParamsEachConstrained(t *testing.T) {
	errs := checkSource(t, `module test

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

fn both[X, Y](x: X, y: Y) where X: Printable, Y: Printable:
    return None

fn main():
    let a = A(v: 1)
    let b = B(v: 2)
    both(a, b)
`)
	expectNoErrors(t, errs)
}

func TestConstraintTwoTypeParamsOneNotSatisfied(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

struct A:
    v: Int

struct B:
    v: Int

impl Printable for A:
    fn display(self: A) -> String:
        return "a"

fn both[X, Y](x: X, y: Y) where X: Printable, Y: Printable:
    return None

fn main():
    let a = A(v: 1)
    let b = B(v: 2)
    both(a, b)
`)
	expectErrorCode(t, errs, ErrConstraintNotSatisfied)
}

// --- No Constraint on Generic Fn ---

func TestConstraintUnconstrainedGenericUnaffected(t *testing.T) {
	errs := checkSource(t, `module test

fn identity[T](x: T) -> T:
    return x

fn main():
    let n = identity(42)
`)
	expectNoErrors(t, errs)
}

// --- Constraint in Trait Signature ---

func TestConstraintInTraitSignatureAccepted(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

trait Container:
    fn show[T]() -> String where T: Printable
`)
	expectNoErrors(t, errs)
}

// --- Refinement Types Still Work After Guard ---

func TestConstraintRefinementTypeUnaffected(t *testing.T) {
	// Verify the WHERE guard in parseOptionOrRefinementType didn't break refinement type parsing
	errs := checkSource(t, `module test

type Age = Int where self >= 0
type Name = String where self.length() > 0
`)
	expectNoErrors(t, errs)
}
