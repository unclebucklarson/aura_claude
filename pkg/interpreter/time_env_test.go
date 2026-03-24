package interpreter

import (
	"os"
	"testing"
	"time"
)

// ============================================================================
// TimeProvider Tests
// ============================================================================

func TestRealTimeProvider_Now(t *testing.T) {
	tp := &RealTimeProvider{}
	before := time.Now().Unix()
	got := tp.Now()
	after := time.Now().Unix()
	if got < before || got > after {
		t.Errorf("RealTimeProvider.Now() = %d, want between %d and %d", got, before, after)
	}
}

func TestRealTimeProvider_NowNano(t *testing.T) {
	tp := &RealTimeProvider{}
	before := time.Now().UnixNano()
	got := tp.NowNano()
	after := time.Now().UnixNano()
	if got < before || got > after {
		t.Errorf("RealTimeProvider.NowNano() = %d, want between %d and %d", got, before, after)
	}
}

func TestRealTimeProvider_Sleep(t *testing.T) {
	tp := &RealTimeProvider{}
	start := time.Now()
	tp.Sleep(50) // 50ms
	elapsed := time.Since(start)
	if elapsed < 40*time.Millisecond {
		t.Errorf("RealTimeProvider.Sleep(50) only waited %v", elapsed)
	}
}

func TestMockTimeProvider_Now(t *testing.T) {
	tp := NewMockTimeProvider()
	if got := tp.Now(); got != 1000000 {
		t.Errorf("MockTimeProvider.Now() = %d, want 1000000", got)
	}
}

func TestMockTimeProvider_NowNano(t *testing.T) {
	tp := NewMockTimeProvider()
	expected := int64(1000000 * 1e9)
	if got := tp.NowNano(); got != expected {
		t.Errorf("MockTimeProvider.NowNano() = %d, want %d", got, expected)
	}
}

func TestMockTimeProvider_SetTime(t *testing.T) {
	tp := NewMockTimeProvider()
	tp.SetTime(2000000)
	if got := tp.Now(); got != 2000000 {
		t.Errorf("After SetTime(2000000), Now() = %d, want 2000000", got)
	}
	expected := int64(2000000 * 1e9)
	if got := tp.NowNano(); got != expected {
		t.Errorf("After SetTime(2000000), NowNano() = %d, want %d", got, expected)
	}
}

func TestMockTimeProvider_Sleep(t *testing.T) {
	tp := NewMockTimeProvider()
	tp.Sleep(1500) // 1500ms = 1.5s
	if got := tp.Now(); got != 1000001 {
		t.Errorf("After Sleep(1500), Now() = %d, want 1000001", got)
	}
	log := tp.SleepLog()
	if len(log) != 1 || log[0] != 1500 {
		t.Errorf("SleepLog = %v, want [1500]", log)
	}
}

func TestMockTimeProvider_MultipleSleeps(t *testing.T) {
	tp := NewMockTimeProvider()
	tp.Sleep(1000)
	tp.Sleep(2000)
	tp.Sleep(500)
	if got := tp.Now(); got != 1000003 {
		t.Errorf("After multiple sleeps, Now() = %d, want 1000003", got)
	}
	log := tp.SleepLog()
	if len(log) != 3 {
		t.Fatalf("SleepLog length = %d, want 3", len(log))
	}
	if log[0] != 1000 || log[1] != 2000 || log[2] != 500 {
		t.Errorf("SleepLog = %v, want [1000 2000 500]", log)
	}
}

func TestMockTimeProvider_DeterministicTime(t *testing.T) {
	tp1 := NewMockTimeProvider()
	tp2 := NewMockTimeProvider()
	// Both should return the same time
	if tp1.Now() != tp2.Now() {
		t.Error("Two fresh MockTimeProviders should return the same time")
	}
	tp1.SetTime(12345)
	tp2.SetTime(12345)
	if tp1.Now() != tp2.Now() {
		t.Error("Two MockTimeProviders set to same time should be equal")
	}
}

// ============================================================================
// EnvProvider Tests
// ============================================================================

func TestRealEnvProvider_GetSet(t *testing.T) {
	ep := &RealEnvProvider{}
	key := "AURA_TEST_ENV_VAR_12345"
	// Should not exist initially
	_, ok := ep.Get(key)
	if ok {
		t.Skip("Test env var already exists, skipping")
	}
	// Set it
	ep.Set(key, "hello")
	defer os.Unsetenv(key)
	val, ok := ep.Get(key)
	if !ok || val != "hello" {
		t.Errorf("After Set, Get(%q) = (%q, %v), want (\"hello\", true)", key, val, ok)
	}
}

