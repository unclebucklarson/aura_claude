# Aura Development Guide

This document covers everything you need to know to contribute to the Aura toolchain.

> 🤖 **Aura is an AI-first language.** All development decisions should be evaluated through the lens of AI-human collaboration. See [AI_MISSION.md](AI_MISSION.md) for the full mission statement.

---

## Designing for AI Developers

Aura's primary audience is **AI agents generating and reasoning about code**, with human developers as reviewers and collaborators. Every feature, API, and error message should be designed with this in mind.

### Design Decision Framework

When faced with trade-offs, apply this priority order:

1. **AI flow** — Does this make AI code generation faster and more accurate?
2. **Compiler verifiability** — Can the compiler check this automatically?
3. **Human readability** — Is this clear for human review?
4. **Brevity** — Is this concise? (Lowest priority — clarity always wins over conciseness)

### Code Review Checklist: AI-First Design

When reviewing PRs or designing features, ask these questions:

- [ ] **Does this feature help AI understand intent?** — Can an AI read the syntax/output and know exactly what to do without surrounding context?
- [ ] **Is the representation structured?** — Prefer structured data (spec blocks, typed annotations) over freeform text (comments, naming conventions).
- [ ] **Are error messages machine-parseable?** — Error output should include structured fields (error code, location, expected vs actual) that AI agents can parse and act on automatically.
- [ ] **Does this integrate with specs?** — Every new feature should consider how it interacts with the specification system. Can specs reference it? Can the compiler validate it?
- [ ] **Are effects explicit?** — If a feature introduces side effects, are they tracked in the effect system?
- [ ] **Is it deterministic?** — Given the same input, does the feature always produce the same output? AI agents depend on deterministic behavior.

### Testing: AI Code Generation Scenarios

When writing tests for new features, include scenarios that validate AI-relevant use cases:

- **Spec-to-implementation validation** — Test that code satisfying a spec actually passes all spec checks.
- **Round-trip stability** — AI-generated code, when formatted, should be identical to human-written canonical form.
- **Error message quality** — Test that error messages include enough information for an AI to fix the issue automatically (error code, location, suggestion).
- **Effect tracking accuracy** — Test that the effect system correctly identifies all effects, especially for complex call graphs that AI might generate.
- **Edge cases from AI generation** — AI may produce valid but unusual code patterns. Test that these are handled correctly (e.g., deeply nested expressions, max-length identifiers, unusual but valid type combinations).

---

## Architecture Overview

### Pipeline

Aura source code flows through the toolchain in stages:

```
Source (.aura)
    │
    ▼
┌──────────┐     ┌──────────┐     ┌────────────┐     ┌───────────┐
│  Lexer   │────▶│  Parser  │────▶│  Semantic  │────▶│  CodeGen  │
│ (tokens) │     │  (AST)   │     │  Analysis  │     │ (output)  │
└──────────┘     └──────────┘     └────────────┘     └───────────┘
    │                 │                 │                   │
    ▼                 ▼                 ▼                   ▼
  Token stream    Raw AST         Typed AST           Executable
                      │
                      ▼
                ┌────────────┐
                │ Formatter  │
                │ (source)   │
                └────────────┘
```

### Package Layout

