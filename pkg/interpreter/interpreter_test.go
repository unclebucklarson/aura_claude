package interpreter

import (
        "strings"
        "testing"

        "github.com/unclebucklarson/aura/pkg/lexer"
        "github.com/unclebucklarson/aura/pkg/parser"
        "github.com/unclebucklarson/aura/pkg/ast"
)

// --- Test Helpers ---

func parseModule(t *testing.T, src string) *ast.Module {
        t.Helper()
        l := lexer.New(src, "test.aura")
        tokens, lexErrors := l.Tokenize()
        if len(lexErrors) > 0 {
                t.Fatalf("lex errors: %v", lexErrors)
        }
        p := parser.New(tokens, "test.aura")
        module, parseErrors := p.Parse()
        if len(parseErrors) > 0 {
                t.Fatalf("parse errors: %v", parseErrors)
        }
        return module
}

func runModule(t *testing.T, src string) *Interpreter {
        t.Helper()
        module := parseModule(t, src)
        interp := New(module)
        _, err := interp.Run()
        if err != nil {
                t.Fatalf("run error: %v", err)
        }
        return interp
}

func runFunc(t *testing.T, src string, fnName string, args []Value) Value {
        t.Helper()
        interp := runModule(t, src)
        result, err := interp.RunFunction(fnName, args)
        if err != nil {
                t.Fatalf("RunFunction error: %v", err)
        }
        return result
}

func expectInt(t *testing.T, val Value, expected int64) {
        t.Helper()
        iv, ok := val.(*IntVal)
        if !ok {
                t.Fatalf("expected IntVal, got %T (%v)", val, val)
        }
        if iv.Val != expected {
                t.Fatalf("expected %d, got %d", expected, iv.Val)
        }
}

func expectFloat(t *testing.T, val Value, expected float64) {
        t.Helper()
        fv, ok := val.(*FloatVal)
        if !ok {
                t.Fatalf("expected FloatVal, got %T (%v)", val, val)
        }
        if fv.Val != expected {
                t.Fatalf("expected %g, got %g", expected, fv.Val)
        }
}

func expectString(t *testing.T, val Value, expected string) {
        t.Helper()
        sv, ok := val.(*StringVal)
        if !ok {
                t.Fatalf("expected StringVal, got %T (%v)", val, val)
        }
        if sv.Val != expected {
                t.Fatalf("expected %q, got %q", expected, sv.Val)
        }
}

func expectBool(t *testing.T, val Value, expected bool) {
        t.Helper()
        bv, ok := val.(*BoolVal)
        if !ok {
                t.Fatalf("expected BoolVal, got %T (%v)", val, val)
        }
        if bv.Val != expected {
                t.Fatalf("expected %v, got %v", expected, bv.Val)
        }
}

func expectNone(t *testing.T, val Value) {
        t.Helper()
        if _, ok := val.(*NoneVal); !ok {
                t.Fatalf("expected NoneVal, got %T (%v)", val, val)
        }
}

func expectRuntimeError(t *testing.T, src string, fnName string, substr string) {
        t.Helper()
        module := parseModule(t, src)
        interp := New(module)
        _, runErr := interp.Run()
        if runErr != nil {
                if strings.Contains(runErr.Error(), substr) {
                        return
                }
                t.Fatalf("unexpected run error: %v", runErr)
        }
        _, err := interp.RunFunction(fnName, nil)
        if err == nil {
                t.Fatalf("expected runtime error containing %q, got no error", substr)
        }
        if !strings.Contains(err.Error(), substr) {
                t.Fatalf("expected runtime error containing %q, got: %v", substr, err)
        }
}

// --- Value System Tests ---

func TestIntVal(t *testing.T) {
        v := &IntVal{Val: 42}
        if v.Type() != TypeInt {
                t.Fatalf("expected TypeInt")
        }
        if v.String() != "42" {
                t.Fatalf("expected '42', got %q", v.String())
        }
}

func TestFloatVal(t *testing.T) {
        v := &FloatVal{Val: 3.14}
        if v.Type() != TypeFloat {
                t.Fatalf("expected TypeFloat")
        }
        if v.String() != "3.14" {
                t.Fatalf("expected '3.14', got %q", v.String())
        }
}

func TestStringVal(t *testing.T) {
        v := &StringVal{Val: "hello"}
        if v.Type() != TypeString {
                t.Fatalf("expected TypeString")
        }
        if v.String() != "hello" {
                t.Fatalf("expected 'hello', got %q", v.String())
        }
}

func TestBoolVal(t *testing.T) {
        v := &BoolVal{Val: true}
        if v.String() != "true" {
                t.Fatalf("expected 'true'")
        }
        v2 := &BoolVal{Val: false}
        if v2.String() != "false" {
                t.Fatalf("expected 'false'")
        }
}

func TestNoneVal(t *testing.T) {
        v := &NoneVal{}
        if v.Type() != TypeNone {
                t.Fatalf("expected TypeNone")
        }
        if v.String() != "none" {
                t.Fatalf("expected 'none'")
        }
}

func TestIsTruthy(t *testing.T) {
        cases := []struct {
                val      Value
                expected bool
        }{
                {&BoolVal{Val: true}, true},
                {&BoolVal{Val: false}, false},
                {&IntVal{Val: 1}, true},
                {&IntVal{Val: 0}, false},
                {&StringVal{Val: "hi"}, true},
                {&StringVal{Val: ""}, false},
                {&NoneVal{}, false},
                {&ListVal{Elements: []Value{&IntVal{Val: 1}}}, true},
                {&ListVal{Elements: []Value{}}, false},
                {nil, false},
        }
        for i, c := range cases {
                got := IsTruthy(c.val)
                if got != c.expected {
                        t.Errorf("case %d: expected %v, got %v", i, c.expected, got)
                }
        }
}

