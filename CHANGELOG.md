# Changelog

All notable changes to the Aura toolchain are documented here.

Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

---

## [v1.0.0] — 2026-03-24

### Issue #11: String Concatenation O(n²) Fix

Expression-level string concatenation chains (`a + b + c + d + …`) now run in O(n) time and produce a single allocation via `strings.Builder` instead of one intermediate string per `+`.

### Added
- `collectConcatLeaves(expr ast.Expr) []ast.Expr` in `pkg/interpreter/eval.go` — walks the left-associative `+` spine and returns all leaf expressions in order
- `evalConcatChain(e, leaves, env)` — evaluates all leaves then builds result with `strings.Builder` for string chains; falls back to left-fold via `evalAdd` for non-string types
- `TestStringConcatChain` in `pkg/interpreter/interpreter_test.go`

### Changed
- `evalBinaryOp` — intercepts `+` chains of 3+ operands before general evaluation, routing them through `evalConcatChain`

### Phase 3.3 Chunk 4: Improved Type Inference

Bidirectional type inference for `let` bindings with annotations. When a `let` has both a type annotation and a value, the annotation is used as a *hint* to check constructor arguments and infer empty collection element types.

### Added
- `inferExprWithHint(expr ast.Expr, hint *types.Type) *types.Type` in `pkg/checker/checker.go` — targeted bidirectional inference; handles `[]` → `[T]`, `{}` → `{K:V}`, `Some(x)` → `Option[T]`, `Ok(x)`/`Err(x)` → `Result[T,E]`
- `pkg/checker/inference_test.go` — **NEW**: 13 tests (empty list/map with annotation, `Some/Ok/Err` match/mismatch, generic type aliases)
- `pkg/interpreter/inference_test.go` — **NEW**: 11 tests (empty collections at runtime, constructor match via annotated binding, generic alias runtime behavior)

### Changed
- `checkLetStmt` — resolves annotation first when present; passes resolved type as hint to `inferExprWithHint` instead of calling `inferExpr` directly

### Files Changed
- `pkg/checker/checker.go` — Added `inferExprWithHint`; updated `checkLetStmt`
- `pkg/checker/inference_test.go` — **NEW**: 13 tests
- `pkg/interpreter/inference_test.go` — **NEW**: 11 tests
- `pkg/interpreter/interpreter_test.go` — Added `TestStringConcatChain`
- `pkg/interpreter/eval.go` — Added `collectConcatLeaves`, `evalConcatChain`

**Total tests:** 1121 (all passing, race-clean)

---

## [v1.0.0-alpha.3] — 2026-03-24

### Phase 3.3 Chunk 3: Type Constraints and `where` Clauses

Generic functions can now declare type constraints: `fn show[T](x: T) -> String where T: Printable`. The compiler enforces at every call site that the concrete type satisfies the required interface.

### Added
- `TypeConstraint struct { TypeParam, TraitName string }` in `pkg/ast/ast.go`
- `Constraints []TypeConstraint` field on `FnDef` and `FnSignature`
- `parseWhereConstraints()` in `pkg/parser/parser.go` — parses comma-separated `T: TraitName` entries after effects
- Disambiguation guard in `parseOptionOrRefinementType` — `WHERE TYPE_IDENT COLON` pattern is left for `parseFnDef` to consume (not treated as a refinement); preserves existing `Int where self >= 0` behavior
- `ErrConstraintNotSatisfied ErrorCode = "CONSTRAINT_NOT_SATISFIED"` in `pkg/checker/errors.go`
- `fnConstraints map[string][]ast.TypeConstraint` field on `Checker` (parallel to `fnEffects`)
- `validateConstraintDeclarations()` pass (Pass 3.6) — validates all where-clause trait names exist and are `KindInterface`
- `pkg/checker/constraints_test.go` — **NEW**: 14 tests (declaration, undefined trait, non-trait, call-site satisfied/unsatisfied, multiple constraints, effects+constraints, multi-TypeParam, unconstrained generic, trait signature, refinement guard)
- `pkg/interpreter/constraints_test.go` — **NEW**: 10 tests (runtime execution of constrained fns, multiple constraints, two TypeParams, impl method dispatch in body, refinement types still enforced)

