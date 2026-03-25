package checker

import "testing"

// --- Trait Declaration ---

func TestInterfaceSimpleTraitAccepted(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String
`)
	expectNoErrors(t, errs)
}

func TestInterfaceMultiMethodTraitAccepted(t *testing.T) {
	errs := checkSource(t, `module test

trait Shape:
    fn area() -> Float
    fn perimeter() -> Float
    fn name() -> String
`)
	expectNoErrors(t, errs)
}

// --- Impl Block Validation ---

func TestInterfaceImplSatisfiesTrait(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

struct Point:
    x: Int
    y: Int

impl Printable for Point:
    fn display(self: Point) -> String:
        return "Point"
`)
	expectNoErrors(t, errs)
}

func TestInterfaceImplMissingMethod(t *testing.T) {
	errs := checkSource(t, `module test

trait Shape:
    fn area() -> Float
    fn perimeter() -> Float

struct Circle:
    radius: Float

impl Shape for Circle:
    fn area(self: Circle) -> Float:
        return 3.14
`)
	expectErrorCode(t, errs, ErrMissingMethod)
}

func TestInterfaceImplWrongReturnType(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

struct Point:
    x: Int
    y: Int

impl Printable for Point:
    fn display(self: Point) -> Int:
        return 42
`)
	expectErrorCode(t, errs, ErrTypeMismatch)
}

func TestInterfaceInherentImplAccepted(t *testing.T) {
	errs := checkSource(t, `module test

struct Counter:
    count: Int

impl Counter:
    fn increment(self: Counter) -> Int:
        return self.count
`)
	expectNoErrors(t, errs)
}

// --- Interface as Type Annotation ---

func TestInterfaceLetAnnotationSatisfied(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

struct Point:
    x: Int
    y: Int

impl Printable for Point:
    fn display(self: Point) -> String:
        return "Point"

fn main():
    let p = Point(x: 1, y: 2)
    let q: Printable = p
`)
	expectNoErrors(t, errs)
}

func TestInterfaceLetAnnotationNotSatisfied(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

struct Rect:
    width: Int
    height: Int

fn main():
    let r = Rect(width: 3, height: 4)
    let q: Printable = r
`)
	expectErrorCode(t, errs, ErrTypeMismatch)
}

func TestInterfaceFnParamSatisfied(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

struct Point:
    x: Int
    y: Int

impl Printable for Point:
    fn display(self: Point) -> String:
        return "Point"

fn show(x: Printable):
    return None

fn main():
    let p = Point(x: 1, y: 2)
    show(p)
`)
	expectNoErrors(t, errs)
}

func TestInterfaceFnParamNotSatisfied(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

struct Rect:
    width: Int
    height: Int

fn show(x: Printable):
    return None

fn main():
    let r = Rect(width: 3, height: 4)
    show(r)
`)
	expectErrorCode(t, errs, ErrTypeMismatch)
}

// --- Structural Satisfaction ---

func TestInterfaceSingleMethodSatisfied(t *testing.T) {
	errs := checkSource(t, `module test

trait Greeter:
    fn greet() -> String

struct English:
    name: String

impl Greeter for English:
    fn greet(self: English) -> String:
        return "Hello"
`)
	expectNoErrors(t, errs)
}

func TestInterfaceAllMethodsSatisfied(t *testing.T) {
	errs := checkSource(t, `module test

trait Shape:
    fn area() -> Float
    fn perimeter() -> Float

struct Square:
    side: Float

impl Shape for Square:
    fn area(self: Square) -> Float:
        return self.side
    fn perimeter(self: Square) -> Float:
        return self.side
`)
	expectNoErrors(t, errs)
}

func TestInterfacePartialSatisfactionFails(t *testing.T) {
	errs := checkSource(t, `module test

trait Shape:
    fn area() -> Float
    fn perimeter() -> Float

struct Triangle:
    base: Float

impl Shape for Triangle:
    fn area(self: Triangle) -> Float:
        return self.base
`)
	expectErrorCode(t, errs, ErrMissingMethod)
}

func TestInterfaceReturnTypeChecked(t *testing.T) {
	errs := checkSource(t, `module test

trait Measurable:
    fn measure() -> Int

struct Box:
    size: Float

impl Measurable for Box:
    fn measure(self: Box) -> String:
        return "big"
`)
	expectErrorCode(t, errs, ErrTypeMismatch)
}

// --- Error Cases ---

func TestInterfaceImplForUndefinedTrait(t *testing.T) {
	errs := checkSource(t, `module test

struct Point:
    x: Int

impl NonExistent for Point:
    fn display(self: Point) -> String:
        return "Point"
`)
	expectErrorCode(t, errs, ErrUndefinedType)
}

func TestInterfaceTraitNameResolvesAsType(t *testing.T) {
	errs := checkSource(t, `module test

trait Printable:
    fn display() -> String

struct Point:
    x: Int

impl Printable for Point:
    fn display(self: Point) -> String:
        return "p"

fn test() -> Printable:
    return Point(x: 1)
`)
	expectNoErrors(t, errs)
}
