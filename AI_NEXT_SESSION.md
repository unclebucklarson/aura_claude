# AI Next Session

> See `AI_MISSION.md` §Session Protocol for how to maintain this file.

---

## State

| | |
|--|--|
| **Version** | v2.0.1 |
| **Tests** | 1193 (all passing) |
| **Phases complete** | 1, 2, 3.1, 3.1.1, 3.2, 3.3, 4, Issue #8, Issue #11, Phase 5 (all sub-items ✅), Phase 6.1 ✅ |

| Package | Tests |
|---------|-------|
| pkg/checker | 129 |
| pkg/codegen | 13 |
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

### 3. ROADMAP.md §3.3 — stale checkboxes (5 minutes)

Phase 3.3 Chunks 3 and 4 are complete but still show as `[ ]` unchecked in ROADMAP.md. Fix this to keep the roadmap honest.

---

## Next Major: Phase 6.2 — Bytecode Compiler

After the debt above is cleared, the primary engineering priority is the bytecode compiler.

The primary compilation path going forward is **bytecode → VM** (development) and **bytecode → LLVM** (production). The Go-source compiler (6.1) provides a working `aura build` today; 6.2–6.4 are the real compiler stack.

**Phase 6.2 scope** (`pkg/compiler`):
- Stack-based bytecode instruction set design
- Compiler from typed AST to bytecode IR
- Constant pool, symbol table, function/closure compilation
- Bytecode serialization (`.aurac` files) and disassembler

**Key design decision:** Settle the instruction set before writing any VM or LLVM code — both 6.3 and 6.4 consume the same IR. See `ROADMAP.md` §6.2 for full checklist.

---

## Recent: Polish (v2.0.1)

- **Effect system alignment** — checker now recognizes `io`, `env`, `file` (renamed from `fs`); `db`/`auth` retained as valid capability declarations
- **`std.io`** — added `io.read_line() -> Option[String]` and `io.input(prompt?) -> String` (stdin reading)
- **`std.env`** — added `env.exit(code?)` for process termination; `env.args()` now returns only user-supplied args (toolchain prefix stripped via `NewRealEnvProvider`)
- **Method aliases removed** — `String.length/to_upper/to_lower`, `List.length`, `Map.length/size` all removed; canonical: `len`, `upper`, `lower`
- **`docs/introduction_to_programming_with_aura.md`** — full rewrite: 4 tutorials, all programs verified to run, correct lambda syntax (`|x| -> expr`), import aliases, new Tutorial 4 on structs and Option
