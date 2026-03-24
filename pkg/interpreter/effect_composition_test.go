package interpreter

import (
        "testing"

        "github.com/unclebucklarson/aura/pkg/ast"
)

// =============================================================================
// Effect Context Composition Tests
// =============================================================================

func TestEffectContext_Clone(t *testing.T) {
        ctx := NewMockEffectContext()
        ctx.File().(*MockFileProvider).AddFile("/test.txt", "hello")

        cloned := ctx.Clone()

        // Cloned context should share same providers
        if cloned.File() != ctx.File() {
                t.Error("Clone should share the same file provider reference")
        }
        if cloned.Time() != ctx.Time() {
                t.Error("Clone should share the same time provider reference")
        }
        if cloned.Env() != ctx.Env() {
                t.Error("Clone should share the same env provider reference")
        }
}

func TestEffectContext_Clone_Independence(t *testing.T) {
        ctx := NewMockEffectContext()
        cloned := ctx.Clone()

        // Replacing a provider on clone doesn't affect original
        newFile := NewMockFileProvider()
        cloned2 := cloned.WithFile(newFile)

        if cloned2.File() == ctx.File() {
                t.Error("Derived context should have different file provider")
        }
        if ctx.File() == newFile {
                t.Error("Original context should not be affected by clone modifications")
        }
}

func TestEffectContext_Derive_AllNil(t *testing.T) {
        ctx := NewMockEffectContext()
        derived := ctx.Derive(nil, nil, nil)

        if derived.File() != ctx.File() {
                t.Error("Derive with nil should keep file provider")
        }
        if derived.Time() != ctx.Time() {
                t.Error("Derive with nil should keep time provider")
        }
        if derived.Env() != ctx.Env() {
                t.Error("Derive with nil should keep env provider")
        }
}

func TestEffectContext_Derive_OverrideFile(t *testing.T) {
        ctx := NewMockEffectContext()
        newFile := NewMockFileProvider()
        newFile.AddFile("/new.txt", "new content")

        derived := ctx.Derive(newFile, nil, nil)

        if derived.File() != newFile {
                t.Error("Derive should use new file provider")
        }
        if derived.Time() != ctx.Time() {
                t.Error("Derive should keep original time provider")
        }
        if !derived.File().Exists("/new.txt") {
                t.Error("Derived context should see new file")
        }
}

func TestEffectContext_Derive_OverrideTime(t *testing.T) {
        ctx := NewMockEffectContext()
        newTime := NewMockTimeProvider()
        newTime.SetTime(999999)

        derived := ctx.Derive(nil, newTime, nil)

        if derived.Time() != newTime {
                t.Error("Derive should use new time provider")
        }
        if derived.Time().Now() != 999999 {
                t.Errorf("Derived time should be 999999, got %d", derived.Time().Now())
        }
}

func TestEffectContext_Derive_OverrideEnv(t *testing.T) {
        ctx := NewMockEffectContext()
        newEnv := NewMockEnvProvider()
        newEnv.SetVar("KEY", "VALUE")

        derived := ctx.Derive(nil, nil, newEnv)

        if derived.Env() != newEnv {
                t.Error("Derive should use new env provider")
        }
        val, ok := derived.Env().Get("KEY")
        if !ok || val != "VALUE" {
                t.Error("Derived env should have KEY=VALUE")
        }
}

func TestEffectContext_Derive_OverrideAll(t *testing.T) {
        ctx := NewMockEffectContext()
        newFile := NewMockFileProvider()
        newTime := NewMockTimeProvider()
        newEnv := NewMockEnvProvider()

        derived := ctx.Derive(newFile, newTime, newEnv)

        if derived.File() != newFile {
                t.Error("should override file")
        }
        if derived.Time() != newTime {
                t.Error("should override time")
        }
        if derived.Env() != newEnv {
                t.Error("should override env")
        }
}

// =============================================================================
// Effect Stack Tests
// =============================================================================

func TestEffectStack_New(t *testing.T) {
        ctx := NewMockEffectContext()
        stack := NewEffectStack(ctx)

        if stack.Depth() != 1 {
                t.Errorf("new stack should have depth 1, got %d", stack.Depth())
        }
        if stack.Current() != ctx {
                t.Error("current should be the initial context")
        }
}

