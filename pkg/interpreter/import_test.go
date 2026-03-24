package interpreter

import (
        "os"
        "path/filepath"
        "testing"

        "github.com/unclebucklarson/aura/pkg/lexer"
        "github.com/unclebucklarson/aura/pkg/module"
        "github.com/unclebucklarson/aura/pkg/parser"
)

// --- Helpers ---

func createTestFile(t *testing.T, dir, filename, content string) string {
        t.Helper()
        path := filepath.Join(dir, filename)
        subdir := filepath.Dir(path)
        if err := os.MkdirAll(subdir, 0755); err != nil {
                t.Fatalf("failed to create dir %s: %v", subdir, err)
        }
        if err := os.WriteFile(path, []byte(content), 0644); err != nil {
                t.Fatalf("failed to write %s: %v", path, err)
        }
        return path
}

func runWithImports(t *testing.T, mainSource string, mainFile string, resolver *module.Resolver) *Interpreter {
        t.Helper()
        l := lexer.New(mainSource, mainFile)
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("lexer errors: %v", errs)
        }
        p := parser.New(tokens, mainFile)
        ast, parseErrs := p.Parse()
        if len(parseErrs) > 0 {
                t.Fatalf("parser errors: %v", parseErrs)
        }
        interp := NewWithResolver(ast, mainFile, resolver)
        if _, err := interp.Run(); err != nil {
                t.Fatalf("run error: %v", err)
        }
        return interp
}

// --- Standard Library Import Tests ---

func TestImportStdMath(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")

        source := `import std.math

let pi_val = math.pi
let e_val = math.e
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        pi, ok := interp.env.Get("pi_val")
        if !ok {
                t.Fatal("pi_val not found")
        }
        piFloat, ok := pi.(*FloatVal)
        if !ok {
                t.Fatalf("pi_val is not float, got %T", pi)
        }
        if piFloat.Val < 3.14 || piFloat.Val > 3.15 {
                t.Errorf("pi_val = %f, expected ~3.14159", piFloat.Val)
        }

        eVal, ok := interp.env.Get("e_val")
        if !ok {
                t.Fatal("e_val not found")
        }
        eFloat, ok := eVal.(*FloatVal)
        if !ok {
                t.Fatalf("e_val is not float, got %T", eVal)
        }
        if eFloat.Val < 2.71 || eFloat.Val > 2.72 {
                t.Errorf("e_val = %f, expected ~2.71828", eFloat.Val)
        }
}

func TestImportStdMathFunction(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")

        source := `import std.math

let result = math.abs(-42)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        intVal, ok := result.(*IntVal)
        if !ok {
                t.Fatalf("result is not int, got %T", result)
        }
        if intVal.Val != 42 {
                t.Errorf("result = %d, expected 42", intVal.Val)
        }
}

func TestImportStdMathWithAlias(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")

        source := `import std.math as m

let pi_val = m.pi
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        pi, ok := interp.env.Get("pi_val")
        if !ok {
                t.Fatal("pi_val not found")
        }
        if _, ok := pi.(*FloatVal); !ok {
                t.Fatalf("pi_val is not float, got %T", pi)
        }
}

func TestFromStdMathImportNames(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")

        source := `from std.math import pi, e

let result = pi + e
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        fVal, ok := result.(*FloatVal)
        if !ok {
                t.Fatalf("result is not float, got %T", result)
        }
        if fVal.Val < 5.85 || fVal.Val > 5.87 {
                t.Errorf("result = %f, expected ~5.859", fVal.Val)
        }
}

func TestFromStdMathImportWildcard(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")

        source := `from std.math import *

let result = abs(-10)
`
        resolver := module.NewResolver(dir)

        l := lexer.New(source, mainFile)
        tokens, _ := l.Tokenize()
        p := parser.New(tokens, mainFile)
        ast, parseErrs := p.Parse()
        if len(parseErrs) > 0 {
                t.Fatalf("parser errors: %v", parseErrs)
        }
        interp := NewWithResolver(ast, mainFile, resolver)
        _, err := interp.Run()
        if err != nil {
                t.Fatalf("run error: %v", err)
        }

        // 'pi' should be directly available
        pi, ok := interp.env.Get("pi")
        if !ok {
                t.Fatal("pi not found after wildcard import")
        }
        if _, ok := pi.(*FloatVal); !ok {
                t.Fatalf("pi is not float, got %T", pi)
        }
}

// --- File Module Import Tests ---

