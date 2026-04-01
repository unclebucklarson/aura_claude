# AI Next Session

> See `AI_MISSION.md` §Session Protocol for how to maintain this file.

---

## State

| | |
|--|--|
| **Version** | v2.0.1 |
| **Tests** | 1221 (all passing) |
| **Phases complete** | 1, 2, 3.1, 3.1.1, 3.2, 3.3, 4, Issue #8, Issue #11, Phase 5 (all sub-items ✅), Phase 6.1 ✅, Phase 6.2 (core ✅) |

| Package | Tests |
|---------|-------|
| pkg/checker | 129 |
| pkg/codegen | 13 |
| pkg/compiler | 28 |
| pkg/docgen | 12 |
| pkg/formatter | 9 |
| pkg/goemit | 14 |
| pkg/interpreter | 904 |
| pkg/lexer | 11 |
| pkg/lsp | 20 |
| pkg/module | 17 |
| pkg/parser | 16 |
| pkg/pkgmgr | 17 |
| pkg/symbols | 9 |
| pkg/types | 26 |

---

## Open Debt (work in priority order)

### 1. `user_docs/examples.md` — full rewrite (1 session)

Every example in the file uses completely wrong syntax: curly-brace blocks `fn f(x) { ... }`, `//` comments, and other non-Aura conventions. The file is linked from the method reference and intro tutorial — anyone landing there gets broken, invalid Aura code. Rewrite all examples using correct colon+indent syntax, `#` comments, and verified-working code before anything else ships.

### 2. `pub` visibility enforcement (requires multi-module checker integration — 2–3 sessions)

`pub` is currently decorative. **This cannot be implemented as a quick fix.** Root cause: the checker operates on a single module's AST and never sees imported module types. `pub` is only meaningful at module boundaries, so enforcement requires multi-module type resolution.

**Prerequisite (must be done first):** Integrate the module resolver with the checker so that when checking module B, the checker has access to the type information (fields, visibility) from module A that B imports. Currently `pkg/checker` does not use `pkg/module` at all.

**Then implement:**
- Add `DefinedInModule string` to `types.Type` (set during struct/enum registration)
- Add `ErrPrivateAccess ErrorCode` in `errors.go`
- In `inferFieldAccess`: when `type.DefinedInModule != currentModule`, check `field.Public`; emit error if not public
- Apply similarly to top-level function calls and type references across module boundaries
- Add ~20 tests using two-module programs (requires multi-module checker to exist first)

---

## Next Major: Phase 6.3 — Bytecode VM

Phase 6.2 (bytecode compiler) is now complete. `pkg/compiler` produces a `Chunk` IR from any Aura AST. See ROADMAP §6.3 for the VM scope.

**Phase 6.3 scope** (`pkg/vm`):
- Call frame stack, operand stack management
- Dispatch loop executing all opcodes defined in `pkg/compiler/opcode.go`
- Upvalue capture and closure execution
- Method dispatch (delegate to interpreter's method registry or re-implement)
- Effect system integration (pass EffectContext through VM calls)
- Integration with `aura run` command (as alternative to tree-walk interpreter)

**Key design decision:** The VM consumes the `*compiler.Chunk` IR directly. It does not re-parse or re-check. The interpreter remains as a fallback for development until the VM covers 100% of features.

**Architecture note:** The VM will be a separate package `pkg/vm`. It needs access to:
- `pkg/compiler` (chunk, opcode, constants)
- `pkg/interpreter` value types (`Value`, `ListValue`, `MapValue`, etc.) — OR define its own value representation

The simplest approach is to **reuse interpreter value types** to avoid reimplementing the entire runtime. The VM becomes a new execution engine for the same value domain.

---

## Phase 6.2 Summary — Completed This Session

**What was built:** `pkg/compiler` — a complete bytecode compiler from typed Aura AST to a stack-based IR.

**Files added:**
- `pkg/compiler/opcode.go` — 60 opcodes, fixed 3-byte instruction format, ReadArg/WriteArg helpers
- `pkg/compiler/chunk.go` — Chunk, Constant, UpvalueDesc, LocalInfo, SourceMapEntry
- `pkg/compiler/disassembler.go` — human-readable bytecode listing, nested chunk expansion
- `pkg/compiler/compiler.go` — Compiler struct + CompileModule() + all compileXxx methods
- `pkg/compiler/compiler_test.go` — 28 tests covering all major constructs

**What the compiler handles:**
- All literal types (int, float, bool, string, string interpolation, none)
- All binary/unary operators including short-circuit and/or
- Local variables, globals, upvalue capture (closures)
- if/elif/else statements and if expressions
- match statements and match expressions (patterns: wildcard, binding, literal, constructor, or, list)
- for/while loops, break/continue
- let bindings, assignment to locals/globals/fields/indices
- Function definitions and lambda expressions
- List, map, tuple construction and comprehensions
- Field access, index access, optional chaining, option propagate (?)
- Pipeline operator (|>)
- Some/None/Ok/Err constructors
- Struct construction
- Assert statements

**Remaining 6.2 item:** Bytecode serialization (`.aurac` files) — deferred to a future session as it requires a wire format design decision.

---

## Recent: docs / examples (v2.0.1)

- **`docs/introduction_to_programming_with_aura.md`** — full rewrite: 4 tutorials, all programs verified to run, correct lambda syntax (`|x| -> expr`), import aliases, new Tutorial 4 on structs and Option
- **`user_docs/examples.md`** — complete syntax rewrite (curly-brace → colon+indent, `//` → `#`, `|x| expr` → `|x| -> expr`)
- **ROADMAP.md §3.3** — stale checkboxes fixed

## Recent: Polish (v2.0.1)

- **Effect system alignment** — checker now recognizes `io`, `env`, `file` (renamed from `fs`); `db`/`auth` retained as valid capability declarations
- **`std.io`** — added `io.read_line() -> Option[String]` and `io.input(prompt?) -> String` (stdin reading)
- **`std.env`** — added `env.exit(code?)` for process termination; `env.args()` now returns only user-supplied args (toolchain prefix stripped via `NewRealEnvProvider`)
- **Method aliases removed** — `String.length/to_upper/to_lower`, `List.length`, `Map.length/size` all removed; canonical: `len`, `upper`, `lower`
