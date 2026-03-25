# AI Next Session

> See `AI_MISSION.md` §Session Protocol for how to maintain this file.

---

## State

| | |
|--|--|
| **Version** | v1.2.0 |
| **Tests** | 1163 (all passing) |
| **Phases complete** | 1, 2, 3.1, 3.1.1, 3.2 (all chunks), 3.3 (all chunks ✅), 4, Issue #8, Issue #11, Phase 5.4 ✅, Phase 5.5 ✅, Phase 5.3 ✅, Phase 5.2 ✅ |

| Package | Tests |
|---------|-------|
| pkg/checker | 129 |
| pkg/codegen | 13 |
| pkg/formatter | 9 |
| pkg/docgen | 12 |
| pkg/interpreter | 904 |
| pkg/lexer | 11 |
| pkg/module | 17 |
| pkg/parser | 16 |
| pkg/pkgmgr | 17 |
| pkg/symbols | 9 |
| pkg/types | 26 |

---

## Open Debt

| # | Item | Status |
|---|------|--------|
| 8 | Refinement type runtime enforcement | ✅ Done v0.9.1 |
| 11 | String concatenation O(n²) | ✅ Done — `collectConcatLeaves` + `evalConcatChain` in eval.go |

**No open debt remaining.**

---

## Next: Phase 5 — Advanced Tooling & Ecosystem

**Target:** v1.2.0 | **Roadmap:** `ROADMAP.md` §5

All language features are stable (Phases 1–4, 3.3 complete). Phase 5 builds the developer experience and ecosystem on top of the current interpreter.

### Remaining items

| Section | Item | Complexity | Estimate |
|---------|------|-----------|----------|
| 5.4 | **Documentation Generator** (`aura doc`) | Low-Medium | ✅ Done |
| 5.5 | **REPL** (enhance existing stub) | Medium | ✅ Done |
| 5.3 | **AI Integration** (`aura generate`) | Medium | ✅ Done |
| 5.2 | **Package Manager** | Medium | ✅ Done |
| 5.1 | **LSP** (`cmd/aura-lsp`) | High | 4–6 weeks |

### Phase 5.2 summary (done)

- **`pkg/pkgmgr/manifest.go`** — `Manifest`, `Dep`, `MetaEntry`; `Find`, `Load`, `FindAndLoad`, `Write`, `Init`, `AddDep`, `ApplyToResolver`
- **Manifest format** — `aura.pkg`: `key = value` top-level fields + `[deps]` section; relative paths resolved at load time; `RawPath` preserved for round-trip writing
- **`aura init [name]`** — creates `aura.pkg` in cwd
- **`aura add <alias> <path>`** — adds/updates a local dep in the nearest `aura.pkg`
- **`aura build`** — loads manifest, prints dep list, verifies all dep dirs exist
- **Auto-detection** — `aura run` walks up from source file dir to find `aura.pkg` and calls `ApplyToResolver`; silent if no manifest found
- **17 tests** in `pkg/pkgmgr`

### Phase 5.3 summary (done)

- **`pkg/codegen/codegen.go`** — new package: `ExtractContext`, `FindUnimplementedSpecs`, `BuildPrompt`, `Generate` (Anthropic API), `Validate`, `Result`
- **`pkg/codegen/codegen_test.go`** — 13 tests covering `stripFences`, `FindUnimplementedSpecs`, `ExtractContext`, `BuildPrompt`, `Validate`
- **`cmd/aura/main.go`** — `aura generate [--dry-run] [--json] <file>` subcommand
- Uses `ANTHROPIC_API_KEY` env var; `--dry-run` prints prompt without API call; `--json` for structured output
- Pure stdlib — no external dependencies

### Next: Phase 5.1 — Language Server Protocol (LSP)

**Target:** v1.3.0

This is the most complex remaining item. Requires a full LSP server (`cmd/aura-lsp`) responding to the JSON-RPC protocol over stdin/stdout.

**Recommended starting scope (MVP):**
- `initialize` / `shutdown` lifecycle
- `textDocument/publishDiagnostics` — run type checker on save, push errors
- `textDocument/hover` — show type + doc comment for identifier under cursor
- `textDocument/definition` — go to function/type definition

**Key design decision to settle first:** LSP servers communicate over stdin/stdout JSON-RPC 2.0. The Go stdlib has no LSP library — need to implement the framing (`Content-Length: N\r\n\r\n` + JSON body) manually. This is not much code (~100 lines) and keeps the zero-external-deps rule.

**Key files:** new `cmd/aura-lsp/main.go`, new `pkg/lsp/` package (or directly in cmd)
