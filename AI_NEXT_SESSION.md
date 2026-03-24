# AI Next Session

> See `AI_MISSION.md` §Session Protocol for how to maintain this file.

---

## State

| | |
|--|--|
| **Version** | v0.9.1 |
| **Tests** | 1004 (all passing) |
| **Phases complete** | 1, 2, 3.1, 3.1.1, 3.2 (all chunks), 4, Issue #8 |

| Package | Tests |
|---------|-------|
| pkg/checker | 61 |
| pkg/formatter | 9 |
| pkg/interpreter | 855 |
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
| 11 | String concatenation O(n²) | 📋 Documented — low priority |

---

## Next: Phase 3.3 — Advanced Type Features

**Target:** v1.0.0 | **Est. tests:** +100–150 | **Roadmap:** `ROADMAP.md` §3.3

### Design decisions to settle first

1. **Generic syntax** — use `[T]` notation: `fn identity[T](x: T) -> T`. Confirm before touching the parser (many parse paths affected).
2. **Trait vs Interface** — the AST already has `TraitDef` and the `trait` keyword. Reuse it; call it "interface" only in user docs. No new keyword.
3. **Monomorphization vs type erasure** — use type erasure in the interpreter (pass types at runtime via `Any`-typed dispatch). Monomorphization is a Phase 6 (compiler) concern.
4. **Higher-kinded types** — do NOT implement in Phase 3.3. Defer unless a concrete stdlib use case demands it.

### Chunk 1 — Generic Types and Functions (v1.0.0-alpha.1)

**Effort:** ~1 week | **Tests:** ~40

**What to build:**
- Generic function declarations: `fn identity[T](x: T) -> T`
- Generic struct declarations: `struct Pair[A, B]: first: A; second: B`
- Generic enum declarations: `enum Tree[T]: Leaf; Node(T, Tree[T], Tree[T])`
- Type parameter substitution at instantiation
- Type argument inference at call sites: `identity(42)` infers `T = Int`

**Key files:** `pkg/ast/ast.go`, `pkg/parser/parser.go`, `pkg/types/types.go`, `pkg/checker/checker.go`, `pkg/interpreter/eval.go`

**Non-obvious constraint:** `TypeDef`, `StructDef`, `EnumDef`, and `FnDef` AST nodes already have `TypeParams []string` — the field exists but is never used. The interpreter needs to bind type params during instantiation.

### Chunk 2 — Interface Types (v1.0.0-alpha.2)

**Effort:** ~1 week | **Tests:** ~30

**What to build:**
- Interface declaration reuses `TraitDef` AST node (already exists)
- Structural typing: a struct satisfies an interface if it has all required methods
- Interface type in function signatures: `fn display(x: Printable)`
- Interface satisfaction check in the type checker
- Runtime dispatch through interfaces in the interpreter

### Chunk 3 — Type Constraints and `where` Clauses (v1.0.0-alpha.3)

**Effort:** ~4 days | **Tests:** ~25

**What to build:**
- `where` clauses on generic functions: `fn show[T](x: T) -> String where T: Printable`
- Constraint satisfaction checking at instantiation
- Multiple constraints: `where T: Printable, T: Comparable`
- Built-in constraints: `Eq`, `Ord`, `Display`

**Key file:** need a new `WhereClause` AST node, or extend `FnDef`.

### Chunk 4 — Improved Type Inference (v1.0.0)

**Effort:** ~4 days | **Tests:** ~20

**What to build:**
- Push expected type into sub-expressions (bidirectional inference improvements)
- Type alias with generic parameters: `type Result[T] = Ok(T) | Err(String)`
- Empty collection inference: `let xs: List[Int] = []` infers element type from annotation
