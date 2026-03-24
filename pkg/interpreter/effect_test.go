package interpreter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// Effect Context Tests
// =============================================================================

func TestEffectContext_DefaultProviders(t *testing.T) {
	ec := NewEffectContext()
	if ec.File() == nil {
		t.Fatal("expected non-nil FileProvider")
	}
	if _, ok := ec.File().(*RealFileProvider); !ok {
		t.Fatal("expected RealFileProvider as default")
	}
}

func TestEffectContext_MockProviders(t *testing.T) {
	ec := NewMockEffectContext()
	if ec.File() == nil {
		t.Fatal("expected non-nil FileProvider")
	}
	if _, ok := ec.File().(*MockFileProvider); !ok {
		t.Fatal("expected MockFileProvider")
	}
}

func TestEffectContext_WithFile(t *testing.T) {
	ec := NewEffectContext()
	mock := NewMockFileProvider()
	ec2 := ec.WithFile(mock)
	if _, ok := ec2.File().(*MockFileProvider); !ok {
		t.Fatal("WithFile should replace the file provider")
	}
	// Original should be unchanged
	if _, ok := ec.File().(*RealFileProvider); !ok {
		t.Fatal("original context should be unchanged")
	}
}

// =============================================================================
// Mock File Provider Tests
// =============================================================================

func TestMockFileProvider_ReadWriteFile(t *testing.T) {
	m := NewMockFileProvider()
	// Write
	err := m.WriteFile("/tmp/test.txt", "hello world")
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	// Read
	content, err := m.ReadFile("/tmp/test.txt")
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if content != "hello world" {
		t.Fatalf("expected 'hello world', got %q", content)
	}
}

func TestMockFileProvider_ReadNonExistent(t *testing.T) {
	m := NewMockFileProvider()
	_, err := m.ReadFile("/nonexistent")
	if err == nil {
		t.Fatal("expected error reading nonexistent file")
	}
}

func TestMockFileProvider_AppendFile(t *testing.T) {
	m := NewMockFileProvider()
	m.WriteFile("/tmp/log.txt", "line1\n")
	m.AppendFile("/tmp/log.txt", "line2\n")
	content, _ := m.ReadFile("/tmp/log.txt")
	if content != "line1\nline2\n" {
		t.Fatalf("expected 'line1\\nline2\\n', got %q", content)
	}
}

func TestMockFileProvider_AppendCreatesFile(t *testing.T) {
	m := NewMockFileProvider()
	m.AppendFile("/tmp/new.txt", "content")
	content, err := m.ReadFile("/tmp/new.txt")
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if content != "content" {
		t.Fatalf("expected 'content', got %q", content)
	}
}

func TestMockFileProvider_Exists(t *testing.T) {
	m := NewMockFileProvider()
	if m.Exists("/tmp/x") {
		t.Fatal("should not exist yet")
	}
	m.WriteFile("/tmp/x", "data")
	if !m.Exists("/tmp/x") {
		t.Fatal("should exist after write")
	}
}

func TestMockFileProvider_ExistsDir(t *testing.T) {
	m := NewMockFileProvider()
	m.CreateDir("/tmp/mydir")
	if !m.Exists("/tmp/mydir") {
		t.Fatal("directory should exist")
	}
}

func TestMockFileProvider_Delete(t *testing.T) {
	m := NewMockFileProvider()
	m.WriteFile("/tmp/del.txt", "data")
	err := m.Delete("/tmp/del.txt")
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if m.Exists("/tmp/del.txt") {
		t.Fatal("file should not exist after delete")
	}
}

func TestMockFileProvider_DeleteNonExistent(t *testing.T) {
	m := NewMockFileProvider()
	err := m.Delete("/nonexistent")
	if err == nil {
		t.Fatal("expected error deleting nonexistent file")
	}
}

func TestMockFileProvider_DeleteNonEmptyDir(t *testing.T) {
	m := NewMockFileProvider()
	m.CreateDir("/tmp/dir")
	m.WriteFile("/tmp/dir/file.txt", "data")
	err := m.Delete("/tmp/dir")
	if err == nil {
		t.Fatal("expected error deleting non-empty directory")
	}
}

