package pkgmgr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile writes content to path, creating the file.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// --- Load / parsing ---

func TestParseMinimalManifest(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "aura.pkg"), "name = mypackage\nversion = 0.1.0\n")
	m, err := Load(filepath.Join(dir, "aura.pkg"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m.Name != "mypackage" {
		t.Errorf("Name: got %q, want %q", m.Name, "mypackage")
	}
	if m.Version != "0.1.0" {
		t.Errorf("Version: got %q, want %q", m.Version, "0.1.0")
	}
	if len(m.Deps) != 0 {
		t.Errorf("Deps: got %d, want 0", len(m.Deps))
	}
	if len(m.Meta) != 0 {
		t.Errorf("Meta: got %d, want 0", len(m.Meta))
	}
	if m.Dir != dir {
		t.Errorf("Dir: got %q, want %q", m.Dir, dir)
	}
}

func TestParseFullManifest(t *testing.T) {
	dir := t.TempDir()
	depA := t.TempDir()
	depB := t.TempDir()

	content := "name = myproject\nversion = 1.2.3\nauthor = Alice\n\n[deps]\nmathlib = " + depA + "\nutils = " + depB + "\n"
	writeFile(t, filepath.Join(dir, "aura.pkg"), content)

	m, err := Load(filepath.Join(dir, "aura.pkg"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m.Name != "myproject" {
		t.Errorf("Name: %q", m.Name)
	}
	if m.Version != "1.2.3" {
		t.Errorf("Version: %q", m.Version)
	}
	if len(m.Meta) != 1 || m.Meta[0].Key != "author" || m.Meta[0].Value != "Alice" {
		t.Errorf("Meta: %v", m.Meta)
	}
	if len(m.Deps) != 2 {
		t.Fatalf("Deps: got %d, want 2", len(m.Deps))
	}
	if m.Deps[0].Alias != "mathlib" || m.Deps[0].Path != depA {
		t.Errorf("Deps[0]: %+v", m.Deps[0])
	}
	if m.Deps[1].Alias != "utils" || m.Deps[1].Path != depB {
		t.Errorf("Deps[1]: %+v", m.Deps[1])
	}
}

func TestParseRelativeDep(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "mypkg")
	depDir := filepath.Join(root, "mylib")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(depDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "name = mypkg\nversion = 0.1.0\n\n[deps]\nmylib = ../mylib\n"
	writeFile(t, filepath.Join(pkgDir, "aura.pkg"), content)

	m, err := Load(filepath.Join(pkgDir, "aura.pkg"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(m.Deps) != 1 {
		t.Fatalf("expected 1 dep, got %d", len(m.Deps))
	}
	if m.Deps[0].Path != depDir {
		t.Errorf("resolved path: got %q, want %q", m.Deps[0].Path, depDir)
	}
	if m.Deps[0].RawPath != "../mylib" {
		t.Errorf("raw path: got %q, want %q", m.Deps[0].RawPath, "../mylib")
	}
}

func TestParseManifestMissingName(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "aura.pkg"), "version = 0.1.0\n")
	_, err := Load(filepath.Join(dir, "aura.pkg"))
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention 'name', got: %v", err)
	}
}

func TestParseManifestUnknownSection(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "aura.pkg"), "name = x\nversion = 0.1.0\n\n[unknown]\nfoo = bar\n")
	_, err := Load(filepath.Join(dir, "aura.pkg"))
	if err == nil {
		t.Fatal("expected error for unknown section, got nil")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention 'unknown', got: %v", err)
	}
}

func TestParseCommentsAndBlanks(t *testing.T) {
	dir := t.TempDir()
	content := "# top comment\n\nname = mypkg\n# another comment\nversion = 0.2.0\n\n"
	writeFile(t, filepath.Join(dir, "aura.pkg"), content)
	m, err := Load(filepath.Join(dir, "aura.pkg"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m.Name != "mypkg" || m.Version != "0.2.0" {
		t.Errorf("Name=%q Version=%q", m.Name, m.Version)
	}
}

// --- Find ---

func TestFindManifestWalksUp(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "aura.pkg"), "name = root\nversion = 0.1.0\n")

	found, err := Find(child)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if found != filepath.Join(root, "aura.pkg") {
		t.Errorf("Find: got %q, want %q", found, filepath.Join(root, "aura.pkg"))
	}
}

func TestFindManifestNotFound(t *testing.T) {
	// Use a fresh temp dir with no aura.pkg anywhere above it.
	// (t.TempDir() is inside /tmp which has no aura.pkg)
	dir := t.TempDir()
	found, err := Find(dir)
	if err != nil {
		t.Fatalf("Find: unexpected error: %v", err)
	}
	if found != "" {
		t.Errorf("Find: expected empty string, got %q", found)
	}
}

// --- Write and round-trip ---