func TestEqual(t *testing.T) {
        cases := []struct {
                a, b     Value
                expected bool
        }{
                {&IntVal{Val: 42}, &IntVal{Val: 42}, true},
                {&IntVal{Val: 42}, &IntVal{Val: 43}, false},
                {&IntVal{Val: 5}, &FloatVal{Val: 5.0}, true},
                {&StringVal{Val: "hi"}, &StringVal{Val: "hi"}, true},
                {&StringVal{Val: "hi"}, &StringVal{Val: "bye"}, false},
                {&BoolVal{Val: true}, &BoolVal{Val: true}, true},
                {&BoolVal{Val: true}, &BoolVal{Val: false}, false},
                {&NoneVal{}, &NoneVal{}, true},
                {nil, nil, true},
        }
        for i, c := range cases {
                got := Equal(c.a, c.b)
                if got != c.expected {
                        t.Errorf("case %d: expected %v, got %v", i, c.expected, got)
                }
        }
}

func TestListEqual(t *testing.T) {
        a := &ListVal{Elements: []Value{&IntVal{Val: 1}, &IntVal{Val: 2}}}
        b := &ListVal{Elements: []Value{&IntVal{Val: 1}, &IntVal{Val: 2}}}
        if !Equal(a, b) {
                t.Fatal("expected equal lists")
        }
        c := &ListVal{Elements: []Value{&IntVal{Val: 1}, &IntVal{Val: 3}}}
        if Equal(a, c) {
                t.Fatal("expected unequal lists")
        }
}

// --- Environment Tests ---

func TestEnvironmentDefineGet(t *testing.T) {
        env := NewEnvironment()
        env.Define("x", &IntVal{Val: 10})
        val, ok := env.Get("x")
        if !ok {
                t.Fatal("expected to find x")
        }
        expectInt(t, val, 10)
}

func TestEnvironmentScoping(t *testing.T) {
        parent := NewEnvironment()
        parent.Define("x", &IntVal{Val: 1})
        child := NewEnclosedEnvironment(parent)
        child.Define("y", &IntVal{Val: 2})

        // Child can see parent
        val, ok := child.Get("x")
        if !ok {
                t.Fatal("expected child to see parent's x")
        }
        expectInt(t, val, 1)

        // Parent cannot see child
        _, ok = parent.Get("y")
        if ok {
                t.Fatal("parent should not see child's y")
        }
}

func TestEnvironmentShadowing(t *testing.T) {
        parent := NewEnvironment()
        parent.Define("x", &IntVal{Val: 1})
        child := NewEnclosedEnvironment(parent)
        child.Define("x", &IntVal{Val: 2})

        val, _ := child.Get("x")
        expectInt(t, val, 2)
        val, _ = parent.Get("x")
        expectInt(t, val, 1)
}

func TestEnvironmentImmutable(t *testing.T) {
        env := NewEnvironment()
        env.DefineConst("x", &IntVal{Val: 10})
        err := env.Set("x", &IntVal{Val: 20})
        if err == nil {
                t.Fatal("expected error when setting immutable variable")
        }
}

func TestEnvironmentMutable(t *testing.T) {
        env := NewEnvironment()
        env.Define("x", &IntVal{Val: 10})
        err := env.Set("x", &IntVal{Val: 20})
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        val, _ := env.Get("x")
        expectInt(t, val, 20)
}

// --- Interpreter Expression Tests ---