func TestMockFileProvider_ListDir(t *testing.T) {
	m := NewMockFileProvider()
	m.CreateDir("/tmp/project")
	m.WriteFile("/tmp/project/a.txt", "a")
	m.WriteFile("/tmp/project/b.txt", "b")
	m.CreateDir("/tmp/project/sub")

	entries, err := m.ListDir("/tmp/project")
	if err != nil {
		t.Fatalf("list_dir failed: %v", err)
	}
	// Should be sorted
	expected := []string{"a.txt", "b.txt", "sub"}
	if len(entries) != len(expected) {
		t.Fatalf("expected %d entries, got %d: %v", len(expected), len(entries), entries)
	}
	for i, e := range expected {
		if entries[i] != e {
			t.Fatalf("entry %d: expected %q, got %q", i, e, entries[i])
		}
	}
}

func TestMockFileProvider_ListDirNonExistent(t *testing.T) {
	m := NewMockFileProvider()
	_, err := m.ListDir("/nonexistent")
	if err == nil {
		t.Fatal("expected error listing nonexistent directory")
	}
}

func TestMockFileProvider_CreateDir(t *testing.T) {
	m := NewMockFileProvider()
	err := m.CreateDir("/tmp/a/b/c")
	if err != nil {
		t.Fatalf("create_dir failed: %v", err)
	}
	if !m.IsDir("/tmp/a/b/c") {
		t.Fatal("directory should exist")
	}
	if !m.IsDir("/tmp/a/b") {
		t.Fatal("parent directory should exist")
	}
	if !m.IsDir("/tmp/a") {
		t.Fatal("grandparent directory should exist")
	}
}

func TestMockFileProvider_IsFile(t *testing.T) {
	m := NewMockFileProvider()
	m.WriteFile("/tmp/f.txt", "data")
	m.CreateDir("/tmp/d")
	if !m.IsFile("/tmp/f.txt") {
		t.Fatal("should be a file")
	}
	if m.IsFile("/tmp/d") {
		t.Fatal("directory should not be a file")
	}
	if m.IsFile("/tmp/nonexistent") {
		t.Fatal("nonexistent should not be a file")
	}
}

func TestMockFileProvider_IsDir(t *testing.T) {
	m := NewMockFileProvider()
	m.WriteFile("/tmp/f.txt", "data")
	m.CreateDir("/tmp/d")
	if m.IsDir("/tmp/f.txt") {
		t.Fatal("file should not be a directory")
	}
	if !m.IsDir("/tmp/d") {
		t.Fatal("should be a directory")
	}
	if m.IsDir("/tmp/nonexistent") {
		t.Fatal("nonexistent should not be a directory")
	}
}

func TestMockFileProvider_AddFileHelper(t *testing.T) {
	m := NewMockFileProvider()
	m.AddFile("/tmp/pre.txt", "pre-populated")
	content, err := m.ReadFile("/tmp/pre.txt")
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if content != "pre-populated" {
		t.Fatalf("expected 'pre-populated', got %q", content)
	}
}

func TestMockFileProvider_OverwriteFile(t *testing.T) {
	m := NewMockFileProvider()
	m.WriteFile("/tmp/f.txt", "first")
	m.WriteFile("/tmp/f.txt", "second")
	content, _ := m.ReadFile("/tmp/f.txt")
	if content != "second" {
		t.Fatalf("expected 'second', got %q", content)
	}
}

// =============================================================================
// Real File Provider Tests (using temp directory)
// =============================================================================

func TestRealFileProvider_ReadWriteFile(t *testing.T) {
	fp := &RealFileProvider{}
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	err := fp.WriteFile(path, "hello real")
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	content, err := fp.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if content != "hello real" {
		t.Fatalf("expected 'hello real', got %q", content)
	}
}

func TestRealFileProvider_AppendFile(t *testing.T) {
	fp := &RealFileProvider{}
	dir := t.TempDir()
	path := filepath.Join(dir, "append.txt")

	fp.WriteFile(path, "first\n")
	fp.AppendFile(path, "second\n")
	content, _ := fp.ReadFile(path)
	if content != "first\nsecond\n" {
		t.Fatalf("expected 'first\\nsecond\\n', got %q", content)
	}
}

