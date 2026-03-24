# Aura Language Roadmap

> A phased plan for building Aura from a working parser into a fully functional programming language.

> 🤖 **Aura is an AI-first language.** Every phase and feature in this roadmap is evaluated against a core question: **Does this make AI code generation and AI-human collaboration better?** See [AI_MISSION.md](AI_MISSION.md) for the full mission statement.

---

## AI-First Design Principles

These principles guide every phase of development. When evaluating features, trade-offs, or priorities, apply them in order:

1. **AI parseability first** — Can an AI agent read this feature's output and know exactly what to do? Structured, unambiguous representations always win.
2. **Machine-checkable contracts** — Every constraint, effect, and requirement should be verifiable by the compiler, not dependent on human review alone.
3. **Explicit over implicit** — If information exists (types, effects, error cases, constraints), it must be in the syntax. Hidden conventions are the enemy of AI code generation.
4. **Specs as the interface** — Specs are how humans communicate intent to AI. Every feature should consider: how does this interact with the spec system?
5. **Vibe coding flow** — The human writes *what* (specs), the AI writes *how* (implementation), the compiler validates *correctness*. Features should reinforce this loop.

**Feature evaluation checklist:**
- [ ] Does this feature help AI generate correct code faster?
- [ ] Is the feature's syntax unambiguous and machine-parseable?
- [ ] Does it reduce the need for AI to read surrounding context?
- [ ] Can the compiler validate it automatically?
- [ ] Does it integrate with specs and effects?

---

## Phase Overview

| Phase | Name | Status | Version | Estimated Effort |
|-------|------|--------|---------|------------------|
| 1 | Syntax (Lexer, Parser, Formatter) | ✅ COMPLETE | v0.1 | — |
| 2 | Semantic Analysis | ✅ COMPLETE | v0.2 | — |
| 3.1 | Tree-Walk Interpreter | ✅ COMPLETE | v0.3 | — |
| 3.2 | Pattern Matching (Advanced) | 🔲 Not Started | v0.9.0 | 2–3 weeks |
| 3.3 | Advanced Type Features | 🔲 Not Started | v1.0.0 | 3–4 weeks |
| 4 | Runtime & Standard Library | ✅ COMPLETE (4.1 ✅, 4.2 ✅, 4.3 ✅) | v0.8.0 | — |
| 5 | Advanced Tooling & Ecosystem | 🔲 Not Started | v1.1.0 | 4–6 weeks |
| 6 | Compiler & Native Compilation | 🔲 Not Started | v2.0.0 | 9–12 weeks |

---

## Phase 1: Syntax — ✅ COMPLETE

> 🤖 **AI optimization:** The parser and AST produce structured, unambiguous representations that AI agents can consume directly. The formatter ensures canonical output — AI-generated code always looks the same as human-written code.

The foundation of the Aura toolchain is fully implemented and tested.

### Deliverables

- [x] **Lexer** (`pkg/lexer`) — Indentation-sensitive tokenizer with INDENT/DEDENT, paren-depth tracking, comment handling
- [x] **Parser** (`pkg/parser`) — Recursive descent parser with operator precedence climbing
- [x] **AST** (`pkg/ast`) — Complete node definitions covering all language constructs
- [x] **Formatter** (`pkg/formatter`) — Canonical source formatter with round-trip guarantee
- [x] **CLI** (`cmd/aura`) — `format` and `parse` commands
- [x] **Test suite** — 36 tests across lexer (11), parser (16), formatter (9)

### Key Properties Verified

- Round-trip guarantee: `parse → format → parse → format` produces identical output
- All language constructs parse correctly: structs, enums, traits, impls, specs, functions, control flow, expressions
- Edge cases handled: empty files, blank lines between blocks, nested indentation

---

## Phase 2: Semantic Analysis — ✅ COMPLETE

**Goal:** Validate that parsed programs are meaningful — names resolve, types check, effects are tracked, and specs are verified.

> 🤖 **AI optimization:** This phase is critical for AI code generation. Type checking, effect validation, and spec verification give AI agents **immediate, automated feedback** on whether generated code is correct. Every error message is structured and JSON-serializable for AI to parse and fix automatically.

**Completed:** 2026-03-17

### 2.1 Symbol Table & Scope Management