func TestIntLiteral(t *testing.T) {
        src := `module test
fn get() -> Int:
  return 42
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 42)
}

func TestFloatLiteral(t *testing.T) {
        src := `module test
fn get() -> Float:
  return 3.14
`
        result := runFunc(t, src, "get", nil)
        expectFloat(t, result, 3.14)
}

func TestStringLiteral(t *testing.T) {
        src := `module test
fn get() -> String:
  return "hello"
`
        result := runFunc(t, src, "get", nil)
        expectString(t, result, "hello")
}

func TestBoolLiteral(t *testing.T) {
        src := `module test
fn get() -> Bool:
  return true
`
        result := runFunc(t, src, "get", nil)
        expectBool(t, result, true)
}

func TestNoneLiteral(t *testing.T) {
        src := `module test
fn get() -> Int:
  return none
`
        result := runFunc(t, src, "get", nil)
        expectNone(t, result)
}

func TestArithmeticAdd(t *testing.T) {
        src := `module test
fn add() -> Int:
  return 3 + 4
`
        result := runFunc(t, src, "add", nil)
        expectInt(t, result, 7)
}

func TestArithmeticSub(t *testing.T) {
        src := `module test
fn sub() -> Int:
  return 10 - 3
`
        result := runFunc(t, src, "sub", nil)
        expectInt(t, result, 7)
}

func TestArithmeticMul(t *testing.T) {
        src := `module test
fn mul() -> Int:
  return 6 * 7
`
        result := runFunc(t, src, "mul", nil)
        expectInt(t, result, 42)
}

func TestArithmeticDiv(t *testing.T) {
        src := `module test
fn div() -> Int:
  return 10 / 3
`
        result := runFunc(t, src, "div", nil)
        expectInt(t, result, 3)
}

func TestArithmeticMod(t *testing.T) {
        src := `module test
fn mod_test() -> Int:
  return 10 % 3
`
        result := runFunc(t, src, "mod_test", nil)
        expectInt(t, result, 1)
}

func TestArithmeticPower(t *testing.T) {
        src := `module test
fn pow_test() -> Int:
  return 2 ** 10
`
        result := runFunc(t, src, "pow_test", nil)
        expectInt(t, result, 1024)
}

func TestFloatArithmetic(t *testing.T) {
        src := `module test
fn calc() -> Float:
  return 1.5 + 2.5
`
        result := runFunc(t, src, "calc", nil)
        expectFloat(t, result, 4.0)
}

func TestMixedArithmetic(t *testing.T) {
        src := `module test
fn calc() -> Float:
  return 1 + 2.5
`
        result := runFunc(t, src, "calc", nil)
        expectFloat(t, result, 3.5)
}

func TestStringConcat(t *testing.T) {
        src := `module test
fn greet() -> String:
  return "hello" + " " + "world"
`
        result := runFunc(t, src, "greet", nil)
        expectString(t, result, "hello world")
}

func TestComparisonOps(t *testing.T) {
        src := `module test
fn lt() -> Bool:
  return 1 < 2
fn gt() -> Bool:
  return 2 > 1
fn lte() -> Bool:
  return 2 <= 2
fn gte() -> Bool:
  return 3 >= 2
fn eq() -> Bool:
  return 5 == 5
fn neq() -> Bool:
  return 5 != 6
`
        cases := []struct {
                fn       string
                expected bool
        }{
                {"lt", true},
                {"gt", true},
                {"lte", true},
                {"gte", true},
                {"eq", true},
                {"neq", true},
        }
        interp := runModule(t, src)
        for _, c := range cases {
                result, err := interp.RunFunction(c.fn, nil)
                if err != nil {
                        t.Fatalf("%s: %v", c.fn, err)
                }
                expectBool(t, result, c.expected)
        }
}

func TestLogicalOps(t *testing.T) {
        src := `module test
fn and_test() -> Bool:
  return true and false
fn or_test() -> Bool:
  return false or true
fn not_test() -> Bool:
  return not false
`
        interp := runModule(t, src)
        r, _ := interp.RunFunction("and_test", nil)
        expectBool(t, r, false)
        r, _ = interp.RunFunction("or_test", nil)
        expectBool(t, r, true)
        r, _ = interp.RunFunction("not_test", nil)
        expectBool(t, r, true)
}

func TestUnaryNegate(t *testing.T) {
        src := `module test
fn neg() -> Int:
  return -42
`
        result := runFunc(t, src, "neg", nil)
        expectInt(t, result, -42)
}

// --- Variable Binding Tests ---

func TestLetBinding(t *testing.T) {
        src := `module test
fn get() -> Int:
  let x = 10
  return x
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 10)
}

func TestLetMutable(t *testing.T) {
        src := `module test
fn get() -> Int:
  let mut x = 10
  x = 20
  return x
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 20)
}

func TestVariableScoping(t *testing.T) {
        src := `module test
fn get() -> Int:
  let x = 1
  let y = 2
  return x + y
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 3)
}

// --- Function Tests ---

func TestFunctionCall(t *testing.T) {
        src := `module test
fn double(n: Int) -> Int:
  return n * 2

fn get() -> Int:
  return double(21)
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 42)
}

func TestFunctionMultipleParams(t *testing.T) {
        src := `module test
fn add(a: Int, b: Int) -> Int:
  return a + b

fn get() -> Int:
  return add(3, 4)
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 7)
}

func TestRecursion(t *testing.T) {
        src := `module test
fn factorial(n: Int) -> Int:
  if n <= 1:
    return 1
  return n * factorial(n - 1)

fn get() -> Int:
  return factorial(5)
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 120)
}

func TestClosure(t *testing.T) {
        src := `module test
fn make_adder(n: Int) -> Int:
  let adder = |x| -> x + n
  return adder

fn get() -> Int:
  let add5 = make_adder(5)
  return add5(10)
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 15)
}

// --- Control Flow Tests ---

func TestIfStmt(t *testing.T) {
        src := `module test
fn check(n: Int) -> String:
  if n > 0:
    return "positive"
  elif n < 0:
    return "negative"
  else:
    return "zero"
`
        interp := runModule(t, src)
        r, _ := interp.RunFunction("check", []Value{&IntVal{Val: 5}})
        expectString(t, r, "positive")
        r, _ = interp.RunFunction("check", []Value{&IntVal{Val: -1}})
        expectString(t, r, "negative")
        r, _ = interp.RunFunction("check", []Value{&IntVal{Val: 0}})
        expectString(t, r, "zero")
}

func TestIfExpr(t *testing.T) {
        src := `module test
fn abs(n: Int) -> Int:
  return if n >= 0 then n else -n
`
        interp := runModule(t, src)
        r, _ := interp.RunFunction("abs", []Value{&IntVal{Val: -5}})
        expectInt(t, r, 5)
        r, _ = interp.RunFunction("abs", []Value{&IntVal{Val: 3}})
        expectInt(t, r, 3)
}

func TestForLoop(t *testing.T) {
        src := `module test
fn sum_list() -> Int:
  let items = [1, 2, 3, 4, 5]
  let mut total = 0
  for item in items:
    total = total + item
  return total
`
        result := runFunc(t, src, "sum_list", nil)
        expectInt(t, result, 15)
}

