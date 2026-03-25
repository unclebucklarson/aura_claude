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

## Open Debt

**`user_docs/examples.md` — full rewrite needed.** The file uses completely wrong syntax throughout (curly-brace blocks `fn f(x) { ... }` instead of Aura's colon+indent style, `//` comments instead of `#`, etc.). Every example in the file is invalid Aura. Rewrite all examples to use correct syntax before the file can serve as reference material.

---

## Recent: Polish (v2.0.1)

- **Effect system alignment** — checker now recognizes `io`, `env`, `file` (renamed from `fs`); `db`/`auth` retained as valid capability declarations
- **`std.io`** — added `io.read_line() -> Option[String]` and `io.input(prompt?) -> String` (stdin reading)
- **`std.env`** — added `env.exit(code?)` for process termination; `env.args()` now returns only user-supplied args (toolchain prefix stripped via `NewRealEnvProvider`)
- **Method aliases removed** — `String.length/to_upper/to_lower`, `List.length`, `Map.length/size` all removed; canonical: `len`, `upper`, `lower`

---

## Next: Phase 6.2 — Bytecode Compiler

The primary compilation path going forward is **bytecode → VM** (development) and **bytecode → LLVM** (production). The Go-source compiler (6.1) provides a working `aura build` today; 6.2–6.4 are the real compiler stack.

**Phase 6.2 scope** (`pkg/compiler`):
- Stack-based bytecode instruction set design
- Compiler from typed AST to bytecode IR
- Constant pool, symbol table, function/closure compilation
- Bytecode serialization (`.aurac` files) and disassembler

**Key design decision:** Settle the instruction set before writing any VM or LLVM code — both 6.3 and 6.4 consume the same IR. See `ROADMAP.md` §6.2 for full checklist.