- [x] Hierarchical symbol table with scope kinds (Module, Function, Block, Loop, Test)
- [x] Symbol definition with duplicate detection
- [x] Hierarchical symbol lookup (walks parent scopes)
- [x] Local-only lookup for shadowing semantics
- [x] Loop context detection (`IsInsideLoop`) for break/continue validation
- [x] Enclosing function resolution for return type checking

**Package:** `pkg/symbols` — 9 tests ✅

### 2.2 Type System

- [x] Complete type representation (Primitive, Struct, Enum, Union, Function, Option, Result, Refinement, TypeParam, Never, Any, None, Alias, Intersection, Tuple, List, Map, Set, StringLiteral)
- [x] Built-in primitive singletons (Int, Float, String, Bool, None, Never, Any)
- [x] Type equality checking (`Equal()`)
- [x] Subtyping/assignability rules (`IsAssignableTo()`) including:
  - Never as bottom type, Any as top type
  - None/T to Option[T], refinement to base type
  - String literal to String/Union, Int to Float widening
  - Struct width subtyping
- [x] Alias and refinement unwrapping (`Underlying()`)
- [x] Type registry with built-in pre-population

**Package:** `pkg/types` — 26 tests ✅

### 2.3 Type Checker

- [x] Multi-pass architecture (types → specs → functions → constants → bodies → spec validation → tests)
- [x] Bidirectional type inference for all expression forms
- [x] Struct construction validation (missing/unknown fields)
- [x] Function call type checking with effect propagation
- [x] Pattern matching with enum exhaustiveness checking
- [x] Effect tracking and validation (declared vs. required effects)
- [x] Spec contract validation (inputs, effects)
- [x] Control flow validation (break/continue in loops, return in functions)
- [x] Mutability enforcement (immutable by default)
- [x] AI-parseable structured error output (JSON format with error codes, expected/got, fix suggestions)
- [x] CLI `aura check` command with `--json` flag for AI agents

**Package:** `pkg/checker` — 48 tests ✅

### Deferred to Future Phases

- Refinement predicate static evaluation (Phase 4 — runtime assertions)
- Import resolution and cross-module type checking (Phase 4/5)
- Visibility (`pub`) enforcement across modules (Phase 4/5)
- Generic type argument inference (Phase 3/4 as needed)
- Transitive effect inference for private functions (Phase 4)

### Phase 2 Milestone — ✅ Achieved

`aura check <file>` now:
- ✅ Reports name resolution errors with source locations
- ✅ Reports type errors with clear messages, expected/got info, and fix suggestions
- ✅ Reports effect mismatches and missing capabilities
- ✅ Reports spec contract violations
- ✅ Outputs structured JSON for AI agent consumption (`--json` flag)
- ✅ **83 new tests** across symbols (9), types (26), checker (48) — all passing

---

## Phase 3: Tree-Walk Interpreter — ✅ COMPLETE

**Goal:** Execute Aura programs via a tree-walk interpreter, completing the end-to-end vibe coding loop.

> 🤖 **AI optimization:** Code generation outputs should be deterministic and predictable so AI agents can reason about the compilation process. The interpreter provides structured error output (JSON-friendly) that AI agents can parse for debugging. A "dry-run" mode validates without executing — useful for AI testing loops.

**Dependencies:** Phase 2 ✅ (semantic analysis complete)

### Why Phase 3 Now? — Rationale

The tree-walk interpreter is the **highest-impact next step** for the AI-first mission:

1. **Closes the vibe coding feedback loop** — Without execution, the workflow is: spec → generate → check. With the interpreter: spec → generate → check → **run → see output → iterate**. This is the complete AI development cycle.
2. **Builds directly on Phase 2** — The typed AST from the checker is the interpreter's input. All the type information, scope resolution, and validation work is already done.
3. **Enables AI test-driven development** — AI agents can generate code, run `aura test` blocks, observe failures, and self-correct. This is the killer feature for vibe coding.
4. **Lower complexity than alternatives** — A tree-walk interpreter (~4–6 weeks) is faster to build than bytecode compilation (~6–10 weeks) and more self-contained than Go transpilation (which requires a Go runtime dependency).
5. **Validates the language design** — Running real programs will surface language design issues early, before investing in optimization.