func TestForWithRange(t *testing.T) {
        src := `module test
fn sum_range() -> Int:
  let mut total = 0
  for i in range(5):
    total = total + i
  return total
`
        result := runFunc(t, src, "sum_range", nil)
        expectInt(t, result, 10)
}

func TestWhileLoop(t *testing.T) {
        src := `module test
fn count() -> Int:
  let mut n = 0
  while n < 10:
    n = n + 1
  return n
`
        result := runFunc(t, src, "count", nil)
        expectInt(t, result, 10)
}

func TestBreak(t *testing.T) {
        src := `module test
fn find_first() -> Int:
  let mut result = 0
  for i in range(100):
    if i == 5:
      result = i
      break
  return result
`
        result := runFunc(t, src, "find_first", nil)
        expectInt(t, result, 5)
}

func TestContinue(t *testing.T) {
        src := `module test
fn sum_even() -> Int:
  let mut total = 0
  for i in range(10):
    if i % 2 != 0:
      continue
    total = total + i
  return total
`
        result := runFunc(t, src, "sum_even", nil)
        expectInt(t, result, 20) // 0+2+4+6+8
}

// --- Struct Tests ---

func TestStructCreation(t *testing.T) {
        src := `module test
struct Point:
  x: Int
  y: Int

fn get_x() -> Int:
  let p = Point(x: 10, y: 20)
  return p.x
`
        result := runFunc(t, src, "get_x", nil)
        expectInt(t, result, 10)
}

func TestStructFieldAccess(t *testing.T) {
        src := `module test
struct Point:
  x: Int
  y: Int

fn sum_coords() -> Int:
  let p = Point(x: 3, y: 4)
  return p.x + p.y
`
        result := runFunc(t, src, "sum_coords", nil)
        expectInt(t, result, 7)
}

// --- Enum Tests ---

func TestEnumVariant(t *testing.T) {
        src := `module test
enum Color:
  Red
  Green
  Blue

fn is_red() -> Bool:
  let c = Color.Red
  match c:
    case Color.Red:
      return true
    case _:
      return false
`
        result := runFunc(t, src, "is_red", nil)
        expectBool(t, result, true)
}

func TestEnumWithData(t *testing.T) {
        src := `module test
enum Shape:
  Circle(Float)
  Rect(Float, Float)

fn get_radius() -> Float:
  let s = Shape.Circle(5.0)
  match s:
    case Shape.Circle(r):
      return r
    case _:
      return 0.0
`
        result := runFunc(t, src, "get_radius", nil)
        expectFloat(t, result, 5.0)
}

// --- Match Tests ---

func TestMatchLiteral(t *testing.T) {
        src := `module test
fn describe(n: Int) -> String:
  match n:
    case 0:
      return "zero"
    case 1:
      return "one"
    case _:
      return "other"
`
        interp := runModule(t, src)
        r, _ := interp.RunFunction("describe", []Value{&IntVal{Val: 0}})
        expectString(t, r, "zero")
        r, _ = interp.RunFunction("describe", []Value{&IntVal{Val: 1}})
        expectString(t, r, "one")
        r, _ = interp.RunFunction("describe", []Value{&IntVal{Val: 99}})
        expectString(t, r, "other")
}

func TestMatchBinding(t *testing.T) {
        src := `module test
fn identity(n: Int) -> Int:
  match n:
    case x:
      return x
`
        result := runFunc(t, src, "identity", []Value{&IntVal{Val: 42}})
        expectInt(t, result, 42)
}

// --- List Tests ---

func TestListCreation(t *testing.T) {
        src := `module test
fn get() -> Int:
  let items = [10, 20, 30]
  return items[1]
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 20)
}

func TestListConcat(t *testing.T) {
        src := `module test
fn get() -> Int:
  let a = [1, 2]
  let b = [3, 4]
  let c = a + b
  return len(c)
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 4)
}

func TestListComprehension(t *testing.T) {
        src := `module test
fn get() -> Int:
  let nums = [1, 2, 3, 4, 5]
  let doubled = [x * 2 for x in nums]
  return doubled[2]
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 6)
}

func TestListComprehensionWithFilter(t *testing.T) {
        src := `module test
fn get() -> Int:
  let nums = [1, 2, 3, 4, 5, 6]
  let evens = [x for x in nums if x % 2 == 0]
  return len(evens)
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 3)
}

// --- Map Tests ---

func TestMapCreation(t *testing.T) {
        src := `module test
fn get() -> Int:
  let m = {"a": 1, "b": 2}
  return m["b"]
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 2)
}

// --- Ok/Err/Some/None Tests ---

func TestOkResult(t *testing.T) {
        src := `module test
fn get() -> Int:
  let r = Ok(42)
  match r:
    case Ok(v):
      return v
    case Err(e):
      return 0
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 42)
}

func TestErrResult(t *testing.T) {
        src := `module test
fn get() -> String:
  let r = Err("fail")
  match r:
    case Ok(v):
      return "ok"
    case Err(e):
      return e
`
        result := runFunc(t, src, "get", nil)
        expectString(t, result, "fail")
}

func TestSomeOption(t *testing.T) {
        src := `module test
fn get() -> Int:
  let o = Some(42)
  match o:
    case Some(v):
      return v
    case None:
      return 0
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 42)
}

// --- Built-in Function Tests ---

func TestLenBuiltin(t *testing.T) {
        src := `module test
fn get() -> Int:
  return len([1, 2, 3])
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 3)
}

