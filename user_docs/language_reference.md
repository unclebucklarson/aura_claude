# Aura Language Reference

Formal reference for the Aura programming language — types, syntax, effects, and specifications.

---

## Table of Contents

1. [Lexical Structure](#lexical-structure)
2. [Module System](#module-system)
3. [Type System](#type-system)
4. [Expressions](#expressions)
5. [Statements](#statements)
6. [Declarations](#declarations)
7. [Effect System](#effect-system)
8. [Specifications](#specifications)
9. [Operator Precedence](#operator-precedence)
10. [Reserved Keywords](#reserved-keywords)

---

## Lexical Structure

### Source Encoding

- **Encoding:** UTF-8 (no BOM)
- **File extension:** `.aura`
- **Line endings:** LF (`\n`)
- **Indentation:** 4 spaces per level (tabs are a syntax error)

### Tokens

| Token Type | Examples |
|------------|----------|
| Identifier | `task`, `create_task`, `x` |
| Type Identifier | `Task`, `String`, `TaskError` |
| Integer Literal | `0`, `42`, `-1` |
| Float Literal | `3.14`, `-0.5`, `1.0e10` |
| String Literal | `"hello"`, `"Task {id}"` |
| Boolean Literal | `true`, `false` |
| None Literal | `none` |
| Keyword | `fn`, `let`, `struct`, etc. |
| Operator | `+`, `-`, `*`, `/`, `==`, `|>`, etc. |
| Delimiter | `(`, `)`, `[`, `]`, `{`, `}`, `:`, `,` |
| INDENT | Increase in indentation level |
| DEDENT | Decrease in indentation level |
| NEWLINE | End of line |
| Comment | `# ...` |
| Doc Comment | `## ...` |

### Indentation Rules

1. Blocks are opened by `:` followed by NEWLINE and an increase in indentation
2. Blocks are closed when indentation returns to the previous level
3. Inside `()`, `[]`, and `{}`, newlines and indentation changes are ignored
4. Blank lines between blocks are handled correctly

### String Interpolation

Strings support embedded expressions with `{}`:

```aura
"Hello, {name}!"             # simple variable
"Priority: {task.priority}"  # field access
"Sum: {a + b}"               # expression
```

Escape sequences: `\n`, `\t`, `\\`, `\"`, `\{`, `\}`

---

## Module System

### Module Declaration

```ebnf
module_decl = "module" qualified_name NEWLINE
```

Every file must begin with a module declaration. The module name is a dot-separated path.

### Imports

```ebnf
import_stmt = "import" qualified_name ["as" IDENT] NEWLINE
            | "from" qualified_name "import" import_list NEWLINE

import_list = IDENT {"," IDENT} | "*"
```

**Examples:**

```aura
import std.time as time          # import module with alias
from std.collections import List  # import specific names
from std.io import *              # import all (wildcard)
```

---

## Type System

### Primitive Types

| Type | Description | Default Value | Example |
|------|-------------|---------------|----------|
| `Int` | Arbitrary-precision integer | `0` | `42`, `-7` |
| `Float` | 64-bit IEEE 754 floating point | `0.0` | `3.14`, `1.0e10` |
| `String` | UTF-8 string | `""` | `"hello"` |
| `Bool` | Boolean | `false` | `true`, `false` |
| `None` | Unit / absence of value | `none` | `none` |

### Composite Types

| Syntax | Description | Example |
|--------|-------------|----------|
| `[T]` | List of T | `[Int]`, `[String]` |
| `{T}` | Set of T | `{String}` |
| `{K: V}` | Map from K to V | `{String: Int}` |
| `(T, U)` | Tuple | `(Int, String)` |
| `T?` | Optional (sugar for `Option[T]`) | `String?` |
| `T \| U` | Union type | `"pending" \| "done"` |
| `fn(T) -> U` | Function type | `fn(Int) -> String` |

### Generic Types

Type parameters use square brackets:

```aura
type NonEmptyList[T] = [T] where len >= 1
pub struct Pair[A, B]:
    pub first: A
    pub second: B
```

### Refinement Types

Attach predicates to constrain values:

```ebnf
refinement_type = type_expr "where" predicate
predicate       = predicate "and" predicate
                | predicate "or" predicate
                | "not" predicate
                | IDENT comp_op expr
                | IDENT "matches" STRING_LIT
                | IDENT "in" expr
```

**Built-in predicate vocabulary:**

| Predicate | Applies To | Meaning |
|-----------|------------|----------|
| `len >= n` | String, List | Minimum length |
| `len <= n` | String, List | Maximum length |
| `self >= n` | Int, Float | Minimum value |
| `self <= n` | Int, Float | Maximum value |
| `self matches r"..."` | String | Regex match |
| `self in [...]` | Any equatable | Set membership |

### Option and Result Types

```
Option[T] = Some(T) | None
Result[T, E] = Ok(T) | Err(E)
```

- `T?` is syntactic sugar for `Option[T]`
- The `?` postfix operator propagates errors from `Result`
- The `!` postfix operator force-unwraps an `Option` (panics on `None`)

### Structural Subtyping

Aura uses structural typing for compatibility:

- A struct with **more** fields is a subtype of a struct with **fewer** fields (if shared fields are compatible)
- `(T where P) <: T` always holds
- `(T where P) <: (T where Q)` if P implies Q

### Type Inference

Aura uses bidirectional type inference:

- **Infer mode:** The type of `let x = expr` is inferred from `expr`
- **Check mode:** `let x: T = expr` checks that `typeof(expr) <: T`
- Lambda parameter types are inferred from calling context
- Generic type arguments are inferred from usage

---

## Expressions

### Literals

```aura
42                   # Int
3.14                 # Float
"hello"              # String
true                 # Bool
none                 # None
```

### Binary Operators

| Operator | Operation | Types |
|----------|-----------|-------|
| `+` | Addition | Int, Float, String (concat) |
| `-` | Subtraction | Int, Float |
| `*` | Multiplication | Int, Float |
| `/` | Division | Int, Float |
| `%` | Modulo | Int |
| `**` | Exponentiation | Int, Float |
| `==`, `!=` | Equality | Any |
| `<`, `>`, `<=`, `>=` | Comparison | Int, Float, String |
| `and`, `or` | Logical | Bool |
| `is` | Type/identity check | Any |
| `in` | Membership | Any iterable |
| `\|>` | Pipeline | Any |

### Unary Operators

| Operator | Operation |
|----------|----------|
| `-` | Numeric negation |
| `not` | Logical negation |

### Postfix Operators

| Operator | Operation | Example |
|----------|-----------|----------|
| `.field` | Field access | `task.title` |
| `(args)` | Function call | `add(1, 2)` |
| `[index]` | Index access | `list[0]` |
| `?` | Option chain / error propagate | `user?.name` |
| `!` | Force unwrap | `maybe_task!` |

### Pipeline Operator

```aura
let result = data |> validate |> transform |> save
# equivalent to: save(transform(validate(data)))
```

### List Comprehension

```aura
[expr for variable in iterable]           # basic
[expr for variable in iterable if cond]   # with filter
```

### Lambda Expression

```aura
|params| -> expr                    # expression body
|params|:                           # block body
    statement
    statement
```

### If Expression

```aura
if condition then true_expr else false_expr
```

### Struct Construction

```aura
TypeName(field1: value1, field2: value2)
```

---

## Statements

### Let Binding

```aura
let name = expr                     # immutable, inferred type
let name: Type = expr               # immutable, explicit type
let mut name = expr                  # mutable
```

### Assignment

```aura
variable = expr                     # simple
object.field = expr                 # field
list[index] = expr                  # index
```

### Return

```aura
return expr
return                               # returns None
```

### If Statement

```aura
if condition:
    body
elif condition:
    body
else:
    body
```

### Match Statement

```aura
match expr:
    case pattern [if guard]:
        body
    case pattern:
        body
```

### For Loop

```aura
for variable in iterable:
    body
```

### While Loop

```aura
while condition:
    body
```

### Assert

```aura
assert condition
assert condition, "error message"
```

---

## Declarations

### Type Alias

```aura
type Name [type_params] = type_expr
```

### Struct

```aura
[pub] struct Name [type_params]:
    [pub] field: Type [= default]
```

### Enum

```aura
[pub] enum Name [type_params]:
    Variant
    Variant(Type, Type)
```

### Function

```aura
[pub] fn name[type_params](params) [-> ReturnType] [with effects] [satisfies SpecName]:
    body
```

### Trait

```aura
[pub] trait Name [type_params]:
    fn method(self, params) -> ReturnType     # required method
    fn method(self, params) -> ReturnType:    # default implementation
        body
```

### Implementation

```aura
impl TraitName for Type:
    fn method(self, params) -> ReturnType:
        body

impl Type:                                    # inherent impl
    fn method(self, params) -> ReturnType:
        body
```

### Spec Block

```aura
spec Name:
    doc: "Description"
    inputs:
        param: Type - "description"
    guarantees:
        - "postcondition"
    effects: capability1, capability2
    errors:
        ErrorVariant(Type) - "when this happens"
```

### Test Block

```aura
test "description":
    body
```

### Module-Level Constant

```aura
[pub] let NAME: Type = expr
```

---

## Effect System

### Core Capabilities

| Capability | Description | Example Operations |
|------------|-------------|--------------------|
| `db` | Database access | query, insert, update, delete |
| `net` | Network I/O | HTTP requests, sockets |
| `fs` | File system access | read, write, list |
| `time` | Clock access | current time, sleep |
| `random` | Non-determinism | random numbers, UUIDs |
| `auth` | Authentication | check permissions, get user |
| `log` | Logging | log messages, metrics |

### Rules

1. **Pure by default:** Functions without `with` are pure
2. **Explicit on public:** `pub fn` must declare all effects
3. **Inferred on private:** Private `fn` can have effects inferred
4. **Transitive:** If `f` calls `g`, then `effects(f) ⊇ effects(g)`
5. **Exact for specs:** `satisfies` requires effects to exactly match the spec

### Syntax

```aura
pub fn name(params) -> ReturnType with effect1, effect2:
    body
```

---

## Specifications

### Structure

```
spec Name:
    doc: "..."              # optional: human-readable description
    inputs:                  # optional: typed inputs with descriptions
        name: Type - "desc"
    guarantees:              # optional: postconditions
        - "condition"
    effects: cap1, cap2      # optional: allowed effects
    errors:                  # optional: error conditions
        Variant(T) - "desc"
```

### Compiler Validation

When a function declares `satisfies SpecName`:

| Check | Rule |
|-------|------|
| Inputs match | Function params must match spec input names and types |
| Effects match | Function effects must exactly equal spec effects |
| Errors covered | Function's error type must include all spec error variants |
| Uniqueness | One function per spec, one spec per function |

---

## Operator Precedence

Highest to lowest:

| Level | Operators | Associativity |
|-------|-----------|---------------|
| 1 | `.` `()` `[]` `?` `!` | Left |
| 2 | `-` (unary) `not` | Right |
| 3 | `**` | Right |
| 4 | `*` `/` `%` | Left |
| 5 | `+` `-` | Left |
| 6 | `==` `!=` `<` `>` `<=` `>=` `is` `in` | Left |
| 7 | `and` | Left |
| 8 | `or` | Left |
| 9 | `\|>` | Left |

---

## Reserved Keywords

```
and       as        assert    break     case      continue
db        do        doc       elif      else      enum
effects   errors    false     fn        for       from
guarantees if       impl      import    in        inputs
is        let       log       match     module    mut
net       none      not       or        pub       random
return    satisfies spec      struct    test      time
trait     true      type      while     with
```

---

*Aura Language Reference — v0.1*