### 3.1 Tree-Walk Interpreter (Primary Target)

**Complexity:** Medium-High | **Estimate:** 4–6 weeks

#### 3.1.1 Value System
- [x] Implement Aura value types (AuraInt, AuraFloat, AuraString, AuraBool, AuraNone)
- [x] Implement composite values (AuraList, AuraMap, AuraSet, AuraTuple)
- [x] Implement AuraStruct with field access and construction
- [x] Implement AuraEnum with variant matching
- [x] Implement AuraOption (Some/None) and AuraResult (Ok/Err)
- [x] Implement AuraFunction (closures with captured environment)

#### 3.1.2 Expression Evaluation
- [x] Evaluate literals (int, float, string, bool, none)
- [x] Evaluate binary and unary operations with type-appropriate semantics
- [x] Evaluate function calls with argument binding and default values
- [x] Evaluate field access and index operations
- [x] Evaluate string interpolation at runtime
- [x] Evaluate list comprehensions and lambda expressions
- [x] Evaluate `?` propagation for Option/Result types
- [x] Evaluate pipeline operator (`|>`)

#### 3.1.3 Statement Execution
- [x] Execute `let` bindings (immutable and mutable)
- [x] Execute assignments (with mutability checking)
- [x] Execute `return`, `break`, `continue` (as control flow signals)
- [x] Execute `if`/`elif`/`else` chains
- [x] Execute `match` with pattern matching evaluation
- [x] Execute `for ... in` loops with iterator protocol
- [x] Execute `while` loops

#### 3.1.4 Effect & Runtime Infrastructure
- [x] Environment/scope management for interpreter state
- [x] Effect capability injection via `with` blocks
- [x] Built-in print/assert functions
- [x] Structured runtime error output (JSON for AI agents)
- [x] Test block runner (`aura test <file>`)

**Package:** `pkg/interpreter` — 112 tests ✅

#### 3.1.5 CLI Integration
- [x] `aura run <file>` — execute an Aura program
- [x] `aura test <file>` — run test blocks with pass/fail reporting
- [x] `--json` flag for structured output (AI agent consumption)
- [x] `--dry-run` flag for validation without execution

### 3.2 Pattern Matching (Advanced) — 🔲 NOT STARTED

**Complexity:** Medium | **Estimate:** 2–3 weeks | **Target:** v0.9.0

> 🤖 **AI optimization:** Advanced pattern matching enables AI to generate more expressive and concise code. Exhaustiveness checking provides compile-time guarantees that AI-generated match expressions handle all cases — a critical safety property.

**Note:** Basic `match` with enum variants and literals is already implemented in Phase 3.1. This phase completes the pattern matching system with advanced features.

- [ ] Nested patterns (patterns within patterns, e.g., `Some(Ok(x))`)
- [ ] Guard clauses (`when` conditions on match arms)
- [ ] Or-patterns (`A | B => ...` — match multiple patterns with one arm)
- [ ] Binding patterns (`x @ Pattern` — bind a name while matching)
- [ ] Exhaustiveness checking for all pattern types (not just enums)
- [ ] Destructuring in `let` bindings (e.g., `let (x, y) = tuple`)
- [ ] Wildcard patterns with type narrowing
- [ ] Struct destructuring in match arms

**Package:** `pkg/interpreter`, `pkg/checker`

**Estimated test additions:** ~80–120 new tests

### 3.3 Advanced Type Features — 🔲 NOT STARTED

**Complexity:** Medium-High | **Estimate:** 3–4 weeks | **Target:** v1.0.0

> 🤖 **AI optimization:** Generics and improved type inference allow AI to generate reusable, type-safe code. Interface types provide the contract system that AI agents need to understand API boundaries without reading implementation details.

- [ ] Generic types and functions (type parameters with constraints)
- [ ] Improved type inference (bidirectional, constraint-based)
- [ ] Interface types (structural typing, trait-like behavior)
- [ ] Type constraints (`where` clauses for generics)
- [ ] Higher-kinded types (if needed for stdlib design)
- [ ] Type aliases with generic parameters
- [ ] Refinement type static evaluation (deferred from Phase 2)

**Package:** `pkg/types`, `pkg/checker`, `pkg/interpreter`