```
aura-toolchain/
├── cmd/
│   └── aura/
│       └── main.go              # CLI entry point
├── pkg/
│   ├── token/
│   │   └── token.go             # Token types, positions, spans
│   ├── lexer/
│   │   ├── lexer.go             # Indentation-sensitive lexer
│   │   └── lexer_test.go        # 11 tests
│   ├── ast/
│   │   └── ast.go               # Complete AST node definitions
│   ├── parser/
│   │   ├── parser.go            # Recursive descent parser
│   │   └── parser_test.go       # 16 tests
│   ├── formatter/
│   │   ├── formatter.go         # AST → canonical source
│   │   └── formatter_test.go    # 9 tests (incl. round-trip)
│   ├── symbols/
│   │   ├── symbols.go           # Symbol table & scope management
│   │   └── symbols_test.go      # 9 tests
│   ├── types/
│   │   ├── types.go             # Type system representation
│   │   └── types_test.go        # 26 tests
│   ├── checker/
│   │   ├── checker.go           # Multi-pass type checker
│   │   ├── errors.go            # AI-parseable structured errors
│   │   └── checker_test.go      # 48 tests
│   ├── module/                  # [Phase 4.2] Module system ✅
│   │   └── resolver.go          # Import resolution, cycle detection, init ordering
│   └── interpreter/             # [Phase 3+4] Interpreter + Runtime + Stdlib + Effects ✅
│       └── value.go             # Value system (Int, Float, String, Bool, etc.)
│       └── env.go               # Environment with scope chain
│       └── eval.go              # Expression/statement evaluator
│       └── interpreter.go       # Module execution, builtins & stdlib registration
│       └── test.go              # Test runner
│       └── methods.go           # Method dispatch registry
│       └── methods_string.go    # 22 String methods
│       └── methods_list.go      # 27 List methods + helpers
│       └── methods_map.go       # 24 Map methods
│       └── methods_option.go    # 17 Option + 18 Result methods
│       └── effect.go            # Effect system: EffectContext, 5 providers
│       └── stdlib_*.go          # 16 stdlib module implementations
│       └── interpreter_test.go  # Core interpreter tests
│       └── methods_test.go      # 222 method tests
│       └── import_advanced_test.go  # 64 module system tests
│       └── stdlib_complete_test.go  # 65 stdlib tests
│       └── effect_test.go       # 48 effect foundation tests
│       └── time_env_test.go     # 66 time/env tests
│       └── effect_composition_test.go  # 54 composition tests
│       └── net_log_test.go      # 54 network/logging tests
├── testdata/                    # Sample .aura files
├── user_docs/                   # User-facing documentation
├── ROADMAP.md                   # Development roadmap
├── DEVELOPMENT.md               # This file
└── README.md                    # Project overview
```

### Key Design Decisions

1. **Go implementation** — Chosen for fast compilation, easy cross-compilation, and strong tooling support.
2. **Indentation-sensitive lexer** — The lexer emits INDENT/DEDENT tokens, so the parser never deals with whitespace directly.
3. **Paren-depth tracking** — Inside `()`, `[]`, `{}`, newlines and indentation are suppressed. This allows multi-line expressions without explicit line continuation.
4. **Recursive descent parser** — Simple, predictable, and easy to extend. Operator precedence climbing handles expression parsing.
5. **Round-trip formatting** — The formatter produces deterministic output from any valid AST, ensuring `parse → format → parse → format` is stable.

---

## Getting Started

### Prerequisites

- Go 1.22 or later
- Git

### Clone & Build

```bash
git clone https://github.com/unclebucklarson/aura.git
cd aura
go build -o aura ./cmd/aura
```

### Run Tests

```bash
# Run all tests
go test ./... -v

# Run tests for a specific package
go test ./pkg/lexer -v
go test ./pkg/parser -v
go test ./pkg/formatter -v
go test ./pkg/symbols -v
go test ./pkg/types -v
go test ./pkg/checker -v
go test ./pkg/interpreter -v

# Run with race detection
go test ./... -race

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### CLI Usage

```bash
# Format a file (print to stdout)
./aura format testdata/models.aura

# Format in-place
./aura format -w testdata/models.aura

# Parse and dump tokens + AST
./aura parse testdata/specs.aura

# Type-check a file (human-readable output)
./aura check testdata/service.aura

# Type-check with AI-parseable JSON output
./aura check --json testdata/service.aura

# Run an Aura program (executes main() function)
./aura run program.aura

# Run test blocks in a file
./aura test testdata/models.aura

