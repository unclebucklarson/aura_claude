# AI Next Session - Aura Language

---

## 🔧 Code Quality — Issues Identified 2026-03-23

> Issues found during codebase audit. Address before advancing to Phase 3.2 Chunk 3 or later phases.
> Priority order: HIGH first, then MEDIUM, then LOW.

### Issue Summary

| # | Severity | Category | Location | Issue | Status |
|---|----------|----------|----------|-------|--------|
| 1 | 🔴 HIGH | Correctness | `pkg/interpreter/eval.go` `intPow` | Integer overflow silently corrupts results; negative exponents return `0` with no error | ✅ Fixed |
| 2 | 🔴 HIGH | Correctness | `pkg/interpreter/eval.go` `evalIndexExpr` | String indexing operates on bytes, not runes — multi-byte UTF-8 characters are mangled | ✅ Fixed |
| 3 | 🔴 HIGH | Correctness | `pkg/interpreter/eval.go` `evalCallExpr` | Enum variant arity is never checked — `Color.Red(1,2,3)` accepted silently | ✅ Fixed |
| 4 | 🔴 HIGH | Correctness | `pkg/lexer/lexer.go` | Unmatched closing paren clamps `parenDepth` to 0 silently; no lexer error emitted | ✅ Fixed |
| 5 | 🔴 HIGH | Correctness | `pkg/interpreter/eval.go` `evalIndexExpr` | Map dot-access returns `NoneVal` for missing keys with no error — hides typos completely | ✅ Fixed |
| 6 | 🟡 MEDIUM | Architecture | `pkg/interpreter/eval.go`, `env.go` | Inconsistent error strategy: some paths `runtimePanic`, others return `error` — recovery is impossible | ✅ Fixed |
| 7 | 🟡 MEDIUM | Architecture | `pkg/parser/parser.go` | `blockExprJustEnded bool` can't track nested block expressions — should be a stack | ✅ Fixed |
| 8 | 🟡 MEDIUM | Semantics | `pkg/interpreter/eval.go` | Refinement types (`Int where x > 0`) parsed and type-checked but never validated at runtime | 📋 Deferred |
| 9 | 🟡 MEDIUM | Semantics | `pkg/interpreter/eval.go` `evalListComp` | List comprehensions only iterate `ListVal` — `[c for c in "hello"]` panics | ✅ Fixed |
| 10 | 🔵 LOW | Performance | `pkg/interpreter/eval.go` | `MapVal` uses O(n) linear search with parallel arrays — every lookup scans all keys | ✅ Fixed |
| 11 | 🔵 LOW | Performance | `pkg/interpreter/eval.go` | String concatenation (`+`) allocates a new string each time — O(n²) in loops | 📋 Documented |

---

### Issue Details

#### Issue 1 — Integer Overflow in `intPow` 🔴
**File:** `pkg/interpreter/eval.go`
- `result *= base` and `base *= base` can silently overflow `int64` with no warning
- `intPow(2, 63)` produces wrong result with no error
- Negative exponents return `0` — should error or return float
- **Fix:** Add overflow detection before each multiply; handle negative exponents explicitly

#### Issue 2 — UTF-8 String Indexing 🔴
**File:** `pkg/interpreter/eval.go`
- `string[i]` indexes bytes: `string(v.Val[i])` breaks multi-byte characters
- `"café"[-1]` returns a garbage byte, not `"é"`
- **Fix:** Decide byte vs rune semantics, convert to `[]rune` for rune-based indexing

#### Issue 3 — Enum Variant Arity Not Checked 🔴
**File:** `pkg/interpreter/eval.go`
- Variants map stores arity but it is never compared against `len(args)` at call time
- `Color.Red(1, 2, 3)` silently creates malformed `EnumVal` with wrong `Data` slice
- **Fix:** Extract expected arity from variant definition; panic with clear message on mismatch

#### Issue 4 — Unmatched Paren Silently Clamped in Lexer 🔴
**File:** `pkg/lexer/lexer.go`
- `l.parenDepth--` followed by `if l.parenDepth < 0 { l.parenDepth = 0 }` emits no error
- Downstream parser sees wrong NEWLINE/INDENT/DEDENT tokens and fails mysteriously
- **Fix:** Call `l.addError(...)` before clamping so the user gets a meaningful lexer error