**Estimated test additions:** ~100–150 new tests

### Phase 3 Deliverables

- ✅ `pkg/interpreter/` with 5 source files: value system, environment, evaluator, module runner, test runner
- ✅ CLI commands: `aura run`, `aura test`, `aura repl`
- ✅ Full expression/statement evaluation (arithmetic, comparison, logic, control flow, structs, enums, match, closures, lambdas, list comprehensions)
- ✅ 14 built-in functions (print, len, str, int, float, range, type_of, abs, min, max, Ok, Err, Some, None)
- ✅ **112 interpreter tests** — all passing (232 total across all packages)
- ✅ String interpolation with full expression support
- ✅ Pipeline operator (`|>`) evaluation
- ✅ Option chaining (`?.`) with None short-circuiting

### 3.1.6 Post-Phase-3.1 Feature Additions

Features added to complete the interpreter's expression support:

- [x] **String interpolation** — Fully implemented in lexer, parser, and interpreter. Strings with `{expr}` syntax are parsed and evaluated at runtime.
- [x] **Pipeline operator (`|>`)** — Lexer tokenizes `|>` as `PIPE_GT`, parser handles pipeline expressions with correct precedence (lower than all other operators, higher than assignment), interpreter evaluates by calling the right-hand function with the left-hand value as argument. Supports chained pipelines and lambdas.
- [x] **Option chaining (`?` operator)** — The existing `?` postfix operator provides sufficient option chaining support for propagating `Option` and `Result` types.

**Tests:** 14 new pipeline operator tests added (225 total across all packages)

---

## Phase 4: Runtime & Standard Library — ✅ COMPLETE

**Goal:** Provide the standard library and runtime support needed for real programs.

> 🤖 **AI optimization:** Standard library APIs should have spec blocks defining their contracts, making them instantly understandable by AI agents. Effect providers should be mockable by default, enabling AI to generate testable code without external dependencies. Library functions should follow consistent patterns so AI can predict APIs for unfamiliar modules.

**Dependencies:** Phase 3 (code generation)

### 4.1 Core Runtime Methods — ✅ COMPLETE

**Complexity:** Medium | **Completed:** 2026-03-20

Implemented **108+ methods** across 5 core types via a centralized method dispatch registry (`pkg/interpreter/methods.go`). All methods follow a consistent `RegisterMethod(Type, "name", func)` pattern.

#### 4.1.1 String Methods (22 methods) — `methods_string.go`
- [x] Core: `len`, `upper`/`to_upper`, `lower`/`to_lower`, `contains`, `split`, `trim`, `trim_start`, `trim_end`
- [x] Search: `starts_with`, `ends_with`, `index_of` (→ Option), `replace`
- [x] Transform: `repeat`, `reverse`, `chars`, `slice` (with bounds checking)
- [x] Aliases: `length` → `len`

#### 4.1.2 List Methods (27 methods) — `methods_list.go`
- [x] Core: `len`/`length`, `append`/`push`, `contains`, `is_empty`
- [x] Accessors: `first`, `last`, `get` (all → Option for safety)
- [x] Mutation: `pop` (→ Option), `remove`
- [x] Transform: `reverse`, `slice` (negative indices), `join`, `index_of` (→ Option)
- [x] Higher-order: `map`, `filter`, `reduce`, `for_each`, `flat_map`, `flatten`
- [x] Predicates: `any`, `all`, `count`
- [x] Utilities: `unique`, `sum`, `min`/`max` (→ Option), `sort`, `zip`, `enumerate`

#### 4.1.3 Map Methods (24 methods) — `methods_map.go`
- [x] Size: `len`/`length`/`size`, `is_empty`
- [x] Access: `keys`, `values`, `entries`, `get` (→ Option), `get_or`, `has`/`contains_key`, `contains_value`
- [x] Mutation: `set`, `remove` (→ Option), `delete` (→ Bool), `clear`, `merge`
- [x] Higher-order: `filter`, `map`, `for_each`, `reduce`, `any`, `all`, `count`
- [x] Utilities: `to_list`, `find` (→ Option)

