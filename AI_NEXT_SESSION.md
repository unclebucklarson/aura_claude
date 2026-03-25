# AI Next Session

> See `AI_MISSION.md` §Session Protocol for how to maintain this file.

---

## State

| | |
|--|--|
| **Version** | v1.0.0 |
| **Tests** | 1121 (all passing) |
| **Phases complete** | 1, 2, 3.1, 3.1.1, 3.2 (all chunks), 3.3 (all chunks ✅), 4, Issue #8, Issue #11 |

| Package | Tests |
|---------|-------|
| pkg/checker | 129 |
| pkg/formatter | 9 |
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

**Target:** v1.1.0 | **Roadmap:** `ROADMAP.md` §5

All language features are stable (Phases 1–4, 3.3 complete). Phase 5 builds the developer experience and ecosystem on top of the current interpreter.

### Recommended order (lowest to highest complexity)

| Section | Item | Complexity | Estimate |
|---------|------|-----------|----------|
| 5.4 | **Documentation Generator** (`aura doc`) | Low-Medium | 1–2 weeks |
| 5.5 | **REPL** (enhance existing stub) | Medium | 2 weeks |
| 5.3 | **AI Integration** (spec-to-impl pipeline) | Medium | 2–3 weeks |
| 5.2 | **Package Manager** | Medium | 3–4 weeks |
| 5.1 | **LSP** (`cmd/aura-lsp`) | High | 4–6 weeks |

### Design decisions to settle before starting

1. **5.4 Doc Generator** — output format: Markdown (simpler, no external deps) vs. HTML (richer). Recommend Markdown first, HTML as a follow-up.
2. **5.5 REPL** — `aura repl` is already stubbed in `cmd/aura`. Decide scope: expression-only vs. full statement support, `:type` introspection depth.
3. **5.3 AI Integration** — requires Claude API key in environment; not pure stdlib. Decide if this is in-process or a CLI pipeline.
4. **5.1 LSP** — requires `gopls`-style architecture; most impactful but most complex. Can be deferred until after 5.4/5.5/5.3.

### Chunk 1 suggestion — 5.4 Documentation Generator

Start here: low risk, builds directly on the existing AST, no new dependencies.

**What to build:**
- `aura doc <file>` CLI command
- Walk AST, extract `##` doc comments attached to `FnDef`, `StructDef`, `EnumDef`, `TraitDef`, `SpecDef`
- Emit Markdown: function signatures with types + effects + doc text; struct fields; enum variants; spec guarantees/errors
- `aura doc --json <file>` for structured output (AI-consumable)

**Key files:** `cmd/aura/main.go` (new subcommand), new `pkg/docgen/docgen.go`