func TestRangeBuiltin(t *testing.T) {
        src := `module test
fn get() -> Int:
  return len(range(5))
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 5)
}

func TestAbsBuiltin(t *testing.T) {
        src := `module test
fn get() -> Int:
  return abs(-42)
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 42)
}

func TestMinMaxBuiltin(t *testing.T) {
        src := `module test
fn get_min() -> Int:
  return min(3, 1, 2)
fn get_max() -> Int:
  return max(3, 1, 2)
`
        interp := runModule(t, src)
        r, _ := interp.RunFunction("get_min", nil)
        expectInt(t, r, 1)
        r, _ = interp.RunFunction("get_max", nil)
        expectInt(t, r, 3)
}

// --- Error Handling Tests ---

func TestDivisionByZero(t *testing.T) {
        src := `module test
fn div() -> Int:
  return 10 / 0
`
        expectRuntimeError(t, src, "div", "division by zero")
}

func TestIndexOutOfBounds(t *testing.T) {
        src := `module test
fn get() -> Int:
  let items = [1, 2, 3]
  return items[10]
`
        expectRuntimeError(t, src, "get", "out of bounds")
}

func TestUndefinedVariable(t *testing.T) {
        src := `module test
fn get() -> Int:
  return x
`
        expectRuntimeError(t, src, "get", "undefined variable")
}

func TestImmutableAssignment(t *testing.T) {
        src := `module test
fn get() -> Int:
  let x = 10
  x = 20
  return x
`
        expectRuntimeError(t, src, "get", "immutable")
}

func TestAssertPass(t *testing.T) {
        src := `module test
fn check() -> Int:
  assert 1 == 1
  return 42
`
        result := runFunc(t, src, "check", nil)
        expectInt(t, result, 42)
}

func TestAssertFail(t *testing.T) {
        src := `module test
fn check() -> Int:
  assert 1 == 2, "math is broken"
  return 42
`
        expectRuntimeError(t, src, "check", "math is broken")
}

// --- Test Block Tests ---

func TestTestBlockPass(t *testing.T) {
        src := `module test

fn add(a: Int, b: Int) -> Int:
  return a + b

test "addition works":
  assert add(2, 3) == 5
  assert add(0, 0) == 0
`
        module := parseModule(t, src)
        results := RunTests(module)
        if len(results) != 1 {
                t.Fatalf("expected 1 test result, got %d", len(results))
        }
        if !results[0].Passed {
                t.Fatalf("test should pass: %s", results[0].Error)
        }
}

func TestTestBlockFail(t *testing.T) {
        src := `module test

fn add(a: Int, b: Int) -> Int:
  return a + b

test "bad assertion":
  assert add(2, 3) == 6
`
        module := parseModule(t, src)
        results := RunTests(module)
        if len(results) != 1 {
                t.Fatalf("expected 1 test result, got %d", len(results))
        }
        if results[0].Passed {
                t.Fatal("test should fail")
        }
}

func TestMultipleTestBlocks(t *testing.T) {
        src := `module test

fn double(n: Int) -> Int:
  return n * 2

test "double positive":
  assert double(5) == 10

test "double zero":
  assert double(0) == 0

test "double negative":
  assert double(-3) == -6
`
        module := parseModule(t, src)
        results := RunTests(module)
        if len(results) != 3 {
                t.Fatalf("expected 3 test results, got %d", len(results))
        }
        for _, r := range results {
                if !r.Passed {
                        t.Fatalf("test %q should pass: %s", r.Name, r.Error)
                }
        }
}

// --- With Block Tests ---

func TestWithBlock(t *testing.T) {
        src := `module test
fn get() -> Int:
  let db = 42
  with db:
    return db
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 42)
}

// --- Constant Tests ---

func TestTopLevelConstant(t *testing.T) {
        src := `module test
let max_val = 100

fn get() -> Int:
  return max_val
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 100)
}

// --- Complex Integration Tests ---

func TestFibonacci(t *testing.T) {
        src := `module test
fn fib(n: Int) -> Int:
  if n <= 1:
    return n
  return fib(n - 1) + fib(n - 2)

fn get() -> Int:
  return fib(10)
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 55)
}

func TestFizzBuzz(t *testing.T) {
        src := `module test
fn fizzbuzz(n: Int) -> String:
  if n % 15 == 0:
    return "FizzBuzz"
  elif n % 3 == 0:
    return "Fizz"
  elif n % 5 == 0:
    return "Buzz"
  else:
    return "other"

fn get() -> String:
  return fizzbuzz(15)
`
        result := runFunc(t, src, "get", nil)
        expectString(t, result, "FizzBuzz")
}

func TestStructWithEnum(t *testing.T) {
        src := `module test
enum Status:
  Active
  Inactive

struct User:
  name: String
  status: Status

fn is_active() -> Bool:
  let u = User(name: "Alice", status: Status.Active)
  match u.status:
    case Status.Active:
      return true
    case _:
      return false
`
        result := runFunc(t, src, "is_active", nil)
        expectBool(t, result, true)
}

func TestNestedFunctionCalls(t *testing.T) {
        src := `module test
fn add(a: Int, b: Int) -> Int:
  return a + b

fn mul(a: Int, b: Int) -> Int:
  return a * b

fn get() -> Int:
  return add(mul(3, 4), mul(5, 6))
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 42)
}

