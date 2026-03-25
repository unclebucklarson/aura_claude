# AI Next Session

> See `AI_MISSION.md` §Session Protocol for how to maintain this file.

---

## State

| | |
|--|--|
| **Version** | v1.3.0 |
| **Tests** | 1183 (all passing) |
| **Phases complete** | 1, 2, 3.1, 3.1.1, 3.2 (all chunks), 3.3 (all chunks ✅), 4, Issue #8, Issue #11, Phase 5.4 ✅, Phase 5.5 ✅, Phase 5.3 ✅, Phase 5.2 ✅, Phase 5.1.1 ✅, Phase 5.1.2 ✅, Phase 5.1.3 ✅, Phase 5.1.4 ✅ |

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
| pkg/lsp | 20 |
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
| 5.1 | **LSP** (`cmd/aura-lsp`) | High | 4 sub-chunks |

### Phase 5.1 sub-chunks

| Chunk | Scope | Key deliverables | Status |
|-------|-------|-----------------|--------|
| 5.1.1 | JSON-RPC server + lifecycle | `pkg/lsp/rpc.go`, `pkg/lsp/server.go`, `pkg/lsp/types.go`, `cmd/aura-lsp/main.go`; `initialize`/`shutdown`/`exit` | ✅ Done |
| 5.1.2 | Diagnostics | `didOpen`/`didChange`/`didClose` doc buffer; run checker; `publishDiagnostics` | ✅ Done |
| 5.1.3 | Hover | Position-to-AST-node lookup; `textDocument/hover` returns type + doc comment | ✅ Done |
| 5.1.4 | Go-to-definition | `textDocument/definition`; trace identifier to definition site | ✅ Done |

**All 4 sub-chunks delivered in one session. See Phase 5.1 summary below.**

### Phase 5.1 summary (done)

- **`pkg/lsp/types.go`** — full LSP 3.17 type surface: `RequestMessage`, `ResponseMessage`, `NotificationMessage`, `Position`, `Range`, `Location`, `Diagnostic`, `Hover`, `InitializeResult`, `ServerCapabilities`, all `textDocument/*` param types
- **`pkg/lsp/rpc.go`** — `Content-Length` framing: `ReadMessage`, `WriteMessage`, `OKResponse`, `ErrResponse`, `Notification`, `MessageID`, `MessageMethod`
- **`pkg/lsp/server.go`** — dispatch loop; lifecycle (`initialize`/`shutdown`/`exit`); `didOpen`/`didChange`/`didClose` doc buffer; `publishDiagnostics` runs full lex+parse+typecheck pipeline; hover and definition forwarded to `locate.go`
- **`pkg/lsp/locate.go`** — `wordAt` cursor extraction; `computeHover` (fn signature + doc comment); `computeDefinition` (top-level definition location); `typeExprStr` renderer; `checkSource` helper
- **`cmd/aura-lsp/main.go`** — entry point; `lsp.NewServer(os.Stdin, os.Stdout).Run()`
- **20 tests** in `pkg/lsp`; zero external dependencies

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

### Phase 5 — Complete ✅

All five sub-items of Phase 5 (5.1–5.5) are now done. The language is feature-complete with a full ecosystem: formatter, type checker, interpreter, REPL, doc generator, AI integration, package manager, and LSP server.

### Next: Phase 6 — Compiler & Native Compilation

**Target:** v2.0.0

Phase 6 compiles Aura to native code (via LLVM or Go codegen). This is the most ambitious phase and is best tackled in a new session with fresh context.

**Recommended first chunk:** Emit Go source from the Aura AST (no LLVM dependency). This gives a working compiler immediately — Aura programs run as compiled Go binaries — and defers the LLVM toolchain requirement.

**Target:** v1.3.0

This is the most complex remaining item. Requires a full LSP server (`cmd/aura-lsp`) responding to the JSON-RPC protocol over stdin/stdout.

**Recommended starting scope (MVP):**
- `initialize` / `shutdown` lifecycle
- `textDocument/publishDiagnostics` — run type checker on save, push errors
- `textDocument/hover` — show type + doc comment for identifier under cursor
- `textDocument/definition` — go to function/type definition

**Key design decision to settle first:** LSP servers communicate over stdin/stdout JSON-RPC 2.0. The Go stdlib has no LSP library — need to implement the framing (`Content-Length: N\r\n\r\n` + JSON body) manually. This is not much code (~100 lines) and keeps the zero-external-deps rule.

**Key files:** new `cmd/aura-lsp/main.go`, new `pkg/lsp/` package (or directly in cmd)