# Start interactive REPL
./aura repl
```

---

## Implementation Checklists

### Phase 2: Semantic Analysis ✅ COMPLETE

> Implemented as three unified packages (`pkg/symbols`, `pkg/types`, `pkg/checker`) with 83 total tests.
> CLI: `aura check [--json] <file.aura>`

#### 2.1 Symbol Table & Scope Management (`pkg/symbols`) ✅

```
[x] Define Scope and Symbol types (Variable, Function, Type, Struct, Enum, Trait, Spec, etc.)
[x] Implement hierarchical scope chain (Module → Function → Block → Loop)
[x] Walk the AST to register all declarations
[x] Walk the AST to resolve all references
[x] Handle qualified name resolution (e.g., `TaskError.NotFound`)
[x] Report errors: undefined names, duplicate declarations
[x] Write tests (9 tests)
```

#### 2.2 Type System (`pkg/types`) ✅

```
[x] Define internal Type representations (19 TypeKinds: Primitive, Struct, Enum, Union, Function,
    Tuple, List, Map, Set, Option, Result, Refinement, StringLit, TypeParam, Never, Any, None, Alias, Intersection)
[x] Implement type registry with built-in types (Int, Float, String, Bool, None, Any, Never)
[x] Implement type equality (Equal)
[x] Implement subtyping rules (IsAssignableTo):
    - Never <: everything, everything <: Any
    - None <: Option[T], T <: Option[T]
    - Refinement <: base type, StringLit <: String
    - Union subtyping, struct width subtyping
    - Int → Float widening
[x] Constructors for all type kinds
[x] String() representation for error messages
[x] Write comprehensive tests (26 tests)
```

#### 2.3 Type Checker (`pkg/checker`) ✅

```
[x] Multi-pass architecture:
    Pass 1: Register types (struct, enum, type aliases)
    Pass 2: Register spec blocks
    Pass 3: Register functions
    Pass 4: Register constants
    Pass 5: Check function bodies
    Pass 6: Validate spec contracts
    Pass 7: Check test blocks
[x] Type-check literals (Int, Float, String, Bool, None)
[x] Type-check binary operators (arithmetic, comparison, logical, string concat)
[x] Type-check unary operators (negation, not)
[x] Type-check function calls (argument count, return types)
[x] Type-check field access on structs
[x] Type-check index expressions on lists and maps
[x] Type-check struct construction (field validation)
[x] Type-check pattern matching with enum exhaustiveness checking
[x] Type-check list comprehensions (element type + iterator variable inference)
[x] Type-check if/elif/else, for, while, match statements
[x] Type-check break/continue (loop context validation)
[x] Type-check return statements (against function return type)
[x] Type-check with statements (effect capability injection)
[x] Type-check assert statements
[x] Effect tracking and validation (db, net, fs, time, random, auth, log)
[x] Spec validation (input names/types, effects subset check)
[x] Built-in constructors (Ok, Err, Some)
[x] Variable type tracking across scopes
[x] AI-parseable structured errors (18 error codes, JSON output, suggested fixes)
[x] Write comprehensive tests (48 tests)
```

#### Future Enhancements (deferred to later phases)

```
[ ] Import resolution across modules
[ ] Refinement predicate evaluation at compile time
[ ] Generic type parameter instantiation
[ ] Transitive effect closure via call graph
[ ] Lambda parameter type inference from context
[ ] ? propagation operator type checking
```

### Phase 3: Tree-Walk Interpreter ✅ COMPLETE

#### 3.1 Tree-Walk Interpreter (`pkg/interpreter`) — 112 tests

```
[x] Value system: IntVal, FloatVal, StringVal, BoolVal, NoneVal, ListVal, MapVal,
    SetVal, TupleVal, StructVal, EnumVal, FunctionVal, LambdaVal, BuiltinFnVal,
    OptionVal (Some/None), ResultVal (Ok/Err)