func TestRealEnvProvider_Has(t *testing.T) {
	ep := &RealEnvProvider{}
	// PATH should exist on most systems
	if !ep.Has("PATH") {
		t.Error("RealEnvProvider.Has(\"PATH\") = false, want true")
	}
	if ep.Has("AURA_UNLIKELY_VAR_999999") {
		t.Error("RealEnvProvider.Has(\"AURA_UNLIKELY_VAR_999999\") = true, want false")
	}
}

func TestRealEnvProvider_List(t *testing.T) {
	ep := &RealEnvProvider{}
	vars := ep.List()
	if len(vars) == 0 {
		t.Error("RealEnvProvider.List() returned empty map")
	}
	// PATH should be in the list
	if _, ok := vars["PATH"]; !ok {
		t.Error("RealEnvProvider.List() missing PATH")
	}
}

func TestRealEnvProvider_Cwd(t *testing.T) {
	ep := &RealEnvProvider{}
	cwd, err := ep.Cwd()
	if err != nil {
		t.Fatalf("RealEnvProvider.Cwd() error: %v", err)
	}
	if cwd == "" {
		t.Error("RealEnvProvider.Cwd() returned empty string")
	}
}

func TestRealEnvProvider_Args(t *testing.T) {
	ep := &RealEnvProvider{}
	args := ep.Args()
	if len(args) == 0 {
		t.Error("RealEnvProvider.Args() returned empty slice")
	}
}

func TestMockEnvProvider_GetSet(t *testing.T) {
	ep := NewMockEnvProvider()
	// Initially empty
	_, ok := ep.Get("FOO")
	if ok {
		t.Error("Fresh mock should not have FOO")
	}
	ep.Set("FOO", "bar")
	val, ok := ep.Get("FOO")
	if !ok || val != "bar" {
		t.Errorf("After Set, Get(\"FOO\") = (%q, %v), want (\"bar\", true)", val, ok)
	}
}

func TestMockEnvProvider_SetVar(t *testing.T) {
	ep := NewMockEnvProvider()
	ep.SetVar("KEY1", "val1")
	ep.SetVar("KEY2", "val2")
	val, ok := ep.Get("KEY1")
	if !ok || val != "val1" {
		t.Errorf("Get(\"KEY1\") = (%q, %v), want (\"val1\", true)", val, ok)
	}
}

func TestMockEnvProvider_Has(t *testing.T) {
	ep := NewMockEnvProvider()
	if ep.Has("MISSING") {
		t.Error("Has(\"MISSING\") should be false")
	}
	ep.Set("PRESENT", "yes")
	if !ep.Has("PRESENT") {
		t.Error("Has(\"PRESENT\") should be true after Set")
	}
}

func TestMockEnvProvider_List(t *testing.T) {
	ep := NewMockEnvProvider()
	ep.SetVar("A", "1")
	ep.SetVar("B", "2")
	ep.SetVar("C", "3")
	vars := ep.List()
	if len(vars) != 3 {
		t.Errorf("List() length = %d, want 3", len(vars))
	}
	if vars["A"] != "1" || vars["B"] != "2" || vars["C"] != "3" {
		t.Errorf("List() = %v, want {A:1, B:2, C:3}", vars)
	}
}

func TestMockEnvProvider_Cwd(t *testing.T) {
	ep := NewMockEnvProvider()
	cwd, err := ep.Cwd()
	if err != nil {
		t.Fatalf("Cwd() error: %v", err)
	}
	if cwd != "/mock/cwd" {
		t.Errorf("Cwd() = %q, want \"/mock/cwd\"", cwd)
	}
	ep.SetCwd("/new/path")
	cwd, _ = ep.Cwd()
	if cwd != "/new/path" {
		t.Errorf("After SetCwd, Cwd() = %q, want \"/new/path\"", cwd)
	}
}