### Changed
- `registerFnDef` — stores `fn.Constraints` in `fnConstraints`
- `registerImplMethod` — stores `fn.Constraints` in `fnConstraints` under qualified name
- `inferCallExpr` generic branch — after `collectTypeBindings` infers type arg mappings, checks each constraint: emits `ErrConstraintNotSatisfied` if concrete type is a struct that doesn't satisfy the required interface
- `Check()` — wires in `validateConstraintDeclarations()` as Pass 3.6 between `validateImplBlocks()` and `registerConstants()`

### Files Changed
- `pkg/ast/ast.go` — Added `TypeConstraint`; added `Constraints` to `FnDef` and `FnSignature`
- `pkg/parser/parser.go` — Added `parseWhereConstraints`, disambiguation guard, wired into `parseFnDef` and `parseFnSignature`
- `pkg/checker/checker.go` — Added `fnConstraints` field, `validateConstraintDeclarations`, constraint check in `inferCallExpr`
- `pkg/checker/errors.go` — Added `ErrConstraintNotSatisfied`
- `pkg/checker/constraints_test.go` — **NEW**: 14 tests
- `pkg/interpreter/constraints_test.go` — **NEW**: 10 tests

**Total tests:** 1096 (all passing, race-clean)

---

## [v1.0.0-alpha.2] — 2026-03-23

### Phase 3.3 Chunk 2: Interface Types

Structural interfaces ("Go-style traits") are now fully supported. A struct satisfies an interface if it has all required methods — checked statically by the type checker and dispatched dynamically at runtime.

### Added
- `KindInterface` type kind and `NewInterfaceType(name string, methods []*Field) *Type` constructor in `pkg/types/types.go`
- `parseFnSignature(vis Visibility) *FnSignature` in `pkg/parser/parser.go` — parses function signatures without bodies for use in trait/interface declarations
- `ErrMissingMethod ErrorCode = "MISSING_METHOD"` in `pkg/checker/errors.go`
- `validateImplBlocks()` pass in checker (Pass 3.5) — verifies all trait methods are implemented and return types match
- `structSatisfiesInterface(structType, ifaceType *types.Type) bool` checker helper — used in `checkLetStmt` and `inferCallExpr`
- `resolveImplTargetName(te ast.TypeExpr) string` and `registerImplMethod(typeName string, fn *ast.FnDef)` in checker — registers impl methods as `"TypeName.methodName"` to avoid collisions
- `implMethods map[string]map[string]*FunctionVal` field on `Environment` — global dispatch table, stored at root env
- `DefineImplMethod(typeName, methodName string, fn *FunctionVal)` and `GetImplMethod(typeName, methodName string) (*FunctionVal, bool)` on `Environment`
- `pkg/checker/interfaces_test.go` — **NEW**: 16 tests (trait declarations, impl validation, interface annotations, structural satisfaction, error cases)
- `pkg/interpreter/interfaces_test.go` — **NEW**: 12 tests (method calls, field access, multiple types, inherent impls, interface dispatch, multi-param methods)

### Changed
- `parseTraitDef` — now calls `parseFnSignature` instead of `parseFnDef` for trait members (no body expected in trait)
- `registerTraitDef` — now builds a real `KindInterface` type and registers it in `typeReg` (was a no-op)
- `registerFunctions` — impl block handling now calls `registerImplMethod` with qualified names (was skipped)
- `IsAssignableTo` — `KindInterface` + `KindStruct` from-type returns `true` (structural check is deferred to checker)
- `evalFieldAccess` `*StructVal` case — checks `env.GetImplMethod` before panicking; returns `*BuiltinFnVal` closure that prepends receiver as first arg
- Module init loop in interpreter — `*ast.ImplBlock` now registers all methods via `DefineImplMethod`