#### 4.1.4 Option Methods (17 methods) — `methods_option.go`
- [x] Predicates: `is_some`, `is_none`
- [x] Extraction: `unwrap`, `expect`, `unwrap_or`, `unwrap_or_else`
- [x] Transform: `map`, `flat_map`, `and_then`, `filter`, `flatten`
- [x] Combinators: `or_else`, `or`, `and`, `zip`
- [x] Query: `contains`
- [x] Conversion: `to_result`

#### 4.1.5 Result Methods (18 methods) — `methods_option.go`
- [x] Predicates: `is_ok`, `is_err`
- [x] Extraction: `unwrap`, `unwrap_err`, `expect`, `unwrap_or`, `unwrap_or_else`
- [x] Transform: `map`, `map_err`, `and_then`, `or_else`, `flatten`
- [x] Combinators: `or`, `and`
- [x] Query: `contains`, `contains_err`
- [x] Conversion: `ok`, `err`, `to_option`

**Infrastructure:** Method dispatch registry (`methods.go`), `callValue()` helper for invoking Aura lambdas from Go, `cmpValues()` for type-safe ordering.

**Tests:** 222 method-specific tests in `methods_test.go` — **468 total tests** across all packages ✅

### 4.2 Complete Module System and Standard Library — ✅ COMPLETE

**Complexity:** Medium | **Completed:** 2026-03-21

Implemented complete module/import system and **12 pure computation stdlib modules** with 89 functions (before effect modules):

- [x] **Module System** — Import resolution, namespace management, aliasing, cycle detection
- [x] `std.math` — 8 functions: abs, max, min, floor, ceil, round, sqrt, pow + constants (pi, e, inf, nan)
- [x] `std.string` — 4 functions: join, split, replace, repeat
- [x] `std.io` — 3 functions: print, println, format
- [x] `std.testing` — 10 base functions: assert, assert_eq, assert_ne, assert_true, assert_false, assert_some, assert_none, assert_ok, assert_err, run_tests
- [x] `std.json` — 2 functions: parse, stringify (with pretty-print)
- [x] `std.regex` — 6 functions: match, find, find_all, replace, split, compile
- [x] `std.collections` — 9 functions: range, zip_with, partition, group_by, chunk, take, drop, take_while, drop_while
- [x] `std.random` — 6 functions: int, float, choice, shuffle, sample, seed
- [x] `std.format` — 7 functions: pad_left, pad_right, center, truncate, wrap, indent, dedent
- [x] `std.result` — 5 functions: all_ok, any_ok, collect, partition_results, from_option
- [x] `std.option` — 5 functions: all_some, any_some, collect, first_some, from_result
- [x] `std.iter` — 5 functions: cycle, repeat, chain, interleave, pairwise

**Tests:** 146 new tests (module: 17, import_advanced: 64, stdlib_complete: 65) — **614 total tests** across all packages ✅

### 4.3 Effect Runtime — ✅ COMPLETE

**Complexity:** Medium-High | **Completed:** 2026-03-22

Full effect system with 5 providers, 34 stdlib functions, and comprehensive mocking framework:

- [x] **EffectContext** with provider pattern (File, Time, Env, Net, Log)
- [x] **FileProvider** (Real + Mock) — std.file: 9 functions
- [x] **TimeProvider** (Real + Mock) — std.time: 8 functions
- [x] **EnvProvider** (Real + Mock) — std.env: 6 functions
- [x] **NetProvider** (Real + Mock) — std.net: 5 functions (HTTP client)
- [x] **LogProvider** (Real + Mock) — std.log: 6 functions (structured logging)
- [x] **Effect Composition** — Clone, Derive, EffectStack, MockBuilder (fluent API)
- [x] **Testing Helpers** — 13 effect-aware std.testing functions
- [x] **Pre-configured Fixtures** — EmptyMockContext, FixtureWithFiles, etc.
- [x] **222 effect-related tests** across 4 test files

**Tests:** 875 total across all packages ✅

### Phase 4 Milestone — ✅ Achieved

**Completed:** 2026-03-22

Phase 4 delivers a complete runtime and standard library for Aura:

