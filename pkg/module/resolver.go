// Package module implements module resolution and loading for the Aura language.
package module

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/unclebucklarson/aura/pkg/ast"
	"github.com/unclebucklarson/aura/pkg/lexer"
	"github.com/unclebucklarson/aura/pkg/parser"
)

// InitState tracks the initialization state of a module.
type InitState int

const (
	InitNone       InitState = iota // Not yet initialized
	InitInProgress                  // Currently being initialized
	InitComplete                    // Initialization done
	InitError                       // Initialization failed
)

// Resolver handles module path resolution and caching.
type Resolver struct {
	// searchPaths are directories to search for modules
	searchPaths []string
	// cache stores already-parsed modules by their resolved path
	cache map[string]*CachedModule
	// loading tracks modules currently being loaded (circular dependency detection)
	loading map[string]bool
	// loadStack tracks the import chain for better error messages
	loadStack []string
	// initState tracks initialization state per module
	initState map[string]InitState
	mu        sync.Mutex
}

// CachedModule represents a parsed and cached module.
type CachedModule struct {
	Path    string          // resolved file path
	AST     *ast.Module     // parsed AST
	Exports map[string]bool // exported symbol names
}

// ResolveError represents an error during module resolution.
type ResolveError struct {
	Message string
	Path    string
}

func (e *ResolveError) Error() string {
	return fmt.Sprintf("module resolution error: %s (path: %s)", e.Message, e.Path)
}

// NewResolver creates a new module resolver.
// basePath is the directory of the main/entry file.
func NewResolver(basePath string) *Resolver {
	return &Resolver{
		searchPaths: []string{basePath},
		cache:       make(map[string]*CachedModule),
		loading:     make(map[string]bool),
		loadStack:   nil,
		initState:   make(map[string]InitState),
	}
}

// AddSearchPath adds a directory to search for modules.
func (r *Resolver) AddSearchPath(path string) {
	r.searchPaths = append(r.searchPaths, path)
}

// Resolve resolves a module path to a CachedModule.
// importPath is like "utils", "./helpers", "../common", "std.testing"
// fromDir is the directory of the file doing the import.
func (r *Resolver) Resolve(importPath string, fromDir string) (*CachedModule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Resolve the file path
	filePath, err := r.resolveFilePath(importPath, fromDir)
	if err != nil {
		return nil, err
	}

	// Check cache
	if cached, ok := r.cache[filePath]; ok {
		return cached, nil
	}

	// Check for circular dependency with detailed path
	if r.loading[filePath] {
		cyclePath := r.buildCyclePath(filePath)
		return nil, &ResolveError{
			Message: fmt.Sprintf("circular dependency detected: %s", cyclePath),
			Path:    importPath,
		}
	}

	// Mark as loading and push to stack
	r.loading[filePath] = true
	r.loadStack = append(r.loadStack, filePath)
	defer func() {
		delete(r.loading, filePath)
		if len(r.loadStack) > 0 {
			r.loadStack = r.loadStack[:len(r.loadStack)-1]
		}
	}()

	// Read and parse the file
	cached, err := r.loadModule(filePath)
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.cache[filePath] = cached
	return cached, nil
}

// buildCyclePath builds a human-readable cycle path string.
func (r *Resolver) buildCyclePath(target string) string {
	parts := make([]string, 0, len(r.loadStack)+1)
	inCycle := false
	for _, p := range r.loadStack {
		if p == target {
			inCycle = true
		}
		if inCycle {
			parts = append(parts, pathToName(p))
		}
	}
	parts = append(parts, pathToName(target))
	return strings.Join(parts, " -> ")
}

// pathToName extracts a readable name from a file path.
func pathToName(p string) string {
	if strings.HasPrefix(p, "@std/") {
		return strings.TrimPrefix(p, "@std/")
	}
	base := filepath.Base(p)
	return strings.TrimSuffix(base, ".aura")
}

// GetInitState returns the initialization state of a module.
func (r *Resolver) GetInitState(filePath string) InitState {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.initState[filePath]
}

// SetInitState sets the initialization state of a module.
func (r *Resolver) SetInitState(filePath string, state InitState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.initState[filePath] = state
}

// resolveFilePath converts an import path to an absolute file path.
func (r *Resolver) resolveFilePath(importPath string, fromDir string) (string, error) {
	parts := strings.Split(importPath, ".")

	// Handle standard library imports (std.*)
	if len(parts) > 0 && parts[0] == "std" {
		return r.resolveStdLib(importPath)
	}

	// Handle relative imports (starts with . or ..)
	if strings.HasPrefix(importPath, ".") {
		return r.resolveRelative(importPath, fromDir)
	}

	// Handle dotted paths as directory separators
	relPath := strings.Join(parts, string(filepath.Separator))

	// Search in all search paths
	for _, searchPath := range r.searchPaths {
		candidates := []string{
			filepath.Join(searchPath, relPath+".aura"),
			filepath.Join(searchPath, relPath, "mod.aura"),
		}
		for _, candidate := range candidates {
			absPath, err := filepath.Abs(candidate)
			if err != nil {
				continue
			}
			if _, err := os.Stat(absPath); err == nil {
				return absPath, nil
			}
		}
	}

	// Also search relative to fromDir
	candidates := []string{
		filepath.Join(fromDir, relPath+".aura"),
		filepath.Join(fromDir, relPath, "mod.aura"),
	}
	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath, nil
		}
	}

	return "", &ResolveError{
		Message: fmt.Sprintf("module not found: '%s'", importPath),
		Path:    importPath,
	}
}