func TestEffectStack_PushPop(t *testing.T) {
        base := NewMockEffectContext()
        stack := NewEffectStack(base)

        child := NewMockEffectContext()
        child.File().(*MockFileProvider).AddFile("/child.txt", "child")

        stack.Push(child)

        if stack.Depth() != 2 {
                t.Errorf("after push, depth should be 2, got %d", stack.Depth())
        }
        if stack.Current() != child {
                t.Error("current should be the child context after push")
        }

        popped := stack.Pop()
        if popped != child {
                t.Error("popped context should be the child")
        }
        if stack.Depth() != 1 {
                t.Errorf("after pop, depth should be 1, got %d", stack.Depth())
        }
        if stack.Current() != base {
                t.Error("after pop, current should be the base context")
        }
}

func TestEffectStack_PopProtectsBase(t *testing.T) {
        base := NewMockEffectContext()
        stack := NewEffectStack(base)

        result := stack.Pop()
        if result != nil {
                t.Error("popping the base context should return nil")
        }
        if stack.Depth() != 1 {
                t.Errorf("base context should remain, depth=%d", stack.Depth())
        }
}

func TestEffectStack_NestedPushPop(t *testing.T) {
        base := NewMockEffectContext()
        stack := NewEffectStack(base)

        ctx1 := NewMockEffectContext()
        ctx2 := NewMockEffectContext()
        ctx3 := NewMockEffectContext()

        stack.Push(ctx1)
        stack.Push(ctx2)
        stack.Push(ctx3)

        if stack.Depth() != 4 {
                t.Errorf("expected depth 4, got %d", stack.Depth())
        }

        if stack.Current() != ctx3 {
                t.Error("current should be ctx3")
        }

        stack.Pop()
        if stack.Current() != ctx2 {
                t.Error("current should be ctx2 after popping ctx3")
        }

        stack.Pop()
        if stack.Current() != ctx1 {
                t.Error("current should be ctx1 after popping ctx2")
        }

        stack.Pop()
        if stack.Current() != base {
                t.Error("current should be base after popping ctx1")
        }
}

// =============================================================================
// Mock Builder Tests
// =============================================================================

func TestMockBuilder_Basic(t *testing.T) {
        ctx := NewMockBuilder().Build()

        if ctx == nil {
                t.Fatal("build should return a non-nil context")
        }
        if ctx.File() == nil || ctx.Time() == nil || ctx.Env() == nil {
                t.Error("all providers should be initialized")
        }
}

func TestMockBuilder_WithFile(t *testing.T) {
        ctx := NewMockBuilder().
                WithFile("/test.txt", "hello world").
                Build()

        content, err := ctx.File().ReadFile("/test.txt")
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        if content != "hello world" {
                t.Errorf("expected 'hello world', got %q", content)
        }
}

func TestMockBuilder_WithDir(t *testing.T) {
        ctx := NewMockBuilder().
                WithDir("/mydir").
                Build()

        if !ctx.File().IsDir("/mydir") {
                t.Error("directory should exist")
        }
}

func TestMockBuilder_WithTime(t *testing.T) {
        ctx := NewMockBuilder().
                WithTime(1234567890).
                Build()

        if ctx.Time().Now() != 1234567890 {
                t.Errorf("expected time 1234567890, got %d", ctx.Time().Now())
        }
}

func TestMockBuilder_WithEnvVar(t *testing.T) {
        ctx := NewMockBuilder().
                WithEnvVar("HOME", "/home/user").
                Build()

        val, ok := ctx.Env().Get("HOME")
        if !ok || val != "/home/user" {
                t.Errorf("expected HOME=/home/user, got %q, %v", val, ok)
        }
}

func TestMockBuilder_WithCwd(t *testing.T) {
        ctx := NewMockBuilder().
                WithCwd("/app").
                Build()

        cwd, _ := ctx.Env().Cwd()
        if cwd != "/app" {
                t.Errorf("expected cwd /app, got %q", cwd)
        }
}