func TestRealFileProvider_Exists(t *testing.T) {
	fp := &RealFileProvider{}
	dir := t.TempDir()
	path := filepath.Join(dir, "exists.txt")

	if fp.Exists(path) {
		t.Fatal("should not exist yet")
	}
	fp.WriteFile(path, "data")
	if !fp.Exists(path) {
		t.Fatal("should exist after write")
	}
}

func TestRealFileProvider_Delete(t *testing.T) {
	fp := &RealFileProvider{}
	dir := t.TempDir()
	path := filepath.Join(dir, "del.txt")

	fp.WriteFile(path, "data")
	err := fp.Delete(path)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if fp.Exists(path) {
		t.Fatal("should not exist after delete")
	}
}

func TestRealFileProvider_ListDir(t *testing.T) {
	fp := &RealFileProvider{}
	dir := t.TempDir()

	fp.WriteFile(filepath.Join(dir, "a.txt"), "a")
	fp.WriteFile(filepath.Join(dir, "b.txt"), "b")
	os.Mkdir(filepath.Join(dir, "sub"), 0755)

	entries, err := fp.ListDir(dir)
	if err != nil {
		t.Fatalf("list_dir failed: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d: %v", len(entries), entries)
	}
}

func TestRealFileProvider_CreateDir(t *testing.T) {
	fp := &RealFileProvider{}
	dir := t.TempDir()
	newDir := filepath.Join(dir, "a", "b", "c")

	err := fp.CreateDir(newDir)
	if err != nil {
		t.Fatalf("create_dir failed: %v", err)
	}
	if !fp.IsDir(newDir) {
		t.Fatal("directory should exist")
	}
}

func TestRealFileProvider_IsFileIsDir(t *testing.T) {
	fp := &RealFileProvider{}
	dir := t.TempDir()
	filePath := filepath.Join(dir, "file.txt")
	dirPath := filepath.Join(dir, "subdir")

	fp.WriteFile(filePath, "data")
	os.Mkdir(dirPath, 0755)

	if !fp.IsFile(filePath) {
		t.Fatal("should be a file")
	}
	if fp.IsFile(dirPath) {
		t.Fatal("directory should not be a file")
	}
	if !fp.IsDir(dirPath) {
		t.Fatal("should be a directory")
	}
	if fp.IsDir(filePath) {
		t.Fatal("file should not be a directory")
	}
}

// =============================================================================
// std.file Module Tests (using Mock Provider)
// =============================================================================

func TestStdFile_Read(t *testing.T) {
	mock := NewMockFileProvider()
	mock.AddFile("/data/hello.txt", "Hello, Aura!")
	exports := createStdFileExports(mock)

	readFn := exports["read"].(*BuiltinFnVal).Fn
	result := readFn([]Value{&StringVal{Val: "/data/hello.txt"}})
	r := result.(*ResultVal)
	if !r.IsOk {
		t.Fatalf("expected Ok, got Err(%s)", r.Val.String())
	}
	if r.Val.(*StringVal).Val != "Hello, Aura!" {
		t.Fatalf("expected 'Hello, Aura!', got %q", r.Val.String())
	}
}

func TestStdFile_ReadNotFound(t *testing.T) {
	mock := NewMockFileProvider()
	exports := createStdFileExports(mock)

	readFn := exports["read"].(*BuiltinFnVal).Fn
	result := readFn([]Value{&StringVal{Val: "/nonexistent"}})
	r := result.(*ResultVal)
	if r.IsOk {
		t.Fatal("expected Err for nonexistent file")
	}
	if !strings.Contains(r.Val.(*StringVal).Val, "not found") {
		t.Fatalf("expected 'not found' in error, got %q", r.Val.String())
	}
}

func TestStdFile_Write(t *testing.T) {
	mock := NewMockFileProvider()
	exports := createStdFileExports(mock)

	writeFn := exports["write"].(*BuiltinFnVal).Fn
	result := writeFn([]Value{&StringVal{Val: "/data/out.txt"}, &StringVal{Val: "written content"}})
	r := result.(*ResultVal)
	if !r.IsOk {
		t.Fatalf("expected Ok, got Err(%s)", r.Val.String())
	}
	// Verify content was written
	content, _ := mock.ReadFile("/data/out.txt")
	if content != "written content" {
		t.Fatalf("expected 'written content', got %q", content)
	}
}