func TestMockEnvProvider_Args(t *testing.T) {
	ep := NewMockEnvProvider()
	args := ep.Args()
	if len(args) != 1 || args[0] != "aura" {
		t.Errorf("Args() = %v, want [\"aura\"]", args)
	}
	ep.SetArgs([]string{"aura", "--flag", "value"})
	args = ep.Args()
	if len(args) != 3 || args[1] != "--flag" {
		t.Errorf("After SetArgs, Args() = %v, want [\"aura\", \"--flag\", \"value\"]", args)
	}
}

func TestMockEnvProvider_Overwrite(t *testing.T) {
	ep := NewMockEnvProvider()
	ep.Set("KEY", "old")
	ep.Set("KEY", "new")
	val, _ := ep.Get("KEY")
	if val != "new" {
		t.Errorf("After overwrite, Get(\"KEY\") = %q, want \"new\"", val)
	}
}

// ============================================================================
// EffectContext Integration Tests
// ============================================================================

func TestEffectContext_TimeProvider(t *testing.T) {
	ec := NewEffectContext()
	if ec.Time() == nil {
		t.Error("NewEffectContext().Time() should not be nil")
	}
}

func TestEffectContext_EnvProvider(t *testing.T) {
	ec := NewEffectContext()
	if ec.Env() == nil {
		t.Error("NewEffectContext().Env() should not be nil")
	}
}

func TestMockEffectContext_TimeProvider(t *testing.T) {
	ec := NewMockEffectContext()
	tp := ec.Time()
	if tp == nil {
		t.Fatal("NewMockEffectContext().Time() should not be nil")
	}
	// Should be a MockTimeProvider
	if _, ok := tp.(*MockTimeProvider); !ok {
		t.Error("NewMockEffectContext().Time() should be *MockTimeProvider")
	}
}

func TestMockEffectContext_EnvProvider(t *testing.T) {
	ec := NewMockEffectContext()
	ep := ec.Env()
	if ep == nil {
		t.Fatal("NewMockEffectContext().Env() should not be nil")
	}
	if _, ok := ep.(*MockEnvProvider); !ok {
		t.Error("NewMockEffectContext().Env() should be *MockEnvProvider")
	}
}

func TestEffectContext_WithTime(t *testing.T) {
	ec := NewEffectContext()
	mockTP := NewMockTimeProvider()
	ec2 := ec.WithTime(mockTP)
	if ec2.Time() != mockTP {
		t.Error("WithTime should replace time provider")
	}
	// File should be preserved
	if ec2.File() != ec.File() {
		t.Error("WithTime should preserve file provider")
	}
}

func TestEffectContext_WithEnv(t *testing.T) {
	ec := NewEffectContext()
	mockEP := NewMockEnvProvider()
	ec2 := ec.WithEnv(mockEP)
	if ec2.Env() != mockEP {
		t.Error("WithEnv should replace env provider")
	}
	if ec2.File() != ec.File() {
		t.Error("WithEnv should preserve file provider")
	}
}

// ============================================================================
// std.time Function Tests
// ============================================================================

func TestStdTime_Now(t *testing.T) {
	tp := NewMockTimeProvider()
	tp.SetTime(1700000000)
	exports := createStdTimeExports(tp)
	fn := exports["now"].(*BuiltinFnVal)
	result := fn.Fn([]Value{})
	iv, ok := result.(*IntVal)
	if !ok {
		t.Fatalf("time.now() returned %T, want *IntVal", result)
	}
	if iv.Val != 1700000000 {
		t.Errorf("time.now() = %d, want 1700000000", iv.Val)
	}
}

func TestStdTime_Unix(t *testing.T) {
	tp := NewMockTimeProvider()
	tp.SetTime(1700000000)
	exports := createStdTimeExports(tp)
	fn := exports["unix"].(*BuiltinFnVal)
	result := fn.Fn([]Value{})
	iv := result.(*IntVal)
	if iv.Val != 1700000000 {
		t.Errorf("time.unix() = %d, want 1700000000", iv.Val)
	}
}

func TestStdTime_Sleep(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["sleep"].(*BuiltinFnVal)
	result := fn.Fn([]Value{&IntVal{Val: 100}})
	if _, ok := result.(*NoneVal); !ok {
		t.Errorf("time.sleep() returned %T, want *NoneVal", result)
	}
	log := tp.SleepLog()
	if len(log) != 1 || log[0] != 100 {
		t.Errorf("SleepLog = %v, want [100]", log)
	}
}

