# AI Next Session - Aura Language

---

## Current State — v0.9.0 (2026-03-23)

| Metric | Value |
|--------|-------|
| Total tests | **994** (all passing) |
| Built-in methods | 120+ across 6 types (String, List, Map, Option, Result, Tuple) |
| Standard library modules | 17 |
| Standard library functions | 117 |
| Effect providers | 5 (File, Time, Env, Net, Log) |
| Phases complete | 1, 2, 3.1, 3.1.1, 3.2, 4 |

### Test Breakdown

| Package | Tests |
|---------|-------|
| pkg/checker | 61 |
| pkg/formatter | 9 |
| pkg/lexer | 11 |
| pkg/module | 17 |
| pkg/parser | 16 |
| pkg/symbols | 9 |
| pkg/types | 26 |
| pkg/interpreter | 845 |
| **Total** | **994** |

---

## Open Technical Debt

### Issue #8 — Refinement Type Runtime Enforcement 🟡 DEFERRED

**File:** `pkg/interpreter/eval.go`, `pkg/types/types.go`

- `Int where x > 0` is parsed, stored in the AST, and type-checked — but the predicate is **never evaluated at runtime**
- A value of `-5` can be assigned to a `PositiveInt` variable without error
- There is a `TODO` comment in `execLetStmt` noting this

**Recommended:** Tackle this before Phase 3.3. It is ~1 day, self-contained, and clears the last piece of open technical debt. Refinement types become meaningless without runtime enforcement, and Phase 3.3 will build on the type system.

**Implementation sketch:**
1. In `execLetStmt` (and assignment), when the declared type is a refinement (`types.KindRefinement`), extract the predicate expression from the type definition
2. Bind the refinement variable name to the assigned value in a temporary environment
3. Evaluate the predicate; if false, call `runtimePanic` with a clear message
4. Also enforce in function argument binding (`evalCallExpr`) when param type is a refinement

**Estimated tests:** ~8 new interpreter tests

---

## 🚀 NEXT PRIORITY — Phase 3.3: Advanced Type Features

> **⚡ START HERE after clearing Issue #8.**
>
> Phase 3.2 complete (v0.9.0). Phase 3.3 advances the type system with generics,
> improved inference, interface types, and type constraints.
> Target: v1.0.0 — the first "language complete" milestone.

**Roadmap entry:** `ROADMAP.md` Phase 3.3

**Estimated effort:** 3–4 weeks
**Estimated test additions:** 100–150 new tests
**Target version:** v1.0.0

### Phase 3.3 Chunks

#### Chunk 1 — Generic Types and Functions (v1.0.0-alpha.1)

**Effort:** ~1 week | **Tests:** ~40

**What to build:**
- Generic function declarations: `fn identity[T](x: T) -> T`
- Generic struct declarations: `struct Pair[A, B]: first: A; second: B`
- Generic enum declarations: `enum Tree[T]: Leaf; Node(T, Tree[T], Tree[T])`
- Type parameter substitution during instantiation
- Type argument inference at call sites: `identity(42)` infers `T = Int`
- Monomorphic instantiation in the interpreter (no type erasure needed at this stage)

**Files:** `pkg/ast/ast.go` (TypeParam nodes), `pkg/parser/parser.go` (parse `[T]` syntax), `pkg/types/types.go` (type parameter representation), `pkg/checker/checker.go` (generic inference), `pkg/interpreter/eval.go` (instantiation)

**Key test cases:**
- `fn identity[T](x: T) -> T` works for Int, String, Bool
- `struct Pair[A, B]` constructs and accesses fields correctly
- `fn first[T](xs: List[T]) -> Option[T]` returns correct type
- Type mismatch caught when wrong type argument used

#### Chunk 2 — Interface Types (v1.0.0-alpha.2)

**Effort:** ~1 week | **Tests:** ~30

**What to build:**
- Interface declaration: `interface Printable: fn to_string(self) -> String`
- Structural typing: a struct satisfies an interface if it has all required methods
- Interface type in function signatures: `fn display(x: Printable)`
- Interface satisfaction checking in the type checker
- Runtime dispatch through interfaces in the interpreter

**Files:** `pkg/ast/ast.go`, `pkg/checker/checker.go`, `pkg/interpreter/eval.go`

**Key test cases:**
- Struct with `to_string` method satisfies `Printable`
- Struct missing method fails at type check with clear error
- Function accepting interface works with multiple concrete types
- Interface used as a generic constraint (see Chunk 3)

#### Chunk 3 — Type Constraints and `where` Clauses (v1.0.0-alpha.3)

**Effort:** ~4 days | **Tests:** ~25