func TestMockBuilder_WithArgs(t *testing.T) {
        ctx := NewMockBuilder().
                WithArgs([]string{"aura", "run", "main.aura"}).
                Build()

        args := ctx.Env().Args()
        if len(args) != 3 || args[0] != "aura" || args[2] != "main.aura" {
                t.Errorf("unexpected args: %v", args)
        }
}

func TestMockBuilder_WithFiles(t *testing.T) {
        ctx := NewMockBuilder().
                WithFiles(map[string]string{
                        "/a.txt": "content a",
                        "/b.txt": "content b",
                }).
                Build()

        a, _ := ctx.File().ReadFile("/a.txt")
        b, _ := ctx.File().ReadFile("/b.txt")

        if a != "content a" || b != "content b" {
                t.Errorf("unexpected file contents: %q, %q", a, b)
        }
}

func TestMockBuilder_WithEnvVars(t *testing.T) {
        ctx := NewMockBuilder().
                WithEnvVars(map[string]string{
                        "A": "1",
                        "B": "2",
                }).
                Build()

        a, _ := ctx.Env().Get("A")
        b, _ := ctx.Env().Get("B")

        if a != "1" || b != "2" {
                t.Errorf("unexpected env vars: A=%q, B=%q", a, b)
        }
}

func TestMockBuilder_FluentChaining(t *testing.T) {
        ctx := NewMockBuilder().
                WithFile("/config.json", `{"key":"val"}`).
                WithTime(1700000000).
                WithEnvVar("MODE", "test").
                WithCwd("/project").
                WithArgs([]string{"aura", "--test"}).
                Build()

        content, _ := ctx.File().ReadFile("/config.json")
        if content != `{"key":"val"}` {
                t.Error("file content mismatch")
        }
        if ctx.Time().Now() != 1700000000 {
                t.Error("time mismatch")
        }
        val, _ := ctx.Env().Get("MODE")
        if val != "test" {
                t.Error("env var mismatch")
        }
        cwd, _ := ctx.Env().Cwd()
        if cwd != "/project" {
                t.Error("cwd mismatch")
        }
        args := ctx.Env().Args()
        if len(args) != 2 {
                t.Error("args mismatch")
        }
}

func TestMockBuilder_WithFileProvider(t *testing.T) {
        custom := NewMockFileProvider()
        custom.AddFile("/custom.txt", "custom")

        ctx := NewMockBuilder().
                WithFileProvider(custom).
                Build()

        content, _ := ctx.File().ReadFile("/custom.txt")
        if content != "custom" {
                t.Errorf("expected 'custom', got %q", content)
        }
}

func TestMockBuilder_WithTimeProvider(t *testing.T) {
        custom := NewMockTimeProvider()
        custom.SetTime(42)

        ctx := NewMockBuilder().
                WithTimeProvider(custom).
                Build()

        if ctx.Time().Now() != 42 {
                t.Errorf("expected 42, got %d", ctx.Time().Now())
        }
}

func TestMockBuilder_WithEnvProvider(t *testing.T) {
        custom := NewMockEnvProvider()
        custom.SetVar("CUSTOM", "yes")

        ctx := NewMockBuilder().
                WithEnvProvider(custom).
                Build()

        val, _ := ctx.Env().Get("CUSTOM")
        if val != "yes" {
                t.Errorf("expected 'yes', got %q", val)
        }
}

// =============================================================================
// Fixture Tests
// =============================================================================

func TestEmptyMockContext(t *testing.T) {
        ctx := EmptyMockContext()

        if ctx.File().Exists("/anything") {
                t.Error("empty mock should have no files")
        }
        if ctx.Time().Now() != 1000000 {
                t.Errorf("empty mock time should be 1000000, got %d", ctx.Time().Now())
        }
}

func TestFixtureWithFiles(t *testing.T) {
        ctx := FixtureWithFiles(map[string]string{
                "/data.csv":   "a,b,c",
                "/config.yml": "key: val",
        })

        if !ctx.File().Exists("/data.csv") {
                t.Error("data.csv should exist")
        }
        content, _ := ctx.File().ReadFile("/config.yml")
        if content != "key: val" {
                t.Error("config content mismatch")
        }
}