func TestStdTime_Sleep_NegativeError(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["sleep"].(*BuiltinFnVal)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("time.sleep(-1) should panic")
		}
	}()
	fn.Fn([]Value{&IntVal{Val: -1}})
}

func TestStdTime_Sleep_WrongType(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["sleep"].(*BuiltinFnVal)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("time.sleep(\"100\") should panic")
		}
	}()
	fn.Fn([]Value{&StringVal{Val: "100"}})
}

func TestStdTime_Millis(t *testing.T) {
	tp := NewMockTimeProvider()
	tp.SetTime(1700000000)
	exports := createStdTimeExports(tp)
	fn := exports["millis"].(*BuiltinFnVal)
	result := fn.Fn([]Value{})
	iv := result.(*IntVal)
	expected := int64(1700000000 * 1e9 / 1e6)
	if iv.Val != expected {
		t.Errorf("time.millis() = %d, want %d", iv.Val, expected)
	}
}

func TestStdTime_Format(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["format"].(*BuiltinFnVal)
	// 1700000000 = 2023-11-14 22:13:20 UTC
	result := fn.Fn([]Value{
		&IntVal{Val: 1700000000},
		&StringVal{Val: "%Y-%m-%d %H:%M:%S"},
	})
	sv := result.(*StringVal)
	if sv.Val != "2023-11-14 22:13:20" {
		t.Errorf("time.format(1700000000, \"%%Y-%%m-%%d %%H:%%M:%%S\") = %q, want \"2023-11-14 22:13:20\"", sv.Val)
	}
}

func TestStdTime_Format_DateOnly(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["format"].(*BuiltinFnVal)
	result := fn.Fn([]Value{
		&IntVal{Val: 0}, // Unix epoch
		&StringVal{Val: "%Y-%m-%d"},
	})
	sv := result.(*StringVal)
	if sv.Val != "1970-01-01" {
		t.Errorf("time.format(0, \"%%Y-%%m-%%d\") = %q, want \"1970-01-01\"", sv.Val)
	}
}

func TestStdTime_Parse(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["parse"].(*BuiltinFnVal)
	result := fn.Fn([]Value{
		&StringVal{Val: "2023-11-14 22:13:20"},
		&StringVal{Val: "%Y-%m-%d %H:%M:%S"},
	})
	rv := result.(*ResultVal)
	if !rv.IsOk {
		t.Fatalf("time.parse() returned Err: %s", rv.Val.String())
	}
	iv := rv.Val.(*IntVal)
	if iv.Val != 1700000000 {
		t.Errorf("time.parse() = %d, want 1700000000", iv.Val)
	}
}

func TestStdTime_Parse_Invalid(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["parse"].(*BuiltinFnVal)
	result := fn.Fn([]Value{
		&StringVal{Val: "not-a-date"},
		&StringVal{Val: "%Y-%m-%d"},
	})
	rv := result.(*ResultVal)
	if rv.IsOk {
		t.Error("time.parse(\"not-a-date\") should return Err")
	}
}

func TestStdTime_Add(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["add"].(*BuiltinFnVal)
	result := fn.Fn([]Value{&IntVal{Val: 1000}, &IntVal{Val: 3600}})
	iv := result.(*IntVal)
	if iv.Val != 4600 {
		t.Errorf("time.add(1000, 3600) = %d, want 4600", iv.Val)
	}
}

func TestStdTime_Add_Negative(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["add"].(*BuiltinFnVal)
	result := fn.Fn([]Value{&IntVal{Val: 5000}, &IntVal{Val: -1000}})
	iv := result.(*IntVal)
	if iv.Val != 4000 {
		t.Errorf("time.add(5000, -1000) = %d, want 4000", iv.Val)
	}
}

func TestStdTime_Diff(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["diff"].(*BuiltinFnVal)
	result := fn.Fn([]Value{&IntVal{Val: 5000}, &IntVal{Val: 3000}})
	iv := result.(*IntVal)
	if iv.Val != 2000 {
		t.Errorf("time.diff(5000, 3000) = %d, want 2000", iv.Val)
	}
}

func TestStdTime_Diff_Negative(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["diff"].(*BuiltinFnVal)
	result := fn.Fn([]Value{&IntVal{Val: 1000}, &IntVal{Val: 5000}})
	iv := result.(*IntVal)
	if iv.Val != -4000 {
		t.Errorf("time.diff(1000, 5000) = %d, want -4000", iv.Val)
	}
}