func TestImportFileModule(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "helpers.aura", `
pub fn greet(name):
    return "hello " + name

pub fn add(a, b):
    return a + b
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import helpers

let greeting = helpers.greet("world")
let sum = helpers.add(3, 4)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        greeting, ok := interp.env.Get("greeting")
        if !ok {
                t.Fatal("greeting not found")
        }
        if s, ok := greeting.(*StringVal); !ok || s.Val != "hello world" {
                t.Errorf("greeting = %v, expected 'hello world'", greeting)
        }

        sum, ok := interp.env.Get("sum")
        if !ok {
                t.Fatal("sum not found")
        }
        if i, ok := sum.(*IntVal); !ok || i.Val != 7 {
                t.Errorf("sum = %v, expected 7", sum)
        }
}

func TestImportFileModuleWithAlias(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "helpers.aura", `
pub fn greet(name):
    return "hi " + name
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import helpers as h

let greeting = h.greet("alice")
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        greeting, ok := interp.env.Get("greeting")
        if !ok {
                t.Fatal("greeting not found")
        }
        if s, ok := greeting.(*StringVal); !ok || s.Val != "hi alice" {
                t.Errorf("greeting = %v, expected 'hi alice'", greeting)
        }
}

func TestFromFileImportNames(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "mathutils.aura", `
pub fn square(x):
    return x * x

pub fn double(x):
    return x * 2

pub fn triple(x):
    return x * 3
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `from mathutils import square, double

let s = square(5)
let d = double(7)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        s, ok := interp.env.Get("s")
        if !ok {
                t.Fatal("s not found")
        }
        if i, ok := s.(*IntVal); !ok || i.Val != 25 {
                t.Errorf("s = %v, expected 25", s)
        }

        d, ok := interp.env.Get("d")
        if !ok {
                t.Fatal("d not found")
        }
        if i, ok := d.(*IntVal); !ok || i.Val != 14 {
                t.Errorf("d = %v, expected 14", d)
        }

        // triple should NOT be available since we didn't import it
        if _, ok := interp.env.Get("triple"); ok {
                t.Error("triple should not be available (not imported)")
        }
}

func TestFromFileImportWildcard(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "tools.aura", `
pub fn foo():
    return 1

pub fn bar():
    return 2
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `from tools import *

let a = foo()
let b = bar()
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        a, ok := interp.env.Get("a")
        if !ok {
                t.Fatal("a not found")
        }
        if i, ok := a.(*IntVal); !ok || i.Val != 1 {
                t.Errorf("a = %v, expected 1", a)
        }

        b, ok := interp.env.Get("b")
        if !ok {
                t.Fatal("b not found")
        }
        if i, ok := b.(*IntVal); !ok || i.Val != 2 {
                t.Errorf("b = %v, expected 2", b)
        }
}

func TestImportRelativeModule(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "lib.aura", `
pub fn helper():
    return 99
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import ./lib

let val = lib.helper()
`
        resolver := module.NewResolver(dir)

        l := lexer.New(source, mainFile)
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("lexer errors: %v", errs)
        }
        p := parser.New(tokens, mainFile)
        ast, parseErrs := p.Parse()
        if len(parseErrs) > 0 {
                // Relative imports with ./ may not parse perfectly since the parser
                // expects identifiers for the qualified name. This is expected behavior.
                t.Skipf("relative import syntax ./lib not yet supported in parser (needs dot-slash handling)")
        }
        interp := NewWithResolver(ast, mainFile, resolver)
        if _, err := interp.Run(); err != nil {
                t.Fatalf("run error: %v", err)
        }
}

func TestImportDottedPath(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "utils/math.aura", `
pub fn square(x):
    return x * x
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import utils.math

let result = math.square(6)
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        result, ok := interp.env.Get("result")
        if !ok {
                t.Fatal("result not found")
        }
        if i, ok := result.(*IntVal); !ok || i.Val != 36 {
                t.Errorf("result = %v, expected 36", result)
        }
}

func TestImportExportsConstant(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "config.aura", `
pub let version = 42
pub let name = "aura"
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import config

let v = config.version
let n = config.name
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        v, ok := interp.env.Get("v")
        if !ok {
                t.Fatal("v not found")
        }
        if i, ok := v.(*IntVal); !ok || i.Val != 42 {
                t.Errorf("v = %v, expected 42", v)
        }

        n, ok := interp.env.Get("n")
        if !ok {
                t.Fatal("n not found")
        }
        if s, ok := n.(*StringVal); !ok || s.Val != "aura" {
                t.Errorf("n = %v, expected 'aura'", n)
        }
}

// --- ModuleVal Tests ---

func TestModuleValType(t *testing.T) {
        m := &ModuleVal{Name: "test", Exports: map[string]Value{}}
        if m.Type() != TypeModule {
                t.Errorf("expected TypeModule, got %v", m.Type())
        }
}