func TestWriteRoundTrip(t *testing.T) {
	dir := t.TempDir()
	depDir := t.TempDir()

	m := &Manifest{
		Dir:     dir,
		Name:    "roundtrip",
		Version: "1.0.0",
		Meta:    []MetaEntry{{Key: "author", Value: "Bob"}},
		Deps:    []Dep{{Alias: "mylib", Path: depDir, RawPath: depDir}},
	}
	if err := Write(m); err != nil {
		t.Fatalf("Write: %v", err)
	}

	m2, err := Load(filepath.Join(dir, "aura.pkg"))
	if err != nil {
		t.Fatalf("Load after Write: %v", err)
	}
	if m2.Name != "roundtrip" {
		t.Errorf("Name: %q", m2.Name)
	}
	if m2.Version != "1.0.0" {
		t.Errorf("Version: %q", m2.Version)
	}
	if len(m2.Meta) != 1 || m2.Meta[0].Key != "author" {
		t.Errorf("Meta: %v", m2.Meta)
	}
	if len(m2.Deps) != 1 || m2.Deps[0].Alias != "mylib" {
		t.Errorf("Deps: %v", m2.Deps)
	}
}

// --- Init ---

func TestInitCreatesFile(t *testing.T) {
	dir := t.TempDir()
	if err := Init(dir, "newpkg"); err != nil {
		t.Fatalf("Init: %v", err)
	}
	m, err := Load(filepath.Join(dir, "aura.pkg"))
	if err != nil {
		t.Fatalf("Load after Init: %v", err)
	}
	if m.Name != "newpkg" {
		t.Errorf("Name: %q", m.Name)
	}
	if m.Version != "0.1.0" {
		t.Errorf("Version: %q", m.Version)
	}
}

func TestInitErrorIfExists(t *testing.T) {
	dir := t.TempDir()
	if err := Init(dir, "first"); err != nil {
		t.Fatalf("first Init: %v", err)
	}
	if err := Init(dir, "second"); err == nil {
		t.Fatal("expected error on second Init, got nil")
	}
}

func TestInitInvalidName(t *testing.T) {
	dir := t.TempDir()
	if err := Init(dir, "bad-name!"); err == nil {
		t.Fatal("expected error for invalid name, got nil")
	}
}

// --- AddDep ---

func TestAddDepAppendsNew(t *testing.T) {
	dir := t.TempDir()
	depDir := t.TempDir()
	m := &Manifest{Dir: dir, Name: "mypkg", Version: "0.1.0"}
	if err := AddDep(m, "mylib", depDir); err != nil {
		t.Fatalf("AddDep: %v", err)
	}
	if len(m.Deps) != 1 || m.Deps[0].Alias != "mylib" {
		t.Errorf("Deps: %v", m.Deps)
	}
}

func TestAddDepUpdatesExisting(t *testing.T) {
	dir := t.TempDir()
	depA := t.TempDir()
	depB := t.TempDir()
	m := &Manifest{
		Dir: dir, Name: "mypkg", Version: "0.1.0",
		Deps: []Dep{{Alias: "mylib", Path: depA, RawPath: depA}},
	}
	if err := AddDep(m, "mylib", depB); err != nil {
		t.Fatalf("AddDep update: %v", err)
	}
	if len(m.Deps) != 1 {
		t.Fatalf("expected 1 dep after update, got %d", len(m.Deps))
	}
	if m.Deps[0].Path != depB {
		t.Errorf("path not updated: %q", m.Deps[0].Path)
	}
}

func TestAddDepInvalidAlias(t *testing.T) {
	dir := t.TempDir()
	depDir := t.TempDir()
	m := &Manifest{Dir: dir, Name: "mypkg", Version: "0.1.0"}
	if err := AddDep(m, "bad alias!", depDir); err == nil {
		t.Fatal("expected error for invalid alias, got nil")
	}
}

func TestAddDepPathMustBeDir(t *testing.T) {
	dir := t.TempDir()
	// Create a file, not a directory
	filePath := filepath.Join(dir, "somefile.aura")
	writeFile(t, filePath, "module x\n")
	m := &Manifest{Dir: dir, Name: "mypkg", Version: "0.1.0"}
	if err := AddDep(m, "x", filePath); err == nil {
		t.Fatal("expected error for file path, got nil")
	}
}

// --- ApplyToResolver ---

type fakeResolver struct {
	paths []string
}

func (r *fakeResolver) AddSearchPath(p string) {
	r.paths = append(r.paths, p)
}

func TestApplyToResolver(t *testing.T) {
	depA := t.TempDir()
	depB := t.TempDir()
	m := &Manifest{
		Deps: []Dep{
			{Alias: "a", Path: depA},
			{Alias: "b", Path: depB},
		},
	}
	r := &fakeResolver{}
	ApplyToResolver(m, r)
	if len(r.paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(r.paths))
	}
	if r.paths[0] != depA || r.paths[1] != depB {
		t.Errorf("paths: %v", r.paths)
	}
}