### Files Changed
- `pkg/parser/parser.go` — Added `parseFnSignature`; changed `parseTraitDef` to use it
- `pkg/types/types.go` — Added `KindInterface`, `NewInterfaceType`; updated `String()`, `SubstituteTypeParams`, `IsAssignableTo`
- `pkg/checker/checker.go` — Added `validateImplBlocks`, `structSatisfiesInterface`, `resolveImplTargetName`, `registerImplMethod`; wired `validateImplBlocks` into `Check()`; updated `checkLetStmt` and `inferCallExpr`
- `pkg/checker/errors.go` — Added `ErrMissingMethod`
- `pkg/interpreter/env.go` — Added `implMethods` field, `root()`, `DefineImplMethod`, `GetImplMethod`
- `pkg/interpreter/interpreter.go` — `*ast.ImplBlock` case now registers impl methods
- `pkg/interpreter/eval.go` — `*StructVal` case in `evalFieldAccess` now dispatches to impl methods
- `pkg/checker/interfaces_test.go` — **NEW**: 16 tests
- `pkg/interpreter/interfaces_test.go` — **NEW**: 12 tests

**Total tests:** 1072 (all passing, race-clean)

---

## [v1.0.0-alpha.1] — 2026-03-23

### Phase 3.3 Chunk 1: Generic Types and Functions

Generic type parameters are now fully supported across the checker and work transparently in the interpreter (type erasure).

### Added
- `SubstituteTypeParams(bindings map[string]*Type) *Type` on `*types.Type` — recursively replaces `KindTypeParam` nodes with bound concrete types; handles all 19 type kinds; never mutates the registry type
- `withTypeParams(params []string, fn func())` on `Checker` — temporarily activates type parameter bindings; used by all 4 registration passes and body checking so `T` resolves correctly instead of emitting `ErrUndefinedType`
- `collectTypeBindings(param, arg *types.Type, bindings map[string]*types.Type)` — package-level helper that infers `{T → ConcreteType}` substitution mappings by walking parameter/argument type pairs at call sites
- `typeParamBindings map[string]*types.Type` field on `Checker` (nil = not in generic context)
- `IsAssignableTo`: `KindTypeParam` is assignable to/from anything (type erasure); List/Set/Option covariance added
- `pkg/checker/generics_test.go` — **NEW**: 25 checker tests (declarations, body checking, call-site inference, struct/enum instantiation, error cases)
- `pkg/interpreter/generics_test.go` — **NEW**: 15 interpreter tests (generic fns, structs, enums, collections, polymorphic use)

### Changed
- `resolveNamedType` — checks `typeParamBindings` before registry lookup; applies full `SubstituteTypeParams` on generic instantiation (was shallow copy); validates type arg count and emits `ErrTypeParamCount`
- `registerTypeDef`, `registerStructDef`, `registerEnumDef`, `registerFnDef` — wrapped with `withTypeParams` so field/param type resolution succeeds for generic definitions
- `checkFnBody` — wrapped with `withTypeParams`; skips default-value type check when param type is a TypeParam placeholder
- `inferCallExpr` — infers type arg bindings and substitutes through return type for generic function calls

### Files Changed
- `pkg/types/types.go` — Added `SubstituteTypeParams`, TypeParam/covariance rules in `IsAssignableTo`
- `pkg/checker/checker.go` — Added `typeParamBindings` field, `withTypeParams`, `collectTypeBindings`; modified `resolveNamedType`, `checkFnBody`, `inferCallExpr`, all 4 registration functions
- `pkg/checker/generics_test.go` — **NEW**: 25 tests
- `pkg/interpreter/generics_test.go` — **NEW**: 15 tests

**No changes to:** `pkg/ast/`, `pkg/parser/`, `pkg/interpreter/` runtime code (scaffolding was already in place; interpreter works dynamically)