func TestFixtureWithTime(t *testing.T) {
        ctx := FixtureWithTime(1609459200) // 2021-01-01

        if ctx.Time().Now() != 1609459200 {
                t.Errorf("expected 1609459200, got %d", ctx.Time().Now())
        }
}

func TestFixtureWithEnv(t *testing.T) {
        ctx := FixtureWithEnv(map[string]string{
                "DB_HOST": "localhost",
                "DB_PORT": "5432",
        })

        host, _ := ctx.Env().Get("DB_HOST")
        port, _ := ctx.Env().Get("DB_PORT")

        if host != "localhost" || port != "5432" {
                t.Errorf("env vars mismatch: host=%q, port=%q", host, port)
        }
}

func TestFixtureComplete(t *testing.T) {
        ctx := FixtureComplete(
                map[string]string{"/app.aura": "fn main() {}"},
                1700000000,
                map[string]string{"ENV": "test"},
        )

        content, _ := ctx.File().ReadFile("/app.aura")
        if content != "fn main() {}" {
                t.Error("file content mismatch")
        }
        if ctx.Time().Now() != 1700000000 {
                t.Error("time mismatch")
        }
        val, _ := ctx.Env().Get("ENV")
        if val != "test" {
                t.Error("env var mismatch")
        }
}

// =============================================================================
// Assertion Helper Tests
// =============================================================================

func TestAssertFileExists(t *testing.T) {
        ctx := NewMockBuilder().WithFile("/test.txt", "content").Build()

        if !AssertFileExists(ctx, "/test.txt") {
                t.Error("file should exist")
        }
        if AssertFileExists(ctx, "/missing.txt") {
                t.Error("file should not exist")
        }
}

func TestAssertFileContent(t *testing.T) {
        ctx := NewMockBuilder().WithFile("/test.txt", "hello").Build()

        if !AssertFileContent(ctx, "/test.txt", "hello") {
                t.Error("content should match")
        }
        if AssertFileContent(ctx, "/test.txt", "world") {
                t.Error("content should not match")
        }
        if AssertFileContent(ctx, "/missing.txt", "anything") {
                t.Error("missing file should not match")
        }
}

func TestAssertEnvVar(t *testing.T) {
        ctx := NewMockBuilder().WithEnvVar("KEY", "val").Build()

        if !AssertEnvVar(ctx, "KEY", "val") {
                t.Error("env var should match")
        }
        if AssertEnvVar(ctx, "KEY", "wrong") {
                t.Error("wrong value should not match")
        }
        if AssertEnvVar(ctx, "MISSING", "anything") {
                t.Error("missing var should not match")
        }
}

func TestAssertMockTime(t *testing.T) {
        ctx := NewMockBuilder().WithTime(12345).Build()

        if !AssertMockTime(ctx, 12345) {
                t.Error("time should match")
        }
        if AssertMockTime(ctx, 99999) {
                t.Error("wrong time should not match")
        }
}

func TestGetMockFileProvider(t *testing.T) {
        mockCtx := NewMockEffectContext()
        if GetMockFileProvider(mockCtx) == nil {
                t.Error("should return MockFileProvider from mock context")
        }

        realCtx := NewEffectContext()
        if GetMockFileProvider(realCtx) != nil {
                t.Error("should return nil for real context")
        }
}

func TestGetMockTimeProvider(t *testing.T) {
        mockCtx := NewMockEffectContext()
        if GetMockTimeProvider(mockCtx) == nil {
                t.Error("should return MockTimeProvider from mock context")
        }

        realCtx := NewEffectContext()
        if GetMockTimeProvider(realCtx) != nil {
                t.Error("should return nil for real context")
        }
}

func TestGetMockEnvProvider(t *testing.T) {
        mockCtx := NewMockEffectContext()
        if GetMockEnvProvider(mockCtx) == nil {
                t.Error("should return MockEnvProvider from mock context")
        }

        realCtx := NewEffectContext()
        if GetMockEnvProvider(realCtx) != nil {
                t.Error("should return nil for real context")
        }
}

// =============================================================================
// Testing Integration Tests (stdlib_testing effect helpers via Go)
// =============================================================================