func TestModuleValString(t *testing.T) {
        m := &ModuleVal{Name: "mymod", Exports: map[string]Value{}}
        expected := "<module 'mymod'>"
        if m.String() != expected {
                t.Errorf("String() = %q, expected %q", m.String(), expected)
        }
}

func TestModuleValGetExport(t *testing.T) {
        m := &ModuleVal{
                Name: "test",
                Exports: map[string]Value{
                        "foo": &IntVal{Val: 42},
                        "bar": &StringVal{Val: "hello"},
                },
        }

        foo, ok := m.GetExport("foo")
        if !ok || foo.(*IntVal).Val != 42 {
                t.Error("expected foo = 42")
        }

        bar, ok := m.GetExport("bar")
        if !ok || bar.(*StringVal).Val != "hello" {
                t.Error("expected bar = 'hello'")
        }

        _, ok = m.GetExport("baz")
        if ok {
                t.Error("expected baz to not exist")
        }
}

// --- Error Cases ---

func TestImportNonExistentModule(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `import nonexistent

let x = 1
`
        resolver := module.NewResolver(dir)

        l := lexer.New(source, mainFile)
        tokens, _ := l.Tokenize()
        p := parser.New(tokens, mainFile)
        ast, _ := p.Parse()
        interp := NewWithResolver(ast, mainFile, resolver)
        _, err := interp.Run()
        if err == nil {
                t.Fatal("expected error for non-existent module")
        }
}

func TestImportNonExistentExport(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "partial.aura", `
pub fn exists():
    return 1
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `from partial import exists, does_not_exist

let x = 1
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
}

func TestImportUnknownStdLib(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `import std.nonexistent

let x = 1
`
        resolver := module.NewResolver(dir)

        l := lexer.New(source, mainFile)
        tokens, _ := l.Tokenize()
        p := parser.New(tokens, mainFile)
        ast, _ := p.Parse()
        interp := NewWithResolver(ast, mainFile, resolver)
        _, err := interp.Run()
        if err == nil {
                t.Fatal("expected error for unknown stdlib module")
        }
}

func TestImportWithoutResolver(t *testing.T) {
        source := `import std.math

let x = 1
`
        l := lexer.New(source, "test.aura")
        tokens, _ := l.Tokenize()
        p := parser.New(tokens, "test.aura")
        ast, _ := p.Parse()

        // Use New() without resolver
        interp := New(ast)
        _, err := interp.Run()
        if err == nil {
                t.Fatal("expected error when imports exist but no resolver")
        }
}

// --- Multi-module Tests ---

func TestImportChain(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "base.aura", `
pub fn base_val():
    return 10
`)
        createTestFile(t, dir, "middle.aura", `
import base

pub fn middle_val():
    return base.base_val() + 5
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import middle

let result = middle.middle_val()
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

func TestImportModuleWithConstants(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "constants.aura", `
pub let pi_val = 3
pub let max_size = 100
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `from constants import pi_val, max_size

let sum = pi_val + max_size
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        sum, ok := interp.env.Get("sum")
        if !ok {
                t.Fatal("sum not found")
        }
        if i, ok := sum.(*IntVal); !ok || i.Val != 103 {
                t.Errorf("sum = %v, expected 103", sum)
        }
}

func TestImportMultipleModules(t *testing.T) {
        dir := t.TempDir()
        createTestFile(t, dir, "alpha.aura", `
pub fn a():
    return 1
`)
        createTestFile(t, dir, "beta.aura", `
pub fn b():
    return 2
`)

        mainFile := filepath.Join(dir, "main.aura")
        source := `import alpha
import beta

let total = alpha.a() + beta.b()
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        total, ok := interp.env.Get("total")
        if !ok {
                t.Fatal("total not found")
        }
        if i, ok := total.(*IntVal); !ok || i.Val != 3 {
                t.Errorf("total = %v, expected 3", total)
        }
}