**Total tests:** 1044 (all passing)

---

## [v0.9.1] — 2026-03-23

### Refinement Type Runtime Enforcement (Issue #8)

Refinement type predicates are now evaluated at runtime, not just at compile time.

### Added
- `enforceRefinement()` in `pkg/interpreter/eval.go` — evaluates the predicate from `ast.RefinementType` with `self` bound to the assigned value; panics with a clear message on violation
- Named type alias resolution — `type Priority = Int where self >= 1 and self <= 5` is enforced anywhere `Priority` is used as a type hint
- `typeExprs map[string]ast.TypeExpr` in `Environment` — registered during module init for both main and imported modules
- Enforcement at `let` binding sites and function parameter binding
- 10 new tests in `pkg/interpreter/refinement_test.go`

### Files Changed
- `pkg/interpreter/env.go` — Added `typeExprs` map, `DefineTypeExpr`, `GetTypeExpr`
- `pkg/interpreter/eval.go` — Added `enforceRefinement()`, wired into `execLetStmt` and `callUserFn`
- `pkg/interpreter/interpreter.go` — `TypeDef` bodies now registered during module init
- `pkg/interpreter/refinement_test.go` — **NEW**: 10 refinement enforcement tests

**Total tests:** 1004 (all passing)

---

## [v0.9.0] — 2026-03-23

### Phase 3.2 Chunk 4: Exhaustiveness Checking

Static exhaustiveness checking for match expressions and statements.

### Added
- `patternCoversVariants()` — recursive helper, handles `OrPattern` and `AsPattern`
- `patternCoversBoolLiterals()` — Bool exhaustiveness helper
- `checkBoolExhaustivenessPats()` — Bool exhaustiveness for stmt and expr forms
- `inferMatchExpr()` — full type inference for match expressions with exhaustiveness
- `checkMatchExprEnumExhaustiveness()` — enum exhaustiveness for expression form
- `*ast.MatchExpr` case wired into `inferExpr` switch
- Bool exhaustiveness check added to `checkMatchStmt`
- 12 new tests in `pkg/checker/match_exhaustiveness_test.go`

**Total tests:** 994 (all passing)

---

## [v0.9.0-alpha.3] — 2026-03-23

### Phase 3.2 Chunk 3: Guard Clauses, Or-patterns, Binding Patterns

### Added
- **Guard clauses**: `x if x > 0 -> "positive"` on match arms
- **Or-patterns**: `1 | 2 | 3 -> "low"` — multiple patterns share one arm
- **Binding patterns**: `pattern as name` — bind the matched value to a name
- Parser: `parsePattern()` wrapper handles `|` chaining and `as` binding
- Parser: `blockExprDepth int` counter replaces `blockExprJustEnded bool`
- Interpreter: `matchPattern()` handles `OrPattern` and `AsPattern`
- `MatchArm` gets optional `Guard Expr` field
- 19 new tests in `pkg/interpreter/match_chunk3_test.go`

**Total tests:** 982 (all passing)

---

## [v0.9.0-alpha.2] — 2026-03-23

### Phase 3.2 Chunk 2: Structured Data Patterns

Structured data pattern matching — tuple, list, constructor, spread, and nested patterns.

### Added
- **Tuple patterns**: `(0, 0) -> "origin"`, `(x, y) -> "point"`, `(_, 2) -> ...`
  - Match by structure and element count
  - Support literal values, variable bindings, and wildcards in elements
  - Nested tuple matching: `((1, 2), 3) -> ...`
- **List patterns**: `[] -> "empty"`, `[x] -> "single"`, `[x, y] -> "pair"`
  - Exact-length matching with element patterns
  - Support literal values, variable bindings, and wildcards
- **Spread patterns**: `[first, ...rest] -> ...`, `[a, b, ...middle, last] -> ...`
  - New `...` (ellipsis) token in lexer
  - Binds remaining list elements to a variable as a new list
  - Works with elements before and after the spread
  - Empty rest produces an empty list