- ✅ **108+ built-in methods** across String (22), List (27), Map (24), Option (17), Result (18)
- ✅ **17 standard library modules** with **117 functions** (12 pure computation + 5 effect-based)
- ✅ **Complete effect system** with 5 providers (File, Time, Env, Net, Log), each with Real + Mock implementations
- ✅ **Full mocking framework** — MockBuilder, EffectStack, Clone/Derive composition
- ✅ **Module system** — Import resolution, namespaces, aliasing, cycle detection
- ✅ **875 tests** across all packages — all passing
- ✅ The AuraTask example from the spec can run end-to-end with mocked effects in tests

---

## Phase 5: Advanced Tooling & Ecosystem — 🔲 NOT STARTED

**Goal:** Build the developer experience and ecosystem around Aura (LSP, Package Manager, AI Integration, Build System).

> 🤖 **AI optimization:** This phase is where AI-first design pays off most. The LSP should expose spec/effect/type information as structured data for AI agents. The spec-to-implementation pipeline (§5.3) is the flagship AI feature — it's the full realization of the vibe coding workflow. Package metadata should be machine-readable so AI can discover and use libraries without human guidance.

**Dependencies:** Phases 3.2, 3.3 (language features should be stable before tooling)

**Target:** v1.1.0

**Note:** This phase focuses on *tooling and ecosystem* — developer experience, IDE support, package management, and AI workflow integration. For *compilation and performance optimization*, see Phase 6.

### 5.1 Language Server Protocol (LSP)

**Complexity:** High | **Estimate:** 4–6 weeks

- [ ] Implement an LSP server for Aura
- [ ] Go-to-definition, find-references, rename
- [ ] Hover information (types, docs, effects)
- [ ] Diagnostics (errors and warnings from semantic analysis)
- [ ] Auto-completion for identifiers, types, and keywords
- [ ] Signature help for function calls
- [ ] Code actions (quick fixes for common errors)

**Package:** `cmd/aura-lsp`

### 5.2 Package Manager

**Complexity:** Medium | **Estimate:** 3–4 weeks

- [ ] Module resolution and dependency management
- [ ] Package manifest file format (`aura.toml`)
- [ ] Version resolution and lock files
- [ ] Registry or Git-based package fetching

### 5.3 AI Integration

**Complexity:** Medium | **Estimate:** 2–3 weeks

- [ ] Spec-to-implementation generation pipeline
- [ ] AST-aware code generation prompts
- [ ] Automatic spec validation for AI-generated code
- [ ] Structured output format for AI consumption

### 5.4 Documentation Generator

**Complexity:** Low-Medium | **Estimate:** 1–2 weeks

- [ ] Extract doc comments (`##`) from source files
- [ ] Generate HTML/Markdown documentation from AST
- [ ] Include type signatures, effects, and spec information
- [ ] Cross-reference linking between modules

### 5.5 REPL

**Complexity:** Medium | **Estimate:** 2 weeks

- [ ] Interactive Aura evaluation loop
- [ ] Expression evaluation and pretty-printing
- [ ] History, auto-completion, and multi-line input
- [ ] `:type` and `:effects` introspection commands

---

## Phase 6: Compiler & Native Compilation — 🔲 NOT STARTED

**Goal:** Compile Aura programs to **native executables** via LLVM, with a bytecode VM for fast development iteration. This is the culmination of Aura's vision — a language with Python's expressiveness that compiles to standalone, zero-dependency binaries.

> 🤖 **AI optimization:** A bytecode compiler produces deterministic, inspectable output that AI agents can reason about. Native compilation via LLVM means AI-generated code runs at C-level speed without manual optimization — the AI writes correct code, the compiler makes it fast. The dual-mode approach (VM for development, native for production) gives AI agents the best of both worlds: instant feedback during generation and maximum performance for deployment.

**Dependencies:** Phases 3.2, 3.3, 5 (language features should be stable before compilation)

**Target:** v2.0.0 | **Total Estimate:** 9–12 weeks

---

### 🔥 Why Native Compilation Matters

Aura is what Python should have been. Python proved that expressive, readable syntax wins — but it pays a permanent performance tax because it was never designed for compilation. Aura is different:

- **Static typing with inference** — Aura knows every type at compile time, enabling aggressive optimization without type annotations everywhere.
- **Explicit effect system** — Side effects are tracked in the type system, so the compiler knows exactly which code is pure and can be optimized freely.
- **No GIL, no interpreter overhead** — Native executables run on bare metal, not through a runtime layer.
- **Zero-dependency deployment** — Ship a single binary. No `pip install`, no virtual environments, no "works on my machine."