// resolveRelative resolves a relative import path.
func (r *Resolver) resolveRelative(importPath string, fromDir string) (string, error) {
	// Convert dot-separated path to file path
	// "./helpers" -> "./helpers.aura"
	// "../common" -> "../common.aura"
	// "./utils.math" -> "./utils/math.aura"

	// Split by dots but preserve leading ./ or ../
	var prefix string
	rest := importPath

	if strings.HasPrefix(importPath, "../") {
		prefix = "../"
		rest = importPath[3:]
	} else if strings.HasPrefix(importPath, "./") {
		prefix = "./"
		rest = importPath[2:]
	} else if strings.HasPrefix(importPath, "..") {
		prefix = ".."
		rest = importPath[2:]
		if strings.HasPrefix(rest, ".") {
			rest = rest[1:]
		}
	} else if importPath == "." {
		prefix = "."
		rest = ""
	}

	// Convert remaining dots to path separators
	if rest != "" {
		parts := strings.Split(rest, ".")
		rest = strings.Join(parts, string(filepath.Separator))
	}

	relPath := prefix + rest

	// Try with .aura extension and as directory
	candidates := []string{
		filepath.Join(fromDir, relPath+".aura"),
		filepath.Join(fromDir, relPath, "mod.aura"),
	}

	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath, nil
		}
	}

	return "", &ResolveError{
		Message: fmt.Sprintf("module not found: '%s'", importPath),
		Path:    importPath,
	}
}

// resolveStdLib resolves a standard library import.
func (r *Resolver) resolveStdLib(importPath string) (string, error) {
	// std library modules are resolved specially
	// For now, return a virtual path that the interpreter will handle
	return "@std/" + importPath, nil
}

// loadModule reads and parses an Aura source file.
func (r *Resolver) loadModule(filePath string) (*CachedModule, error) {
	// Standard library modules are handled by the interpreter
	if strings.HasPrefix(filePath, "@std/") {
		return &CachedModule{
			Path:    filePath,
			AST:     nil, // stdlib modules have no AST; handled natively
			Exports: nil,
		}, nil
	}

	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, &ResolveError{
			Message: fmt.Sprintf("cannot read module file: %v", err),
			Path:    filePath,
		}
	}

	// Lex
	l := lexer.New(string(source), filePath)
	tokens, lexErrors := l.Tokenize()
	if len(lexErrors) > 0 {
		return nil, &ResolveError{
			Message: fmt.Sprintf("lexer errors in module: %v", lexErrors[0]),
			Path:    filePath,
		}
	}

	// Parse
	p := parser.New(tokens, filePath)
	moduleAST, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		return nil, &ResolveError{
			Message: fmt.Sprintf("parser errors in module: %v", parseErrors[0]),
			Path:    filePath,
		}
	}

	// Collect exports (pub items)
	exports := make(map[string]bool)
	for _, item := range moduleAST.Items {
		switch it := item.(type) {
		case *ast.FnDef:
			if it.Visibility == ast.Public {
				exports[it.Name] = true
			}
		case *ast.StructDef:
			if it.Visibility == ast.Public {
				exports[it.Name] = true
			}
		case *ast.EnumDef:
			if it.Visibility == ast.Public {
				exports[it.Name] = true
			}
		case *ast.ConstDef:
			if it.Visibility == ast.Public {
				exports[it.Name] = true
			}
		case *ast.TypeDef:
			if it.Visibility == ast.Public {
				exports[it.Name] = true
			}
		case *ast.TraitDef:
			if it.Visibility == ast.Public {
				exports[it.Name] = true
			}
		}
	}

	// Also export all top-level items without visibility (default export for simple modules)
	// This allows modules without explicit pub to still be imported
	if len(exports) == 0 {
		for _, item := range moduleAST.Items {
			switch it := item.(type) {
			case *ast.FnDef:
				exports[it.Name] = true
			case *ast.StructDef:
				exports[it.Name] = true
			case *ast.EnumDef:
				exports[it.Name] = true
			case *ast.ConstDef:
				exports[it.Name] = true
			case *ast.TypeDef:
				exports[it.Name] = true
			case *ast.TraitDef:
				exports[it.Name] = true
			}
		}
	}

	return &CachedModule{
		Path:    filePath,
		AST:     moduleAST,
		Exports: exports,
	}, nil
}

// IsStdLib returns true if the import path is a standard library module.
func IsStdLib(importPath string) bool {
	return strings.HasPrefix(importPath, "std.")
}

// GetModuleName extracts the module name from a qualified path.
// "std.testing" -> "testing"
// "utils.math" -> "math"
// "helpers" -> "helpers"
func GetModuleName(importPath string) string {
	parts := strings.Split(importPath, ".")
	return parts[len(parts)-1]
}

// IsCached returns true if a module is already cached.
func (r *Resolver) IsCached(importPath string, fromDir string) bool {
	filePath, err := r.resolveFilePath(importPath, fromDir)
	if err != nil {
		return false
	}
	_, ok := r.cache[filePath]
	return ok
}

// CacheCount returns the number of cached modules.
func (r *Resolver) CacheCount() int {
	return len(r.cache)
}

// GetDependencies extracts import paths from a module's AST.
func GetDependencies(mod *ast.Module) []string {
	if mod == nil || mod.Imports == nil {
		return nil
	}
	deps := make([]string, 0, len(mod.Imports))
	for _, imp := range mod.Imports {
		deps = append(deps, imp.Path.String())
	}
	return deps
}