func TestTestingEffectExports_WithMockEffects(t *testing.T) {
        ctx := NewMockEffectContext()
        interp := createTestInterpreter(ctx)

        exports := createStdTestingEffectExports(interp)
        fn := exports["with_mock_effects"].(*BuiltinFnVal)

        // Create a simple function to call
        called := false
        callback := &BuiltinFnVal{
                Name: "callback",
                Fn: func(args []Value) Value {
                        called = true
                        return &IntVal{Val: 42}
                },
        }

        result := fn.Fn([]Value{callback})
        if !called {
                t.Error("callback should have been called")
        }
        if r, ok := result.(*IntVal); !ok || r.Val != 42 {
                t.Errorf("expected 42, got %v", result)
        }
}

func TestTestingEffectExports_ResetEffects(t *testing.T) {
        ctx := NewMockBuilder().WithFile("/old.txt", "old").Build()
        interp := createTestInterpreter(ctx)

        exports := createStdTestingEffectExports(interp)
        resetFn := exports["reset_effects"].(*BuiltinFnVal)

        // Verify file exists before reset
        if !interp.effects.File().Exists("/old.txt") {
                t.Error("file should exist before reset")
        }

        resetFn.Fn(nil)

        // After reset, file should be gone (fresh mock)
        if interp.effects.File().Exists("/old.txt") {
                t.Error("file should not exist after reset")
        }
}

func TestTestingEffectExports_MockTime(t *testing.T) {
        ctx := NewMockEffectContext()
        interp := createTestInterpreter(ctx)

        exports := createStdTestingEffectExports(interp)
        mockTimeFn := exports["mock_time"].(*BuiltinFnVal)

        mockTimeFn.Fn([]Value{&IntVal{Val: 5000000}})

        if interp.effects.Time().Now() != 5000000 {
                t.Errorf("expected time 5000000, got %d", interp.effects.Time().Now())
        }
}

func TestTestingEffectExports_AdvanceTime(t *testing.T) {
        ctx := NewMockEffectContext()
        interp := createTestInterpreter(ctx)

        // Set initial time
        GetMockTimeProvider(interp.effects).SetTime(1000)

        exports := createStdTestingEffectExports(interp)
        advanceFn := exports["advance_time"].(*BuiltinFnVal)

        advanceFn.Fn([]Value{&IntVal{Val: 60}}) // advance 60 seconds

        if interp.effects.Time().Now() != 1060 {
                t.Errorf("expected time 1060, got %d", interp.effects.Time().Now())
        }
}

func TestTestingEffectExports_AssertFileExists(t *testing.T) {
        ctx := NewMockBuilder().WithFile("/exists.txt", "yes").Build()
        interp := createTestInterpreter(ctx)

        exports := createStdTestingEffectExports(interp)
        fn := exports["assert_file_exists"].(*BuiltinFnVal)

        // Should not panic for existing file
        result := fn.Fn([]Value{&StringVal{Val: "/exists.txt"}})
        if r, ok := result.(*BoolVal); !ok || !r.Val {
                t.Error("assert_file_exists should return true")
        }

        // Should panic for missing file
        defer func() {
                if r := recover(); r == nil {
                        t.Error("should panic for missing file")
                }
        }()
        fn.Fn([]Value{&StringVal{Val: "/missing.txt"}})
}

func TestTestingEffectExports_AssertFileContent(t *testing.T) {
        ctx := NewMockBuilder().WithFile("/test.txt", "hello world").Build()
        interp := createTestInterpreter(ctx)

        exports := createStdTestingEffectExports(interp)
        fn := exports["assert_file_content"].(*BuiltinFnVal)

        // Should succeed
        result := fn.Fn([]Value{&StringVal{Val: "/test.txt"}, &StringVal{Val: "hello world"}})
        if r, ok := result.(*BoolVal); !ok || !r.Val {
                t.Error("should return true for matching content")
        }

        // Should panic for mismatch
        defer func() {
                if r := recover(); r == nil {
                        t.Error("should panic for content mismatch")
                }
        }()
        fn.Fn([]Value{&StringVal{Val: "/test.txt"}, &StringVal{Val: "wrong"}})
}