func TestStdTime_Now_NoArgs(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["now"].(*BuiltinFnVal)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("time.now(1) should panic")
		}
	}()
	fn.Fn([]Value{&IntVal{Val: 1}})
}

func TestStdTime_Format_WrongArgs(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["format"].(*BuiltinFnVal)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("time.format() with wrong args should panic")
		}
	}()
	fn.Fn([]Value{&StringVal{Val: "not an int"}, &StringVal{Val: "%Y"}})
}

func TestStdTime_Add_WrongArgs(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["add"].(*BuiltinFnVal)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("time.add with wrong types should panic")
		}
	}()
	fn.Fn([]Value{&StringVal{Val: "x"}, &IntVal{Val: 1}})
}

// ============================================================================
// std.env Function Tests
// ============================================================================

func TestStdEnv_Get_Exists(t *testing.T) {
	ep := NewMockEnvProvider()
	ep.SetVar("HOME", "/home/test")
	exports := createStdEnvExports(ep)
	fn := exports["get"].(*BuiltinFnVal)
	result := fn.Fn([]Value{&StringVal{Val: "HOME"}})
	opt := result.(*OptionVal)
	if !opt.IsSome {
		t.Fatal("env.get(\"HOME\") should return Some")
	}
	sv := opt.Val.(*StringVal)
	if sv.Val != "/home/test" {
		t.Errorf("env.get(\"HOME\") = %q, want \"/home/test\"", sv.Val)
	}
}

func TestStdEnv_Get_Missing(t *testing.T) {
	ep := NewMockEnvProvider()
	exports := createStdEnvExports(ep)
	fn := exports["get"].(*BuiltinFnVal)
	result := fn.Fn([]Value{&StringVal{Val: "MISSING"}})
	opt := result.(*OptionVal)
	if opt.IsSome {
		t.Error("env.get(\"MISSING\") should return None")
	}
}

func TestStdEnv_Set(t *testing.T) {
	ep := NewMockEnvProvider()
	exports := createStdEnvExports(ep)
	setFn := exports["set"].(*BuiltinFnVal)
	getFn := exports["get"].(*BuiltinFnVal)
	result := setFn.Fn([]Value{&StringVal{Val: "NEW_VAR"}, &StringVal{Val: "new_value"}})
	if _, ok := result.(*NoneVal); !ok {
		t.Errorf("env.set() returned %T, want *NoneVal", result)
	}
	// Verify it was set
	opt := getFn.Fn([]Value{&StringVal{Val: "NEW_VAR"}}).(*OptionVal)
	if !opt.IsSome || opt.Val.(*StringVal).Val != "new_value" {
		t.Error("env.set() did not persist the value")
	}
}

func TestStdEnv_Has_True(t *testing.T) {
	ep := NewMockEnvProvider()
	ep.SetVar("EXISTS", "yes")
	exports := createStdEnvExports(ep)
	fn := exports["has"].(*BuiltinFnVal)
	result := fn.Fn([]Value{&StringVal{Val: "EXISTS"}})
	bv := result.(*BoolVal)
	if !bv.Val {
		t.Error("env.has(\"EXISTS\") should be true")
	}
}

func TestStdEnv_Has_False(t *testing.T) {
	ep := NewMockEnvProvider()
	exports := createStdEnvExports(ep)
	fn := exports["has"].(*BuiltinFnVal)
	result := fn.Fn([]Value{&StringVal{Val: "NOPE"}})
	bv := result.(*BoolVal)
	if bv.Val {
		t.Error("env.has(\"NOPE\") should be false")
	}
}

func TestStdEnv_List(t *testing.T) {
	ep := NewMockEnvProvider()
	ep.SetVar("A", "1")
	ep.SetVar("B", "2")
	exports := createStdEnvExports(ep)
	fn := exports["list"].(*BuiltinFnVal)
	result := fn.Fn([]Value{})
	mv := result.(*MapVal)
	if len(mv.Keys) != 2 {
		t.Errorf("env.list() returned map with %d entries, want 2", len(mv.Keys))
	}
}

func TestStdEnv_Cwd(t *testing.T) {
	ep := NewMockEnvProvider()
	exports := createStdEnvExports(ep)
	fn := exports["cwd"].(*BuiltinFnVal)
	result := fn.Fn([]Value{})
	sv := result.(*StringVal)
	if sv.Val != "/mock/cwd" {
		t.Errorf("env.cwd() = %q, want \"/mock/cwd\"", sv.Val)
	}
}

