# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Is

**Aura** is an AI-first programming language toolchain built entirely in Go. It provides a complete pipeline: lexing → parsing → formatting → type checking → interpretation. The language is designed for "vibe coding" — human writes specs, AI generates implementations, compiler validates correctness.

See `AI_MISSION.md` for the full design philosophy. See `AI_NEXT_SESSION.md` for current state and next tasks. See `ROADMAP.md` for the phased plan.

## Commands

```bash
# Build
go build -o aura ./cmd/aura

# Run full test suite
go test ./... -v

# Run a specific package
go test ./pkg/interpreter -v

# Run a single test by name
go test ./pkg/checker -run TestTypeCheckFunctionCall -v

# Run with race detection
go test ./... -race

# Coverage
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

# CLI usage
./aura format <file>        # canonical formatting
./aura check [--json] <file> # type-check (--json for AI-parseable errors)
./aura run <file>            # execute (calls main())
./aura test <file>           # run test blocks
./aura repl                  # interactive REPL
```

**No external dependencies** — pure Go stdlib. Requires Go 1.22+.

## Architecture

The pipeline is strictly linear; each stage produces input for the next:

```
Source (.aura)
  ↓ pkg/token, pkg/lexer     — indentation-sensitive tokenizer, emits INDENT/DEDENT
  ↓ pkg/parser               — recursive descent, builds AST (pkg/ast)
  ↓ pkg/formatter            — AST → canonical source (round-trip guarantee)
  ↓ pkg/symbols, pkg/types   — symbol tables and 19-kind type system
  ↓ pkg/checker              — 7-pass type checker with effect tracking
  ↓ pkg/interpreter          — tree-walk interpreter with full stdlib
```

### Key architectural details

**Lexer** (`pkg/lexer`): Tracks paren depth to suppress INDENT/DEDENT inside `()`, `[]`, `{}`. Uses `##` for doc comments vs `#` for line comments.

**Parser** (`pkg/parser`): Operator precedence climbing. Parses spec blocks with sections: `doc`, `inputs`, `guarantees`, `effects`, `errors`.

**Type checker** (`pkg/checker`): Multi-pass — 7 passes in order: types → specs → functions → constants → bodies → spec validation → tests. Errors are structured JSON with error codes and fix suggestions (`--json` flag). Effect tracking validates that function bodies only use declared effects (`with db, time, net, ...`).

**Type system** (`pkg/types`): 19 type kinds. Subtyping rules: Never as bottom, Any as top, refinements strip predicates for base-type checks, Int widens to Float, unions use set intersection. Refinement types enforce predicates at runtime (v0.9.1).

**Interpreter** (`pkg/interpreter`): Effect system uses `EffectContext` with 5 providers (File, Time, Env, Net, Log), each with Real + Mock implementations. `MockBuilder` fluent API for tests. Standard library split across 17 `stdlib_*.go` files (12 pure + 5 effect-based). Methods dispatched through central registry in `methods.go`.

**Module system** (`pkg/module`): Import resolution with cycle detection and namespace aliasing.

### Session protocol

Per `AI_MISSION.md`:
- **Before editing:** Read the relevant file(s). Never edit blind.
- **After every change:** Build with `go build`. Fix any compile error before proceeding.
- **Before marking done:** Run tests. All 1005 must pass.
- **Forward notes:** Update `AI_NEXT_SESSION.md` at end of session.
- **Design priority order:** AI flow → compiler verifiability → human readability → brevity.

### Operating modes

There are two distinct modes of work in this repo:

1. **Toolchain mode (Go):** Editing `.go` files in `pkg/` or `cmd/`. Use Go idioms, standard library only.
2. **Aura code generation mode (`.aura`):** Writing example or test `.aura` files. Follow Aura language conventions — specs first, explicit effects, refinement types.

Do not confuse these modes. When modifying the interpreter or checker, you are in toolchain mode even if you're handling Aura-specific constructs.