#### Issue 5 — Map Dot-Access Returns `None` Silently 🔴
**File:** `pkg/interpreter/eval.go`
- `mymap.missingkey` returns `NoneVal{}` instead of panicking
- Typos in field names are invisible at runtime
- **Fix:** Error (or panic) when key is not found and field name is not a known method

#### Issue 6 — Inconsistent Error Handling Strategy 🟡
**Files:** `pkg/interpreter/eval.go`, `pkg/interpreter/env.go`
- `eval.go` uses `runtimePanic()` (requires panic recovery to handle)
- `env.go` returns `error` values
- Mixing the two makes it impossible to handle errors consistently in calling code
- **Fix:** Settle on one strategy; returning `(Value, error)` pairs is more testable and composable

#### Issue 7 — `blockExprJustEnded` Boolean Is a Fragile Hack 🟡
**File:** `pkg/parser/parser.go`
- Single `bool` cannot correctly track nested block expressions
- `if ... match ... if ...` can misfire the flag and generate wrong DEDENT handling
- **Fix:** Replace with a stack (`[]bool`) — push on enter, pop on exit

#### Issue 8 — Refinement Types Not Enforced at Runtime 🟡
**Files:** `pkg/types/types.go`, `pkg/interpreter/eval.go`
- `Int where x > 0` is parsed and type-checked but the predicate is never evaluated at runtime
- A value of `-5` can be assigned to a `PositiveInt` variable with no error
- **Fix:** Evaluate refinement predicate on assignment; return runtime error if violated

#### Issue 9 — List Comprehensions Don't Support String/Map Iteration 🟡
**File:** `pkg/interpreter/eval.go` `evalListComp`
- Comprehension iterator requires `*ListVal`; all other iterables panic
- `[c for c in "hello"]` and `[k for k in mymap]` both fail
- For-loops and comprehensions use separate (divergent) iteration logic
- **Fix:** Extract a shared iterator protocol; reuse it in both for-loops and comprehensions

#### Issue 10 — Map Lookup Is O(n) 🔵
**File:** `pkg/interpreter/eval.go`
- `MapVal` stores `Keys []Value` and `Values []Value` as parallel slices
- Every get/set/has operation calls `Equal(k, index)` in a linear scan
- **Fix:** Use `map[string]Value` for string-keyed maps; fall back to linear only for non-hashable keys

#### Issue 11 — String Concatenation Is O(n²) 🔵
**File:** `pkg/interpreter/eval.go`
- `lv.Val + rv.Val` allocates a new string for every `+` operation
- `s = s + chunk` in a loop is quadratic
- **Fix:** Detect chain concatenation and use `strings.Builder`; or document the limitation

---

## Status: Phase 3.2 Chunk 4 COMPLETE ✅ — Exhaustiveness Checking

**Version:** v0.9.0
**Total Tests:** 994 (all passing)
**Date:** 2026-03-23

---

## Phase 4 Achievement Summary

### Phase 4.1: Core Runtime Methods ✅ (Completed 2026-03-20)
- 108+ built-in methods across 5 core types
- Method dispatch registry infrastructure
- String (22), List (27), Map (24), Option (17), Result (18) methods

### Phase 4.2: Module System & Standard Library ✅ (Completed 2026-03-21)
- Complete import/module system (resolution, namespaces, aliasing, cycle detection)
- 12 pure computation stdlib modules with 70 functions
- Modules: math, string, io, testing, json, regex, collections, random, format, result, option, iter

### Phase 4.3: Effect Runtime ✅ (Completed 2026-03-22)
- EffectContext with 5 providers (File, Time, Env, Net, Log)
- Each provider has Real + Mock implementation
- 5 effect-based stdlib modules with 34 functions
- Effect composition: Clone, Derive, EffectStack, MockBuilder
- 13 effect-aware std.testing functions
- 222 effect-related tests across 4 test files

---

## Key Statistics

| Metric | Value |
|--------|-------|
| Built-in methods | 120+ across 6 types (incl. Tuple) |
| Standard library modules | 17 |
| Standard library functions | 117 |
| Effect providers | 5 (File, Time, Env, Net, Log) |
| Total tests | 963 |
| Interpreter tests | 830 |
| Phases complete | 1, 2, 3, 3.1.1, 3.2-chunk1, 3.2-chunk2, 4 |