- **Constructor patterns**: `Some(x) -> ...`, `None -> ...`, `Ok(v) -> ...`, `Err(e) -> ...`
  - Match Option types: `Some(value)` and `None`
  - Match Result types: `Ok(value)` and `Err(error)`
  - Support literal matching inside constructors: `Some(42) -> "exact"`
  - Support wildcard inside constructors: `Some(_) -> "has something"`
- **Nested patterns**: Compose all pattern types freely
  - Constructor + tuple: `Some((x, y)) -> "point"`
  - Constructor + list + spread: `Some([first, ...rest]) -> ...`
  - Tuple + constructor: `(Some(1), None) -> ...`
  - List + tuple: `[(a, b)] -> "pair"`
  - Deep nesting: `Ok(Some((x, y))) -> ...`
- **33 new tests** for structured data patterns, bringing total to **963 passing tests**

### Token Changes
- New `DOTDOTDOT` token type (`...`) for spread patterns

### AST Changes
- New `SpreadPattern` node with `Name` field for variable binding

### Files Changed
- `pkg/token/token.go` — Added `DOTDOTDOT` token type and string mapping
- `pkg/lexer/lexer.go` — Lexer recognizes `...` as `DOTDOTDOT` token
- `pkg/ast/ast.go` — Added `SpreadPattern` type implementing `Pattern` interface
- `pkg/parser/parser.go` — Parse `...ident` as spread pattern in list patterns
- `pkg/interpreter/eval.go` — New `matchListPattern()` with spread support
- `pkg/interpreter/match_structured_test.go` — **NEW**: 33 structured pattern tests

---

## [v0.9.0-alpha.1] — 2026-03-23

### Phase 3.2 Chunk 1: Pattern Matching Core Infrastructure

First chunk of pattern matching support — match expressions with basic patterns.

### Added
- **Match expression syntax**: `match value: pattern -> expr` as an expression (returns a value)
  - Distinct from existing `match` statement (which uses `case` keyword + block body)
  - Can be used in `let` bindings: `let result = match x: 1 -> "one" ...`
  - Can be used in `return` statements: `return match x: ...`
  - Supports nesting: `match x: 0 -> "zero" _ -> match y: ...`
- **Literal patterns**: Match against `Int`, `Float`, `String`, `Bool`, `None` values
- **Variable binding patterns**: `n -> n * 2` binds matched value to a variable
- **Wildcard pattern**: `_ -> default_value` matches anything
- **First-match semantics**: Returns result of first matching arm
- **No-match runtime error**: Panics with descriptive error if no pattern matches
- **Float literal matching**: Added missing float support to `matchLiteralPattern`
- **Parser disambiguation**: Automatically detects `case`-based (statement) vs `arrow`-based (expression) match syntax
- **25 new tests** for match expressions, bringing total to **930 passing tests**

### AST Changes
- New `MatchExpr` node (expression returning a value)
- New `MatchArm` node (pattern -> expression pair)

### Files Changed
- `pkg/ast/ast.go` — Added `MatchExpr`, `MatchArm` types
- `pkg/parser/parser.go` — Added `parseMatchExpr()`, `parseMatchStmtOrExpr()` disambiguation, `blockExprJustEnded` flag for indentation handling
- `pkg/interpreter/eval.go` — Added `evalMatchExpr()`, float literal matching in `matchLiteralPattern`
- `pkg/interpreter/match_test.go` — **NEW**: 25 comprehensive match expression tests

---

## [v0.8.1] — 2026-03-23

### Phase 3.1.1: Tuple Literal Syntax & Destructuring

Quick-win release adding first-class tuple support to the Aura language.