func TestStdEnv_Cwd_Custom(t *testing.T) {
	ep := NewMockEnvProvider()
	ep.SetCwd("/custom/dir")
	exports := createStdEnvExports(ep)
	fn := exports["cwd"].(*BuiltinFnVal)
	result := fn.Fn([]Value{})
	sv := result.(*StringVal)
	if sv.Val != "/custom/dir" {
		t.Errorf("env.cwd() = %q, want \"/custom/dir\"", sv.Val)
	}
}

func TestStdEnv_Args(t *testing.T) {
	ep := NewMockEnvProvider()
	exports := createStdEnvExports(ep)
	fn := exports["args"].(*BuiltinFnVal)
	result := fn.Fn([]Value{})
	lv := result.(*ListVal)
	if len(lv.Elements) != 1 || lv.Elements[0].(*StringVal).Val != "aura" {
		t.Errorf("env.args() = %v, want [\"aura\"]", result)
	}
}

func TestStdEnv_Args_Custom(t *testing.T) {
	ep := NewMockEnvProvider()
	ep.SetArgs([]string{"aura", "test.aura", "--verbose"})
	exports := createStdEnvExports(ep)
	fn := exports["args"].(*BuiltinFnVal)
	result := fn.Fn([]Value{})
	lv := result.(*ListVal)
	if len(lv.Elements) != 3 {
		t.Fatalf("env.args() length = %d, want 3", len(lv.Elements))
	}
	if lv.Elements[2].(*StringVal).Val != "--verbose" {
		t.Errorf("env.args()[2] = %q, want \"--verbose\"", lv.Elements[2].String())
	}
}

func TestStdEnv_Get_WrongType(t *testing.T) {
	ep := NewMockEnvProvider()
	exports := createStdEnvExports(ep)
	fn := exports["get"].(*BuiltinFnVal)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("env.get(123) should panic")
		}
	}()
	fn.Fn([]Value{&IntVal{Val: 123}})
}

func TestStdEnv_Set_WrongTypes(t *testing.T) {
	ep := NewMockEnvProvider()
	exports := createStdEnvExports(ep)
	fn := exports["set"].(*BuiltinFnVal)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("env.set(123, 456) should panic")
		}
	}()
	fn.Fn([]Value{&IntVal{Val: 123}, &IntVal{Val: 456}})
}

func TestStdEnv_Has_WrongType(t *testing.T) {
	ep := NewMockEnvProvider()
	exports := createStdEnvExports(ep)
	fn := exports["has"].(*BuiltinFnVal)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("env.has(123) should panic")
		}
	}()
	fn.Fn([]Value{&IntVal{Val: 123}})
}

func TestStdEnv_List_NoArgs(t *testing.T) {
	ep := NewMockEnvProvider()
	exports := createStdEnvExports(ep)
	fn := exports["list"].(*BuiltinFnVal)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("env.list(\"x\") should panic")
		}
	}()
	fn.Fn([]Value{&StringVal{Val: "x"}})
}

// ============================================================================
// Time Format Tests
// ============================================================================

func TestAuraToGoTimeFormat(t *testing.T) {
	tests := []struct {
		aura string
		goFmt string
	}{
		{"%Y-%m-%d", "2006-01-02"},
		{"%H:%M:%S", "15:04:05"},
		{"%Y-%m-%d %H:%M:%S", "2006-01-02 15:04:05"},
		{"%Y/%m/%d", "2006/01/02"},
		{"%d.%m.%Y", "02.01.2006"},
	}
	for _, tt := range tests {
		got := auraToGoTimeFormat(tt.aura)
		if got != tt.goFmt {
			t.Errorf("auraToGoTimeFormat(%q) = %q, want %q", tt.aura, got, tt.goFmt)
		}
	}
}