---

## Effect System Architecture (Complete)

```
EffectContext
├── FileProvider  (Real: os    | Mock: in-memory filesystem)
├── TimeProvider  (Real: time  | Mock: controllable clock)
├── EnvProvider   (Real: os    | Mock: in-memory env vars)
├── NetProvider   (Real: http  | Mock: configurable responses)
└── LogProvider   (Real: stdout| Mock: in-memory log storage)

Composition:
├── Clone()          — Copy context with shared providers
├── Derive()         — Override file/time/env providers
├── DeriveWithNetLog() — Override net/log providers
├── EffectStack      — Nested effect scopes
└── MockBuilder      — Fluent API for test contexts

Standard Library Modules (17 total):
├── Pure: math, string, io, json, regex, collections, random, format, result, option, iter
├── Testing: testing (23 functions incl. effect-aware)
└── Effect: file (9), time (8), env (6), net (5), log (6)
```

---

## Test Breakdown

| Package | Tests | Coverage |
|---------|-------|----------|
| pkg/checker | 61 | Type checking, effects, specs, exhaustiveness |
| pkg/formatter | 9 | Round-trip formatting |
| pkg/lexer | 11 | Tokenization |
| pkg/module | 17 | Module resolution |
| pkg/parser | 16 | All language constructs |
| pkg/symbols | 9 | Symbol table, scopes |
| pkg/types | 26 | Type system, subtyping |
| pkg/interpreter | 845 | Full runtime + stdlib + effects + tuples + match expr + structured patterns + guards/or-patterns/as |
| **Total** | **994** | **All passing** |

---

## ✅ MERGED — Phase 3.1.1: Tuple Literal Syntax (v0.8.1)

> Completed 2026-03-23. PR #3 merged to main 2026-03-23. Branch deleted.

### What Was Delivered

- [x] `let t = (1, 2, 3)` creates a tuple
- [x] `let (x, y) = (1, 2)` destructures correctly
- [x] `t.len()` returns 3
- [x] `t.get(0)` returns `Some(1)`
- [x] `t.to_list()` returns `[1, 2, 3]`
- [x] `(1, 2) == (1, 2)` is `true`
- [x] `(1,)` is a single-element tuple, not a grouped expression
- [x] `()` empty tuple support
- [x] All existing 870 tests still pass
- [x] 34 new tuple tests added and passing (905 total)
- [x] 12 tuple methods: len, length, get, to_list, is_empty, contains, first, last, reverse, map, for_each, enumerate, zip
- [x] Tuple destructuring with wildcards: `let (x, _, z) = t`
- [x] List destructuring support: `let (a, b) = [5, 10]`
- [x] Tuple iteration: `for x in tuple:`

### Files Changed

| File | Action |
|------|--------|
| `pkg/ast/ast.go` | Added `TupleLiteral` and `LetTupleDestructure` nodes |
| `pkg/parser/parser.go` | Tuple parsing, destructuring in let statements |
| `pkg/interpreter/eval.go` | Tuple evaluation, destructuring, for-loop iteration |
| `pkg/interpreter/methods_tuple.go` | **NEW** — 12 tuple methods |
| `pkg/interpreter/tuple_test.go` | **NEW** — 34 tests |

---

## ✅ COMPLETE — Phase 3.2 Chunk 1: Pattern Matching Core Infrastructure

> Completed 2026-03-23. PR #4 merged. 930 tests passing on v0.9.0-alpha.1.

### What Was Delivered (Chunk 1)
- [x] `MatchExpr` AST node (expression form with `->` syntax)
- [x] `MatchArm` AST node (pattern -> expression pair)
- [x] Parser: `parseMatchExpr()` for expression-form match
- [x] Parser: `parseMatchStmtOrExpr()` auto-disambiguation (case vs arrow)
- [x] Parser: `blockExprJustEnded` flag for proper indentation handling
- [x] Interpreter: `evalMatchExpr()` with first-match semantics
- [x] Float literal pattern matching (was missing)
- [x] Literal patterns: Int, Float, String, Bool, None
- [x] Variable binding patterns
- [x] Wildcard patterns
- [x] No-match runtime error
- [x] 25 new tests added (930 total)