func TestTestingEffectExports_AssertEnvVar(t *testing.T) {
        ctx := NewMockBuilder().WithEnvVar("APP", "aura").Build()
        interp := createTestInterpreter(ctx)

        exports := createStdTestingEffectExports(interp)
        fn := exports["assert_env_var"].(*BuiltinFnVal)

        result := fn.Fn([]Value{&StringVal{Val: "APP"}, &StringVal{Val: "aura"}})
        if r, ok := result.(*BoolVal); !ok || !r.Val {
                t.Error("should return true for matching env var")
        }

        defer func() {
                if r := recover(); r == nil {
                        t.Error("should panic for missing env var")
                }
        }()
        fn.Fn([]Value{&StringVal{Val: "MISSING"}, &StringVal{Val: "anything"}})
}

func TestTestingEffectExports_AssertFileContains(t *testing.T) {
        ctx := NewMockBuilder().WithFile("/log.txt", "ERROR: something went wrong").Build()
        interp := createTestInterpreter(ctx)

        exports := createStdTestingEffectExports(interp)
        fn := exports["assert_file_contains"].(*BuiltinFnVal)

        result := fn.Fn([]Value{&StringVal{Val: "/log.txt"}, &StringVal{Val: "ERROR"}})
        if r, ok := result.(*BoolVal); !ok || !r.Val {
                t.Error("should return true for contained substring")
        }

        defer func() {
                if r := recover(); r == nil {
                        t.Error("should panic for missing substring")
                }
        }()
        fn.Fn([]Value{&StringVal{Val: "/log.txt"}, &StringVal{Val: "SUCCESS"}})
}

func TestTestingEffectExports_AssertNoFile(t *testing.T) {
        ctx := NewMockBuilder().WithFile("/exists.txt", "yes").Build()
        interp := createTestInterpreter(ctx)

        exports := createStdTestingEffectExports(interp)
        fn := exports["assert_no_file"].(*BuiltinFnVal)

        // Should succeed for missing file
        result := fn.Fn([]Value{&StringVal{Val: "/missing.txt"}})
        if r, ok := result.(*BoolVal); !ok || !r.Val {
                t.Error("should return true for non-existing file")
        }

        defer func() {
                if r := recover(); r == nil {
                        t.Error("should panic for existing file")
                }
        }()
        fn.Fn([]Value{&StringVal{Val: "/exists.txt"}})
}

func TestTestingEffectExports_GetMockTime(t *testing.T) {
        ctx := NewMockBuilder().WithTime(9999).Build()
        interp := createTestInterpreter(ctx)

        exports := createStdTestingEffectExports(interp)
        fn := exports["get_mock_time"].(*BuiltinFnVal)

        result := fn.Fn(nil)
        if r, ok := result.(*IntVal); !ok || r.Val != 9999 {
                t.Errorf("expected 9999, got %v", result)
        }
}

func TestTestingEffectExports_GetEnv(t *testing.T) {
        ctx := NewMockBuilder().WithEnvVar("KEY", "val").Build()
        interp := createTestInterpreter(ctx)

        exports := createStdTestingEffectExports(interp)
        fn := exports["get_env"].(*BuiltinFnVal)

        result := fn.Fn([]Value{&StringVal{Val: "KEY"}})
        opt, ok := result.(*OptionVal)
        if !ok || !opt.IsSome {
                t.Error("should return Some for existing key")
        }
        if s, ok := opt.Val.(*StringVal); !ok || s.Val != "val" {
                t.Error("value mismatch")
        }

        result2 := fn.Fn([]Value{&StringVal{Val: "MISSING"}})
        opt2, ok := result2.(*OptionVal)
        if !ok || opt2.IsSome {
                t.Error("should return None for missing key")
        }
}

// =============================================================================
// Integration: Effect composition with interpreter
// =============================================================================

func TestIntegration_EffectCompositionWithInterpreter(t *testing.T) {
        // Create a mock context with specific state
        ctx := NewMockBuilder().
                WithFile("/data.txt", "important data").
                WithTime(1700000000).
                WithEnvVar("APP_MODE", "production").
                Build()

        // Create an interpreter with this context
        interp := createTestInterpreter(ctx)

        // Verify the interpreter uses the correct effects
        content, err := interp.effects.File().ReadFile("/data.txt")
        if err != nil || content != "important data" {
                t.Error("interpreter should use mock file provider")
        }
        if interp.effects.Time().Now() != 1700000000 {
                t.Error("interpreter should use mock time provider")
        }
        val, _ := interp.effects.Env().Get("APP_MODE")
        if val != "production" {
                t.Error("interpreter should use mock env provider")
        }
}