func TestWhileWithBreak(t *testing.T) {
        src := `module test
fn first_divisible() -> Int:
  let mut i = 1
  while true:
    if i % 7 == 0:
      return i
    i = i + 1
  return 0
`
        result := runFunc(t, src, "first_divisible", nil)
        expectInt(t, result, 7)
}

func TestLambdaExpression(t *testing.T) {
        src := `module test
fn get() -> Int:
  let triple = |n| -> n * 3
  return triple(14)
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 42)
}

func TestNegativeIndex(t *testing.T) {
        src := `module test
fn get() -> Int:
  let items = [10, 20, 30]
  return items[-1]
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 30)
}

func TestFormatTestResults(t *testing.T) {
        results := []TestResult{
                {Name: "test1", Passed: true},
                {Name: "test2", Passed: false, Error: "assertion failed"},
        }
        output := FormatTestResults(results)
        if !strings.Contains(output, "✓ test1") {
                t.Fatal("expected pass marker for test1")
        }
        if !strings.Contains(output, "✗ test2") {
                t.Fatal("expected fail marker for test2")
        }
        if !strings.Contains(output, "1 passed") {
                t.Fatal("expected pass count")
        }
}

func TestOptionVal(t *testing.T) {
        some := &OptionVal{IsSome: true, Val: &IntVal{Val: 42}}
        none := &OptionVal{IsSome: false}
        if some.String() != "Some(42)" {
                t.Fatalf("expected 'Some(42)', got %q", some.String())
        }
        if none.String() != "None" {
                t.Fatalf("expected 'None', got %q", none.String())
        }
        if !IsTruthy(some) {
                t.Fatal("Some should be truthy")
        }
        if IsTruthy(none) {
                t.Fatal("None should be falsy")
        }
}

func TestResultVal(t *testing.T) {
        ok := &ResultVal{IsOk: true, Val: &IntVal{Val: 42}}
        err := &ResultVal{IsOk: false, Val: &StringVal{Val: "fail"}}
        if ok.String() != "Ok(42)" {
                t.Fatalf("expected 'Ok(42)', got %q", ok.String())
        }
        if err.String() != `Err("fail")` {
                t.Fatalf("expected 'Err(\"fail\")', got %q", err.String())
        }
}

func TestEnumVal(t *testing.T) {
        v := &EnumVal{TypeName: "Color", VariantName: "Red"}
        if v.String() != "Color.Red" {
                t.Fatalf("expected 'Color.Red', got %q", v.String())
        }
        v2 := &EnumVal{TypeName: "Shape", VariantName: "Circle", Data: []Value{&FloatVal{Val: 5.0}}}
        if v2.String() != "Shape.Circle(5)" {
                t.Fatalf("expected 'Shape.Circle(5)', got %q", v2.String())
        }
}

func TestStructVal(t *testing.T) {
        v := &StructVal{TypeName: "Point", Fields: map[string]Value{"x": &IntVal{Val: 1}}}
        s := v.String()
        if !strings.Contains(s, "Point(") || !strings.Contains(s, "x: 1") {
                t.Fatalf("unexpected struct string: %q", s)
        }
}

func TestMapVal(t *testing.T) {
        v := &MapVal{
                Keys:   []Value{&StringVal{Val: "a"}},
                Values: []Value{&IntVal{Val: 1}},
        }
        if v.String() != `{"a": 1}` {
                t.Fatalf("unexpected map string: %q", v.String())
        }
}

func TestListVal(t *testing.T) {
        v := &ListVal{Elements: []Value{&IntVal{Val: 1}, &IntVal{Val: 2}}}
        if v.String() != "[1, 2]" {
                t.Fatalf("unexpected list string: %q", v.String())
        }
}

func TestTupleVal(t *testing.T) {
        v := &TupleVal{Elements: []Value{&IntVal{Val: 1}, &StringVal{Val: "hi"}}}
        expected := `(1, "hi")`
        if v.String() != expected {
                t.Fatalf("expected %q, got %q", expected, v.String())
        }
}

func TestSetVal(t *testing.T) {
        v := &SetVal{Elements: []Value{&IntVal{Val: 1}, &IntVal{Val: 2}}}
        if v.String() != "{1, 2}" {
                t.Fatalf("unexpected set string: %q", v.String())
        }
}

func TestFunctionVal(t *testing.T) {
        v := &FunctionVal{Name: "add"}
        if v.String() != "<fn add>" {
                t.Fatalf("expected '<fn add>', got %q", v.String())
        }
}

func TestRuntimeError(t *testing.T) {
        err := &RuntimeError{Message: "test error"}
        if err.Error() != "runtime error: test error" {
                t.Fatalf("unexpected error string: %q", err.Error())
        }
}


// --- Pipeline Operator Tests ---

func TestPipelineBasic(t *testing.T) {
        // Basic pipeline: value |> func
        src := `
fn double(x: Int) -> Int:
    return x * 2

fn test_pipeline() -> Int:
    return 5 |> double
`
        result := runFunc(t, src, "test_pipeline", nil)
        expectInt(t, result, 10)
}

func TestPipelineChained(t *testing.T) {
        // Chained pipelines: data |> f1 |> f2 |> f3
        src := `
fn add_one(x: Int) -> Int:
    return x + 1

fn double(x: Int) -> Int:
    return x * 2

fn square(x: Int) -> Int:
    return x * x

fn test_chain() -> Int:
    return 3 |> add_one |> double |> square
`
        // 3 -> add_one -> 4 -> double -> 8 -> square -> 64
        result := runFunc(t, src, "test_chain", nil)
        expectInt(t, result, 64)
}