### Added
- **Tuple literal syntax**: `(a, b, c)` creates tuples directly in expressions
- **Single-element tuples**: `(42,)` with trailing comma disambiguation
- **Empty tuples**: `()` creates an empty tuple
- **Nested tuples**: `((1, 2), (3, 4))` fully supported
- **Tuple destructuring**: `let (x, y) = point` in let bindings
- **Mutable destructuring**: `let mut (a, b) = (1, 2)` for mutable bindings
- **Wildcard in destructuring**: `let (x, _, z) = tuple` to skip elements
- **List destructuring**: `let (a, b) = [5, 10]` also works with lists
- **Tuple iteration**: `for x in tuple:` support in for loops
- **12 tuple methods**: `len`, `length`, `get`, `to_list`, `is_empty`, `contains`, `first`, `last`, `reverse`, `map`, `for_each`, `enumerate`, `zip`
- **34 new tests** for tuple functionality (905 total tests)

### AST Changes
- Added `TupleLiteral` expression node
- Added `LetTupleDestructure` statement node

### Files Changed
- `pkg/ast/ast.go` — New AST nodes
- `pkg/parser/parser.go` — Tuple parsing and let destructuring
- `pkg/interpreter/eval.go` — Tuple evaluation, destructuring, iteration
- `pkg/interpreter/methods_tuple.go` — New file with 12 tuple methods
- `pkg/interpreter/tuple_test.go` — New file with 34 tests

---

## [v0.8.0] — 2026-03-22

### Phase 4 Complete: Runtime & Standard Library — COMPLETE

Major release completing all of Phase 4 — the full runtime, standard library, and effect system for Aura.

### Added

#### Phase 4.2: Module System & Standard Library (v0.6.0)

##### Module System (`pkg/module/resolver.go`)
- Complete import resolution with namespace management
- Named imports and aliasing (`import std.math as m`)
- Import cycle detection with path reporting
- Module initialization ordering
- 17 tests in `pkg/module/`

##### Pure Computation Standard Library — 12 modules, 70 functions
- `std.math` — 8 functions: `abs`, `max`, `min`, `floor`, `ceil`, `round`, `sqrt`, `pow` + constants (`pi`, `e`, `inf`, `nan`)
- `std.string` — 4 functions: `join`, `split`, `replace`, `repeat`
- `std.io` — 3 functions: `print`, `println`, `format`
- `std.testing` — 10 base functions: `assert`, `assert_eq`, `assert_ne`, `assert_true`, `assert_false`, `assert_some`, `assert_none`, `assert_ok`, `assert_err`, `run_tests`
- `std.json` — 2 functions: `parse`, `stringify` (with pretty-print support)
- `std.regex` — 6 functions: `match`, `find`, `find_all`, `replace`, `split`, `compile`
- `std.collections` — 9 functions: `range`, `zip_with`, `partition`, `group_by`, `chunk`, `take`, `drop`, `take_while`, `drop_while`
- `std.random` — 6 functions: `int`, `float`, `choice`, `shuffle`, `sample`, `seed`
- `std.format` — 7 functions: `pad_left`, `pad_right`, `center`, `truncate`, `wrap`, `indent`, `dedent`
- `std.result` — 5 functions: `all_ok`, `any_ok`, `collect`, `partition_results`, `from_option`
- `std.option` — 5 functions: `all_some`, `any_some`, `collect`, `first_some`, `from_result`
- `std.iter` — 5 functions: `cycle`, `repeat`, `chain`, `interleave`, `pairwise`

##### Tests
- 64 advanced import/module system tests in `import_advanced_test.go`
- 65 stdlib tests in `stdlib_complete_test.go`
- 17 module resolution tests in `pkg/module/`

#### Phase 4.3: Effect Runtime (v0.8.0)

##### Effect System Infrastructure (`effect.go`)
- `EffectContext` with provider pattern — thread-safe capability injection
- 5 effect providers, each with Real + Mock implementations:
  - **FileProvider** — Real: `os` filesystem | Mock: in-memory filesystem
  - **TimeProvider** — Real: `time` package | Mock: controllable clock
  - **EnvProvider** — Real: `os` env vars | Mock: in-memory variables
  - **NetProvider** — Real: `net/http` | Mock: configurable responses with request logging
  - **LogProvider** — Real: stdout | Mock: in-memory storage with query methods