[x] Environment with scope chain (parent lookup, const/mutable tracking)
[x] Expression evaluation:
    [x] Literals (int, float, string, bool, none)
    [x] Identifiers (variable lookup)
    [x] Binary operators (+, -, *, /, %, **, ==, !=, <, >, <=, >=, and, or)
    [x] Unary operators (-, not)
    [x] Function calls (user-defined, builtins, lambdas)
    [x] Field access (structs, enums)
    [x] Index access (lists, maps, negative indexing)
    [x] Struct construction
    [x] List literals and list comprehensions (with filter)
    [x] Map literals
    [x] Lambda expressions (|params| -> expr)
    [x] If expressions (if/then/else)
    [x] String concatenation
[x] Statement execution:
    [x] Let bindings (mutable/immutable)
    [x] Assignment
    [x] Return
    [x] If/elif/else
    [x] Match/case (literals, bindings, wildcards)
    [x] For loops (with range)
    [x] While loops
    [x] Break/continue
    [x] Assert
    [x] Expression statements
    [x] With blocks (effect capabilities)
[x] Function definition and calling convention
[x] Closure support (capturing enclosing environment)
[x] Builtins: print, len, str, int, float, range, type_of, abs, min, max,
    Ok, Err, Some, None
[x] Test block runner (RunTests, FormatTestResults)
[x] CLI integration: run, test, repl commands
[x] 112 comprehensive tests
```

#### Recently Completed (v0.3.1)
```
[x] String interpolation with full expression support
[x] Pipeline operator (|>) evaluation
[x] Option chaining (?.) with None short-circuiting
```

#### Deferred to future phases
```
[ ] Effect capability enforcement at runtime
[ ] Import/module resolution
```

### Phase 4.1: Core Runtime Methods ✅ COMPLETE

> Implemented 108+ methods across 5 core types via centralized method dispatch registry.
> 222 new tests in `methods_test.go` — 468 total tests across all packages.

#### Method Dispatch Infrastructure (`methods.go`)
```
[x] RegisterMethod(ValueType, "name", func) registry system
[x] LookupMethod() and resolveMethod() for dispatch
[x] Integration with eval.go FieldAccess evaluation
[x] callValue() helper for invoking Aura functions from Go
[x] cmpValues() helper for type-safe ordering
```

#### String Methods (`methods_string.go`) — 22 methods
```
[x] Core: len, upper/to_upper, lower/to_lower, contains, split, trim, trim_start, trim_end
[x] Search: starts_with, ends_with, index_of (→ Option), replace
[x] Transform: repeat, reverse, chars, slice (with bounds checking)
[x] Aliases: length → len
```

#### List Methods (`methods_list.go`) — 27 methods
```
[x] Core: len/length, append/push, contains, is_empty
[x] Safe accessors: first, last, get (all → Option)
[x] Mutation: pop (→ Option), remove
[x] Transform: reverse, slice, join, index_of (→ Option)
[x] Higher-order: map, filter, reduce, for_each, flat_map, flatten
[x] Predicates: any, all, count
[x] Utilities: unique, sum, min/max (→ Option), sort, zip, enumerate
```

#### Map Methods (`methods_map.go`) — 24 methods
```
[x] Size: len/length/size, is_empty
[x] Access: keys, values, entries, get (→ Option), get_or, has/contains_key, contains_value
[x] Mutation: set, remove (→ Option), delete (→ Bool), clear, merge
[x] Higher-order: filter, map, for_each, reduce, any, all, count
[x] Utilities: to_list, find (→ Option)
```

#### Option Methods (`methods_option.go`) — 17 methods
```
[x] Predicates: is_some, is_none
[x] Extraction: unwrap, expect, unwrap_or, unwrap_or_else
[x] Monadic: map, flat_map, and_then, filter, flatten
[x] Combinators: or_else, or, and, zip
[x] Query/conversion: contains, to_result
```

#### Result Methods (`methods_option.go`) — 18 methods
```
[x] Predicates: is_ok, is_err
[x] Extraction: unwrap, unwrap_err, expect, unwrap_or, unwrap_or_else
[x] Monadic: map, map_err, and_then, or_else, flatten
[x] Combinators: or, and
[x] Query/conversion: contains, contains_err, ok, err, to_option
```

### Phase 4.2: Module System & Standard Library ✅ COMPLETE

> Implemented complete module/import system and 12 pure computation stdlib modules with 70 functions.
> 146 new tests — 614 total tests across all packages.

#### Module System (`pkg/module/resolver.go`)
```
[x] Import resolution with namespace management
[x] Named imports and aliasing
[x] Import cycle detection with path reporting
[x] Module initialization ordering
[x] Deep dependency chain handling
[x] 17 tests
```

#### Standard Library Modules (12 pure computation modules)
```
[x] std.math — 8 functions + constants (pi, e, inf, nan)
[x] std.string — 4 functions (join, split, replace, repeat)
[x] std.io — 3 functions (print, println, format)
[x] std.testing — 10 base functions (assert, assert_eq, etc.)
[x] std.json — 2 functions (parse, stringify with pretty-print)
[x] std.regex — 6 functions (match, find, find_all, replace, split, compile)
[x] std.collections — 9 functions (range, zip_with, partition, group_by, etc.)
[x] std.random — 6 functions (int, float, choice, shuffle, sample, seed)
[x] std.format — 7 functions (pad_left, pad_right, center, truncate, wrap, indent, dedent)
[x] std.result — 5 functions (all_ok, any_ok, collect, partition_results, from_option)
[x] std.option — 5 functions (all_some, any_some, collect, first_some, from_result)
[x] std.iter — 5 functions (cycle, repeat, chain, interleave, pairwise)
```

### Phase 4.3: Effect Runtime ✅ COMPLETE

> Implemented complete effect system with 5 providers (Real + Mock), 34 effect-based stdlib functions,
> effect composition, and comprehensive mocking framework.
> 222 effect-related tests — 875 total tests across all packages.

#### Effect System Infrastructure (`effect.go`)
```
[x] EffectContext with provider pattern
[x] FileProvider (Real: os filesystem | Mock: in-memory)
[x] TimeProvider (Real: time package | Mock: controllable clock)
[x] EnvProvider (Real: os env vars | Mock: in-memory)
[x] NetProvider (Real: net/http | Mock: configurable responses)
[x] LogProvider (Real: stdout | Mock: in-memory storage)
[x] Clone, Derive, DeriveWithNetLog context manipulation
[x] EffectStack for nested effect scopes
[x] MockBuilder fluent API for test contexts
[x] Pre-configured fixtures (EmptyMockContext, FixtureWithFiles, etc.)
```

#### Effect-Based Standard Library (5 modules, 34 functions)
```
[x] std.file — 9 functions (read, write, append, exists, delete, list_dir, create_dir, is_file, is_dir)
[x] std.time — 8 functions (now, unix, millis, sleep, format, parse, add, diff)
[x] std.env — 6 functions (get, set, remove, has, all, args)
[x] std.net — 5 functions (get, post, put, delete, request)
[x] std.log — 6 functions (info, warn, error, debug, with_context, get_logs)
```

#### Effect-Aware Testing (`stdlib_testing.go` additions)
```
[x] with_mock_effects(fn) — execute in empty mock context
[x] with_effects(config, fn) — execute with custom mock config
[x] assert_file_exists(path), assert_file_content(path, expected)
[x] assert_file_contains(path, substring), assert_no_file(path)
[x] assert_env_var(key, expected)
[x] mock_time(timestamp), advance_time(seconds), get_mock_time()
[x] reset_effects(), get_env(key)
```

---

## Testing Strategy

### Test Categories

1. **Unit tests** — Each package has `_test.go` files testing individual functions and components in isolation.

2. **Round-trip tests** — The formatter tests verify `source → parse → format → parse → format` stability. This catches bugs in both the parser and formatter.

3. **Integration tests** — Test the full pipeline from source to output. Located alongside the package tests or in a dedicated `integration/` directory.

4. **Testdata files** — The `testdata/` directory contains representative `.aura` files covering all language constructs:
   - `simple.aura` — Minimal: type, struct, enum, function
   - `models.aura` — Structs with refinement types, enums
   - `specs.aura` — Spec blocks with all sections
   - `service.aura` — Functions with effects, satisfies, complex bodies
   - `control_flow.aura` — if/elif/else, match, for, while
   - `expressions.aura` — Pipelines, comprehensions, lambdas, operators
   - `comments.aura` — Comment handling edge cases
   - `empty.aura` — Empty module edge case

### Writing Tests

- Use table-driven tests where appropriate
- Test both success cases and error cases
- Include source location validation in error tests
- Use `testdata/` files for integration-level tests
- Aim for >80% code coverage on new packages

### Running Tests

```bash
# All tests, verbose
go test ./... -v