**What to build:**
- `where` clauses on generic functions: `fn show[T](x: T) -> String where T: Printable`
- Constraint satisfaction checking at instantiation
- Multiple constraints: `where T: Printable, T: Comparable`
- Built-in constraints: `Eq` (equality), `Ord` (ordering), `Display` (to_string)

**Files:** `pkg/ast/ast.go` (WhereClause node), `pkg/parser/parser.go`, `pkg/checker/checker.go`

**Key test cases:**
- `where T: Eq` allows `==` inside generic function body
- Instantiating with a type that doesn't satisfy constraint → type error
- Multiple constraints combined correctly

#### Chunk 4 — Improved Type Inference + Refinement Types (v1.0.0)

**Effort:** ~4 days | **Tests:** ~20–30

**What to build:**
- Bidirectional type inference improvements (push expected type into sub-expressions)
- Type alias with generic parameters: `type Result[T] = Ok(T) | Err(String)`
- Refinement type static evaluation (carry forward Issue #8 if not done separately)
- Return type inference for simple functions (optional, low priority)

**Files:** `pkg/checker/checker.go`, `pkg/types/types.go`

**Key test cases:**
- `let xs: List[Int] = []` — empty list infers element type from context
- `type StringResult = Result[String]` works as alias
- Refinement predicates evaluated at assignment (if not done in Issue #8)

---

## Phase 3.3 — Implementation Notes

### Design Decisions to Settle First

1. **Generic syntax** — The roadmap uses `[T]` notation. This is the current direction. Confirm before implementing the parser changes, as it touches many parse paths.

2. **Interface vs Trait** — The current codebase has `trait` in the AST (from Phase 1 syntax). Decide: unify `trait` and `interface`, or keep them as separate concepts? The roadmap uses "interface" for structural typing. **Recommendation:** Reuse the `trait` AST node and keyword; call it "interface" in user docs.

3. **Monomorphization vs type erasure** — For the interpreter, monomorphization (generate one copy per concrete type) is simpler but creates more code. Type erasure (pass types at runtime) is more consistent with the existing `Any`-typed dispatch. **Recommendation:** Use type erasure in the interpreter; monomorphization is a Phase 6 (compiler) concern.

4. **Higher-kinded types** — Listed in the roadmap as "if needed for stdlib design." Do not implement in Phase 3.3 unless a concrete stdlib use case demands it. Defer.

---

## Completed Phases — Reference

| Phase | Version | Tests | Highlights |
|-------|---------|-------|------------|
| 1 — Syntax | v0.1 | 36 | Lexer, parser, formatter, round-trip guarantee |
| 2 — Semantic Analysis | v0.2 | 83 | Type checker, effect tracking, spec validation, JSON errors |
| 3.1 — Interpreter | v0.3 | 112 | Tree-walk interpreter, closures, effects, CLI |
| 3.1.1 — Tuples | v0.8.1 | 34 | Tuple literals, destructuring, 12 methods |
| 3.2 — Pattern Matching | v0.9.0 | 89 | Full pattern matching: literals, tuples, lists, constructors, guards, or-patterns, as-patterns, exhaustiveness |
| 4.1 — Runtime Methods | v0.4.0 | 222 | 108+ built-in methods across 5 types |
| 4.2 — Module System | v0.6.0 | 146 | 12 stdlib modules, 70 functions, import resolution |
| 4.3 — Effect Runtime | v0.8.0 | 222 | 5 effect providers (Real + Mock), MockBuilder, EffectStack |

---

## Files Summary

### Core Implementation
- `pkg/interpreter/eval.go` — Tree-walk interpreter (main evaluator)
- `pkg/interpreter/value.go` — Value types (MapVal with lazy strIdx cache)
- `pkg/interpreter/methods_*.go` — Method dispatch for each type
- `pkg/interpreter/stdlib_*.go` — 16 stdlib module files
- `pkg/interpreter/effect.go` — EffectContext, 5 providers
- `pkg/checker/checker.go` — Type checker with pattern exhaustiveness
- `pkg/parser/parser.go` — Recursive descent parser
- `pkg/ast/ast.go` — AST node definitions (incl. OrPattern, AsPattern, MatchExpr)
- `pkg/module/resolver.go` — Module resolution system

### Documentation
- `ROADMAP.md` — Full development roadmap, all phases
- `CHANGELOG.md` — Detailed changelog with all versions
- `DEVELOPMENT.md` — Architecture, checklists, contribution guide
- `README.md` — Project overview
- `user_docs/method_reference.md` — Complete method & stdlib reference
- `AI_MISSION.md` — AI-first design principles (follow in all work)
- `AI_NEXT_SESSION.md` — This file
