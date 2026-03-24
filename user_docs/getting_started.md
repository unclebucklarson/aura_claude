# Getting Started with Aura

Welcome to Aura — a Python-inspired, statically typed programming language with specification-driven development, algebraic types, and effect tracking.

This guide will help you set up the Aura toolchain and write your first program.

---

## Installation

### Prerequisites

- [Go 1.22+](https://go.dev/dl/) (to build from source)
- Git

### Build from Source

```bash
# Clone the repository
git clone https://github.com/unclebucklarson/aura.git
cd aura

# Build the CLI
go build -o aura ./cmd/aura

# (Optional) Move to your PATH
sudo mv aura /usr/local/bin/
```

### Verify Installation

```bash
./aura format testdata/simple.aura
```

You should see formatted Aura source code printed to stdout.

---

## Your First Aura Program

Create a file called `hello.aura`:

```aura
module hello

type Name = String

pub struct Greeting:
    pub message: String
    pub recipient: Name

pub fn greet(name: Name) -> Greeting:
    let msg = "Hello, {name}!"
    return Greeting(message: msg, recipient: name)
```

Now format it:

```bash
./aura format hello.aura
```

Or parse it to see the token stream and AST:

```bash
./aura parse hello.aura
```

---

## Key Concepts

### Indentation-Based Blocks

Aura uses **4-space indentation** to define blocks, similar to Python. Tabs are not allowed.

```aura
pub fn add(a: Int, b: Int) -> Int:
    return a + b
```

### Static Typing

All values have types, and the compiler checks types at compile time:

```aura
let x: Int = 42
let name: String = "Aura"
let active: Bool = true
```

Type annotations can often be inferred:

```aura
let x = 42           # inferred as Int
let name = "Aura"    # inferred as String
```

### Specs Before Code

Aura encourages writing **specifications** before implementations. A spec describes what a function should do:

```aura
spec AddNumbers:
    doc: "Adds two integers and returns their sum."

    inputs:
        a: Int - "First number"
        b: Int - "Second number"

    guarantees:
        - "Returns the arithmetic sum of a and b"
```

Then implement with `satisfies`:

```aura
pub fn add(a: Int, b: Int) -> Int satisfies AddNumbers:
    return a + b
```

### Effects Are Explicit

Functions that perform side effects must declare them:

```aura
# Pure function — no side effects
pub fn double(x: Int) -> Int:
    return x * 2

# Effectful function — declares what it needs
pub fn save_record(data: String) -> Result[Bool, Error] with db, log:
    log.info("Saving: {data}")
    db.insert("records", data)
    return Ok(true)
```

---

## Available CLI Commands

| Command | Description |
|---------|-------------|
| `aura format <file>` | Format an Aura file and print to stdout |
| `aura format -w <file>` | Format an Aura file in-place |
| `aura parse <file>` | Parse a file and dump tokens + AST |

---

## Next Steps

- 📖 **[Language Guide](language_guide.md)** — Tutorial covering all language features
- 📋 **[Language Reference](language_reference.md)** — Formal syntax and type system reference
- 💡 **[Examples](examples.md)** — Complete working examples
- 🗺️ **[Roadmap](../ROADMAP.md)** — What's coming next
