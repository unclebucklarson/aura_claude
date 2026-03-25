# AI Next Session

> See `AI_MISSION.md` §Session Protocol for how to maintain this file.

---

## State

| | |
|--|--|
| **Version** | v1.1.0 |
| **Tests** | 1146 (all passing) |
| **Phases complete** | 1, 2, 3.1, 3.1.1, 3.2 (all chunks), 3.3 (all chunks ✅), 4, Issue #8, Issue #11, Phase 5.4 ✅, Phase 5.5 ✅, Phase 5.3 ✅ |

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
| 5.2 | **Package Manager** | Medium | 3–4 weeks |
| 5.1 | **LSP** (`cmd/aura-lsp`) | High | 4–6 weeks |

### Phase 5.3 summary (done)

- **`pkg/codegen/codegen.go`** — new package: `ExtractContext`, `FindUnimplementedSpecs`, `BuildPrompt`, `Generate` (Anthropic API), `Validate`, `Result`
- **`pkg/codegen/codegen_test.go`** — 13 tests covering `stripFences`, `FindUnimplementedSpecs`, `ExtractContext`, `BuildPrompt`, `Validate`
- **`cmd/aura/main.go`** — `aura generate [--dry-run] [--json] <file>` subcommand
- Uses `ANTHROPIC_API_KEY` env var; `--dry-run` prints prompt without API call; `--json` for structured output
- Pure stdlib — no external dependencies

### Next: Phase 5.2 — Package Manager

**Recommended first steps:**

1. Define package manifest format (`aura.toml` or `aura.pkg`) — name, version, dependencies
2. `aura init` — scaffold a new package
3. `aura add <package>` — add a dependency (initially local path deps)
4. `aura build` — resolve imports across packages

**Key design decision:** local-path-only deps first (no registry), then registry later. Keeps it simple.

**Key files:** new `cmd/aura/main.go` subcommands, new `pkg/pkgmgr/` package
