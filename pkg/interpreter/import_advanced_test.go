package interpreter

import (
        "path/filepath"
        "strings"
        "testing"

        "github.com/unclebucklarson/aura/pkg/lexer"
        "github.com/unclebucklarson/aura/pkg/module"
        "github.com/unclebucklarson/aura/pkg/parser"
)

// === Advanced Namespace Management Tests ===

func TestNamespaceIsolation(t *testing.T) {
        // Importing a module should not pollute the global namespace
        dir := t.TempDir()
        createTestFile(t, dir, "modA.aura", `
pub fn compute():
    return 100

fn internal_helper():
    return 42
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import modA

let result = modA.compute()
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if i, ok := result.(*IntVal); !ok || i.Val != 100 {
                t.Errorf("result = %v, expected 100", result)
        }

        // internal_helper should NOT be accessible
        if _, ok := interp.env.Get("internal_helper"); ok {
                t.Error("internal_helper should not be accessible in main namespace")
        }
}

func TestQualifiedAccess(t *testing.T) {
        // Module.symbol access pattern
        dir := t.TempDir()
        createTestFile(t, dir, "config.aura", `
pub let version = 1
pub let name = "test"
pub fn get_info():
    return "info"
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import config

let v = config.version
let n = config.name
let info = config.get_info()
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        v, ok := interp.env.Get("v")
        if !ok || v.(*IntVal).Val != 1 {
                t.Errorf("config.version = %v, expected 1", v)
        }
        n, ok := interp.env.Get("n")
        if !ok || n.(*StringVal).Val != "test" {
                t.Errorf("config.name = %v, expected 'test'", n)
        }
        info, ok := interp.env.Get("info")
        if !ok || info.(*StringVal).Val != "info" {
                t.Errorf("config.get_info() = %v, expected 'info'", info)
        }
}

func TestNamedImportScoping(t *testing.T) {
        // from X import a, b should only bring a and b into scope
        dir := t.TempDir()
        createTestFile(t, dir, "utils.aura", `
pub fn alpha():
    return 1

pub fn beta():
    return 2

pub fn gamma():
    return 3
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `from utils import alpha, beta

let a = alpha()
let b = beta()
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        a, ok := interp.env.Get("a")
        if !ok || a.(*IntVal).Val != 1 {
                t.Errorf("a = %v, expected 1", a)
        }
        b, ok := interp.env.Get("b")
        if !ok || b.(*IntVal).Val != 2 {
                t.Errorf("b = %v, expected 2", b)
        }
        // gamma should NOT be available
        if _, ok := interp.env.Get("gamma"); ok {
                t.Error("gamma should not be in scope (not imported)")
        }
}

func TestAliasedImportDoesNotExposeOriginalName(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "mylib.aura", `
pub fn helper():
    return 42
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import mylib as ml

let result = ml.helper()
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok || result.(*IntVal).Val != 42 {
                t.Errorf("result = %v, expected 42", result)
        }

        // 'mylib' should NOT be accessible when aliased
        if _, ok := interp.env.Get("mylib"); ok {
                t.Error("original module name 'mylib' should not be accessible when aliased as 'ml'")
        }
}

func TestImprovedErrorForUndefinedExport(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "small.aura", `
pub fn exists():
    return 1

pub fn also_exists():
    return 2
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `from small import exists, nonexistent
`
        resolver := module.NewResolver(dir)

        l := lexer.New(source, mainFile)
        tokens, _ := l.Tokenize()
        p := parser.New(tokens, mainFile)
        ast, _ := p.Parse()
        interp := NewWithResolver(ast, mainFile, resolver)
        _, err := interp.Run()
        if err == nil {
                t.Fatal("expected error for non-existent export")
        }
        // Error should mention available exports
        errStr := err.Error()
        if !strings.Contains(errStr, "nonexistent") {
                t.Errorf("error should mention 'nonexistent', got: %s", errStr)
        }
        if !strings.Contains(errStr, "available") || !strings.Contains(errStr, "exists") {
                t.Errorf("error should list available exports, got: %s", errStr)
        }
}

// === Module Initialization Ordering Tests ===

func TestModuleInitOrder(t *testing.T) {
        // Module dependencies should be initialized before dependents
        dir := t.TempDir()
        createTestFile(t, dir, "base.aura", `
pub let base_value = 10
`)
        createTestFile(t, dir, "derived.aura", `
import base

pub let derived_value = base.base_value + 5
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import derived

let result = derived.derived_value
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if i, ok := result.(*IntVal); !ok || i.Val != 15 {
                t.Errorf("result = %v, expected 15", result)
        }
}

func TestDeepModuleChain(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "level1.aura", `
pub fn val():
    return 1
`)
        createTestFile(t, dir, "level2.aura", `
import level1

pub fn val():
    return level1.val() + 1
`)
        createTestFile(t, dir, "level3.aura", `
import level2

pub fn val():
    return level2.val() + 1
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import level3

let result = level3.val()
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if i, ok := result.(*IntVal); !ok || i.Val != 3 {
                t.Errorf("result = %v, expected 3", result)
        }
}

func TestMultipleModulesShareDependency(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "shared.aura", `
pub let value = 42
`)
        createTestFile(t, dir, "userA.aura", `
import shared

pub fn get_a():
    return shared.value + 1
`)
        createTestFile(t, dir, "userB.aura", `
import shared

pub fn get_b():
    return shared.value + 2
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import userA
import userB

let a = userA.get_a()
let b = userB.get_b()
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        a, ok := interp.env.Get("a")
        if !ok || a.(*IntVal).Val != 43 {
                t.Errorf("a = %v, expected 43", a)
        }
        b, ok := interp.env.Get("b")
        if !ok || b.(*IntVal).Val != 44 {
                t.Errorf("b = %v, expected 44", b)
        }
}

// === Enhanced Import Cycle Prevention Tests ===

func TestCircularDependencyDetection(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "cycleA.aura", `
import cycleB

pub fn a():
    return 1
`)
        createTestFile(t, dir, "cycleB.aura", `
import cycleA

pub fn b():
    return 2
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import cycleA
`
        resolver := module.NewResolver(dir)

        l := lexer.New(source, mainFile)
        tokens, _ := l.Tokenize()
        p := parser.New(tokens, mainFile)
        ast, _ := p.Parse()
        interp := NewWithResolver(ast, mainFile, resolver)
        _, err := interp.Run()
        if err == nil {
                t.Fatal("expected error for circular dependency")
        }
        errStr := err.Error()
        if !strings.Contains(errStr, "circular") {
                t.Errorf("error should mention 'circular', got: %s", errStr)
        }
}

func TestCircularDependencyPathShown(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "pathA.aura", `
import pathB

pub fn a():
    return 1
`)
        createTestFile(t, dir, "pathB.aura", `
import pathC

pub fn b():
    return 2
`)
        createTestFile(t, dir, "pathC.aura", `
import pathA

pub fn c():
    return 3
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import pathA
`
        resolver := module.NewResolver(dir)

        l := lexer.New(source, mainFile)
        tokens, _ := l.Tokenize()
        p := parser.New(tokens, mainFile)
        ast, _ := p.Parse()
        interp := NewWithResolver(ast, mainFile, resolver)
        _, err := interp.Run()
        if err == nil {
                t.Fatal("expected error for circular dependency")
        }
        errStr := err.Error()
        if !strings.Contains(errStr, "circular") {
                t.Errorf("error should mention 'circular', got: %s", errStr)
        }
}

// === Package-Level Initialization Tests ===

func TestModuleConstantsInitialized(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "constants.aura", `
pub let pi_approx = 3
pub let tau_approx = pi_approx * 2
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import constants

let pi = constants.pi_approx
let tau = constants.tau_approx
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        pi, ok := interp.env.Get("pi")
        if !ok || pi.(*IntVal).Val != 3 {
                t.Errorf("pi = %v, expected 3", pi)
        }
        tau, ok := interp.env.Get("tau")
        if !ok || tau.(*IntVal).Val != 6 {
                t.Errorf("tau = %v, expected 6", tau)
        }
}

func TestModuleWithFunctionsDependingOnConstants(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "mymod.aura", `
pub let factor = 10

pub fn scaled(x):
    return x * factor
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import mymod

let result = mymod.scaled(5)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok || result.(*IntVal).Val != 50 {
                t.Errorf("result = %v, expected 50", result)
        }
}

// === std.testing Tests ===

func TestStdTestingImport(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `import std.testing
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        testingMod, ok := interp.env.Get("testing")
        if !ok {
                t.Fatal("testing module not found")
        }
        modVal, ok := testingMod.(*ModuleVal)
        if !ok {
                t.Fatalf("testing is not a ModuleVal, got %T", testingMod)
        }

        expectedExports := []string{"assert", "assert_eq", "assert_ne", "assert_true", "assert_false",
                "assert_none", "assert_some", "assert_ok", "assert_err", "test", "run_tests"}
        for _, name := range expectedExports {
                if _, ok := modVal.Exports[name]; !ok {
                        t.Errorf("testing.%s not found in exports", name)
                }
        }
}

func TestStdTestingAssert(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.testing import assert

let result = assert(true, "should pass")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if b, ok := result.(*BoolVal); !ok || !b.Val {
                t.Errorf("assert(true) should return true, got %v", result)
        }
}

func TestStdTestingAssertFails(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.testing import assert

let x = assert(false, "expected failure")
`
        resolver := module.NewResolver(dir)

        l := lexer.New(source, mainFile)
        tokens, _ := l.Tokenize()
        p := parser.New(tokens, mainFile)
        ast, _ := p.Parse()
        interp := NewWithResolver(ast, mainFile, resolver)
        _, err := interp.Run()
        if err == nil {
                t.Fatal("expected assertion error")
        }
        if !strings.Contains(err.Error(), "expected failure") {
                t.Errorf("error should contain message, got: %s", err.Error())
        }
}

func TestStdTestingAssertEq(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.testing import assert_eq

let r1 = assert_eq(42, 42)
let r2 = assert_eq("hello", "hello")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        r1, ok := interp.env.Get("r1")
        if !ok || !r1.(*BoolVal).Val {
                t.Error("assert_eq(42, 42) should pass")
        }
        r2, ok := interp.env.Get("r2")
        if !ok || !r2.(*BoolVal).Val {
                t.Error("assert_eq('hello', 'hello') should pass")
        }
}

func TestStdTestingAssertEqFails(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.testing import assert_eq

let x = assert_eq(1, 2)
`
        resolver := module.NewResolver(dir)

        l := lexer.New(source, mainFile)
        tokens, _ := l.Tokenize()
        p := parser.New(tokens, mainFile)
        ast, _ := p.Parse()
        interp := NewWithResolver(ast, mainFile, resolver)
        _, err := interp.Run()
        if err == nil {
                t.Fatal("expected assertion error")
        }
        if !strings.Contains(err.Error(), "expected") && !strings.Contains(err.Error(), "got") {
                t.Errorf("error should show expected vs got, got: %s", err.Error())
        }
}

func TestStdTestingAssertNe(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.testing import assert_ne

let r = assert_ne(1, 2)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        r, ok := interp.env.Get("r")
        if !ok || !r.(*BoolVal).Val {
                t.Error("assert_ne(1, 2) should pass")
        }
}

func TestStdTestingAssertNeFails(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.testing import assert_ne

let x = assert_ne(42, 42)
`
        resolver := module.NewResolver(dir)

        l := lexer.New(source, mainFile)
        tokens, _ := l.Tokenize()
        p := parser.New(tokens, mainFile)
        ast, _ := p.Parse()
        interp := NewWithResolver(ast, mainFile, resolver)
        _, err := interp.Run()
        if err == nil {
                t.Fatal("expected assertion error for equal values")
        }
        if !strings.Contains(err.Error(), "should not be equal") {
                t.Errorf("error should mention inequality, got: %s", err.Error())
        }
}

func TestStdTestingAssertTrueAndFalse(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.testing import assert_true, assert_false

let r1 = assert_true(1)
let r2 = assert_false(0)
let r3 = assert_true("non-empty")
let r4 = assert_false("")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        for _, name := range []string{"r1", "r2", "r3", "r4"} {
                r, ok := interp.env.Get(name)
                if !ok || !r.(*BoolVal).Val {
                        t.Errorf("%s should be true", name)
                }
        }
}

func TestStdTestingAssertNone(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.testing import assert_none

let r = assert_none(None)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        r, ok := interp.env.Get("r")
        if !ok || !r.(*BoolVal).Val {
                t.Error("assert_none(None) should pass")
        }
}

func TestStdTestingAssertSome(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.testing import assert_some

let r = assert_some(Some(42))
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        r, ok := interp.env.Get("r")
        if !ok || r.(*IntVal).Val != 42 {
                t.Errorf("assert_some(Some(42)) should return 42, got %v", r)
        }
}

func TestStdTestingAssertOk(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.testing import assert_ok

let r = assert_ok(Ok("success"))
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        r, ok := interp.env.Get("r")
        if !ok || r.(*StringVal).Val != "success" {
                t.Errorf("assert_ok(Ok('success')) should return 'success', got %v", r)
        }
}

func TestStdTestingAssertErr(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.testing import assert_err

let r = assert_err(Err("oops"))
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        r, ok := interp.env.Get("r")
        if !ok || r.(*StringVal).Val != "oops" {
                t.Errorf("assert_err(Err('oops')) should return 'oops', got %v", r)
        }
}

func TestStdTestingWithAlias(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `import std.testing as t

let r = t.assert_eq(1, 1)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        r, ok := interp.env.Get("r")
        if !ok || !r.(*BoolVal).Val {
                t.Error("t.assert_eq(1, 1) should pass")
        }
}

// === std.json Tests ===

func TestStdJsonImport(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `import std.json
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        jsonMod, ok := interp.env.Get("json")
        if !ok {
                t.Fatal("json module not found")
        }
        modVal, ok := jsonMod.(*ModuleVal)
        if !ok {
                t.Fatalf("json is not a ModuleVal, got %T", jsonMod)
        }
        for _, name := range []string{"parse", "stringify"} {
                if _, ok := modVal.Exports[name]; !ok {
                        t.Errorf("json.%s not found in exports", name)
                }
        }
}

func TestJsonParseInteger(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import parse

let result = parse("42")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if i, ok := result.(*IntVal); !ok || i.Val != 42 {
                t.Errorf("parse('42') = %v, expected 42", result)
        }
}

func TestJsonParseFloat(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import parse

let result = parse("3.14")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        f, ok := result.(*FloatVal)
        if !ok {
                t.Fatalf("expected FloatVal, got %T", result)
        }
        if f.Val < 3.13 || f.Val > 3.15 {
                t.Errorf("parse('3.14') = %f, expected ~3.14", f.Val)
        }
}

func TestJsonParseString(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import parse

let result = parse("\"hello world\"")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if s, ok := result.(*StringVal); !ok || s.Val != "hello world" {
                t.Errorf("parse('\"hello world\"') = %v, expected 'hello world'", result)
        }
}

func TestJsonParseBooleans(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import parse

let t_val = parse("true")
let f_val = parse("false")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        tVal, ok := interp.env.Get("t_val")
        if !ok || !tVal.(*BoolVal).Val {
                t.Error("parse('true') should be true")
        }
        fVal, ok := interp.env.Get("f_val")
        if !ok || fVal.(*BoolVal).Val {
                t.Error("parse('false') should be false")
        }
}

func TestJsonParseNull(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import parse

let result = parse("null")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if _, ok := result.(*NoneVal); !ok {
                t.Errorf("parse('null') should be NoneVal, got %T", result)
        }
}

func TestJsonParseArray(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import parse

let result = parse("[1, 2, 3]")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        list, ok := result.(*ListVal)
        if !ok {
                t.Fatalf("expected ListVal, got %T", result)
        }
        if len(list.Elements) != 3 {
                t.Fatalf("expected 3 elements, got %d", len(list.Elements))
        }
        for i, expected := range []int64{1, 2, 3} {
                if list.Elements[i].(*IntVal).Val != expected {
                        t.Errorf("element[%d] = %v, expected %d", i, list.Elements[i], expected)
                }
        }
}

func TestJsonParseEmptyArray(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import parse

let result = parse("[]")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        list, ok := result.(*ListVal)
        if !ok {
                t.Fatalf("expected ListVal, got %T", result)
        }
        if len(list.Elements) != 0 {
                t.Errorf("expected empty list, got %d elements", len(list.Elements))
        }
}

func TestJsonParseObject(t *testing.T) {
        // Test JSON object parsing directly through Go API to avoid Aura string interpolation
        exports := createStdJsonExports()
        parseFn := exports["parse"].(*BuiltinFnVal)
        result := parseFn.Fn([]Value{&StringVal{Val: `{"name": "Aura", "version": 1}`}})
        m, ok := result.(*MapVal)
        if !ok {
                t.Fatalf("expected MapVal, got %T", result)
        }
        if len(m.Keys) != 2 {
                t.Fatalf("expected 2 keys, got %d", len(m.Keys))
        }
        if m.Keys[0].(*StringVal).Val != "name" {
                t.Errorf("first key = %v, expected 'name'", m.Keys[0])
        }
        if m.Values[0].(*StringVal).Val != "Aura" {
                t.Errorf("first value = %v, expected 'Aura'", m.Values[0])
        }
}

func TestJsonParseEmptyObject(t *testing.T) {
        exports := createStdJsonExports()
        parseFn := exports["parse"].(*BuiltinFnVal)
        result := parseFn.Fn([]Value{&StringVal{Val: `{}`}})
        m, ok := result.(*MapVal)
        if !ok {
                t.Fatalf("expected MapVal, got %T", result)
        }
        if len(m.Keys) != 0 {
                t.Errorf("expected empty map, got %d entries", len(m.Keys))
        }
}

func TestJsonParseNestedObject(t *testing.T) {
        exports := createStdJsonExports()
        parseFn := exports["parse"].(*BuiltinFnVal)
        result := parseFn.Fn([]Value{&StringVal{Val: `{"data": {"x": 1}, "list": [true, null]}`}})
        m, ok := result.(*MapVal)
        if !ok {
                t.Fatalf("expected MapVal, got %T", result)
        }
        if len(m.Keys) != 2 {
                t.Fatalf("expected 2 keys, got %d", len(m.Keys))
        }
        // Verify nested object
        inner, ok := m.Values[0].(*MapVal)
        if !ok {
                t.Fatalf("expected nested MapVal, got %T", m.Values[0])
        }
        if inner.Values[0].(*IntVal).Val != 1 {
                t.Errorf("nested x = %v, expected 1", inner.Values[0])
        }
        // Verify nested list
        list, ok := m.Values[1].(*ListVal)
        if !ok {
                t.Fatalf("expected nested ListVal, got %T", m.Values[1])
        }
        if len(list.Elements) != 2 {
                t.Errorf("expected 2 list elements, got %d", len(list.Elements))
        }
}

func TestJsonParseStringEscapes(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import parse

let result = parse("\"hello\\nworld\"")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        s, ok := result.(*StringVal)
        if !ok {
                t.Fatalf("expected StringVal, got %T", result)
        }
        if s.Val != "hello\nworld" {
                t.Errorf("expected 'hello\\nworld', got %q", s.Val)
        }
}

func TestJsonParseNegativeNumber(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import parse

let result = parse("-42")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if i, ok := result.(*IntVal); !ok || i.Val != -42 {
                t.Errorf("parse('-42') = %v, expected -42", result)
        }
}

func TestJsonParseScientificNotation(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import parse

let result = parse("1.5e2")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        f, ok := result.(*FloatVal)
        if !ok {
                t.Fatalf("expected FloatVal, got %T", result)
        }
        if f.Val != 150.0 {
                t.Errorf("parse('1.5e2') = %f, expected 150.0", f.Val)
        }
}

func TestJsonParseInvalidInput(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import parse

let result = parse("invalid")
`
        resolver := module.NewResolver(dir)

        l := lexer.New(source, mainFile)
        tokens, _ := l.Tokenize()
        p := parser.New(tokens, mainFile)
        ast, _ := p.Parse()
        interp := NewWithResolver(ast, mainFile, resolver)
        _, err := interp.Run()
        if err == nil {
                t.Fatal("expected error for invalid JSON")
        }
}

func TestJsonStringifyInt(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import stringify

let result = stringify(42)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if s, ok := result.(*StringVal); !ok || s.Val != "42" {
                t.Errorf("stringify(42) = %v, expected '42'", result)
        }
}

func TestJsonStringifyString(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import stringify

let result = stringify("hello")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if s, ok := result.(*StringVal); !ok || s.Val != `"hello"` {
                t.Errorf("stringify('hello') = %v, expected '\"hello\"'", result)
        }
}

func TestJsonStringifyBool(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import stringify

let r1 = stringify(true)
let r2 = stringify(false)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        r1, _ := interp.env.Get("r1")
        if r1.(*StringVal).Val != "true" {
                t.Errorf("stringify(true) = %v, expected 'true'", r1)
        }
        r2, _ := interp.env.Get("r2")
        if r2.(*StringVal).Val != "false" {
                t.Errorf("stringify(false) = %v, expected 'false'", r2)
        }
}

func TestJsonStringifyNone(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import stringify

let result = stringify(None)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if s, ok := result.(*StringVal); !ok || s.Val != "null" {
                t.Errorf("stringify(None) = %v, expected 'null'", result)
        }
}

func TestJsonStringifyList(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import stringify

let result = stringify([1, 2, 3])
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        s, ok := result.(*StringVal)
        if !ok {
                t.Fatalf("expected StringVal, got %T", result)
        }
        if s.Val != "[1, 2, 3]" {
                t.Errorf("stringify([1,2,3]) = %q, expected '[1, 2, 3]'", s.Val)
        }
}

func TestJsonStringifyOptionNone(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import stringify

let result = stringify(Some(42))
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if s, ok := result.(*StringVal); !ok || s.Val != "42" {
                t.Errorf("stringify(Some(42)) = %v, expected '42'", result)
        }
}

func TestJsonRoundTrip(t *testing.T) {
        exports := createStdJsonExports()
        parseFn := exports["parse"].(*BuiltinFnVal)
        stringifyFn := exports["stringify"].(*BuiltinFnVal)

        original := `{"name": "Aura", "version": 1}`
        parsed := parseFn.Fn([]Value{&StringVal{Val: original}})
        back := stringifyFn.Fn([]Value{parsed})

        s, ok := back.(*StringVal)
        if !ok {
                t.Fatalf("expected StringVal, got %T", back)
        }
        if !strings.Contains(s.Val, "name") || !strings.Contains(s.Val, "Aura") {
                t.Errorf("round-trip should preserve data, got: %s", s.Val)
        }
}

func TestJsonParseMixedArray(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import parse

let result = parse("[1, \"two\", true, null]")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        list, ok := result.(*ListVal)
        if !ok {
                t.Fatalf("expected ListVal, got %T", result)
        }
        if len(list.Elements) != 4 {
                t.Fatalf("expected 4 elements, got %d", len(list.Elements))
        }
        if _, ok := list.Elements[0].(*IntVal); !ok {
                t.Error("element[0] should be IntVal")
        }
        if _, ok := list.Elements[1].(*StringVal); !ok {
                t.Error("element[1] should be StringVal")
        }
        if _, ok := list.Elements[2].(*BoolVal); !ok {
                t.Error("element[2] should be BoolVal")
        }
        if _, ok := list.Elements[3].(*NoneVal); !ok {
                t.Error("element[3] should be NoneVal")
        }
}

// === std.math Enhanced Tests ===

func TestStdMathFloorCeilRound(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.math import floor, ceil, round

let f = floor(3.7)
let c = ceil(3.2)
let r = round(3.5)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        f, ok := interp.env.Get("f")
        if !ok || f.(*IntVal).Val != 3 {
                t.Errorf("floor(3.7) = %v, expected 3", f)
        }
        c, ok := interp.env.Get("c")
        if !ok || c.(*IntVal).Val != 4 {
                t.Errorf("ceil(3.2) = %v, expected 4", c)
        }
        r, ok := interp.env.Get("r")
        if !ok || r.(*IntVal).Val != 4 {
                t.Errorf("round(3.5) = %v, expected 4", r)
        }
}

func TestStdMathSqrt(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.math import sqrt

let result = sqrt(16)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        f, ok := result.(*FloatVal)
        if !ok {
                t.Fatalf("expected FloatVal, got %T", result)
        }
        if f.Val != 4.0 {
                t.Errorf("sqrt(16) = %f, expected 4.0", f.Val)
        }
}

func TestStdMathPow(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.math import pow

let result = pow(2, 10)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        f, ok := result.(*FloatVal)
        if !ok {
                t.Fatalf("expected FloatVal, got %T", result)
        }
        if f.Val != 1024.0 {
                t.Errorf("pow(2, 10) = %f, expected 1024.0", f.Val)
        }
}

// === Module Resolver Tests ===

func TestResolverGetDependencies(t *testing.T) {
        source := `import alpha
import beta
from gamma import x
`
        l := lexer.New(source, "test.aura")
        tokens, _ := l.Tokenize()
        p := parser.New(tokens, "test.aura")
        mod, _ := p.Parse()

        deps := module.GetDependencies(mod)
        if len(deps) != 3 {
                t.Fatalf("expected 3 dependencies, got %d", len(deps))
        }
        if deps[0] != "alpha" || deps[1] != "beta" || deps[2] != "gamma" {
                t.Errorf("deps = %v, expected [alpha, beta, gamma]", deps)
        }
}

func TestResolverInitState(t *testing.T) {
        resolver := module.NewResolver(t.TempDir())
        if resolver.GetInitState("test.aura") != module.InitNone {
                t.Error("initial state should be InitNone")
        }
        resolver.SetInitState("test.aura", module.InitInProgress)
        if resolver.GetInitState("test.aura") != module.InitInProgress {
                t.Error("state should be InitInProgress")
        }
        resolver.SetInitState("test.aura", module.InitComplete)
        if resolver.GetInitState("test.aura") != module.InitComplete {
                t.Error("state should be InitComplete")
        }
}

// === std.string Enhanced Tests ===

func TestStdStringSplit(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.string import split

let result = split("a,b,c", ",")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        list, ok := result.(*ListVal)
        if !ok {
                t.Fatalf("expected ListVal, got %T", result)
        }
        if len(list.Elements) != 3 {
                t.Fatalf("expected 3 elements, got %d", len(list.Elements))
        }
}

func TestStdStringReplace(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.string import replace

let result = replace("hello world", "world", "aura")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if s, ok := result.(*StringVal); !ok || s.Val != "hello aura" {
                t.Errorf("replace = %v, expected 'hello aura'", result)
        }
}

func TestStdStringRepeat(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.string import repeat

let result = repeat("ab", 3)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if s, ok := result.(*StringVal); !ok || s.Val != "ababab" {
                t.Errorf("repeat('ab', 3) = %v, expected 'ababab'", result)
        }
}

// === std.io Enhanced Tests ===

func TestStdIoFormat(t *testing.T) {
        // Test io.format directly to avoid Aura {} interpolation conflict
        exports := createStdIoExports()
        formatFn := exports["format"].(*BuiltinFnVal)
        result := formatFn.Fn([]Value{
                &StringVal{Val: "hello {} and {}"},
                &StringVal{Val: "world"},
                &StringVal{Val: "universe"},
        })
        if s, ok := result.(*StringVal); !ok || s.Val != "hello world and universe" {
                t.Errorf("format = %v, expected 'hello world and universe'", result)
        }
}

// === Combined Feature Tests ===

func TestImportStdTestingAndJsonTogether(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `import std.testing
import std.json

let parsed = json.parse("[1, 2, 3]")
let check = testing.assert_eq(3, 3)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        parsed, ok := interp.env.Get("parsed")
        if !ok {
                t.Fatal("parsed not found")
        }
        if _, ok := parsed.(*ListVal); !ok {
                t.Errorf("expected ListVal, got %T", parsed)
        }

        check, ok := interp.env.Get("check")
        if !ok || !check.(*BoolVal).Val {
                t.Error("assert_eq(3, 3) should pass")
        }
}

func TestModuleWithStdLibDependency(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "mymath.aura", `
import std.math

pub fn circle_area(r):
    return math.pi * r * r
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import mymath

let area = mymath.circle_area(1)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        area, ok := interp.env.Get("area")
        if !ok {
                t.Fatal("area not found")
        }
        f, ok := area.(*FloatVal)
        if !ok {
                t.Fatalf("expected FloatVal, got %T", area)
        }
        if f.Val < 3.14 || f.Val > 3.15 {
                t.Errorf("circle_area(1) = %f, expected ~3.14159", f.Val)
        }
}

func TestJsonStringifyPretty(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import stringify

let result = stringify([1, 2], true)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        s, ok := result.(*StringVal)
        if !ok {
                t.Fatalf("expected StringVal, got %T", result)
        }
        if !strings.Contains(s.Val, "\n") {
                t.Error("pretty stringify should contain newlines")
        }
}

func TestWildcardImportFromStdTesting(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.testing import *

let r = assert_eq(1, 1)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        r, ok := interp.env.Get("r")
        if !ok || !r.(*BoolVal).Val {
                t.Error("wildcard import: assert_eq(1, 1) should pass")
        }
}

func TestJsonParseWhitespace(t *testing.T) {
        exports := createStdJsonExports()
        parseFn := exports["parse"].(*BuiltinFnVal)
        result := parseFn.Fn([]Value{&StringVal{Val: `  { "key" :  "value" }  `}})
        m, ok := result.(*MapVal)
        if !ok {
                t.Fatalf("expected MapVal, got %T", result)
        }
        if len(m.Keys) != 1 {
                t.Errorf("expected 1 key, got %d", len(m.Keys))
        }
        if m.Keys[0].(*StringVal).Val != "key" {
                t.Errorf("key = %v, expected 'key'", m.Keys[0])
        }
        if m.Values[0].(*StringVal).Val != "value" {
                t.Errorf("value = %v, expected 'value'", m.Values[0])
        }
}

func TestJsonStringifyFloat(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import stringify

let result = stringify(3.14)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        s, ok := result.(*StringVal)
        if !ok {
                t.Fatalf("expected StringVal, got %T", result)
        }
        if !strings.Contains(s.Val, "3.14") {
                t.Errorf("stringify(3.14) = %q, expected to contain '3.14'", s.Val)
        }
}

func TestJsonStringifyEmptyList(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `from std.json import stringify

let result = stringify([])
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if s, ok := result.(*StringVal); !ok || s.Val != "[]" {
                t.Errorf("stringify([]) = %v, expected '[]'", result)
        }
}