func TestStdFile_Append(t *testing.T) {
	mock := NewMockFileProvider()
	mock.AddFile("/data/log.txt", "line1\n")
	exports := createStdFileExports(mock)

	appendFn := exports["append"].(*BuiltinFnVal).Fn
	result := appendFn([]Value{&StringVal{Val: "/data/log.txt"}, &StringVal{Val: "line2\n"}})
	r := result.(*ResultVal)
	if !r.IsOk {
		t.Fatalf("expected Ok, got Err(%s)", r.Val.String())
	}
	content, _ := mock.ReadFile("/data/log.txt")
	if content != "line1\nline2\n" {
		t.Fatalf("expected 'line1\\nline2\\n', got %q", content)
	}
}

func TestStdFile_Exists(t *testing.T) {
	mock := NewMockFileProvider()
	mock.AddFile("/data/exists.txt", "data")
	exports := createStdFileExports(mock)

	existsFn := exports["exists"].(*BuiltinFnVal).Fn
	result := existsFn([]Value{&StringVal{Val: "/data/exists.txt"}})
	if !result.(*BoolVal).Val {
		t.Fatal("expected true for existing file")
	}
	result = existsFn([]Value{&StringVal{Val: "/data/nope.txt"}})
	if result.(*BoolVal).Val {
		t.Fatal("expected false for nonexistent file")
	}
}

func TestStdFile_Delete(t *testing.T) {
	mock := NewMockFileProvider()
	mock.AddFile("/data/del.txt", "data")
	exports := createStdFileExports(mock)

	deleteFn := exports["delete"].(*BuiltinFnVal).Fn
	result := deleteFn([]Value{&StringVal{Val: "/data/del.txt"}})
	r := result.(*ResultVal)
	if !r.IsOk {
		t.Fatalf("expected Ok, got Err(%s)", r.Val.String())
	}
	if mock.Exists("/data/del.txt") {
		t.Fatal("file should be deleted")
	}
}

func TestStdFile_DeleteNotFound(t *testing.T) {
	mock := NewMockFileProvider()
	exports := createStdFileExports(mock)

	deleteFn := exports["delete"].(*BuiltinFnVal).Fn
	result := deleteFn([]Value{&StringVal{Val: "/nonexistent"}})
	r := result.(*ResultVal)
	if r.IsOk {
		t.Fatal("expected Err for nonexistent file")
	}
}

func TestStdFile_ListDir(t *testing.T) {
	mock := NewMockFileProvider()
	mock.CreateDir("/data/project")
	mock.AddFile("/data/project/main.aura", "fn main(): ...")
	mock.AddFile("/data/project/utils.aura", "fn helper(): ...")
	mock.AddDir("/data/project/tests")
	exports := createStdFileExports(mock)

	listFn := exports["list_dir"].(*BuiltinFnVal).Fn
	result := listFn([]Value{&StringVal{Val: "/data/project"}})
	r := result.(*ResultVal)
	if !r.IsOk {
		t.Fatalf("expected Ok, got Err(%s)", r.Val.String())
	}
	list := r.Val.(*ListVal)
	if len(list.Elements) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(list.Elements))
	}
	// Sorted: main.aura, tests, utils.aura
	names := make([]string, len(list.Elements))
	for i, e := range list.Elements {
		names[i] = e.(*StringVal).Val
	}
	expected := []string{"main.aura", "tests", "utils.aura"}
	for i, exp := range expected {
		if names[i] != exp {
			t.Fatalf("entry %d: expected %q, got %q", i, exp, names[i])
		}
	}
}

func TestStdFile_ListDirNotFound(t *testing.T) {
	mock := NewMockFileProvider()
	exports := createStdFileExports(mock)

	listFn := exports["list_dir"].(*BuiltinFnVal).Fn
	result := listFn([]Value{&StringVal{Val: "/nonexistent"}})
	r := result.(*ResultVal)
	if r.IsOk {
		t.Fatal("expected Err for nonexistent directory")
	}
}

func TestStdFile_CreateDir(t *testing.T) {
	mock := NewMockFileProvider()
	exports := createStdFileExports(mock)

	createFn := exports["create_dir"].(*BuiltinFnVal).Fn
	result := createFn([]Value{&StringVal{Val: "/data/new/nested/dir"}})
	r := result.(*ResultVal)
	if !r.IsOk {
		t.Fatalf("expected Ok, got Err(%s)", r.Val.String())
	}
	if !mock.IsDir("/data/new/nested/dir") {
		t.Fatal("directory should exist")
	}
}