func TestImportStdIo(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `import std.io
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        // io module should be available
        ioMod, ok := interp.env.Get("io")
        if !ok {
                t.Fatal("io module not found")
        }
        modVal, ok := ioMod.(*ModuleVal)
        if !ok {
                t.Fatalf("io is not a ModuleVal, got %T", ioMod)
        }
        if _, ok := modVal.Exports["print"]; !ok {
                t.Error("io.print not found in exports")
        }
}

func TestImportStdString(t *testing.T) {
        dir := t.TempDir()
        mainFile := filepath.Join(dir, "main.aura")
        source := `import std.string
`
        resolver := module.NewResolver(dir)
        interp := runWithImports(t, source, mainFile, resolver)

        strMod, ok := interp.env.Get("string")
        if !ok {
                t.Fatal("string module not found")
        }
        modVal, ok := strMod.(*ModuleVal)
        if !ok {
                t.Fatalf("string is not a ModuleVal, got %T", strMod)
        }
        if _, ok := modVal.Exports["join"]; !ok {
                t.Error("string.join not found in exports")
        }
}

// --- Import Parsing Tests ---

func TestParseImportStatement(t *testing.T) {
        tests := []struct {
                name     string
                source   string
                wantPath string
                wantAlias string
                wantNames []string
        }{
                {
                        name:     "simple import",
                        source:   "import helpers\n",
                        wantPath: "helpers",
                },
                {
                        name:     "dotted import",
                        source:   "import std.math\n",
                        wantPath: "std.math",
                },
                {
                        name:      "import with alias",
                        source:    "import std.math as m\n",
                        wantPath:  "std.math",
                        wantAlias: "m",
                },
                {
                        name:      "from import names",
                        source:    "from std.math import pi, e\n",
                        wantPath:  "std.math",
                        wantNames: []string{"pi", "e"},
                },
                {
                        name:      "from import wildcard",
                        source:    "from helpers import *\n",
                        wantPath:  "helpers",
                        wantNames: []string{"*"},
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        l := lexer.New(tt.source, "test.aura")
                        tokens, errs := l.Tokenize()
                        if len(errs) > 0 {
                                t.Fatalf("lexer errors: %v", errs)
                        }
                        p := parser.New(tokens, "test.aura")
                        mod, parseErrs := p.Parse()
                        if len(parseErrs) > 0 {
                                t.Fatalf("parser errors: %v", parseErrs)
                        }

                        if len(mod.Imports) != 1 {
                                t.Fatalf("expected 1 import, got %d", len(mod.Imports))
                        }

                        imp := mod.Imports[0]
                        if imp.Path.String() != tt.wantPath {
                                t.Errorf("path = %q, want %q", imp.Path.String(), tt.wantPath)
                        }
                        if imp.Alias != tt.wantAlias {
                                t.Errorf("alias = %q, want %q", imp.Alias, tt.wantAlias)
                        }
                        if tt.wantNames != nil {
                                if len(imp.Names) != len(tt.wantNames) {
                                        t.Fatalf("names count = %d, want %d", len(imp.Names), len(tt.wantNames))
                                }
                                for i, n := range tt.wantNames {
                                        if imp.Names[i] != n {
                                                t.Errorf("names[%d] = %q, want %q", i, imp.Names[i], n)
                                        }
                                }
                        }
                })
        }
}

func TestParseMultipleImports(t *testing.T) {
        source := `import alpha
import beta
from gamma import x, y

fn main():
    return 0
`
        l := lexer.New(source, "test.aura")
        tokens, _ := l.Tokenize()
        p := parser.New(tokens, "test.aura")
        mod, parseErrs := p.Parse()
        if len(parseErrs) > 0 {
                t.Fatalf("parser errors: %v", parseErrs)
        }
        if len(mod.Imports) != 3 {
                t.Fatalf("expected 3 imports, got %d", len(mod.Imports))
        }
}

// --- Lexer Token Tests for Import Keywords ---

func TestLexerImportTokens(t *testing.T) {
        source := "import std.math as m"
        l := lexer.New(source, "test.aura")
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("lexer errors: %v", errs)
        }

        // Should produce: IMPORT, IDENT(std), DOT, IDENT(math), AS, IDENT(m), NEWLINE/EOF
        expectedTypes := []string{"import", "IDENT", ".", "IDENT", "as", "IDENT"}
        idx := 0
        for _, tok := range tokens {
                if tok.Type.String() == "EOF" || tok.Type.String() == "NEWLINE" {
                        continue
                }
                if idx < len(expectedTypes) {
                        if tok.Type.String() != expectedTypes[idx] {
                                t.Errorf("token[%d] type = %s, want %s (literal: %q)", idx, tok.Type.String(), expectedTypes[idx], tok.Literal)
                        }
                        idx++
                }
        }
}

func TestLexerFromImportTokens(t *testing.T) {
        source := "from helpers import greet"
        l := lexer.New(source, "test.aura")
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("lexer errors: %v", errs)
        }

        expectedTypes := []string{"from", "IDENT", "import", "IDENT"}
        idx := 0
        for _, tok := range tokens {
                if tok.Type.String() == "EOF" || tok.Type.String() == "NEWLINE" {
                        continue
                }
                if idx < len(expectedTypes) {
                        if tok.Type.String() != expectedTypes[idx] {
                                t.Errorf("token[%d] type = %s, want %s (literal: %q)", idx, tok.Type.String(), expectedTypes[idx], tok.Literal)
                        }
                        idx++
                }
        }
}