The dual-mode approach gives developers the best workflow:

| Mode | Command | Use Case | Speed |
|------|---------|----------|-------|
| **Development** | `aura run main.aura` | Fast iteration, debugging, REPL | Instant startup, interpreted |
| **Production** | `aura build main.aura --release` | Deployment, distribution | Native speed, single binary |

```bash
# Development: instant feedback loop (VM mode)
$ aura run main.aura
Hello, World!

# Production: compile to native executable
$ aura build main.aura --release
Compiling main.aura → main (x86_64-linux)
   Optimizing with LLVM O2...
   Done in 1.2s

$ ./main
Hello, World!

$ file main
main: ELF 64-bit LSB executable, x86-64, statically linked

$ ls -lh main
-rwxr-xr-x 1 user user 1.8M Mar 22 2026 main
```

This is the end state: **write Aura, ship native binaries.** No runtime. No dependencies. Just your program.

---

### 6.1 Bytecode Compiler (2 weeks)

**Complexity:** High

Design and implement a stack-based intermediate representation (IR) and bytecode compiler.

- [ ] Design stack-based bytecode instruction set
- [ ] Implement compiler from typed AST to bytecode IR
- [ ] Constant pool for literals, identifiers, and type info
- [ ] Symbol table generation for functions, closures, and modules
- [ ] Function/closure compilation with upvalue capture
- [ ] Module-level bytecode compilation
- [ ] Debug information embedding (source locations, variable names)
- [ ] Bytecode serialization/deserialization (`.aurac` files)
- [ ] Bytecode disassembler for inspection and debugging

**Package:** `pkg/compiler`

**Expected outcome:** Aura programs compile to a portable bytecode format that serves as the shared IR for both the VM and native backends.

### 6.2 Virtual Machine — Development Mode (2 weeks)

**Complexity:** High

Build a stack-based virtual machine for **fast development iteration**. This is the `aura run` experience — instant startup, rich debugging, rapid feedback.

- [ ] Stack-based bytecode interpreter
- [ ] Call frame management (functions, closures, methods)
- [ ] Upvalue handling for closures
- [ ] Effect system integration (capability passing through VM)
- [ ] Runtime type checking (for dynamic dispatch)
- [ ] Exception/panic handling with stack unwinding
- [ ] Built-in function dispatch (bridge to Go runtime)
- [ ] Standard library integration via VM opcodes
- [ ] Rich debugging support (breakpoints, step-through, variable inspection)
- [ ] Hot reload for development workflow

**Package:** `pkg/vm`

**Expected outcome:** `aura run` executes 10–50x faster than the tree-walk interpreter while maintaining identical semantics. Development mode provides instant startup and rich debugging — the go-to mode during coding.

### 6.3 LLVM Native Backend — Production Mode (3–4 weeks) ⭐

**Complexity:** Very High | **This is the key differentiator.**

Compile Aura bytecode IR to native machine code via LLVM. This is what makes Aura a **real systems-capable language**, not just another scripting language with nice syntax.

- [ ] LLVM IR generation from Aura bytecode/AST
- [ ] Type mapping (Aura types → LLVM types)
- [ ] Function compilation (including closures and lambdas)
- [ ] Native compilation to platform-specific executables
  - [ ] x86_64 (Linux, macOS, Windows)
  - [ ] ARM64 (macOS Apple Silicon, Linux ARM)
  - [ ] Cross-compilation support
- [ ] Zero-runtime executables (no interpreter, no VM needed)
- [ ] Static linking for single-binary deployment
- [ ] Effect system compilation (capabilities as compile-time constructs)
- [ ] Struct layout optimization (field reordering, padding)
- [ ] Enum compilation with tagged unions
- [ ] Pattern matching compilation (decision trees)
- [ ] Standard library native compilation (inline where possible)
- [ ] C FFI bridge (call C libraries from Aura)
- [ ] Production deployment mode (`aura build --release`)

**Package:** `pkg/codegen/llvm`