func TestIntegration_DerivedContextChain(t *testing.T) {
        // Build a chain of derived contexts
        base := NewMockBuilder().
                WithFile("/base.txt", "base").
                WithTime(1000).
                WithEnvVar("LEVEL", "0").
                Build()

        level1 := base.Derive(nil, nil, nil)
        // level1 shares base's providers, so file is visible
        if !AssertFileExists(level1, "/base.txt") {
                t.Error("level1 should see base files")
        }

        // Override time at level 2
        newTime := NewMockTimeProvider()
        newTime.SetTime(2000)
        level2 := level1.Derive(nil, newTime, nil)

        if level2.Time().Now() != 2000 {
                t.Error("level2 should have overridden time")
        }
        if !AssertFileExists(level2, "/base.txt") {
                t.Error("level2 should still see base files")
        }
}

func TestIntegration_StackWithDerivedContexts(t *testing.T) {
        baseTime := NewMockTimeProvider()
        baseTime.SetTime(100)
        base := NewMockBuilder().WithTimeProvider(baseTime).Build()
        stack := NewEffectStack(base)

        // Push a derived context with a completely new time provider
        childTime := NewMockTimeProvider()
        childTime.SetTime(200)
        derived := base.Derive(nil, childTime, nil)
        stack.Push(derived)

        if stack.Current().Time().Now() != 200 {
                t.Errorf("expected 200, got %d", stack.Current().Time().Now())
        }

        stack.Pop()
        if stack.Current().Time().Now() != 100 {
                t.Errorf("expected 100, got %d", stack.Current().Time().Now())
        }
}

// =============================================================================
// Error handling / edge case tests
// =============================================================================

func TestMockBuilder_WithFile_EmptyContent(t *testing.T) {
        ctx := NewMockBuilder().WithFile("/empty.txt", "").Build()

        content, err := ctx.File().ReadFile("/empty.txt")
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        if content != "" {
                t.Errorf("expected empty content, got %q", content)
        }
}

func TestMockBuilder_WithEnvVar_EmptyValue(t *testing.T) {
        ctx := NewMockBuilder().WithEnvVar("EMPTY", "").Build()

        val, ok := ctx.Env().Get("EMPTY")
        if !ok {
                t.Error("var should exist")
        }
        if val != "" {
                t.Errorf("expected empty value, got %q", val)
        }
}

func TestEffectContext_WithFile_Immutability(t *testing.T) {
        original := NewMockEffectContext()
        newFP := NewMockFileProvider()

        modified := original.WithFile(newFP)

        // Original should not be modified
        if original.File() == newFP {
                t.Error("original should not be modified")
        }
        if modified.File() != newFP {
                t.Error("modified should use new provider")
        }
}

func TestEffectContext_WithTime_Immutability(t *testing.T) {
        original := NewMockEffectContext()
        newTP := NewMockTimeProvider()

        modified := original.WithTime(newTP)

        if original.Time() == newTP {
                t.Error("original should not be modified")
        }
        if modified.Time() != newTP {
                t.Error("modified should use new provider")
        }
}

func TestEffectContext_WithEnv_Immutability(t *testing.T) {
        original := NewMockEffectContext()
        newEP := NewMockEnvProvider()

        modified := original.WithEnv(newEP)

        if original.Env() == newEP {
                t.Error("original should not be modified")
        }
        if modified.Env() != newEP {
                t.Error("modified should use new provider")
        }
}

// Helper function to create a simple test interpreter with mock effects
func createTestInterpreter(ctx *EffectContext) *Interpreter {
        mod := createEmptyModule()
        return NewWithEffects(mod, ctx)
}

func createEmptyModule() *ast.Module {
        return &ast.Module{
                Items: []ast.TopLevelItem{},
        }
}
