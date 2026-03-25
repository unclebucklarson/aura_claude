package interpreter

import "testing"

// --- Method Calls ---

func TestInterfaceMethodCallOnStruct(t *testing.T) {
	val := runFunc(t, `module test

struct Point:
    x: Int
    y: Int

impl Point:
    fn get_x(self: Point) -> Int:
        return self.x

fn get_x_of_point() -> Int:
    let p = Point(x: 42, y: 7)
    return p.get_x()
`, "get_x_of_point", nil)
	expectInt(t, val, 42)
}

func TestInterfaceMethodAccessesFields(t *testing.T) {
	val := runFunc(t, `module test

struct Point:
    x: Int
    y: Int

impl Point:
    fn sum(self: Point) -> Int:
        return self.x + self.y

fn test() -> Int:
    let p = Point(x: 10, y: 20)
    return p.sum()
`, "test", nil)
	expectInt(t, val, 30)
}

func TestInterfaceMethodWithArgs(t *testing.T) {
	val := runFunc(t, `module test

struct Counter:
    value: Int

impl Counter:
    fn add(self: Counter, n: Int) -> Int:
        return self.value + n

fn test() -> Int:
    let c = Counter(value: 5)
    return c.add(3)
`, "test", nil)
	expectInt(t, val, 8)
}

func TestInterfaceMethodReturnsString(t *testing.T) {
	val := runFunc(t, `module test

struct Person:
    name: String

impl Person:
    fn greet(self: Person) -> String:
        return self.name

fn test() -> String:
    let p = Person(name: "Alice")
    return p.greet()
`, "test", nil)
	expectString(t, val, "Alice")
}

// --- Multiple Types Same Method Name ---

func TestInterfaceMultipleTypesDispatchCorrectly(t *testing.T) {
	val := runFunc(t, `module test

trait Describable:
    fn describe() -> Int

struct Dog:
    legs: Int

struct Spider:
    legs: Int

impl Describable for Dog:
    fn describe(self: Dog) -> Int:
        return self.legs

impl Describable for Spider:
    fn describe(self: Spider) -> Int:
        return self.legs

fn test() -> Int:
    let d = Dog(legs: 4)
    let s = Spider(legs: 8)
    return d.describe() + s.describe()
`, "test", nil)
	expectInt(t, val, 12)
}

func TestInterfaceDispatchDoesNotCrossTypes(t *testing.T) {
	val := runFunc(t, `module test

struct Foo:
    val: Int

struct Bar:
    val: Int

impl Foo:
    fn get(self: Foo) -> Int:
        return self.val + 1

impl Bar:
    fn get(self: Bar) -> Int:
        return self.val + 100

fn test() -> Int:
    let f = Foo(val: 10)
    let b = Bar(val: 10)
    return b.get()
`, "test", nil)
	expectInt(t, val, 110)
}

// --- Inherent Impl ---

func TestInterfaceInherentImplCall(t *testing.T) {
	val := runFunc(t, `module test

struct Rectangle:
    width: Int
    height: Int

impl Rectangle:
    fn area(self: Rectangle) -> Int:
        return self.width * self.height

fn test() -> Int:
    let r = Rectangle(width: 6, height: 7)
    return r.area()
`, "test", nil)
	expectInt(t, val, 42)
}

func TestInterfaceInherentImplMultipleMethods(t *testing.T) {
	val := runFunc(t, `module test

struct Box:
    w: Int
    h: Int
    d: Int

impl Box:
    fn volume(self: Box) -> Int:
        return self.w * self.h * self.d
    fn surface(self: Box) -> Int:
        return 2 * (self.w * self.h + self.h * self.d + self.w * self.d)

fn test() -> Int:
    let b = Box(w: 2, h: 3, d: 4)
    return b.volume()
`, "test", nil)
	expectInt(t, val, 24)
}

// --- Trait Interface Dispatch ---

func TestInterfaceTraitMethodDispatch(t *testing.T) {
	val := runFunc(t, `module test

trait Printable:
    fn display() -> String

struct Label:
    text: String

impl Printable for Label:
    fn display(self: Label) -> String:
        return self.text

fn test() -> String:
    let l = Label(text: "hello")
    return l.display()
`, "test", nil)
	expectString(t, val, "hello")
}

func TestInterfacePassStructToInterfaceParam(t *testing.T) {
	val := runFunc(t, `module test

trait Measurable:
    fn size() -> Int

struct Vec:
    len: Int

impl Measurable for Vec:
    fn size(self: Vec) -> Int:
        return self.len

fn get_size(m: Measurable) -> Int:
    return m.size()

fn test() -> Int:
    let v = Vec(len: 7)
    return get_size(v)
`, "test", nil)
	expectInt(t, val, 7)
}

// --- Method Chaining ---

func TestInterfaceMethodChaining(t *testing.T) {
	val := runFunc(t, `module test

struct Num:
    val: Int

impl Num:
    fn doubled(self: Num) -> Int:
        return self.val * 2

fn test() -> Int:
    let n = Num(val: 5)
    return n.doubled()
`, "test", nil)
	expectInt(t, val, 10)
}

// --- Method with Multiple Params ---

func TestInterfaceMethodMultipleParams(t *testing.T) {
	val := runFunc(t, `module test

struct Calc:
    base: Int

impl Calc:
    fn compute(self: Calc, a: Int, b: Int) -> Int:
        return self.base + a * b

fn test() -> Int:
    let c = Calc(base: 10)
    return c.compute(3, 4)
`, "test", nil)
	expectInt(t, val, 22)
}