### Files Changed
| File | Action |
|------|--------|
| `pkg/ast/ast.go` | Added `MatchExpr`, `MatchArm` types |
| `pkg/parser/parser.go` | Added `parseMatchExpr()`, `parseMatchStmtOrExpr()`, `blockExprJustEnded` |
| `pkg/interpreter/eval.go` | Added `evalMatchExpr()`, float literal matching |
| `pkg/interpreter/match_test.go` | **NEW** — 25 match expression tests |

---

## ✅ COMPLETE — Phase 3.2 Chunk 2: Structured Data Patterns

> Completed 2026-03-23. 963 tests passing on v0.9.0-alpha.2.

### What Was Delivered (Chunk 2)
- [x] Tuple patterns: `(0, 0) -> "origin"`, `(x, y) -> "point"`, `(_, 2) -> ...`
- [x] List patterns: `[] -> "empty"`, `[x] -> "single"`, `[x, y] -> "pair"`
- [x] Spread patterns: `[first, ...rest] -> ...`, `[a, ...middle, last] -> ...`
- [x] Constructor patterns: `Some(x)`, `None`, `Ok(v)`, `Err(e)`
- [x] Nested patterns: `Some((x, y))`, `Ok(Some((x, y)))`, `[(a, b)]`
- [x] New `DOTDOTDOT` token and `SpreadPattern` AST node
- [x] 33 new tests added (963 total)

### Files Changed
| File | Action |
|------|--------|
| `pkg/token/token.go` | Added `DOTDOTDOT` token type |
| `pkg/lexer/lexer.go` | Lexer recognizes `...` as `DOTDOTDOT` |
| `pkg/ast/ast.go` | Added `SpreadPattern` type |
| `pkg/parser/parser.go` | Parse `...ident` as spread pattern |
| `pkg/interpreter/eval.go` | New `matchListPattern()` with spread support |
| `pkg/interpreter/match_structured_test.go` | **NEW** — 33 structured pattern tests |

---

## ✅ COMPLETE — Phase 3.2 Chunk 3: Guard Clauses, Or-patterns, Binding Patterns

> Completed 2026-03-23. 982 tests passing on v0.9.0-alpha.3.
> Guard clauses (`if`), or-patterns (`|`), and binding patterns (`as`) all implemented.

### Phase 3.2 Remaining Chunks

| Chunk | Focus | Effort | Tests | Status |
|-------|-------|--------|-------|--------|
| **1** | Core infrastructure + Literal/Variable/Wildcard patterns | 3–4 days | 25 | ✅ DONE |
| **2** | Tuple, List & Constructor patterns + Spread | 3–4 days | 33 | ✅ DONE |
| **3** | Guard clauses, Or-patterns, Binding patterns | 2–3 days | 19 | ✅ DONE |
| **4** | Exhaustiveness checking | 1 day | 12 | ✅ DONE |

### Phase 3.2 COMPLETE — v0.9.0, 994 total tests

---

## ✅ COMPLETE — Phase 3.2 Chunk 4: Exhaustiveness Checking

> Completed 2026-03-23. 994 tests passing on v0.9.0.

### What Was Delivered (Chunk 4)
- [x] `patternCoversVariants()` — recursive helper, handles OrPattern + AsPattern
- [x] `patternCoversBoolLiterals()` — recursive helper for Bool exhaustiveness
- [x] Refactored `checkEnumExhaustiveness` to use `patternCoversVariants`
- [x] `checkBoolExhaustivenessPats()` — shared Bool checker for stmt + expr forms
- [x] `inferMatchExpr()` — full type inference for match expressions with exhaustiveness
- [x] `checkMatchExprEnumExhaustiveness()` — exhaustiveness for expression form
- [x] `*ast.MatchExpr` case wired into `inferExpr` switch
- [x] Bool exhaustiveness added to `checkMatchStmt`
- [x] 12 new checker tests in `match_exhaustiveness_test.go`
- [x] Guard type checking (guard must be Bool) in both stmt + expr forms

### Files Changed
| File | Action |
|------|--------|
| `pkg/checker/checker.go` | Added 6 new functions/methods; wired MatchExpr into inferExpr; added bool exhaustiveness to checkMatchStmt |
| `pkg/checker/match_exhaustiveness_test.go` | **NEW** — 12 exhaustiveness tests |

---

## 🚀 NEXT PRIORITY — Phase 3.3: Advanced Type Features