**Expected outcome:** `aura build --release` produces standalone native executables with:
- **Performance:** Within 2–5x of equivalent C code
- **Binary size:** Small, statically linked executables (< 5MB for typical programs)
- **Startup time:** Instant (no interpreter initialization)
- **Dependencies:** Zero (single binary, no runtime required)
- **Platforms:** x86_64 and ARM64 on Linux, macOS, and Windows

### 6.4 Optimizations (2 weeks)

**Complexity:** Medium-High

Optimization passes that apply to both VM and native compilation paths.

- [ ] **Inlining** — Inline small functions and closures
- [ ] **Constant folding** — Evaluate constant expressions at compile time
- [ ] **Dead code elimination** — Remove unreachable code and unused definitions
- [ ] **Tail call optimization** — Convert tail-recursive calls to loops
- [ ] **Escape analysis** — Stack-allocate values that don't escape their scope
- [ ] **String interning** — Deduplicate string constants
- [ ] **Profile-guided optimization (PGO)** — Use runtime profiles to guide optimization decisions
- [ ] **Link-time optimization (LTO)** — Cross-module optimization during final linking
- [ ] **Garbage collection** — Mark-and-sweep GC for VM mode; reference counting or ownership for native mode
- [ ] **Inline caching** — Fast method dispatch for polymorphic calls (VM mode)

**Package:** `pkg/compiler`, `pkg/vm`, `pkg/codegen/llvm`

**Expected outcome:** Optimized Aura programs achieve predictable, high performance with minimal memory overhead. LTO and PGO enable production builds to approach hand-optimized performance.

---

### The Dual-Mode Architecture

```
                    ┌─────────────────┐
                    │   Aura Source    │
                    │   (.aura files)  │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  Parser + AST   │
                    │  Type Checker   │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │    Bytecode     │
                    │   Compiler      │
                    │  (shared IR)    │
                    └───┬─────────┬───┘
                        │         │
           ┌────────────▼──┐  ┌──▼────────────┐
           │  VM Executor  │  │  LLVM Backend  │
           │  (aura run)   │  │  (aura build)  │
           │               │  │                │
           │  • Fast start │  │  • Native bin  │
           │  • Debugging  │  │  • Zero deps   │
           │  • Hot reload │  │  • Max speed   │
           │  • Dev mode   │  │  • Prod mode   │
           └───────────────┘  └────────────────┘
```

### Phase 6 Milestone Criteria

Phase 6 is complete when:
- [ ] All existing tests pass under both interpreter and VM execution
- [ ] Bytecode compiler produces correct output for all language features
- [ ] VM executes at least 10x faster than tree-walk interpreter on benchmarks
- [ ] `aura build --release` produces native executables for x86_64 and ARM64
- [ ] Native executables run with zero external dependencies
- [ ] Performance within 5x of equivalent C code on compute benchmarks
- [ ] GC handles long-running programs without memory leaks (VM mode)
- [ ] Debug information allows source-level error reporting from both VM and native
- [ ] Cross-compilation works (build Linux binary on macOS and vice versa)

---

## Contributing

See [DEVELOPMENT.md](DEVELOPMENT.md) for setup instructions, architecture overview, and contribution guidelines.

---

## Version History

| Date | Version | Notes |
|------|---------|-------|
| 2026-03-17 | v0.1 | Phase 1 complete; roadmap published |
| 2026-03-17 | v0.2 | Phase 2 complete (type checker, 83 tests); Phase 3 (interpreter) selected as next |
| 2026-03-17 | v0.3 | Phase 3 complete (tree-walk interpreter, 91 tests, 211 total); run/test/repl CLI |
| 2026-03-19 | v0.3.1 | String interpolation, pipeline operator (`\|>`), option chaining (pipeline + interpolation + chaining tests, 232+ total) |
| 2026-03-20 | v0.4.0 | **Phase 4.1 complete** — 108+ core runtime methods (String: 22, List: 27, Map: 24, Option: 17, Result: 18), method dispatch registry, 468 total tests |
| 2026-03-21 | v0.6.0 | **Phase 4.2 complete** — Module system + 12 pure computation stdlib modules, 70 stdlib functions, 614 total tests |
| 2026-03-22 | v0.8.0 | **Phase 4 complete** — Effect Runtime (5 providers), 17 stdlib modules, 117 stdlib functions, MockBuilder, effect composition, 875 total tests |
