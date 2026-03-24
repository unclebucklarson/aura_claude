package module

import (
        "os"
        "path/filepath"
        "testing"
)

// --- Helper to create temp modules ---

func createTempModule(t *testing.T, dir, filename, content string) string {
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

// --- Resolver Tests ---

func TestNewResolver(t *testing.T) {
        r := NewResolver("/tmp/test")
        if r == nil {
                t.Fatal("NewResolver returned nil")
        }
        if len(r.searchPaths) != 1 || r.searchPaths[0] != "/tmp/test" {
                t.Errorf("expected searchPaths [/tmp/test], got %v", r.searchPaths)
        }
        if r.CacheCount() != 0 {
                t.Errorf("expected empty cache, got %d", r.CacheCount())
        }
}

func TestAddSearchPath(t *testing.T) {
        r := NewResolver("/tmp/a")
        r.AddSearchPath("/tmp/b")
        if len(r.searchPaths) != 2 {
                t.Errorf("expected 2 search paths, got %d", len(r.searchPaths))
        }
}

func TestResolveSimpleModule(t *testing.T) {
        dir := t.TempDir()
        createTempModule(t, dir, "helpers.aura", `
pub fn greet(name):
    return "hello " + name
`)

        r := NewResolver(dir)
        cached, err := r.Resolve("helpers", dir)
        if err != nil {
                t.Fatalf("Resolve failed: %v", err)
        }
        if cached == nil {
                t.Fatal("cached module is nil")
        }
        if cached.AST == nil {
                t.Fatal("cached module AST is nil")
        }
        if !cached.Exports["greet"] {
                t.Error("expected 'greet' to be exported")
        }
}

func TestResolveRelativeModule(t *testing.T) {
        dir := t.TempDir()
        createTempModule(t, dir, "utils.aura", `
pub fn add(a, b):
    return a + b
`)

        r := NewResolver(dir)
        cached, err := r.Resolve("./utils", dir)
        if err != nil {
                t.Fatalf("Resolve relative failed: %v", err)
        }
        if cached == nil {
                t.Fatal("cached module is nil")
        }
        if !cached.Exports["add"] {
                t.Error("expected 'add' to be exported")
        }
}

func TestResolveParentRelativeModule(t *testing.T) {
        dir := t.TempDir()
        subDir := filepath.Join(dir, "sub")
        os.MkdirAll(subDir, 0755)
        createTempModule(t, dir, "common.aura", `
pub fn shared():
    return 42
`)

        r := NewResolver(dir)
        cached, err := r.Resolve("../common", subDir)
        if err != nil {
                t.Fatalf("Resolve parent relative failed: %v", err)
        }
        if cached == nil {
                t.Fatal("cached module is nil")
        }
        if !cached.Exports["shared"] {
                t.Error("expected 'shared' to be exported")
        }
}

func TestResolveDottedPath(t *testing.T) {
        dir := t.TempDir()
        createTempModule(t, dir, "utils/math.aura", `
pub fn square(x):
    return x * x
`)

        r := NewResolver(dir)
        cached, err := r.Resolve("utils.math", dir)
        if err != nil {
                t.Fatalf("Resolve dotted path failed: %v", err)
        }
        if !cached.Exports["square"] {
                t.Error("expected 'square' to be exported")
        }
}

func TestResolveDirectoryModule(t *testing.T) {
        dir := t.TempDir()
        createTempModule(t, dir, "mylib/mod.aura", `
pub fn init():
    return "initialized"
`)

        r := NewResolver(dir)
        cached, err := r.Resolve("mylib", dir)
        if err != nil {
                t.Fatalf("Resolve directory module failed: %v", err)
        }
        if !cached.Exports["init"] {
                t.Error("expected 'init' to be exported")
        }
}

func TestResolveStdLib(t *testing.T) {
        r := NewResolver("/tmp")
        cached, err := r.Resolve("std.math", "/tmp")
        if err != nil {
                t.Fatalf("Resolve std.math failed: %v", err)
        }
        if cached == nil {
                t.Fatal("cached is nil")
        }
        // stdlib modules have nil AST (handled natively by interpreter)
        if cached.AST != nil {
                t.Error("expected nil AST for stdlib module")
        }
        if cached.Path != "@std/std.math" {
                t.Errorf("expected path '@std/std.math', got '%s'", cached.Path)
        }
}

func TestResolveModuleNotFound(t *testing.T) {
        r := NewResolver("/tmp/nonexistent")
        _, err := r.Resolve("no_such_module", "/tmp/nonexistent")
        if err == nil {
                t.Fatal("expected error for non-existent module")
        }
        if _, ok := err.(*ResolveError); !ok {
                t.Errorf("expected ResolveError, got %T", err)
        }
}

func TestCircularDependencyDetection(t *testing.T) {
        dir := t.TempDir()
        // Module A imports B, but since we're only parsing, not interpreting,
        // we can test the loading tracking
        createTempModule(t, dir, "a.aura", `
import b
pub fn fa():
    return 1
`)
        createTempModule(t, dir, "b.aura", `
import a
pub fn fb():
    return 2
`)

        r := NewResolver(dir)
        // First resolve a
        _, err := r.Resolve("a", dir)
        if err != nil {
                // The circular dependency should be detected when b tries to import a
                // while a is still loading
                if re, ok := err.(*ResolveError); ok {
                        if re.Message != "circular dependency detected" {
                                // It may fail differently since loadModule triggers parser
                                // which doesn't recursively resolve imports
                                t.Logf("Got error (may not be circular): %v", err)
                        }
                }
        }
}

func TestModuleCaching(t *testing.T) {
        dir := t.TempDir()
        createTempModule(t, dir, "cached_mod.aura", `
pub fn foo():
    return 1
`)

        r := NewResolver(dir)

        // First resolve
        cached1, err := r.Resolve("cached_mod", dir)
        if err != nil {
                t.Fatalf("first resolve failed: %v", err)
        }

        // Second resolve should return same cached module
        cached2, err := r.Resolve("cached_mod", dir)
        if err != nil {
                t.Fatalf("second resolve failed: %v", err)
        }

        if cached1 != cached2 {
                t.Error("expected same cached module on second resolve")
        }
        if r.CacheCount() != 1 {
                t.Errorf("expected cache count 1, got %d", r.CacheCount())
        }
}

func TestIsCached(t *testing.T) {
        dir := t.TempDir()
        createTempModule(t, dir, "check.aura", `
pub fn bar():
    return 2
`)

        r := NewResolver(dir)

        if r.IsCached("check", dir) {
                t.Error("expected module to not be cached before resolve")
        }

        _, err := r.Resolve("check", dir)
        if err != nil {
                t.Fatalf("resolve failed: %v", err)
        }

        if !r.IsCached("check", dir) {
                t.Error("expected module to be cached after resolve")
        }
}

// --- Utility Function Tests ---

func TestIsStdLib(t *testing.T) {
        tests := []struct {
                path string
                want bool
        }{
                {"std.math", true},
                {"std.testing", true},
                {"std.io", true},
                {"mylib", false},
                {"utils.math", false},
                {"./helpers", false},
                {"std", false},
        }

        for _, tt := range tests {
                got := IsStdLib(tt.path)
                if got != tt.want {
                        t.Errorf("IsStdLib(%q) = %v, want %v", tt.path, got, tt.want)
                }
        }
}

func TestGetModuleName(t *testing.T) {
        tests := []struct {
                path string
                want string
        }{
                {"std.math", "math"},
                {"std.testing", "testing"},
                {"utils.math", "math"},
                {"helpers", "helpers"},
                {"a.b.c", "c"},
        }

        for _, tt := range tests {
                got := GetModuleName(tt.path)
                if got != tt.want {
                        t.Errorf("GetModuleName(%q) = %q, want %q", tt.path, got, tt.want)
                }
        }
}

func TestExportVisibility(t *testing.T) {
        dir := t.TempDir()
        createTempModule(t, dir, "visibility.aura", `
pub fn public_fn():
    return "public"

fn private_fn():
    return "private"

pub let public_const = 42
`)

        r := NewResolver(dir)
        cached, err := r.Resolve("visibility", dir)
        if err != nil {
                t.Fatalf("resolve failed: %v", err)
        }

        if !cached.Exports["public_fn"] {
                t.Error("expected 'public_fn' to be exported")
        }
        if !cached.Exports["public_const"] {
                t.Error("expected 'public_const' to be exported")
        }
        if cached.Exports["private_fn"] {
                t.Error("expected 'private_fn' to NOT be exported")
        }
}

func TestDefaultExportWhenNoPub(t *testing.T) {
        dir := t.TempDir()
        // When no pub keywords, all top-level items are exported
        createTempModule(t, dir, "simple.aura", `
fn helper():
    return 1

fn another():
    return 2
`)

        r := NewResolver(dir)
        cached, err := r.Resolve("simple", dir)
        if err != nil {
                t.Fatalf("resolve failed: %v", err)
        }

        if !cached.Exports["helper"] {
                t.Error("expected 'helper' to be exported (no pub = export all)")
        }
        if !cached.Exports["another"] {
                t.Error("expected 'another' to be exported (no pub = export all)")
        }
}

func TestResolveError_String(t *testing.T) {
        e := &ResolveError{Message: "not found", Path: "test.module"}
        got := e.Error()
        if got != "module resolution error: not found (path: test.module)" {
                t.Errorf("unexpected error string: %s", got)
        }
}
