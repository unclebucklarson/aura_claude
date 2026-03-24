# Aura Toolchain — Built by Claude

> This implementation was built entirely by [Claude](https://claude.ai) (Anthropic) via Claude Code,
> working session-by-session with a human collaborator. No human wrote any of the Go source code.
> A parallel implementation built using Abacus AI exists at a separate repository for comparison.

---

## 🤖 AI-First Language — Designed for AI-Human Collaboration

> **Aura is an AI-first programming language.** Every design decision optimizes for AI code generation, AI parseability, and seamless AI-human "vibe coding."

**This is the primary design goal of Aura.** The language exists to make AI agents the best developers they can be, while keeping code clear for human review.

### Why AI-First?

| Aura Feature | How It Helps AI |
|---|---|
| **Spec blocks** | Structured, machine-readable contracts — AI knows *what* to build before writing *how* |
| **Effect annotations** (`with db, time`) | AI knows exactly what side effects are allowed — no hidden state mutations |
| **Refinement types** (`String where len >= 1`) | Data constraints live in the type, not in scattered validation code |
| **`satisfies` clauses** | AI-generated code is automatically verified against the spec |
| **Explicit types everywhere** | Every function boundary is a clear contract — zero guessing |
| **Structured error types** | AI can generate exhaustive error handling from the type definition |

### The Vibe Coding Flow

1. **Human writes the spec** — structured intent, not ambiguous prose
2. **AI generates the implementation** — using the spec as a complete contract
3. **Compiler validates** — types, effects, and spec satisfaction checked automatically
4. **Human reviews** — the spec makes intent clear, so review is fast

📖 **Read the full mission statement: [AI_MISSION.md](AI_MISSION.md)**

---

A complete toolchain for the **Aura programming language** — a Python-inspired, statically typed language with specification-driven development, algebraic types, and effect tracking.

Built in Go. Implements lexing, parsing, AST construction, canonical source formatting, type checking with semantic analysis, tree-walk interpreter, 108+ core runtime methods across String, List, Map, Option, and Result types, 17 standard library modules with 117 functions, and a complete effect system with 5 mockable providers.

## Project Structure

```
aura-toolchain/
├── cmd/aura/main.go              # CLI entry point (format, parse, check, run, test, repl)
├── pkg/
│   ├── token/token.go            # Token types, positions, spans
│   ├── lexer/lexer.go            # Indentation-sensitive lexer (INDENT/DEDENT)
│   ├── ast/ast.go                # Complete AST node definitions
│   ├── parser/parser.go          # Recursive descent parser
│   ├── formatter/formatter.go    # AST → canonical source formatter
│   ├── symbols/symbols.go        # Symbol table & scope management
│   ├── types/types.go            # Type system representation & subtyping
│   ├── checker/                   # Type checker & semantic analysis
│   │   ├── checker.go            # Multi-pass type checker
│   │   └── errors.go             # Structured, AI-parseable error diagnostics
│   ├── module/                    # Module system
│   │   └── resolver.go           # Module resolution, initialization ordering, cycle detection
│   └── interpreter/               # Tree-walk interpreter + Runtime + Standard Library + Effects
│       ├── value.go              # Value types (Int, Float, String, Bool, etc.)
│       ├── env.go                # Environment with scope chain
│       ├── eval.go               # Expression & statement evaluator
│       ├── interpreter.go        # Module execution, builtins & stdlib registration
│       ├── test.go               # Test block runner
│       ├── methods.go            # Method dispatch registry infrastructure
│       ├── methods_string.go     # 22 String methods
│       ├── methods_list.go       # 27 List methods + callValue/cmpValues helpers
│       ├── methods_map.go        # 24 Map methods
│       ├── methods_option.go     # 17 Option + 18 Result methods
│       ├── effect.go             # Effect system: EffectContext, 5 providers (Real + Mock)
│       ├── stdlib_math.go        # std.math (8 functions)
│       ├── stdlib_string.go      # std.string (4 functions)
│       ├── stdlib_io.go          # std.io (3 functions)
│       ├── stdlib_testing.go     # std.testing (23 functions incl. effect-aware)
│       ├── stdlib_json.go        # std.json (2 functions)
│       ├── stdlib_regex.go       # std.regex (6 functions)
│       ├── stdlib_collections.go # std.collections (9 functions)
│       ├── stdlib_random.go      # std.random (6 functions)
│       ├── stdlib_format.go      # std.format (7 functions)
│       ├── stdlib_result.go      # std.result (5 functions)
│       ├── stdlib_option.go      # std.option (5 functions)
│       ├── stdlib_iter.go        # std.iter (5 functions)
│       ├── stdlib_file.go        # std.file (9 functions, effect-based)
│       ├── stdlib_time.go        # std.time (8 functions, effect-based)
│       ├── stdlib_env.go         # std.env (6 functions, effect-based)
│       ├── stdlib_net.go         # std.net (5 functions, effect-based)
│       └── stdlib_log.go         # std.log (6 functions, effect-based)
├── testdata/                      # Sample .aura files
├── user_docs/                     # User-facing documentation
│   ├── getting_started.md        # Installation and first program
│   ├── language_guide.md         # Tutorial-style language guide
│   ├── language_reference.md     # Formal language reference
│   ├── method_reference.md       # Complete method & stdlib reference
│   └── examples.md              # Working examples for all features
└── README.md
```

## Quick Start

### Prerequisites

- Go 1.22+

### Build

```bash
go build -o aura ./cmd/aura
```

### Usage

**Format** an Aura source file (prints canonical formatting to stdout):

```bash
./aura format testdata/models.aura
```

**Parse** an Aura source file (dumps tokens and AST):

```bash
./aura parse testdata/specs.aura
```

**Format in-place** with the `-w` flag:

```bash
./aura format -w testdata/service.aura
```

**Type-check** an Aura source file:

```bash
./aura check testdata/models.aura
```

**Type-check with JSON output** (for AI agents):

```bash
./aura check --json testdata/service.aura
```

**Run** an Aura program (executes `main()` function):

```bash
./aura run program.aura
```

**Run test blocks** in an Aura file:

```bash
./aura test testdata/models.aura
```

**Interactive REPL**:

```bash
./aura repl
```

## Language Features Supported

### Types & Definitions
- **Type aliases** with refinement types: `type TaskId = String where len >= 1`
- **Union types**: `type Status = "pending" | "done" | "cancelled"`
- **Structs** with default values and optional fields: `pub struct Task:`
- **Enums** with data variants: `pub enum TaskError: NotFound(TaskId)`
- **Traits** and **impl blocks**

### Functions
- Named & default parameters, return types
- **Effect tracking**: `fn save() -> Result with db, time`
- **Satisfies clauses**: `fn create_task(...) satisfies CreateNewTask`
- **Guard clauses** with `where`

### Spec Blocks
- `doc`, `inputs`, `guarantees`, `effects`, `errors` sections
- Typed inputs with descriptions
- Guarantee strings and error variant descriptions

### Control Flow
- `if` / `elif` / `else`
- `match` with patterns (literals, wildcards, constructors, destructuring)
- `for ... in` loops
- `while` loops
- `return`, `break`, `continue`

### Expressions
- Binary and unary operators
- Pipeline operator: `data |> transform |> format`
- List comprehensions: `[x * 2 for x in items if x > 0]`
- Lambda expressions: `|x| x + 1`
- Optional chaining: `task.completed_at?`
- Unwrap: `maybe_task!`
- String interpolation: `"Hello, {name}!"`
- Struct construction with named fields

### Built-in Methods (108+)
- **String** (22): `len`, `upper`, `lower`, `contains`, `split`, `trim`, `replace`, `starts_with`, `ends_with`, `index_of`, `slice`, `chars`, `repeat`, `reverse`, and more
- **List** (27): `map`, `filter`, `reduce`, `sort`, `reverse`, `first`, `last`, `get`, `flat_map`, `flatten`, `unique`, `zip`, `enumerate`, `any`, `all`, `sum`, `min`, `max`, and more
- **Map** (24): `keys`, `values`, `entries`, `get`, `set`, `remove`, `merge`, `filter`, `map`, `find`, `has`, `contains_key`, `contains_value`, and more
- **Option** (17): `unwrap`, `expect`, `map`, `flat_map`, `and_then`, `filter`, `or_else`, `zip`, `to_result`, `is_some`, `is_none`, `contains`, and more
- **Result** (18): `unwrap`, `expect`, `map`, `map_err`, `and_then`, `or_else`, `ok`, `err`, `to_option`, `is_ok`, `is_err`, `contains`, and more

### Standard Library (17 modules, 117 functions)
- **Pure Computation:** `std.math` (8), `std.string` (4), `std.io` (3), `std.json` (2), `std.regex` (6), `std.collections` (9), `std.random` (6), `std.format` (7), `std.result` (5), `std.option` (5), `std.iter` (5)
- **Testing:** `std.testing` (23 functions including effect-aware assertions and mock setup)
- **Effect-Based I/O:** `std.file` (9), `std.time` (8), `std.env` (6), `std.net` (5), `std.log` (6)

### Effect System
- **5 effect providers** with Real + Mock implementations: File, Time, Env, Net, Log
- **MockBuilder** fluent API for configuring test contexts
- **Effect composition** via Clone, Derive, and EffectStack
- All I/O operations return `Result` types for explicit error handling
- Full mockability enables AI-generated testable code without external dependencies

### Indentation
- Python-style significant whitespace
- INDENT/DEDENT token generation
- 4-space canonical indentation (enforced by formatter)

## Testing

Run the full test suite:

```bash
go test ./... -v
```

**875 tests total** across all packages:
- `pkg/lexer/` — 11 tests covering tokenization, indentation, comments, edge cases
- `pkg/parser/` — 16 tests covering all language constructs
- `pkg/formatter/` — 9 tests including round-trip verification (parse → format → parse = same AST)
- `pkg/symbols/` — 9 tests covering symbol table, scopes, and lookups
- `pkg/types/` — 26 tests covering type system, equality, subtyping, and registry
- `pkg/checker/` — 49 tests covering type checking, effects, specs, and error diagnostics
- `pkg/module/` — 17 tests covering module resolution, initialization ordering, cycle detection
- `pkg/interpreter/` — 738 tests covering:
  - Core interpreter (values, environment, expressions, statements, control flow, builtins, structs, enums, match, closures, test runner, string interpolation, pipeline operator, option chaining)
  - 222 method-specific tests for String/List/Map/Option/Result
  - 64 advanced import/module system tests
  - 65 stdlib tests (regex, collections, random, format, result, option, iter)
  - 222 effect system tests (file, time, env, net, log, composition, mocking)

### Round-Trip Guarantee

The formatter produces deterministic output. Formatting source code, then parsing and formatting again, always produces identical output:

```
source → parse → AST₁ → format → source₂ → parse → AST₂ → format → source₃
                                  source₂ == source₃  ✓
```

## Architecture

### Lexer (`pkg/lexer`)
Scans Aura source into tokens with full position tracking. Key features:
- **Indentation tracking** via an indent stack — emits `INDENT` and `DEDENT` tokens
- **Paren depth tracking** — suppresses NEWLINE/INDENT/DEDENT inside `()`, `[]`, `{}`
- **Comment handling** — `#` line comments and `##` doc comments
- **Blank line handling** — properly emits DEDENTs across blank-line gaps between blocks

### Parser (`pkg/parser`)
Recursive descent parser that builds a complete AST. Features:
- Operator precedence climbing for expressions
- Pattern matching for `match` cases (wildcards, literals, constructors, lists, tuples)
- Type expression parsing (generics, optionals, result types, maps, tuples)
- Spec block parsing with all section types

### AST (`pkg/ast`)
Complete node definitions for the Aura language including:
- Module-level declarations (types, structs, enums, functions, specs, traits, impls, tests)
- Statements (let, assignment, return, if, match, for, while, expression statements)
- Expressions (literals, binary/unary ops, calls, field access, index, pipe, comprehensions, lambdas)
- Patterns and type expressions

### Formatter (`pkg/formatter`)
Converts AST back to canonical Aura source with:
- Consistent 4-space indentation
- Deterministic output ordering
- Blank line separation between top-level declarations
- Proper handling of all expression precedence (parenthesization where needed)

## Example

Input (`testdata/models.aura`):

```aura
module auratask.models

import std.time as time

type TaskId = String where len >= 1 and len <= 64

pub struct Task:
    pub id: TaskId
    pub title: String where len >= 1 and len <= 200
    pub status: TaskStatus = "pending"
    pub priority: Priority = 3
    pub created_at: time.Instant
    pub tags: [String] = []
```

The formatter preserves this exact canonical form. The parser produces a full AST that can be inspected, transformed, or used for code generation.

## Documentation

### User Documentation

- **[Getting Started](user_docs/getting_started.md)** — Installation, first program, and basic usage
- **[Language Guide](user_docs/language_guide.md)** — Tutorial-style guide covering all language features with examples
- **[Language Reference](user_docs/language_reference.md)** — Formal reference for types, syntax, effects, and specifications
- **[Examples](user_docs/examples.md)** — Complete working examples covering every language feature

### AI-First Mission

- **[AI Mission Statement](AI_MISSION.md)** — Why Aura is AI-first, design principles, and guidelines for AI contributors

### Development

- **[Roadmap](ROADMAP.md)** — Phased development plan from parser to full language
- **[Development Guide](DEVELOPMENT.md)** — Architecture overview, implementation checklists, testing strategy, and contribution guidelines

## License

MIT