##### Effect-Based Standard Library — 5 modules, 34 functions
- `std.file` — 9 functions: `read`, `write`, `append`, `exists`, `delete`, `list_dir`, `create_dir`, `is_file`, `is_dir`
- `std.time` — 8 functions: `now`, `unix`, `millis`, `sleep`, `format`, `parse`, `add`, `diff`
- `std.env` — 6 functions: `get`, `set`, `remove`, `has`, `all`, `args`
- `std.net` — 5 functions: `get`, `post`, `put`, `delete`, `request`
- `std.log` — 6 functions: `info`, `warn`, `error`, `debug`, `with_context`, `get_logs`

##### Effect Composition & Mocking Framework
- `Clone()`, `Derive()`, `DeriveWithNetLog()` — context manipulation
- `EffectStack` — nested effect scopes
- `MockBuilder` — fluent API for configuring test contexts
- Pre-configured fixtures: `EmptyMockContext`, `FixtureWithFiles`, etc.
- 13 additional effect-aware `std.testing` functions (`with_mock_effects`, `with_effects`, `assert_file_exists`, `mock_time`, `advance_time`, etc.)

##### Tests
- 48 effect foundation tests in `effect_test.go`
- 66 time/env tests in `time_env_test.go`
- 54 effect composition tests in `effect_composition_test.go`
- 54 network/logging tests in `net_log_test.go`
- **Total: 875 tests** across all packages (up from 468)

### Summary

| Metric | Value |
|--------|-------|
| Built-in methods | 108+ across 5 types |
| Standard library modules | 17 |
| Standard library functions | 117 |
| Effect providers | 5 (File, Time, Env, Net, Log) |
| Total tests | 875 |

---

## [v0.4.0] — 2026-03-20

### Phase 4.1: Core Runtime Methods — COMPLETE

Major release adding **108+ built-in methods** across 5 core types, delivered in 4 implementation chunks.

### Added

#### Method Dispatch Infrastructure
- Centralized method registry system (`pkg/interpreter/methods.go`) using `RegisterMethod(ValueType, "name", func)` pattern
- `callValue()` helper for invoking Aura lambdas/closures from Go method implementations
- `cmpValues()` helper for type-safe ordering comparisons (Int, Float, String)
- Registry-based method resolution replaces inline switch statements in `eval.go`

#### String Methods (22) — `methods_string.go`
- Core: `len`, `upper`/`to_upper`, `lower`/`to_lower`, `contains`, `split`, `trim`, `trim_start`, `trim_end`
- Search: `starts_with`, `ends_with`, `index_of` (returns Option), `replace`
- Transform: `repeat`, `reverse`, `chars`, `slice` (with bounds checking)
- Aliases: `length` → `len`

#### List Methods (27) — `methods_list.go`
- Core: `len`/`length`, `append`/`push`, `contains`, `is_empty`
- Safe accessors: `first()`, `last()`, `get(index)` — all return Option
- Mutation: `pop()` (returns Option), `remove(index)`
- Transforms: `reverse()`, `slice(start, end?)` (supports negative indices), `join(sep)`, `index_of(item)` (returns Option)
- Higher-order: `map(fn)`, `filter(fn)`, `reduce(init, fn)`, `for_each(fn)`, `flat_map(fn)`, `flatten()`
- Predicates: `any(fn)`, `all(fn)`, `count(fn?)`
- Utilities: `unique()`, `sum()`, `min()`/`max()` (return Option), `sort()`, `zip(other)`, `enumerate()`

