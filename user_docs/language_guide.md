# Aura Language Guide

A tutorial-style guide covering all of Aura's language features with examples.

---

## Aura: Designed for AI-Human Pair Programming

Aura is an **AI-first programming language** — built from the ground up for seamless collaboration between human developers and AI agents.

As you learn Aura, you'll notice features that make it uniquely suited for AI-assisted development:

- **Specifications (specs)** let you describe *what* you want in a structured format that AI can parse unambiguously. Instead of writing vague comments, you write machine-readable contracts — and AI generates the implementation.
- **Explicit effects** (`with db, time, net`) tell AI exactly what side effects a function is allowed to have. AI-generated code can't accidentally introduce database writes or network calls that weren't part of the plan.
- **Refinement types** (`String where len >= 1`) encode data constraints directly in the type system. AI doesn't need to hunt for validation logic — the constraints are right there in the type definition.
- **`satisfies` clauses** link implementations to their specs, so the compiler can automatically verify that AI-generated code meets requirements.

This means the typical workflow in Aura is:

1. **You write the spec** — structured intent, clear contracts
2. **AI generates the implementation** — using the spec as a complete blueprint
3. **The compiler validates** — types, effects, specs all checked automatically
4. **You review** — the spec makes your intent clear, so review is fast

For the full AI-first mission and design philosophy, see [AI_MISSION.md](../AI_MISSION.md).

---

## Table of Contents