> **⚡ START HERE for the next session.**
>
> Phase 3.2 complete. Main is clean with 994 tests passing on v0.9.0.
> Phase 3.3 advances the type system with generics, type inference, and interfaces.

---

## Strategic Plan — Path to v2.0.0

### Recommended Implementation Order

The following order maximizes value delivery while building on completed foundations:

```
Phase 3.1.1 → Phase 3.2 → Phase 3.3 → Phase 5 → Phase 6
Tuple         Pattern      Advanced     Tooling &   Compiler &
Literals      Matching     Type System  Ecosystem   Optimization
(v0.8.1)      (v0.9.0)     (v1.0.0)     (v1.1.0)    (v2.0.0)
```

### Why This Order?

1. **Phase 3.1.1 (Tuple Literals) IMMEDIATE** — Tuples are a foundational data type missing from Phase 3.1. They are a quick 1–2 day implementation that completes Phase 3.1 fully. Pattern matching (Phase 3.2) will heavily use tuple destructuring, so having tuples first avoids rework and enables cleaner pattern matching design.

2. **Phase 3.2 (Pattern Matching) next** — Pattern matching is already partially implemented (basic `match` works). Completing it with exhaustiveness checking, nested patterns, guard clauses, and destructuring (including tuple destructuring!) will immediately improve the expressiveness of existing code. This is a high-impact, bounded-scope enhancement.

3. **Phase 3.3 (Advanced Type Features) third** — Generics, type inference improvements, and interface types build on the type system already in place. These are prerequisites for writing truly reusable libraries and are critical for the AI code generation workflow (AI needs strong types to generate correct code).

4. **Phase 5 (Tooling & Ecosystem) fourth** — LSP, package manager, and AI integration are the "developer experience" layer. They're most valuable *after* the language features are stable. Building tooling on top of an incomplete type system would require rework.

5. **Phase 6 (Compiler & Optimization) last** — The tree-walk interpreter is sufficient for correctness and development. Compilation to bytecode/native is a performance optimization that only matters at scale. It should come after the language is feature-complete.

### Phase Timeline & Milestones

| Phase | Focus | Effort | Version | Key Deliverables |
|-------|-------|--------|---------|------------------|
| **3.1.1** | **Tuple Literals** ✅ | **DONE** | **v0.8.1** | **Tuple parsing, destructuring, 12 methods, 34 tests** |
| **3.2** | Pattern Matching | 2–3 weeks | **v0.9.0** | Exhaustive patterns, guards, nested destructuring, `when` clauses |
| **3.3** | Advanced Type Features | 3–4 weeks | **v1.0.0** | Generics, improved inference, interface types, type constraints |
| **5** | Tooling & Ecosystem | 4–6 weeks | **v1.1.0** | LSP server, package manager, AI integration, doc generator |
| **6** | Compiler & Optimization | 6–8 weeks | **v2.0.0** | Bytecode compiler, VM, GC, performance optimizations |

### Next Session After Tuples: Phase 3.2 — Pattern Matching

After Phase 3.1.1 is complete, Phase 3.2 is the next priority:

Key tasks for Phase 3.2:
1. Nested pattern matching (patterns within patterns)
2. Guard clauses (`when` conditions on match arms)
3. Or-patterns (`A | B => ...`)
4. Binding patterns (`x @ Pattern`)
5. Exhaustiveness checking for all pattern types
6. Destructuring in `let` bindings (building on tuple destructuring from 3.1.1)
7. Wildcard patterns with type narrowing

**Estimated test additions:** ~80–120 new tests

---

## Files Summary

### Core Implementation
- `pkg/interpreter/effect.go` — EffectContext, 5 providers (Real + Mock)
- `pkg/interpreter/stdlib_*.go` — 16 stdlib module files
- `pkg/interpreter/methods_*.go` — 4 method files (108+ methods)
- `pkg/module/resolver.go` — Module resolution system

### Documentation
- `ROADMAP.md` — Full development roadmap, all phases
- `CHANGELOG.md` — Detailed changelog with all versions
- `DEVELOPMENT.md` — Architecture, checklists, contribution guide
- `README.md` — Project overview with complete feature list
- `user_docs/method_reference.md` — Complete method & stdlib reference
- `AI_NEXT_SESSION.md` — This file