func TestStdFile_IsFile(t *testing.T) {
	mock := NewMockFileProvider()
	mock.AddFile("/data/f.txt", "data")
	mock.CreateDir("/data/d")
	exports := createStdFileExports(mock)

	isFileFn := exports["is_file"].(*BuiltinFnVal).Fn
	if !isFileFn([]Value{&StringVal{Val: "/data/f.txt"}}).(*BoolVal).Val {
		t.Fatal("expected true for file")
	}
	if isFileFn([]Value{&StringVal{Val: "/data/d"}}).(*BoolVal).Val {
		t.Fatal("expected false for directory")
	}
	if isFileFn([]Value{&StringVal{Val: "/nope"}}).(*BoolVal).Val {
		t.Fatal("expected false for nonexistent")
	}
}

func TestStdFile_IsDir(t *testing.T) {
	mock := NewMockFileProvider()
	mock.AddFile("/data/f.txt", "data")
	mock.CreateDir("/data/d")
	exports := createStdFileExports(mock)

	isDirFn := exports["is_dir"].(*BuiltinFnVal).Fn
	if isDirFn([]Value{&StringVal{Val: "/data/f.txt"}}).(*BoolVal).Val {
		t.Fatal("expected false for file")
	}
	if !isDirFn([]Value{&StringVal{Val: "/data/d"}}).(*BoolVal).Val {
		t.Fatal("expected true for directory")
	}
	if isDirFn([]Value{&StringVal{Val: "/nope"}}).(*BoolVal).Val {
		t.Fatal("expected false for nonexistent")
	}
}

// =============================================================================
// std.file Module Tests with Real Provider (temp directory)
// =============================================================================

func TestStdFile_RealProvider_ReadWriteDeleteCycle(t *testing.T) {
	fp := &RealFileProvider{}
	exports := createStdFileExports(fp)
	dir := t.TempDir()
	path := filepath.Join(dir, "cycle.txt")

	// Write
	writeFn := exports["write"].(*BuiltinFnVal).Fn
	r := writeFn([]Value{&StringVal{Val: path}, &StringVal{Val: "cycle content"}}).(*ResultVal)
	if !r.IsOk {
		t.Fatalf("write failed: %s", r.Val.String())
	}

	// Read
	readFn := exports["read"].(*BuiltinFnVal).Fn
	r = readFn([]Value{&StringVal{Val: path}}).(*ResultVal)
	if !r.IsOk {
		t.Fatalf("read failed: %s", r.Val.String())
	}
	if r.Val.(*StringVal).Val != "cycle content" {
		t.Fatalf("expected 'cycle content', got %q", r.Val.String())
	}

	// Exists
	existsFn := exports["exists"].(*BuiltinFnVal).Fn
	if !existsFn([]Value{&StringVal{Val: path}}).(*BoolVal).Val {
		t.Fatal("should exist")
	}

	// Delete
	deleteFn := exports["delete"].(*BuiltinFnVal).Fn
	r = deleteFn([]Value{&StringVal{Val: path}}).(*ResultVal)
	if !r.IsOk {
		t.Fatalf("delete failed: %s", r.Val.String())
	}

	// Should not exist
	if existsFn([]Value{&StringVal{Val: path}}).(*BoolVal).Val {
		t.Fatal("should not exist after delete")
	}
}

func TestStdFile_RealProvider_ListDirCreateDir(t *testing.T) {
	fp := &RealFileProvider{}
	exports := createStdFileExports(fp)
	dir := t.TempDir()

	// Create dir
	createFn := exports["create_dir"].(*BuiltinFnVal).Fn
	newDir := filepath.Join(dir, "sub", "nested")
	r := createFn([]Value{&StringVal{Val: newDir}}).(*ResultVal)
	if !r.IsOk {
		t.Fatalf("create_dir failed: %s", r.Val.String())
	}

	// is_dir
	isDirFn := exports["is_dir"].(*BuiltinFnVal).Fn
	if !isDirFn([]Value{&StringVal{Val: newDir}}).(*BoolVal).Val {
		t.Fatal("new directory should exist")
	}

	// Write some files
	writeFn := exports["write"].(*BuiltinFnVal).Fn
	writeFn([]Value{&StringVal{Val: filepath.Join(dir, "a.txt")}, &StringVal{Val: "a"}})
	writeFn([]Value{&StringVal{Val: filepath.Join(dir, "b.txt")}, &StringVal{Val: "b"}})

	// list_dir
	listFn := exports["list_dir"].(*BuiltinFnVal).Fn
	lr := listFn([]Value{&StringVal{Val: dir}}).(*ResultVal)
	if !lr.IsOk {
		t.Fatalf("list_dir failed: %s", lr.Val.String())
	}
	list := lr.Val.(*ListVal)
	if len(list.Elements) != 3 { // a.txt, b.txt, sub
		t.Fatalf("expected 3 entries, got %d", len(list.Elements))
	}
}