func TestPipelineWithLambda(t *testing.T) {
        // Pipeline with inline lambda
        src := `
fn test_pipe_lambda() -> Int:
    return 10 |> |x| -> x + 5
`
        result := runFunc(t, src, "test_pipe_lambda", nil)
        expectInt(t, result, 15)
}

func TestPipelineChainedWithLambdas(t *testing.T) {
        // Chained pipeline with mix of functions and lambdas
        src := `
fn double(x: Int) -> Int:
    return x * 2

fn test_chain_lambdas() -> Int:
    return 3 |> double |> |x| -> x + 10
`
        // 3 -> double -> 6 -> +10 -> 16
        result := runFunc(t, src, "test_chain_lambdas", nil)
        expectInt(t, result, 16)
}

func TestPipelineWithStrings(t *testing.T) {
        // Pipeline with string operations
        src := `
fn exclaim(s: String) -> String:
    return s + "!"

fn greet(name: String) -> String:
    return "Hello, " + name

fn test_pipe_string() -> String:
    return "World" |> greet |> exclaim
`
        // "World" -> greet -> "Hello, World" -> exclaim -> "Hello, World!"
        result := runFunc(t, src, "test_pipe_string", nil)
        expectString(t, result, "Hello, World!")
}

func TestPipelineWithFloat(t *testing.T) {
        // Pipeline with float values
        src := `
fn halve(x: Float) -> Float:
    return x / 2.0

fn test_pipe_float() -> Float:
    return 10.0 |> halve
`
        result := runFunc(t, src, "test_pipe_float", nil)
        expectFloat(t, result, 5.0)
}

func TestPipelineWithBool(t *testing.T) {
        // Pipeline returning bool
        src := `
fn is_positive(x: Int) -> Bool:
    return x > 0

fn test_pipe_bool() -> Bool:
    return 42 |> is_positive
`
        result := runFunc(t, src, "test_pipe_bool", nil)
        expectBool(t, result, true)
}

func TestPipelineInLetBinding(t *testing.T) {
        // Pipeline result stored in variable
        src := `
fn double(x: Int) -> Int:
    return x * 2

fn triple(x: Int) -> Int:
    return x * 3

fn test_pipe_let() -> Int:
    let a = 5 |> double
    let b = a |> triple
    return b
`
        // a = 10, b = 30
        result := runFunc(t, src, "test_pipe_let", nil)
        expectInt(t, result, 30)
}

func TestPipelinePrecedence(t *testing.T) {
        // Pipeline has lower precedence than arithmetic
        // So `2 + 3 |> double` means `(2 + 3) |> double`
        src := `
fn double(x: Int) -> Int:
    return x * 2

fn test_pipe_prec() -> Int:
    return 2 + 3 |> double
`
        // (2 + 3) = 5 |> double = 10
        result := runFunc(t, src, "test_pipe_prec", nil)
        expectInt(t, result, 10)
}

func TestPipelineIdentity(t *testing.T) {
        // Pipeline with identity lambda
        src := `
fn test_pipe_identity() -> Int:
    return 42 |> |x| -> x
`
        result := runFunc(t, src, "test_pipe_identity", nil)
        expectInt(t, result, 42)
}

func TestPipelineLongChain(t *testing.T) {
        // Longer chain to verify left-associativity
        src := `
fn inc(x: Int) -> Int:
    return x + 1

fn test_long_chain() -> Int:
    return 0 |> inc |> inc |> inc |> inc |> inc
`
        result := runFunc(t, src, "test_long_chain", nil)
        expectInt(t, result, 5)
}

func TestPipelineWithNone(t *testing.T) {
        // Pipeline that receives none and handles it
        src := `
fn to_string(x) -> String:
    return str(x)

fn test_pipe_none() -> String:
    return none |> to_string
`
        result := runFunc(t, src, "test_pipe_none", nil)
        expectString(t, result, "none")
}

func TestPipelineLexerToken(t *testing.T) {
        // Verify the lexer correctly tokenizes |>
        src := "5 |> double\n"
        l := lexer.New(src, "test.aura")
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("lexer errors: %v", errs)
        }

        // Expected: INT_LIT("5"), PIPE_GT("|>"), IDENT("double"), NEWLINE, EOF
        found := false
        for _, tok := range tokens {
                if tok.Literal == "|>" {
                        found = true
                        break
                }
        }
        if !found {
                t.Fatal("expected PIPE_GT token in lexer output")
        }
}

func TestPipelineLexerPipeStillWorks(t *testing.T) {
        // Verify that plain | still works (for lambda params, union types)
        src := "|x| x + 1\n"
        l := lexer.New(src, "test.aura")
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("lexer errors: %v", errs)
        }

        // Should have PIPE tokens, not PIPE_GT
        for _, tok := range tokens {
                if tok.Literal == "|>" {
                        t.Fatal("plain | should not produce PIPE_GT")
                }
        }
}

// ============================================================
// --- String Interpolation Tests ---
// ============================================================

func TestStringInterpolation_SimpleVariable(t *testing.T) {
        src := `module test
fn greet() -> String:
  let name = "World"
  return "Hello, {name}!"
`
        result := runFunc(t, src, "greet", nil)
        expectString(t, result, "Hello, World!")
}

func TestStringInterpolation_Expression(t *testing.T) {
        src := `module test
fn calc() -> String:
  let x = 3
  let y = 7
  return "Result: {x + y}"
`
        result := runFunc(t, src, "calc", nil)
        expectString(t, result, "Result: 10")
}

