# 🤖 Aura: AI-First Language Mission

> **Aura is an AI-first programming language designed from the ground up for seamless AI-human collaboration.**
>
> Every design decision in Aura optimizes for AI parseability, code generation, and "vibe coding" — the fluid, intuitive flow between human intent and AI implementation.

---

## Core Mission

**Aura exists to make AI the best developer it can be.**

Traditional programming languages were designed for humans typing code into editors. Aura is designed for a different era — one where AI agents generate, modify, and reason about code alongside human collaborators. The language's syntax, type system, specification blocks, and effect tracking are all engineered to minimize ambiguity and maximize the signal available to AI systems.

### The Problem with Existing Languages

Most languages suffer from **implicit context** — the meaning of code depends on conventions, runtime state, or tribal knowledge that AI systems struggle to access:

- **Implicit side effects** — A function might write to a database, call an API, or mutate global state, and nothing in the signature reveals this.
- **Ambiguous intent** — Code expresses *what* but rarely *why*. AI must guess at requirements.
- **Loose types** — Dynamic types and type inference hide the contract between components.
- **Convention-dependent** — Meaning is buried in naming conventions, comments, and project-specific patterns.

### How Aura Solves This

Aura makes everything **explicit, structured, and machine-readable**:

| Feature | AI Benefit |
|---------|-----------|
| **Specs** | AI reads the spec to understand *what* to build before writing *how* |
| **Effect annotations** | AI knows exactly what capabilities a function needs (`with db, time`) |
| **Refinement types** | AI understands data constraints without reading comments (`String where len >= 1`) |
| **Explicit types** | Every boundary has a clear contract — no guessing |
| **`satisfies` clauses** | AI can verify its own output against the spec automatically |
| **Structured error types** | AI knows every failure mode and can generate exhaustive handling |

---

## Design Principles

### 1. Explicitness Over Brevity

Every piece of information an AI needs to generate correct code should be **in the source**, not inferred from context.

```aura
# BAD (in other languages): What does this do? What can fail? What state does it touch?
def create_task(title, priority=3):
    ...

# GOOD (Aura): Complete contract visible to AI
spec CreateNewTask:
    doc: "Create a new task with validation"
    inputs:
        title: String where len >= 1 and len <= 200 — "Task title"
        priority: Int where self >= 1 and self <= 5 — "Priority level"
    guarantees:
        - "Returns a valid Task with generated ID"
        - "Created timestamp is set to current time"
    effects: [db, time]
    errors:
        ValidationError — "Title or priority out of bounds"

fn create_task(title: String, priority: Int = 3) -> Result[Task, TaskError] with db, time satisfies CreateNewTask:
    ...
```

### 2. Specs as the Source of Truth

Specifications are not documentation — they are **machine-checkable contracts**. When an AI reads a spec block, it has everything needed to generate a correct implementation:

- **Inputs** with types and constraints
- **Guarantees** the output must satisfy
- **Effects** the implementation is allowed to use
- **Errors** that must be handled

This eliminates the "read the whole codebase to understand what this function should do" problem.

### 3. Effects as Capabilities

The effect system (`with db, time, net`) serves as an **AI safety mechanism**:

- AI-generated code cannot accidentally introduce side effects
- The compiler enforces that declared effects match actual effects
- Test code can mock effects without modifying implementation
- AI can reason about function purity and composition

### 4. Types as Documentation

Aura's type system is designed to carry **maximum information density**:

```aura
type TaskId = String where len >= 1 and len <= 64
type Priority = Int where self >= 1 and self <= 5
type Status = "pending" | "in_progress" | "done" | "cancelled"
```

An AI reading these types knows *exactly* what values are valid — no need to search for validation logic or read comments.

### 5. Pattern Matching for Exhaustive Handling

When AI generates code that handles enums or union types, pattern matching with exhaustiveness checking ensures **every case is covered**:

```aura
match result:
    case Ok(task):
        log.info("Created: {task.id}")
    case Err(TaskError.NotFound(id)):
        log.warn("Task {id} not found")
    case Err(TaskError.ValidationError(msg)):
        log.error("Invalid: {msg}")
    # Compiler ensures all cases are handled — AI can't miss one
```

---