// =============================================================================
// Effect Mocking Integration Tests
// =============================================================================

func TestEffectMocking_SwapFileProvider(t *testing.T) {
	// Demonstrate that mock providers can be injected for testing
	mock := NewMockFileProvider()
	mock.AddFile("/config/app.json", `{"port": 8080}`)

	ec := NewEffectContext().WithFile(mock)
	exports := createStdFileExports(ec.File())

	// Read from "filesystem" — actually reads from mock
	readFn := exports["read"].(*BuiltinFnVal).Fn
	result := readFn([]Value{&StringVal{Val: "/config/app.json"}}).(*ResultVal)
	if !result.IsOk {
		t.Fatal("should be able to read from mock")
	}
	if result.Val.(*StringVal).Val != `{"port": 8080}` {
		t.Fatal("wrong content from mock")
	}

	// Write to "filesystem" — only affects mock
	writeFn := exports["write"].(*BuiltinFnVal).Fn
	writeFn([]Value{&StringVal{Val: "/config/app.json"}, &StringVal{Val: `{"port": 9090}`}})

	// Verify mock was updated
	content, _ := mock.ReadFile("/config/app.json")
	if content != `{"port": 9090}` {
		t.Fatalf("mock should be updated, got %q", content)
	}
}

func TestEffectMocking_InterpreterWithMockEffects(t *testing.T) {
	// Test that NewWithEffects correctly threads mock effects
	mock := NewMockFileProvider()
	mock.AddFile("/test/data.txt", "test data")

	ec := NewMockEffectContext()
	ec = ec.WithFile(mock)

	// Verify the effect context is accessible
	if _, ok := ec.File().(*MockFileProvider); !ok {
		t.Fatal("expected MockFileProvider")
	}

	// Verify file operations work through the context
	content, err := ec.File().ReadFile("/test/data.txt")
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if content != "test data" {
		t.Fatalf("expected 'test data', got %q", content)
	}
}

// =============================================================================
// Error Handling / Argument Validation Tests
// =============================================================================

func TestStdFile_ReadWrongArgCount(t *testing.T) {
	mock := NewMockFileProvider()
	exports := createStdFileExports(mock)
	readFn := exports["read"].(*BuiltinFnVal).Fn

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		re, ok := r.(*RuntimeError)
		if !ok {
			t.Fatal("expected RuntimeError")
		}
		if !strings.Contains(re.Message, "requires exactly 1 argument") {
			t.Fatalf("unexpected error: %s", re.Message)
		}
	}()
	readFn([]Value{})
}

func TestStdFile_ReadWrongArgType(t *testing.T) {
	mock := NewMockFileProvider()
	exports := createStdFileExports(mock)
	readFn := exports["read"].(*BuiltinFnVal).Fn

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		re, ok := r.(*RuntimeError)
		if !ok {
			t.Fatal("expected RuntimeError")
		}
		if !strings.Contains(re.Message, "must be a String") {
			t.Fatalf("unexpected error: %s", re.Message)
		}
	}()
	readFn([]Value{&IntVal{Val: 42}})
}

func TestStdFile_WriteWrongArgCount(t *testing.T) {
	mock := NewMockFileProvider()
	exports := createStdFileExports(mock)
	writeFn := exports["write"].(*BuiltinFnVal).Fn

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
	}()
	writeFn([]Value{&StringVal{Val: "path"}})
}