func TestStdTime_FormatParse_Roundtrip(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	formatFn := exports["format"].(*BuiltinFnVal)
	parseFn := exports["parse"].(*BuiltinFnVal)

	timestamps := []int64{0, 1000000, 1700000000, 2000000000}
	for _, ts := range timestamps {
		formatted := formatFn.Fn([]Value{
			&IntVal{Val: ts},
			&StringVal{Val: "%Y-%m-%d %H:%M:%S"},
		}).(*StringVal)
		parsed := parseFn.Fn([]Value{
			formatted,
			&StringVal{Val: "%Y-%m-%d %H:%M:%S"},
		}).(*ResultVal)
		if !parsed.IsOk {
			t.Errorf("Roundtrip failed for ts=%d: parse error", ts)
			continue
		}
		got := parsed.Val.(*IntVal).Val
		if got != ts {
			t.Errorf("Roundtrip for ts=%d: format->parse = %d", ts, got)
		}
	}
}

// ============================================================================
// Integration: std.time and std.env via interpreter
// ============================================================================

func TestStdTime_Integration_NowAndSleep(t *testing.T) {
	mockEffects := NewMockEffectContext()
	mockTP := mockEffects.Time().(*MockTimeProvider)
	mockTP.SetTime(1000)

	exports := createStdTimeExports(mockTP)
	nowFn := exports["now"].(*BuiltinFnVal)
	sleepFn := exports["sleep"].(*BuiltinFnVal)

	// Get initial time
	t1 := nowFn.Fn([]Value{}).(*IntVal).Val
	if t1 != 1000 {
		t.Errorf("Initial time = %d, want 1000", t1)
	}

	// Sleep 2000ms
	sleepFn.Fn([]Value{&IntVal{Val: 2000}})

	// Time should have advanced by 2 seconds
	t2 := nowFn.Fn([]Value{}).(*IntVal).Val
	if t2 != 1002 {
		t.Errorf("After sleep(2000), time = %d, want 1002", t2)
	}
}

func TestStdEnv_Integration_SetGetCycle(t *testing.T) {
	mockEffects := NewMockEffectContext()
	mockEP := mockEffects.Env().(*MockEnvProvider)

	exports := createStdEnvExports(mockEP)
	setFn := exports["set"].(*BuiltinFnVal)
	getFn := exports["get"].(*BuiltinFnVal)
	hasFn := exports["has"].(*BuiltinFnVal)

	// Initially empty
	if hasFn.Fn([]Value{&StringVal{Val: "APP_MODE"}}).(*BoolVal).Val {
		t.Error("APP_MODE should not exist initially")
	}

	// Set and verify
	setFn.Fn([]Value{&StringVal{Val: "APP_MODE"}, &StringVal{Val: "production"}})

	if !hasFn.Fn([]Value{&StringVal{Val: "APP_MODE"}}).(*BoolVal).Val {
		t.Error("APP_MODE should exist after set")
	}

	opt := getFn.Fn([]Value{&StringVal{Val: "APP_MODE"}}).(*OptionVal)
	if !opt.IsSome || opt.Val.(*StringVal).Val != "production" {
		t.Error("env.get(\"APP_MODE\") should return Some(\"production\")")
	}
}

func TestStdEnv_List_EmptyEnv(t *testing.T) {
	ep := NewMockEnvProvider()
	exports := createStdEnvExports(ep)
	fn := exports["list"].(*BuiltinFnVal)
	result := fn.Fn([]Value{})
	mv := result.(*MapVal)
	if len(mv.Keys) != 0 {
		t.Errorf("env.list() on empty env returned %d entries, want 0", len(mv.Keys))
	}
}

func TestStdTime_Diff_Zero(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["diff"].(*BuiltinFnVal)
	result := fn.Fn([]Value{&IntVal{Val: 5000}, &IntVal{Val: 5000}})
	iv := result.(*IntVal)
	if iv.Val != 0 {
		t.Errorf("time.diff(5000, 5000) = %d, want 0", iv.Val)
	}
}

func TestStdTime_Sleep_Zero(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["sleep"].(*BuiltinFnVal)
	result := fn.Fn([]Value{&IntVal{Val: 0}})
	if _, ok := result.(*NoneVal); !ok {
		t.Errorf("time.sleep(0) returned %T, want *NoneVal", result)
	}
}

func TestStdTime_Add_Zero(t *testing.T) {
	tp := NewMockTimeProvider()
	exports := createStdTimeExports(tp)
	fn := exports["add"].(*BuiltinFnVal)
	result := fn.Fn([]Value{&IntVal{Val: 1000}, &IntVal{Val: 0}})
	iv := result.(*IntVal)
	if iv.Val != 1000 {
		t.Errorf("time.add(1000, 0) = %d, want 1000", iv.Val)
	}
}