## The "Vibe Coding" Philosophy

**Vibe coding** is the seamless collaboration between human intent and AI implementation:

1. **Human writes the spec** — Describes *what* they want in structured, natural-ish language
2. **AI generates the implementation** — Using the spec as a complete contract
3. **Compiler validates** — Types, effects, and spec satisfaction are checked automatically
4. **Human reviews and iterates** — The spec makes the intent clear, so review is fast

This flow means:
- Humans focus on **what** (specs, types, requirements)
- AI focuses on **how** (implementation, patterns, optimization)
- The compiler ensures **correctness** (types check, effects match, specs satisfied)

### Why This Matters

In vibe coding, the bottleneck isn't writing code — it's **communicating intent**. Aura's spec system is purpose-built for this: it's structured enough for AI to parse unambiguously, but readable enough for humans to write and review naturally.

---

## Instructions for AI Contributors

If you are an AI agent working on the Aura codebase or generating Aura code, follow these guidelines:

### When Generating Aura Code

1. **Start with the spec** — Always read or write the spec block first. It defines the contract.
2. **Respect effects** — Only use capabilities declared in the function signature.
3. **Use refinement types** — Encode constraints in the type system, not in runtime checks.
4. **Handle all error cases** — Use the spec's `errors` section as a checklist.
5. **Use `satisfies`** — Link every implementation to its spec for automatic validation.

### When Contributing to the Aura Toolchain

1. **Optimize for AI parseability** — Every new feature should ask: "Can an AI read this and know exactly what to do?"
2. **Structured over freeform** — Prefer structured syntax (like spec blocks) over comments or conventions.
3. **Explicit over implicit** — If information exists, it should be in the syntax, not inferred.
4. **Machine-checkable** — Every contract should be verifiable by the compiler, not just by human review.

### Design Decision Framework

When faced with a design choice, apply this priority order:

1. **AI flow** — Does this make AI code generation faster and more accurate?
2. **Compiler verifiability** — Can the compiler check this automatically?
3. **Human readability** — Is this clear for human review?
4. **Brevity** — Is this concise? (Lowest priority — clarity always wins)

---

## How Aura Reduces Ambiguity for AI

### Structured Specs vs. Natural Language Comments

```
# Other languages: AI must parse natural language (error-prone)
# Creates a new task. Title must be non-empty and under 200 chars.
# Priority defaults to 3. Can fail with validation error.
# Writes to database and uses system time.

# Aura: AI reads structured data (unambiguous)
spec CreateNewTask:
    doc: "Create a new task with validation"
    inputs:
        title: String where len >= 1 and len <= 200 — "Task title"
        priority: Int where self >= 1 and self <= 5 — "Priority level"
    guarantees:
        - "Returns a valid Task with generated ID"
    effects: [db, time]
    errors:
        ValidationError — "Title or priority out of bounds"
```

### Explicit Effects vs. Hidden Side Effects

```
# Other languages: AI has no idea what this function touches
def save_task(task):
    db.insert(task)          # Hidden DB write
    notify(task.assignee)    # Hidden network call
    audit_log(task)          # Hidden file write

# Aura: AI knows exactly what capabilities are used
fn save_task(task: Task) -> Result[Unit, SaveError] with db, net, log:
    db.insert(task)?
    notify(task.assignee)?
    audit_log(task)?
```

### Refinement Types vs. Runtime Validation

```
# Other languages: AI must find and read validation code
class Task:
    title: str  # Must be 1-200 chars... somewhere there's a validator

# Aura: Constraints are IN the type
pub struct Task:
    pub title: String where len >= 1 and len <= 200
```

---

## Summary

Aura is not just another programming language. It's a **communication protocol between humans and AI**, designed so that:

- **Specs** replace ambiguous requirements with structured contracts
- **Effects** replace hidden side effects with explicit capabilities
- **Refinement types** replace scattered validation with inline constraints
- **`satisfies`** replaces manual verification with compiler-checked links

**When you work on Aura, you're building the future of AI-human programming.**

---

*See also: [README.md](README.md) · [ROADMAP.md](ROADMAP.md) · [DEVELOPMENT.md](DEVELOPMENT.md) · [Language Guide](user_docs/language_guide.md)*