1. [Modules & Imports](#modules--imports)
2. [Types & Type Aliases](#types--type-aliases)
3. [Structs](#structs)
4. [Enums](#enums)
5. [Functions](#functions)
6. [Variables & Constants](#variables--constants)
7. [Control Flow](#control-flow)
8. [Pattern Matching](#pattern-matching)
9. [Expressions](#expressions)
10. [Traits & Implementations](#traits--implementations)
11. [Specs (Specifications)](#specs-specifications)
12. [Effect System](#effect-system)
13. [Tests](#tests)
14. [Comments](#comments)

---

## Modules & Imports

Every Aura file starts with a module declaration:

```aura
module myapp.models
```

Import other modules:

```aura
import std.time as time
from std.collections import List, Map
```

The module name is a dotted path that represents the logical namespace of the file.

---

## Types & Type Aliases

Create type aliases to give meaningful names to types:

```aura
type TaskId = String
type Priority = Int
```

### Refinement Types

Attach constraints directly to types using `where`:

```aura
type TaskId = String where len >= 1 and len <= 64
type Priority = Int where self >= 1 and self <= 5
type Email = String where self matches r"^[^@]+@[^@]+\.[^@]+$"
```

Refinement predicates are checked at compile time when possible, and at runtime otherwise.

### Union Types (String Literals)

Define types that are one of several literal values:

```aura
type TaskStatus = "pending" | "in_progress" | "done" | "cancelled"
```

---

## Structs

Structs are named record types with typed fields:

```aura
pub struct Task:
    pub id: TaskId
    pub title: String where len >= 1 and len <= 200
    pub description: String = ""
    pub status: TaskStatus = "pending"
    pub priority: Priority = 3
    pub created_at: time.Instant
    pub completed_at: time.Instant? = none
    pub tags: [String] = []
```

**Key features:**
- Fields can have **default values** (e.g., `= "pending"`)
- Fields can be **optional** using `?` (e.g., `time.Instant?`)
- Fields can have **visibility** (`pub` makes them accessible outside the module)
- Fields can have **refinement constraints** inline

### Constructing Structs

Use named fields:

```aura
let task = Task(
    id: "t-001",
    title: "Write documentation",
    created_at: time.now(),
)
```

Fields with defaults can be omitted.

---

## Enums

Enums define types with a fixed set of variants, optionally carrying data:

```aura
# Simple enum (no data)
pub enum Color:
    Red
    Green
    Blue

# Enum with data variants
pub enum TaskError:
    NotFound(TaskId)
    InvalidTitle(String)
    AlreadyCompleted(TaskId)
    Unauthorized(String)

# Enum with multiple fields
pub enum TaskEvent:
    Created(Task)
    StatusChanged(TaskId, TaskStatus, TaskStatus)
    Deleted(TaskId)
```

Access variants with dot notation: `TaskError.NotFound("t-001")`

---

## Functions

Functions are declared with `fn`:

```aura
pub fn add(a: Int, b: Int) -> Int:
    return a + b
```

### Default Parameters

```aura
pub fn create_task(title: String, priority: Priority = 3) -> Task:
    return Task(id: generate_id(), title: title, priority: priority, created_at: time.now())
```

### Effect Annotations

Declare side effects with `with`:

```aura
pub fn save(task: Task) -> Result[Task, Error] with db, log:
    log.info("Saving task: {task.id}")
    db.insert("tasks", task)
    return Ok(task)
```

### Satisfies Clause

Bind a function to a specification:

```aura
pub fn create_task(title: String, priority: Priority = 3) -> Result[Task, TaskError] with db, time satisfies CreateNewTask:
    # implementation...
```

### Return Types

Return types come after `->`. Use `Result[T, E]` for fallible functions:

```aura
pub fn find(id: TaskId) -> Result[Task, TaskError] with db:
    let row = db.query_one("tasks", id)?
    return Ok(Task.from_row(row))
```

The `?` operator propagates errors automatically.

---

## Variables & Constants

### Let Bindings

```aura
let x = 42                    # immutable, type inferred
let name: String = "Aura"     # immutable, type annotated
let mut counter = 0           # mutable
counter = counter + 1         # reassignment OK for mut
```

### Module-Level Constants

```aura
pub let MAX_TITLE_LENGTH: Int = 200
pub let DEFAULT_PRIORITY: Priority = 3
```

---

## Control Flow

### If / Elif / Else

```aura
if priority >= 4:
    log.warn("High priority task")
elif priority >= 2:
    log.info("Medium priority task")
else:
    log.debug("Low priority task")
```

### For Loops

```aura
for task in tasks:
    if task.status == "pending":
        process(task)
```

### While Loops

```aura
let mut count = 0
while count < 10:
    count = count + 1
```

### Break and Continue

```aura
for item in items:
    if item.is_skip():
        continue
    if item.is_done():
        break
    process(item)
```

---

## Pattern Matching

The `match` statement provides powerful pattern matching:

```aura
match result:
    case Ok(task):
        log.info("Got task: {task.title}")
    case Err(TaskError.NotFound(id)):
        log.error("Task not found: {id}")
    case Err(TaskError.InvalidTitle(msg)):
        log.error("Invalid title: {msg}")
    case Err(e):
        log.error("Unexpected error: {e}")
```

### Pattern Types

| Pattern | Example | Matches |
|---------|---------|----------|
| Wildcard | `_` | Anything |
| Binding | `x` | Anything, binds to `x` |
| Literal | `42`, `"hello"`, `true` | Exact value |
| Constructor | `Ok(value)` | Enum variant with data |
| List | `[first, second]` | List with exactly 2 elements |
| Tuple | `(a, b)` | Tuple with 2 elements |

### Guards

Add conditions to match cases:

```aura
match task.priority:
    case p if p >= 4:
        handle_urgent(task)
    case p if p >= 2:
        handle_normal(task)
    case _:
        handle_low(task)
```

---

## Expressions

### Pipeline Operator

Chain operations left-to-right:

```aura
let result = data
    |> validate
    |> transform
    |> save
```

### List Comprehensions

Create lists from iterables with optional filtering:

```aura
let high_priority = [t for t in tasks if t.priority >= 4]
let titles = [t.title for t in tasks]
let doubled = [x * 2 for x in numbers if x > 0]
```

### Lambda Expressions

Anonymous functions:

```aura
let sorted = tasks.sort_by(|a, b| -> a.priority > b.priority)
let names = tasks.map(|t| -> t.title)
```

### If Expressions

Inline conditional values:

```aura
let label = if priority >= 4 then "urgent" else "normal"
```

### String Interpolation

Embed expressions in strings with `{}`:

```aura
let msg = "Task {task.title} has priority {task.priority}"
let greeting = "Hello, {name}! You have {count} tasks."
```

### Option Chaining

Safely navigate through optional values:

```aura
let name = user?.profile?.display_name
```

### Unwrap

Force-unwrap an Option (panics on None):

```aura
let task = maybe_task!
```

---

## Traits & Implementations

### Defining Traits

Traits define shared behavior:

```aura
pub trait Validate:
    fn validate(self) -> Result[self, [String]]
```

### Implementing Traits

```aura
impl Validate for Task:
    fn validate(self) -> Result[Task, [String]]:
        let errors: [String] = []
        if self.title.len == 0:
            errors.push("Title must not be empty")
        if self.priority < 1 or self.priority > 5:
            errors.push("Priority must be between 1 and 5")
        if errors.len > 0:
            return Err(errors)
        return Ok(self)
```

### Inherent Implementations

Add methods directly to a type:

```aura
impl Task:
    fn is_complete(self) -> Bool:
        return self.status == "done"
```

---

## Specs (Specifications)

Specs capture the **intent** of a function before implementation.

```aura
spec CreateNewTask:
    doc: "Creates a new task with the given title and optional priority."

    inputs:
        title: String where len >= 1 and len <= 200 - "The task title"
        priority: Priority = 3 - "Task priority from 1 (low) to 5 (critical)"

    guarantees:
        - "Returns a Task with status 'pending'"
        - "The returned Task has a unique, non-empty id"
        - "The task is persisted in the database"

    effects: db, time

    errors:
        InvalidTitle(String) - "When title is empty or exceeds 200 characters"
```

### Spec Sections

| Section | Purpose |
|---------|----------|
| `doc` | Human-readable description |
| `inputs` | Typed parameters with descriptions |
| `guarantees` | Postconditions the implementation must satisfy |
| `effects` | Side effects the implementation is allowed to use |
| `errors` | Error conditions and their types |

The compiler validates that functions declaring `satisfies` match their spec.

---

## Effect System

Aura tracks side effects as **capabilities**:

| Capability | Description |
|------------|-------------|
| `db` | Database access |
| `net` | Network I/O |
| `fs` | File system access |
| `time` | Clock / time access |
| `random` | Non-determinism |
| `auth` | Authentication context |
| `log` | Logging / observability |

### Rules

1. **Pure functions** have no `with` clause and cannot call effectful functions
2. **Public functions** must declare effects explicitly
3. **Private functions** can have effects inferred
4. **Callers** must declare all effects of their callees

```aura
# Pure — cannot call anything with effects
fn add(a: Int, b: Int) -> Int:
    return a + b

# Has effects — must declare them
pub fn create_task(title: String) -> Result[Task, Error] with db, time:
    let now = time.now()       # OK: 'time' is declared
    db.insert("tasks", task)   # OK: 'db' is declared
    return Ok(task)
```

---

## Tests

Test blocks are built into the language:

```aura
test "creating a task with valid title succeeds":
    let result = create_task("Buy groceries")
    match result:
        case Ok(task):
            assert task.title == "Buy groceries"
            assert task.status == "pending"
        case Err(e):
            assert false, "Expected Ok but got Err"
```

Tests can mock effects using `with` blocks:

```aura
test "create_task inserts into database":
    with mock_db(), mock_time(fixed: "2026-01-01T00:00:00Z"):
        let result = create_task("Test task")
        assert result.is_ok()
```

---

## Comments

```aura
# Single-line comment

## Doc comment — attached to the next declaration
## Can span multiple lines
pub struct Task:
    pub id: String
```

Doc comments (`##`) are preserved in the AST and used for documentation generation.

---

## Next Steps

- 📋 **[Language Reference](language_reference.md)** — Formal syntax and type system details
- 💡 **[Examples](examples.md)** — Complete working examples
- 🚀 **[Getting Started](getting_started.md)** — Installation and setup