# Specific package
go test ./pkg/parser -v -run TestMatchStatement

# With coverage report
go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out
```

---

## Code Organization Guidelines

### Package Responsibilities

Each package should have a **single, clear responsibility**:

- `token` — Token type definitions only. No logic.
- `lexer` — Source text → token stream. No AST knowledge.
- `ast` — AST node type definitions. Minimal logic (just constructors and accessors).
- `parser` — Token stream → AST. No type checking or validation.
- `formatter` — AST → canonical source text. No parsing.
- `symbols` — Symbol table with hierarchical scope management. Used by the checker.
- `types` — Type system representation, equality, subtyping, and registry. Used by the checker.
- `checker` — Multi-pass type checker integrating symbols, types, and effect tracking. Depends on parser output.
- `module` — Module resolution, import cycle detection, initialization ordering. Used by the interpreter. 17 tests.
- `interpreter` — Tree-walk interpreter with complete runtime. Evaluates AST directly: value system, environment/scope chain, expression/statement evaluation, builtins, test runner, string interpolation, pipeline operator, option chaining. Includes method dispatch registry with 108+ built-in methods, 17 stdlib modules with 117 functions, and effect system with 5 mockable providers (File, Time, Env, Net, Log). Depends on `ast`, `token`, `lexer`, `parser`, `module`. 738 tests.

### Naming Conventions

- Go standard conventions: `CamelCase` for exports, `camelCase` for unexported
- AST node types: `PascalCase` matching the grammar (e.g., `StructDef`, `LetStmt`, `BinaryOp`)
- Test functions: `TestComponentName_scenario` (e.g., `TestParser_MatchStatement`)
- Error types: `ErrorCategory` prefix (e.g., `ResolveError`, `TypeError`, `EffectError`)

### Error Reporting

All errors should include:
- Source file path
- Line and column number
- Error code (e.g., `E101`, `E201`, `W301`)
- Clear message describing the problem
- Suggestion for fixing (where possible)

Use the `token.Span` from AST nodes to generate precise error locations.

---

## How to Contribute

### Workflow

1. **Check the roadmap** — Pick a task from [ROADMAP.md](ROADMAP.md)
2. **Create a branch** — `git checkout -b feature/your-feature`
3. **Implement** — Write code following the guidelines above
4. **Test** — Add tests and ensure all existing tests pass (`go test ./...`)
5. **Format** — Run `gofmt -w .` to format Go code
6. **Commit** — Use conventional commit messages:
   - `feat: Add name resolution for imports`
   - `fix: Handle empty match bodies in parser`
   - `test: Add type checker tests for union types`
   - `docs: Update roadmap with Phase 2 progress`
7. **Push & PR** — Push your branch and open a pull request

### Commit Message Format

```
<type>: <short description>

<optional longer description>

<optional references>
```

Types: `feat`, `fix`, `test`, `docs`, `refactor`, `chore`

### Before Submitting a PR

- [ ] All tests pass: `go test ./...`
- [ ] No race conditions: `go test ./... -race`
- [ ] Code is formatted: `gofmt -l .` (should output nothing)
- [ ] New code has tests
- [ ] Commit messages follow conventions
- [ ] ROADMAP.md is updated if a task is completed