func TestStdFile_ExportsAllFunctions(t *testing.T) {
	mock := NewMockFileProvider()
	exports := createStdFileExports(mock)

	expectedFns := []string{"read", "write", "append", "exists", "delete", "list_dir", "create_dir", "is_file", "is_dir"}
	for _, name := range expectedFns {
		if _, ok := exports[name]; !ok {
			t.Fatalf("missing export: %s", name)
		}
		if _, ok := exports[name].(*BuiltinFnVal); !ok {
			t.Fatalf("export %s is not a BuiltinFnVal", name)
		}
	}
	if len(exports) != len(expectedFns) {
		t.Fatalf("expected %d exports, got %d", len(expectedFns), len(exports))
	}
}

func TestStdFile_ModuleRegistration(t *testing.T) {
	// Verify std.file is recognized as a stdlib module
	if !strings.HasPrefix("std.file", "std.") {
		t.Fatal("std.file should be recognized as stdlib")
	}
}

// =============================================================================
// Complex Workflow Tests
// =============================================================================

func TestStdFile_CompleteWorkflow(t *testing.T) {
	mock := NewMockFileProvider()
	exports := createStdFileExports(mock)

	createDirFn := exports["create_dir"].(*BuiltinFnVal).Fn
	writeFn := exports["write"].(*BuiltinFnVal).Fn
	readFn := exports["read"].(*BuiltinFnVal).Fn
	appendFn := exports["append"].(*BuiltinFnVal).Fn
	existsFn := exports["exists"].(*BuiltinFnVal).Fn
	listDirFn := exports["list_dir"].(*BuiltinFnVal).Fn
	isFileFn := exports["is_file"].(*BuiltinFnVal).Fn
	isDirFn := exports["is_dir"].(*BuiltinFnVal).Fn
	deleteFn := exports["delete"].(*BuiltinFnVal).Fn

	// Step 1: Create a project directory
	r := createDirFn([]Value{&StringVal{Val: "/project/src"}}).(*ResultVal)
	if !r.IsOk {
		t.Fatal("create_dir failed")
	}

	// Step 2: Write files
	writeFn([]Value{&StringVal{Val: "/project/src/main.aura"}, &StringVal{Val: "fn main():\n    print(\"hello\")\n"}})
	writeFn([]Value{&StringVal{Val: "/project/src/utils.aura"}, &StringVal{Val: "fn helper(): 42\n"}})
	writeFn([]Value{&StringVal{Val: "/project/README.md"}, &StringVal{Val: "# My Project\n"}})

	// Step 3: Verify structure
	if !isDirFn([]Value{&StringVal{Val: "/project/src"}}).(*BoolVal).Val {
		t.Fatal("src should be a directory")
	}
	if !isFileFn([]Value{&StringVal{Val: "/project/src/main.aura"}}).(*BoolVal).Val {
		t.Fatal("main.aura should be a file")
	}

	// Step 4: List directory
	lr := listDirFn([]Value{&StringVal{Val: "/project"}}).(*ResultVal)
	list := lr.Val.(*ListVal)
	if len(list.Elements) != 2 { // README.md, src
		t.Fatalf("expected 2 entries in /project, got %d", len(list.Elements))
	}

	// Step 5: Append to a file
	appendFn([]Value{&StringVal{Val: "/project/README.md"}, &StringVal{Val: "\nA great project.\n"}})
	content := readFn([]Value{&StringVal{Val: "/project/README.md"}}).(*ResultVal).Val.(*StringVal).Val
	if content != "# My Project\n\nA great project.\n" {
		t.Fatalf("unexpected readme content: %q", content)
	}

	// Step 6: Delete a file
	deleteFn([]Value{&StringVal{Val: "/project/src/utils.aura"}})
	if existsFn([]Value{&StringVal{Val: "/project/src/utils.aura"}}).(*BoolVal).Val {
		t.Fatal("utils.aura should be deleted")
	}

	// Step 7: Verify remaining files
	lr = listDirFn([]Value{&StringVal{Val: "/project/src"}}).(*ResultVal)
	list = lr.Val.(*ListVal)
	if len(list.Elements) != 1 {
		t.Fatalf("expected 1 entry in /project/src after delete, got %d", len(list.Elements))
	}
	if list.Elements[0].(*StringVal).Val != "main.aura" {
		t.Fatalf("expected main.aura, got %q", list.Elements[0].String())
	}
}
