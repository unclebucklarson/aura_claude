// Package pkgmgr implements the Aura package manifest format (aura.pkg)
// and the operations that read, write, and apply it to the module resolver.
//
// Manifest format (aura.pkg):
//
//	# comment
//	name    = mypackage
//	version = 0.1.0
//	author  = Someone
//
//	[deps]
//	mathlib = ../mathlib
//	utils   = /abs/path/to/utils
//
// Local-path dependencies only in this version.
package pkgmgr

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// manifestFileName is the conventional name for the package manifest.
const manifestFileName = "aura.pkg"

// reAlias validates dependency alias names: letters/digits/underscore, must start with letter.
var reAlias = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

// Manifest represents the contents of an aura.pkg file.
type Manifest struct {
	// Dir is the absolute directory where aura.pkg lives. Set by Load/Find; not written.
	Dir string

	// Name is the package name (required).
	Name string

	// Version is the package version string (required).
	Version string

	// Meta holds additional top-level key=value entries in parsed order (author, description, etc.)
	// Preserved for round-trip writing.
	Meta []MetaEntry

	// Deps is the ordered list of [deps] entries.
	Deps []Dep
}

// MetaEntry is a single top-level key=value pair that is not Name or Version.
type MetaEntry struct {
	Key   string
	Value string
}

// Dep is a single entry in the [deps] section.
type Dep struct {
	// Alias is the import alias used in Aura source files (e.g., "mathlib").
	Alias string
	// Path is the resolved absolute path to the dep's root directory.
	Path string
	// RawPath is exactly what was written in aura.pkg (preserved for round-trip writing).
	RawPath string
}

// Find walks up the directory tree from startDir looking for aura.pkg.
// Returns the absolute path to the manifest file, or ("", nil) if not found.
// Returns a non-nil error only on unexpected filesystem failures.
func Find(startDir string) (string, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	dir := abs
	for {
		candidate := filepath.Join(dir, manifestFileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root — not found.
			return "", nil
		}
		dir = parent
	}
}

// Load reads and parses the aura.pkg file at the given absolute path.
// Sets Manifest.Dir to filepath.Dir(manifestPath).
// Dep paths are resolved relative to Manifest.Dir at parse time.
func Load(manifestPath string) (*Manifest, error) {
	f, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", manifestPath, err)
	}
	defer f.Close()

	m := &Manifest{Dir: filepath.Dir(manifestPath)}

	inDeps := false
	lineNum := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineNum++
		line := strings.TrimRight(scanner.Text(), "\r\n \t")

		// Skip blank lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section header.
		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return nil, fmt.Errorf("%s:%d: malformed section header %q", manifestPath, lineNum, line)
			}
			section := line[1 : len(line)-1]
			switch section {
			case "deps":
				inDeps = true
			default:
				return nil, fmt.Errorf("%s:%d: unknown section [%s]", manifestPath, lineNum, section)
			}
			continue
		}

		// Key = value.
		eq := strings.IndexByte(line, '=')
		if eq < 0 {
			return nil, fmt.Errorf("%s:%d: expected key = value, got %q", manifestPath, lineNum, line)
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])

		if inDeps {
			if !reAlias.MatchString(key) {
				return nil, fmt.Errorf("%s:%d: invalid dep alias %q (must match [a-zA-Z][a-zA-Z0-9_]*)", manifestPath, lineNum, key)
			}
			abs, err := resolvePath(val, m.Dir)
			if err != nil {
				return nil, fmt.Errorf("%s:%d: dep %q: %w", manifestPath, lineNum, key, err)
			}
			m.Deps = append(m.Deps, Dep{Alias: key, Path: abs, RawPath: val})
		} else {
			switch key {
			case "name":
				m.Name = val
			case "version":
				m.Version = val
			default:
				m.Meta = append(m.Meta, MetaEntry{Key: key, Value: val})
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", manifestPath, err)
	}

	if m.Name == "" {
		return nil, fmt.Errorf("%s: missing required field \"name\"", manifestPath)
	}
	return m, nil
}

// FindAndLoad combines Find and Load.
// Returns (nil, nil) if no manifest exists anywhere in the directory tree.
func FindAndLoad(startDir string) (*Manifest, error) {
	path, err := Find(startDir)
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, nil
	}
	return Load(path)
}

// Write serializes m to aura.pkg format and writes it to filepath.Join(m.Dir, "aura.pkg").
func Write(m *Manifest) error {
	var sb strings.Builder
	sb.WriteString("# Aura package manifest\n\n")
	sb.WriteString("name    = " + m.Name + "\n")
	sb.WriteString("version = " + m.Version + "\n")
	for _, me := range m.Meta {
		sb.WriteString(me.Key + " = " + me.Value + "\n")
	}
	if len(m.Deps) > 0 {
		sb.WriteString("\n[deps]\n")
		for _, d := range m.Deps {
			sb.WriteString(d.Alias + " = " + d.RawPath + "\n")
		}
	}
	dest := filepath.Join(m.Dir, manifestFileName)
	return os.WriteFile(dest, []byte(sb.String()), 0o644)
}

// Init creates a new minimal aura.pkg in dir with the given package name.
// Errors if aura.pkg already exists in dir (not higher up — only in dir itself).
func Init(dir, name string) error {
	if name == "" {
		return fmt.Errorf("package name must not be empty")
	}
	if !reAlias.MatchString(name) {
		return fmt.Errorf("invalid package name %q (must match [a-zA-Z][a-zA-Z0-9_]*)", name)
	}
	dest := filepath.Join(dir, manifestFileName)
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("aura.pkg already exists in %s", dir)
	}
	m := &Manifest{
		Dir:     dir,
		Name:    name,
		Version: "0.1.0",
	}
	return Write(m)
}

// AddDep adds or updates a dependency entry in the manifest.
// depPath is the raw path as provided by the user; it is resolved relative to cwd
// before verification. RawPath stores the original value for portable manifests.
// If alias already exists, its path is updated.
func AddDep(m *Manifest, alias, depPath string) error {
	if !reAlias.MatchString(alias) {
		return fmt.Errorf("invalid dep alias %q (must match [a-zA-Z][a-zA-Z0-9_]*)", alias)
	}
	abs, err := filepath.Abs(depPath)
	if err != nil {
		return fmt.Errorf("cannot resolve path %q: %w", depPath, err)
	}
	fi, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("dep path %q does not exist: %w", abs, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("dep path %q must be a directory, not a file", abs)
	}
	// Update existing entry or append.
	for i, d := range m.Deps {
		if d.Alias == alias {
			m.Deps[i] = Dep{Alias: alias, Path: abs, RawPath: depPath}
			return nil
		}
	}
	m.Deps = append(m.Deps, Dep{Alias: alias, Path: abs, RawPath: depPath})
	return nil
}

// ApplyToResolver calls resolver.AddSearchPath for each dep in m.
func ApplyToResolver(m *Manifest, resolver interface{ AddSearchPath(string) }) {
	for _, d := range m.Deps {
		resolver.AddSearchPath(d.Path)
	}
}

// resolvePath resolves a path relative to baseDir, returning an absolute path.
// Absolute paths are returned as-is.
func resolvePath(p, baseDir string) (string, error) {
	if filepath.IsAbs(p) {
		return filepath.Clean(p), nil
	}
	return filepath.Abs(filepath.Join(baseDir, p))
}