func TestStringInterpolation_FieldAccess(t *testing.T) {
        src := `module test
struct User:
  name: String
  age: Int

fn info() -> String:
  let u = User(name: "Alice", age: 30)
  return "User: {u.name}"
`
        result := runFunc(t, src, "info", nil)
        expectString(t, result, "User: Alice")
}

func TestStringInterpolation_MultipleInterpolations(t *testing.T) {
        src := `module test
fn multi() -> String:
  let a = "foo"
  let b = "bar"
  let c = 42
  return "{a} and {b} is {c}"
`
        result := runFunc(t, src, "multi", nil)
        expectString(t, result, "foo and bar is 42")
}

func TestStringInterpolation_NestedExpression(t *testing.T) {
        src := `module test
fn nested() -> String:
  let x = 10
  let y = 3
  return "calc: {x * y + 1}"
`
        result := runFunc(t, src, "nested", nil)
        expectString(t, result, "calc: 31")
}

func TestStringInterpolation_BoolAndNone(t *testing.T) {
        src := `module test
fn show() -> String:
  let flag = true
  return "flag is {flag}"
`
        result := runFunc(t, src, "show", nil)
        expectString(t, result, "flag is true")
}

func TestStringInterpolation_NoInterpolation(t *testing.T) {
        src := `module test
fn plain() -> String:
  return "no interpolation here"
`
        result := runFunc(t, src, "plain", nil)
        expectString(t, result, "no interpolation here")
}

// ============================================================
// --- Pipeline Operator Tests ---
// ============================================================

func TestPipeline_SimpleFunction(t *testing.T) {
        src := `module test
fn double(x: Int) -> Int:
  return x * 2

fn get() -> Int:
  return 5 |> double
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 10)
}

func TestPipeline_MultiStage(t *testing.T) {
        src := `module test
fn double(x: Int) -> Int:
  return x * 2

fn add_one(x: Int) -> Int:
  return x + 1

fn get() -> Int:
  return 3 |> double |> add_one
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 7)
}

func TestPipeline_WithLambda(t *testing.T) {
        src := `module test
fn get() -> Int:
  return 4 |> |x| -> x * x
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 16)
}

func TestPipeline_WithBuiltinFunction(t *testing.T) {
        src := `module test
fn get() -> String:
  return 42 |> str
`
        result := runFunc(t, src, "get", nil)
        expectString(t, result, "42")
}

func TestPipeline_ChainedWithLambdas(t *testing.T) {
        src := `module test
fn get() -> Int:
  return 2 |> |x| -> x + 3 |> |x| -> x * 10
`
        result := runFunc(t, src, "get", nil)
        expectInt(t, result, 50)
}

func TestPipeline_WithStringFunction(t *testing.T) {
        src := `module test
fn exclaim(s: String) -> String:
  return s + "!"

fn get() -> String:
  return "hello" |> exclaim
`
        result := runFunc(t, src, "get", nil)
        expectString(t, result, "hello!")
}

func TestPipeline_NonFunctionError(t *testing.T) {
        src := `module test
fn get() -> Int:
  return 5 |> 10
`
        expectRuntimeError(t, src, "get", "cannot call")
}

// ============================================================
// --- Option Chaining Tests ---
// ============================================================

func TestOptionChaining_SuccessfulFieldAccess(t *testing.T) {
        src := `module test
struct User:
  name: String
  age: Int

fn get() -> String:
  let u = User(name: "Alice", age: 25)
  return u?.name
`
        result := runFunc(t, src, "get", nil)
        expectString(t, result, "Alice")
}

func TestOptionChaining_NoneShortCircuit(t *testing.T) {
        src := `module test
fn get() -> None:
  let u = None
  return u?.name
`
        result := runFunc(t, src, "get", nil)
        expectNone(t, result)
}

func TestOptionChaining_NestedChaining(t *testing.T) {
        src := `module test
struct Address:
  city: String

struct User:
  name: String
  address: Address

fn get() -> String:
  let u = User(name: "Bob", address: Address(city: "NYC"))
  return u?.address?.city
`
        result := runFunc(t, src, "get", nil)
        expectString(t, result, "NYC")
}

func TestOptionChaining_SomeOptionUnwrap(t *testing.T) {
        src := `module test
struct User:
  name: String

fn get() -> String:
  let u = Some(User(name: "Carol"))
  return u?.name
`
        result := runFunc(t, src, "get", nil)
        expectString(t, result, "Carol")
}

func TestOptionChaining_NoneOptionValue(t *testing.T) {
        // None is an OptionVal{IsSome: false} — option chaining should short-circuit
        src := `module test
fn get() -> Bool:
  let val = None
  let result = val?.something
  match result:
    case None:
      return true
    case _:
      return false
`
        result := runFunc(t, src, "get", nil)
        expectBool(t, result, true)
}

func TestOptionChaining_MissingField(t *testing.T) {
        src := `module test
struct User:
  name: String

fn get() -> None:
  let u = User(name: "Dave")
  return u?.nonexistent
`
        result := runFunc(t, src, "get", nil)
        expectNone(t, result)
}

func TestOptionChaining_NestedNoneMiddle(t *testing.T) {
        src := `module test
struct Address:
  city: String

fn get() -> None:
  let u = None
  return u?.address?.city
`
        result := runFunc(t, src, "get", nil)
        expectNone(t, result)
}