#### Map Methods (24) — `methods_map.go`
- Size/emptiness: `len`/`length`/`size`, `is_empty`
- Key/value access: `keys()`, `values()`, `entries()`, `get(key)` (returns Option), `get_or(key, default)`
- Lookup: `has(key)`, `contains_key(key)`, `contains_value(value)`
- Mutation: `set(key, value)`, `remove(key)` (returns Option), `delete(key)` (returns Bool), `clear()`, `merge(other)`
- Higher-order: `filter(fn)`, `map(fn)`, `for_each(fn)`, `reduce(init, fn)`, `any(fn)`, `all(fn)`, `count(fn?)`
- Utilities: `to_list()`, `find(fn)` (returns Option)

#### Option Methods (17) — `methods_option.go`
- Predicates: `is_some()`, `is_none()`
- Extraction: `unwrap()`, `expect(msg)`, `unwrap_or(default)`, `unwrap_or_else(fn)`
- Monadic transforms: `map(fn)`, `flat_map(fn)`, `and_then(fn)`, `filter(fn)`, `flatten()`
- Combinators: `or_else(fn)`, `or(alt)`, `and(other)`, `zip(other)`
- Query: `contains(value)`
- Conversion: `to_result(err_val)`

#### Result Methods (18) — `methods_option.go`
- Predicates: `is_ok()`, `is_err()`
- Extraction: `unwrap()`, `unwrap_err()`, `expect(msg)`, `unwrap_or(default)`, `unwrap_or_else(fn)`
- Monadic transforms: `map(fn)`, `map_err(fn)`, `and_then(fn)`, `or_else(fn)`, `flatten()`
- Combinators: `or(alt)`, `and(other)`
- Query: `contains(value)`, `contains_err(value)`
- Conversion: `ok()`, `err()`, `to_option()`

#### Tests
- 222 new method-specific tests in `methods_test.go`
- Total test count: **468 tests** across all packages (up from 232)
- Comprehensive coverage including: success cases, error/panic conditions, None/Err edge cases, method chaining, monadic composition, Option↔Result round-trip conversions

### Changed
- `eval.go`: Refactored FieldAccess evaluation to use method registry instead of inline switch statements
- Interpreter package now contains 12 source files (up from 6)

---

## [v0.3.1] — 2026-03-19

### Added
- **String interpolation** — `"Hello, {name}!"` with full expression support in lexer, parser, and interpreter
- **Pipeline operator** (`|>`) — Lexer tokenization, parser precedence handling, interpreter evaluation with lambda support
- **Option chaining** (`?.`) — None short-circuiting for `?` postfix operator
- 14 new pipeline operator tests (232+ total tests)

---

## [v0.3.0] — 2026-03-17

### Phase 3: Tree-Walk Interpreter — COMPLETE

### Added
- Tree-walk interpreter (`pkg/interpreter/`) with value system, environment, evaluator, module runner, test runner
- CLI commands: `aura run`, `aura test`, `aura repl`
- Full expression/statement evaluation (arithmetic, comparison, logic, control flow, structs, enums, match, closures, lambdas, list comprehensions)
- 14 built-in functions: `print`, `len`, `str`, `int`, `float`, `range`, `type_of`, `abs`, `min`, `max`, `Ok`, `Err`, `Some`, `None`
- 112 interpreter tests (211 total)

---

## [v0.2.0] — 2026-03-17

### Phase 2: Semantic Analysis — COMPLETE

### Added
- Symbol table and scope management (`pkg/symbols/`) — 9 tests
- Type system representation with subtyping (`pkg/types/`) — 26 tests
- Multi-pass type checker (`pkg/checker/`) — 48 tests
- AI-parseable structured error output with JSON format
- CLI `aura check` command with `--json` flag
- 83 new tests (119 total)

---

## [v0.1.0] — 2026-03-17

### Phase 1: Syntax — COMPLETE

### Added
- Indentation-sensitive lexer (`pkg/lexer/`) — 11 tests
- Recursive descent parser (`pkg/parser/`) — 16 tests
- Complete AST node definitions (`pkg/ast/`)
- Canonical source formatter (`pkg/formatter/`) — 9 tests with round-trip guarantee
- CLI entry point with `format` and `parse` commands
- 36 tests total
