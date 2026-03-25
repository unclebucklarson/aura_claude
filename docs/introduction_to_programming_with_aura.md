# Introduction to Programming with Aura

> **Your first steps into the Aura programming language.**
> This tutorial will take you from zero to writing real programs in Aura — no prior Aura experience required!

**Aura version:** v2.0.1
**Difficulty:** Beginner
**Time to complete:** ~60 minutes

---

## Table of Contents

1. [What is Aura?](#what-is-aura)
2. [Getting Started](#getting-started)
3. [Tutorial 1: Hello, World!](#tutorial-1-hello-world)
4. [Tutorial 2: Calculator Program](#tutorial-2-calculator-program)
5. [Tutorial 3: Number Guessing Game](#tutorial-3-number-guessing-game)
6. [Tutorial 4: Structs and Custom Types](#tutorial-4-structs-and-custom-types)
7. [Aura Syntax Quick Reference](#aura-syntax-quick-reference)
8. [Next Steps](#next-steps)

---

## What is Aura?

Aura is a **modern, Python-inspired programming language** designed for clarity, safety, and AI-human collaboration. If you've ever written Python, Aura will feel familiar — but with powerful extras like static types, pattern matching, and an effect system that makes your code safer and easier to reason about.

### Why Learn Aura?

- **Clean syntax** — Indentation-based blocks, just like Python. No curly braces or semicolons.
- **Safe by default** — `Option` and `Result` types prevent null pointer errors and unhandled exceptions.
- **Pattern matching** — Elegant `match` expressions replace messy chains of `if`/`elif`.
- **Batteries included** — 17 standard library modules with 120 functions ready to use.
- **AI-first design** — Built for seamless collaboration between humans and AI agents.

### What You'll Learn

In these four tutorials, you'll progressively learn:

| Tutorial | What You'll Build | Key Concepts |
|---|---|---|
| **Hello World** | A greeting program | Functions, strings, `std.io`, program structure |
| **Calculator** | A math calculator | Pattern matching, arithmetic, functions with parameters |
| **Guessing Game** | A number guessing game | Loops, conditionals, random numbers, mutable variables |
| **Grade Tracker** | A student grade system | Structs, custom types, Option, safe data handling |

Let's dive in! 🚀

---

## Getting Started

### Installing Aura

Aura is built in Go. To get started:

```bash
# 1. Make sure you have Go 1.22+ installed
go version

# 2. Clone the Aura repository
git clone https://github.com/unclebucklarson/aura.git
cd aura

# 3. Build the Aura CLI
go build -o aura ./cmd/aura

# 4. (Optional) Move it to your PATH so you can use it from anywhere
sudo mv aura /usr/local/bin/
```

### Verify It Works

```bash
./aura repl
```

You should see:

```
Aura REPL v1.1.0
Type expressions or statements. Press Ctrl+D to exit.
```

Type `42 + 8` and press Enter — you should see `50`. Press `Ctrl+D` to exit.

### Creating Your First Aura File

Aura source files use the `.aura` extension. You can create them with any text editor:

```bash
touch my_program.aura
```

### Running Aura Programs

To run an Aura program, use the `aura run` command:

```bash
aura run my_program.aura
```

This will:
1. Parse your `.aura` file
2. Type-check it
3. Execute the `main()` function

### The Aura Toolchain

Aura comes with several useful commands:

| Command | What It Does |
|---|---|
| `aura run file.aura` | Execute a program |
| `aura test file.aura` | Run test blocks |
| `aura format file.aura` | Pretty-print your code |
| `aura check file.aura` | Type-check without running |
| `aura build file.aura` | Compile to a native binary |
| `aura repl` | Interactive playground |

Now you're ready to write your first program!

---

## Tutorial 1: Hello, World!

Every programming journey starts with "Hello, World!" — a simple program that prints a message to the screen. In Aura, it's delightfully simple.

### The Complete Program

Create a file called `hello_world.aura`:

```aura
module hello_world

import std.io

fn main():
    std.io.println("Hello, World!")
    std.io.println("Welcome to Aura!")
```

### Running It

```bash
aura run hello_world.aura
```

### Expected Output

```
Hello, World!
Welcome to Aura!
```

Congratulations! You just ran your first Aura program!

### Step-by-Step Explanation

#### Line 1: `module hello_world`

```aura
module hello_world
```

Every Aura file **must** start with a `module` declaration. This gives your file a name — think of it as a label for your code. The module name is typically the same as the filename (without `.aura`).

> 💡 **Key concept:** Modules are how Aura organizes code. Each file is a module.

#### Line 3: `import std.io`

```aura
import std.io
```

This imports the **standard I/O module** — `std.io` — which gives us functions for printing to the screen and reading user input. In Aura, you need to explicitly import the tools you want to use.

Some commonly used standard library modules:
- `std.io` — Input/output (printing and reading input)
- `std.math` — Math functions (sqrt, abs, pow, etc.)
- `std.string` — String utilities
- `std.random` — Random number generation

#### Lines 5–7: The `main()` Function

```aura
fn main():
    std.io.println("Hello, World!")
    std.io.println("Welcome to Aura!")
```

- `fn` — The keyword that declares a function.
- `main` — The function name. When you use `aura run`, Aura automatically looks for and calls a function named `main()`.
- `():` — Empty parentheses mean this function takes no parameters. The colon `:` starts the function body.
- The indented lines are the **function body**. Aura uses **4-space indentation** to define blocks (just like Python).

#### `std.io.println("Hello, World!")`

- `std.io.println` — Calls the `println` function from the `std.io` module. It prints text followed by a newline.
- `"Hello, World!"` — A **string literal** enclosed in double quotes. Strings in Aura always use double quotes `"`.

### Using Import Aliases

When you use a module frequently, you can give it a shorter alias:

```aura
module hello_world

import std.io as io

fn main():
    io.println("Hello, World!")
    io.println("Welcome to Aura!")
```

The `as io` part lets you write `io.println` instead of `std.io.println`. Both forms work — use whichever feels clearer.

### Key Concepts Learned

| Concept | Example | Description |
|---|---|---|
| Module declaration | `module hello_world` | Every file needs one |
| Imports | `import std.io` | Bring in standard library modules |
| Import aliases | `import std.io as io` | Shorter names for modules |
| Functions | `fn main():` | `fn` keyword, name, params, colon |
| Indentation | 4 spaces | Defines code blocks |
| Strings | `"Hello, World!"` | Double-quoted text |
| Printing | `std.io.println(...)` | Output text to screen |

### Exercises

Try these modifications to deepen your understanding:

1. **Change the message:** Print your own name instead of "World".
2. **Add more lines:** Print three lines — your name, your favorite color, and your favorite number.
3. **String interpolation:** Try using Aura's string interpolation:
   ```aura
   let name = "Aura"
   std.io.println("Hello, {name}!")
   ```
   This will print: `Hello, Aura!`
4. **Use an alias:** Rewrite the program using `import std.io as io` and `io.println`.

### Common Mistakes

| Mistake | What Happens | Fix |
|---|---|---|
| Missing `module` line | Parse error | Always start with `module name` |
| Using tabs instead of spaces | Parse error | Use exactly 4 spaces for indentation |
| Forgetting to import `std.io` | Runtime error: undefined | Add `import std.io` at the top |
| Using single quotes `'hello'` | Parse error | Always use double quotes `"hello"` |
| Wrong indentation (2 or 3 spaces) | Unexpected behavior | Always use exactly **4 spaces** |

---

## Tutorial 2: Calculator Program

Now let's build something more useful — a calculator! This program demonstrates functions with parameters, arithmetic operations, and Aura's powerful **pattern matching**.

### The Complete Program

Create a file called `calculator.aura`:

```aura
module calculator

import std.io as io

fn add(a, b):
    return a + b

fn subtract(a, b):
    return a - b

fn multiply(a, b):
    return a * b

fn divide(a, b):
    if b == 0:
        io.println("  Error: Cannot divide by zero!")
        return none
    return a / b

fn calculate(a, op, b):
    io.println("Calculating: {a} {op} {b}")
    let result = match op:
        "add"      -> add(a, b)
        "subtract" -> subtract(a, b)
        "multiply" -> multiply(a, b)
        "divide"   -> divide(a, b)
        _          -> none
    if result == none:
        if op != "add" and op != "subtract" and op != "multiply" and op != "divide":
            io.println("  Unknown operation: {op}")
    else:
        io.println("  Result: {result}")
    return result

fn main():
    io.println("=================================")
    io.println("   Aura Calculator")
    io.println("=================================")
    io.println("")

    # Basic arithmetic
    calculate(10, "add", 5)
    calculate(10, "subtract", 3)
    calculate(10, "multiply", 4)
    calculate(10, "divide", 3)

    io.println("")
    io.println("----- Edge Cases -----")

    # Edge cases
    calculate(10, "divide", 0)
    calculate(7, "modulo", 3)

    io.println("")
    io.println("----- Chained Calculations -----")

    # Chained calculations
    let step1 = add(100, 50)
    io.println("Step 1: 100 + 50 = {step1}")

    let step2 = multiply(step1, 2)
    io.println("Step 2: {step1} * 2 = {step2}")

    let step3 = subtract(step2, 25)
    io.println("Step 3: {step2} - 25 = {step3}")

    io.println("")
    io.println("Final answer: {step3}")
    io.println("=================================")
```

### Running It

```bash
aura run calculator.aura
```

### Expected Output

```
=================================
   Aura Calculator
=================================

Calculating: 10 add 5
  Result: 15
Calculating: 10 subtract 3
  Result: 7
Calculating: 10 multiply 4
  Result: 40
Calculating: 10 divide 3
  Result: 3

----- Edge Cases -----
Calculating: 10 divide 0
  Error: Cannot divide by zero!
Calculating: 7 modulo 3
  Unknown operation: modulo

----- Chained Calculations -----
Step 1: 100 + 50 = 150
Step 2: 150 * 2 = 300
Step 3: 300 - 25 = 275

Final answer: 275
=================================
```

### Step-by-Step Explanation

#### Arithmetic Functions

```aura
fn add(a, b):
    return a + b

fn subtract(a, b):
    return a - b
```

- Each function takes **two parameters** (`a` and `b`).
- The `return` keyword sends a value back to the caller.
- Aura supports all standard arithmetic operators: `+`, `-`, `*`, `/`, `%` (modulo).

> 💡 **Key concept:** Functions are defined with `fn`, take parameters in parentheses, and use `return` to send back results.

#### Safe Division with Error Handling

```aura
fn divide(a, b):
    if b == 0:
        io.println("  Error: Cannot divide by zero!")
        return none
    return a / b
```

- We check `if b == 0` to prevent division by zero.
- `none` is Aura's "no value" — similar to `null` in other languages, but safer because Aura tracks it through the type system.
- The `if` block uses **4-space indentation** for its body.

> 💡 **Key concept:** In Aura, `none` represents the absence of a value. It's part of the `Option` type system that prevents null pointer errors.

#### Pattern Matching with `match`

```aura
let result = match op:
    "add"      -> add(a, b)
    "subtract" -> subtract(a, b)
    "multiply" -> multiply(a, b)
    "divide"   -> divide(a, b)
    _          -> none
```

This is one of Aura's most powerful features — **pattern matching**! Here's how it works:

- `match op:` — Look at the value of `op` and compare it against patterns.
- `"add" -> add(a, b)` — If `op` equals `"add"`, call `add(a, b)` and use the result.
- `_ -> none` — The underscore `_` is a **wildcard** that matches anything. It's the catch-all case.
- The `match` expression returns a value, so we can assign it to `let result`.

> 💡 **Key concept:** `match` is an **expression** — it produces a value. This is cleaner than long `if`/`elif` chains.

#### String Interpolation

```aura
io.println("Calculating: {a} {op} {b}")
```

Aura supports **string interpolation** using curly braces `{}` inside double-quoted strings. Any expression inside `{}` is evaluated and converted to a string automatically.

#### Variables with `let`

```aura
let step1 = add(100, 50)
```

- `let` declares an **immutable variable** — once set, it cannot be changed.
- Aura infers the type automatically, so you don't need to write `let step1: Int = ...`.

### Key Concepts Learned

| Concept | Example | Description |
|---|---|---|
| Function parameters | `fn add(a, b):` | Functions can take inputs |
| Return values | `return a + b` | Send results back to the caller |
| `if` conditionals | `if b == 0:` | Execute code based on conditions |
| `none` value | `return none` | Represents "no value" |
| Pattern matching | `match op: "add" -> ...` | Elegant branching based on values |
| Wildcard pattern | `_ -> none` | Matches anything not matched above |
| String interpolation | `"Result: {result}"` | Embed expressions in strings |
| Immutable variables | `let x = 42` | Variables that can't change |
| Import aliases | `import std.io as io` | Shorter module names |

### Exercises

1. **Add modulo:** Add a `modulo(a, b)` function and handle `"modulo"` in the `match` expression.
2. **Add power:** Add a `power(a, b)` function. Hint: you can import `std.math as math` and use `math.pow(a, b)`.
3. **Floating point:** Try `calculate(10.0, "divide", 3.0)` — does it return `3` or `3.333...`?
4. **Chain more:** Create a sequence of 5 calculations where each step uses the previous result.

### Common Mistakes

| Mistake | What Happens | Fix |
|---|---|---|
| Forgetting `return` | Function returns `none` | Always `return` your result |
| Missing wildcard `_` in match | Runtime error if no pattern matches | Always add a `_` catch-all |
| Wrong comparison `=` vs `==` | Parse error | Use `==` for comparison, assignment uses `let` |
| Indentation inside match arms | Parse error | Match arms use `->` for single-line results |

---

## Tutorial 3: Number Guessing Game

Now for something fun — a number guessing game! This tutorial brings together everything you've learned and introduces **loops**, **mutable variables**, **random numbers**, and **comparisons**.

### The Complete Program

Create a file called `guessing_game.aura`:

```aura
module guessing_game

import std.io as io
import std.random

fn give_hint(guess, secret):
    if guess < secret:
        io.println("  Too low! Try a higher number.")
        return false
    elif guess > secret:
        io.println("  Too high! Try a lower number.")
        return false
    else:
        return true

fn play_round(secret, guesses):
    io.println("")
    io.println("--- Round with {guesses.len()} guesses ---")
    let mut attempts = 0
    let mut found = false
    for guess in guesses:
        attempts = attempts + 1
        io.println("")
        io.println("Guess #{attempts}: {guess}")
        if give_hint(guess, secret):
            found = true
            io.println("  CORRECT! You got it in {attempts} attempt(s)!")
            break
    if found == false:
        io.println("")
        io.println("  Out of guesses! The number was {secret}.")
    return attempts

fn play_binary_search(secret, min_val, max_val):
    io.println("")
    io.println("Now let's watch an AI play with binary search!")
    io.println("   Target: a number between {min_val} and {max_val}")
    io.println("")

    let mut low = min_val
    let mut high = max_val
    let mut attempts = 0
    let mut found = false

    while found == false:
        let guess = (low + high) / 2
        attempts = attempts + 1
        io.println("  Attempt {attempts}: AI guesses {guess}")

        if guess == secret:
            io.println("  AI found it in {attempts} attempts!")
            found = true
        elif guess < secret:
            io.println("    Too low! Searching higher...")
            low = guess + 1
        else:
            io.println("    Too high! Searching lower...")
            high = guess - 1

        if attempts > 20:
            io.println("  Safety limit reached!")
            found = true

    return attempts

fn rate_performance(attempts):
    io.println("")
    if attempts <= 3:
        io.println("Amazing! Only {attempts} attempts!")
    elif attempts <= 7:
        io.println("Good job! {attempts} attempts.")
    elif attempts <= 10:
        io.println("Not bad! {attempts} attempts.")
    else:
        io.println("Keep practicing! {attempts} attempts.")

fn main():
    io.println("=================================")
    io.println("   Aura Number Guessing Game")
    io.println("=================================")
    io.println("")

    # Generate a random secret number between 1 and 100
    let secret = std.random.int(1, 100)
    let min_val = 1
    let max_val = 100

    io.println("Secret number is between {min_val} and {max_val}")
    io.println("   (The secret is {secret})")

    # Simulate a player making guesses with a fixed list
    let guesses = [50, 25, 75, 37, 62, 42, 55, 48, 52, secret]
    play_round(secret, guesses)

    # Watch the binary search AI play
    let ai_attempts = play_binary_search(secret, min_val, max_val)
    rate_performance(ai_attempts)

    io.println("")
    io.println("Thanks for playing!")
```

### Running It

```bash
aura run guessing_game.aura
```

### Example Output

(The secret number is random, so your output will vary!)

```
=================================
   Aura Number Guessing Game
=================================

Secret number is between 1 and 100
   (The secret is 42)

--- Round with 10 guesses ---

Guess #1: 50
  Too high! Try a lower number.

Guess #2: 25
  Too low! Try a higher number.

Guess #3: 75
  Too high! Try a lower number.

Guess #4: 37
  Too low! Try a higher number.

Guess #5: 62
  Too high! Try a lower number.

Guess #6: 42
  CORRECT! You got it in 6 attempt(s)!

Now let's watch an AI play with binary search!
   Target: a number between 1 and 100

  Attempt 1: AI guesses 50
    Too high! Searching lower...
  Attempt 2: AI guesses 25
    Too low! Searching higher...
  Attempt 3: AI guesses 37
    Too low! Searching higher...
  Attempt 4: AI guesses 43
    Too high! Searching lower...
  Attempt 5: AI guesses 40
    Too low! Searching higher...
  Attempt 6: AI guesses 41
    Too low! Searching higher...
  Attempt 7: AI guesses 42
  AI found it in 7 attempts!

Good job! 7 attempts.

Thanks for playing!
```

### Step-by-Step Explanation

#### Importing Multiple Modules

```aura
import std.io as io
import std.random
```

You can import as many modules as you need, one per line. Here we use:
- `std.io` for printing to the screen (aliased as `io`)
- `std.random` for generating random numbers

#### Generating Random Numbers

```aura
let secret = std.random.int(1, 100)
```

`std.random.int(min, max)` returns a random integer between `min` and `max` (inclusive). Every time you run the program, you'll get a different secret number!

Other useful `std.random` functions:
- `std.random.float()` — Random decimal between 0.0 and 1.0
- `std.random.choice(list)` — Pick a random element from a list
- `std.random.shuffle(list)` — Shuffle a list randomly

#### Mutable Variables with `let mut`

```aura
let mut attempts = 0
let mut found = false
```

- `let` creates an **immutable** variable (cannot change).
- `let mut` creates a **mutable** variable (can be reassigned).
- Use `let` by default; only use `let mut` when you need the value to change.

```aura
attempts = attempts + 1    # OK — attempts is mutable
```

> 💡 **Key concept:** Aura encourages immutability. Using `let mut` signals "this value will change" — making your code's intent clearer.

#### Loops with `for` and `while`

**`for` loop** — Iterate over a collection:

```aura
for guess in guesses:
    io.println("Guess: {guess}")
```

This loops through each element in the `guesses` list. On each iteration, `guess` holds the current element.

**`while` loop** — Repeat while a condition is true:

```aura
while found == false:
    let guess = (low + high) / 2
    # ... check the guess ...
```

The `while` loop keeps running as long as `found == false`. Once `found` becomes `true`, the loop stops.

#### `break` — Exit a Loop Early

```aura
if give_hint(guess, secret):
    found = true
    io.println("  CORRECT!")
    break
```

`break` immediately exits the innermost loop. Use it when you've found what you're looking for.

#### `if` / `elif` / `else` Chains

```aura
if guess < secret:
    io.println("  Too low!")
    return false
elif guess > secret:
    io.println("  Too high!")
    return false
else:
    return true
```

- `if` — Check the first condition.
- `elif` — Check additional conditions (short for "else if"). You can chain as many `elif` as you need.
- `else` — Runs if no previous condition was true.

#### Lists

```aura
let guesses = [50, 25, 75, 37, 62, 42, 55, 48, 52, secret]
```

Lists in Aura use square brackets `[]` and can contain any values. You can even include variables like `secret` — its current value is captured.

Useful list methods:
- `list.len()` — Get the number of elements
- `list.first()` — Get the first element (returns `Option`)
- `list.last()` — Get the last element (returns `Option`)
- `list.contains(x)` — Check if `x` is in the list

#### The Binary Search Algorithm

The `play_binary_search` function demonstrates a classic algorithm:

1. Start with the full range (1–100).
2. Guess the middle value: `(low + high) / 2`
3. If too low, search the upper half. If too high, search the lower half.
4. Repeat until found.

Binary search finds any number in at most 7 attempts for a range of 1–100. That's the power of log₂(100) ≈ 7.

### Key Concepts Learned

| Concept | Example | Description |
|---|---|---|
| Random numbers | `std.random.int(1, 100)` | Generate random values |
| Mutable variables | `let mut count = 0` | Variables that can change |
| `for` loops | `for x in list:` | Iterate over collections |
| `while` loops | `while cond:` | Repeat while condition is true |
| `break` | `break` | Exit a loop early |
| `if`/`elif`/`else` | `if x < y:` ... `elif:` ... `else:` | Conditional branching |
| Lists | `[1, 2, 3]` | Ordered collections |
| Comparison operators | `<`, `>`, `==`, `!=`, `<=`, `>=` | Compare values |
| Boolean values | `true`, `false` | Logical values |

### Exercises

1. **Change the range:** Modify the game to use a range of 1–1000. How many attempts does binary search need now?
2. **Add difficulty levels:** Create functions for `easy_game()` (1–10), `medium_game()` (1–100), and `hard_game()` (1–1000).
3. **Random guesses:** Instead of the fixed `guesses` list, use `std.random.int()` to generate random guesses in a `while` loop.
4. **Interactive play:** Use `std.io.input(prompt)` to read guesses from the user in a terminal:
   ```aura
   let guess_str = std.io.input("Your guess: ")
   # Use std.json.parse() to convert the string to a number if needed
   ```
5. **Multiple rounds:** Use a `for` loop to play 5 rounds and track the total attempts across all games.

### Common Mistakes

| Mistake | What Happens | Fix |
|---|---|---|
| Forgetting `mut` | Cannot reassign: "variable is immutable" | Use `let mut` for variables you'll change |
| Infinite `while` loop | Program hangs forever | Always ensure the condition can become false; add a safety limit |
| Off-by-one in `std.random.int` | Wrong range | `std.random.int(1, 100)` includes both 1 and 100 |
| Using `=` instead of `==` | Wrong semantics | Use `==` for comparison |
| Forgetting `break` after finding answer | Loop continues unnecessarily | Add `break` when the goal is achieved |

---

## Tutorial 4: Structs and Custom Types

Now we'll explore one of Aura's most important features: **structs** and the **Option** type. These let you model real-world data precisely and handle the absence of a value safely.

### What We'll Build

A student grade tracker that:
- Defines a `Student` struct with custom fields
- Creates a list of students
- Finds the top student using `Option[Student]`
- Assigns letter grades using pattern matching with guards

### The Complete Program

Create a file called `grade_tracker.aura`:

```aura
module grade_tracker

import std.io as io

pub struct Student:
    pub name: String
    pub grade: Int

fn grade_letter(score: Int) -> String:
    return match score:
        _ if score >= 90 -> "A"
        _ if score >= 80 -> "B"
        _ if score >= 70 -> "C"
        _ if score >= 60 -> "D"
        _                -> "F"

fn describe_student(student: Student) -> String:
    let letter = grade_letter(student.grade)
    return "{student.name}: {student.grade}/100 — Grade {letter}"

fn find_top_student(students: [Student]) -> Option[Student]:
    if students.is_empty():
        return none
    let mut best = students[0]
    for s in students:
        if s.grade > best.grade:
            best = s
    return Some(best)

fn class_average(students: [Student]) -> Float:
    if students.is_empty():
        return 0.0
    let total = students.reduce(0, |acc, s| -> acc + s.grade)
    return total / students.len()

fn main():
    io.println("============================")
    io.println("   Student Grade Tracker")
    io.println("============================")
    io.println("")

    let students = [
        Student(name: "Alice", grade: 92),
        Student(name: "Bob", grade: 78),
        Student(name: "Carol", grade: 85),
        Student(name: "Dave", grade: 61),
        Student(name: "Eve", grade: 95),
    ]

    io.println("--- Individual Reports ---")
    for s in students:
        io.println(describe_student(s))

    io.println("")
    io.println("--- Class Statistics ---")

    let avg = class_average(students)
    io.println("Class average: {avg}")

    io.println("")
    io.println("--- Top Student ---")
    let top = find_top_student(students)
    match top:
        Some(s) -> io.println("Top student: {s.name} with {s.grade}/100")
        None    -> io.println("No students enrolled.")

    io.println("")
    io.println("--- Grade Distribution ---")
    for s in students:
        let bar = "*".repeat(s.grade / 10)
        io.println("{s.name.pad_right(8)} {bar}")

    io.println("")
    io.println("============================")
```

### Running It

```bash
aura run grade_tracker.aura
```

### Expected Output

```
============================
   Student Grade Tracker
============================

--- Individual Reports ---
Alice: 92/100 — Grade A
Bob: 78/100 — Grade C
Carol: 85/100 — Grade B
Dave: 61/100 — Grade D
Eve: 95/100 — Grade A

--- Class Statistics ---
Class average: 82

--- Top Student ---
Top student: Eve with 95/100

--- Grade Distribution ---
Alice    *********
Bob      *******
Carol    ********
Dave     ******
Eve      *********

============================
```

### Step-by-Step Explanation

#### Defining a Struct

```aura
pub struct Student:
    pub name: String
    pub grade: Int
```

A **struct** is a custom data type that groups related values together.

- `pub struct Student` — Declares a public struct named `Student`.
- `pub name: String` — A field named `name` that holds a `String`.
- `pub grade: Int` — A field named `grade` that holds an `Int`.
- `pub` means the field is visible outside the module.

> 💡 **Key concept:** Structs let you define your own types. Instead of passing `name` and `grade` as separate parameters everywhere, you bundle them into a `Student`.

#### Creating Struct Instances

```aura
let students = [
    Student(name: "Alice", grade: 92),
    Student(name: "Bob", grade: 78),
]
```

You create a `Student` by calling it like a function with **named arguments**. Named arguments make the code self-documenting — it's clear what each value means.

#### Accessing Fields

```aura
fn describe_student(student: Student) -> String:
    let letter = grade_letter(student.grade)
    return "{student.name}: {student.grade}/100 — Grade {letter}"
```

Use dot notation to access fields: `student.name`, `student.grade`. Fields can be embedded directly in string interpolation.

#### Type Annotations

```aura
fn grade_letter(score: Int) -> String:
fn describe_student(student: Student) -> String:
fn find_top_student(students: [Student]) -> Option[Student]:
```

Notice the explicit type annotations on these functions:
- `score: Int` — Parameter types
- `-> String` — Return type
- `[Student]` — A list of Student
- `Option[Student]` — Either a Student or nothing

Type annotations help both the compiler and your readers understand exactly what a function expects and returns. They're optional for local variables (Aura infers those), but good practice for function signatures.

#### Match with Guards

```aura
fn grade_letter(score: Int) -> String:
    return match score:
        _ if score >= 90 -> "A"
        _ if score >= 80 -> "B"
        _ if score >= 70 -> "C"
        _ if score >= 60 -> "D"
        _                -> "F"
```

Match arms can include **guards** — extra conditions written with `if`. The pattern `_ if score >= 90` means: match anything, but only if `score >= 90`. Guards are checked top to bottom; the first matching arm wins.

#### The Option Type

```aura
fn find_top_student(students: [Student]) -> Option[Student]:
    if students.is_empty():
        return none
    let mut best = students[0]
    for s in students:
        if s.grade > best.grade:
            best = s
    return Some(best)
```

`Option[Student]` represents "either a Student, or nothing." It forces callers to handle both cases explicitly.

- `return none` — Returns the "nothing" case (the student list was empty).
- `return Some(best)` — Wraps the found student in `Some`, signaling "we found one."

> 💡 **Key concept:** `Option` is how Aura handles values that might not exist. It's much safer than returning `none` without declaring it — the type system ensures callers can't forget to handle the empty case.

#### Matching on Option

```aura
let top = find_top_student(students)
match top:
    Some(s) -> io.println("Top student: {s.name} with {s.grade}/100")
    None    -> io.println("No students enrolled.")
```

When you have an `Option` value, use `match` to handle both cases:
- `Some(s) ->` — The value exists; `s` is bound to the Student inside.
- `None ->` — The value is absent; handle the empty case.

This pattern appears throughout Aura. Methods like `list.first()` and `list.last()` return `Option` too:

```aura
let first = students.first()
match first:
    Some(s) -> io.println("First student: {s.name}")
    None    -> io.println("List is empty!")
```

#### List Functional Methods

```aura
let total = students.reduce(0, |acc, s| acc + s.grade)
```

`reduce` applies a function to accumulate a result across all elements:
- `0` — The starting value (accumulator)
- `|acc, s| -> acc + s.grade` — A **lambda** (anonymous function): `|params| -> expr`
- `acc` is the running total; `s` is the current Student

The `|params| -> expr` syntax is Aura's lambda syntax — a compact way to write small functions inline.

#### Visual Output with String Methods

```aura
let bar = "*".repeat(s.grade / 10)
io.println("{s.name.pad_right(8)} {bar}")
```

- `"*".repeat(n)` — Creates a string of `n` asterisks
- `s.name.pad_right(8)` — Pads the name to 8 characters wide, right-padding with spaces
- `s.grade / 10` — Integer division (92 / 10 = 9 asterisks)

### Key Concepts Learned

| Concept | Example | Description |
|---|---|---|
| Struct definition | `pub struct Student:` | Define custom data types |
| Struct construction | `Student(name: "Alice", grade: 92)` | Create instances with named args |
| Field access | `student.name` | Access values with dot notation |
| Type annotations | `fn f(x: Int) -> String:` | Declare parameter and return types |
| Match with guards | `_ if score >= 90 -> "A"` | Conditional match arms |
| Option type | `Option[Student]` | Values that might not exist |
| Returning Some/None | `return Some(best)` / `return none` | Wrap or signal absence |
| Matching Option | `Some(s) -> ...` / `None -> ...` | Handle both cases |
| Lambdas | `\|acc, s\| -> acc + s.grade` | Inline anonymous functions |
| `reduce` | `list.reduce(0, \|acc, x\| -> ...)` | Accumulate a result |

### Exercises

1. **Add a field:** Add a `subject: String` field to `Student`. Update construction and printing.
2. **Find the bottom student:** Write `find_lowest_student` mirroring `find_top_student`.
3. **Grade statistics:** Count how many students got each letter grade (A, B, C, D, F).
4. **Filtering:** Use `students.filter(|s| -> s.grade >= 80)` to get only the passing students.
5. **Sorting:** Use `students.sort(|a, b| -> b.grade - a.grade)` to sort by grade descending, then print the ranked list.

### Common Mistakes

| Mistake | What Happens | Fix |
|---|---|---|
| Forgetting `pub` on fields | Fields not accessible outside module | Add `pub` to fields that need external access |
| Using positional args for structs | Parse error | Always use named arguments: `Student(name: "x", grade: 90)` |
| Not matching both `Some` and `None` | Runtime error if not exhaustive | Always handle both cases in `match` |
| Using `None` as a value | Parse error | Use lowercase `none` to return the absence value |
| Using `none` in a match pattern | Syntax error | Use uppercase `None` in match patterns |

---

## Aura Syntax Quick Reference

Here's a handy cheat sheet of everything covered in these tutorials:

### Variables

```aura
let name = "Aura"           # Immutable — cannot change
let mut counter = 0          # Mutable — can be reassigned
counter = counter + 1        # OK — counter is mutable
let x: Int = 42              # Explicit type annotation
```

### Functions

```aura
fn greet(name):
    std.io.println("Hello, {name}!")

fn add(a: Int, b: Int) -> Int:
    return a + b

fn safe_div(a: Int, b: Int) -> Option[Int]:
    if b == 0:
        return none
    return Some(a / b)
```

### Custom Data Types

```aura
pub struct Point:
    pub x: Float
    pub y: Float

let p = Point(x: 3.0, y: 4.0)
std.io.println("Point: ({p.x}, {p.y})")
```

### Control Flow

```aura
# If / elif / else
if x > 0:
    std.io.println("positive")
elif x == 0:
    std.io.println("zero")
else:
    std.io.println("negative")

# For loop
for item in [1, 2, 3]:
    std.io.println(item)

# While loop
let mut n = 0
while n < 5:
    n = n + 1

# Match expression
let label = match score:
    100     -> "perfect"
    90      -> "excellent"
    _       -> "good"

# Match with guards
let grade = match score:
    _ if score >= 90 -> "A"
    _ if score >= 70 -> "B"
    _                -> "C"
```

### Option Type

```aura
# Return an optional value
fn find(items: [Int], target: Int) -> Option[Int]:
    for item in items:
        if item == target:
            return Some(item)
    return none

# Use an optional value
let result = find([1, 2, 3], 2)
match result:
    Some(v) -> std.io.println("Found: {v}")
    None    -> std.io.println("Not found")
```

### Data Types

```aura
let x = 42               # Int
let pi = 3.14            # Float
let name = "Aura"        # String
let active = true        # Bool
let nothing = none       # None (absence of value)
let items = [1, 2, 3]   # List[Int]
let pair = (1, "hello")  # Tuple
```

### String Interpolation

```aura
let name = "World"
let greeting = "Hello, {name}!"    # "Hello, World!"
let math = "2 + 2 = {2 + 2}"      # "2 + 2 = 4"
let nested = "{p.x} and {p.y}"    # Fields work too
```

### Operators

| Operator | Description | Example |
|---|---|---|
| `+` `-` `*` `/` `%` | Arithmetic | `10 + 3` → `13` |
| `**` | Exponentiation | `2 ** 8` → `256` |
| `==` `!=` | Equality | `x == 5` |
| `<` `>` `<=` `>=` | Comparison | `x < 10` |
| `and` `or` `not` | Logical | `x > 0 and x < 100` |
| `\|>` | Pipeline | `data \|> transform \|> print` |

### Lambdas

```aura
# Inline anonymous function: |params| -> expr
let double = |x| -> x * 2

# With multiple parameters
let add = |a, b| -> a + b

# Used with list methods
let evens = [1, 2, 3, 4].filter(|x| -> x % 2 == 0)   # [2, 4]
let doubled = [1, 2, 3].map(|x| -> x * 2)              # [2, 4, 6]
```

### Printing and Input

```aura
import std.io as io

io.println("Hello!")           # Print with newline
io.print("No newline")         # Print without newline
let name = io.input("Name: ")  # Read a line from the user (returns String)
```

---

## Next Steps

Congratulations on completing these tutorials! You now know the fundamentals of Aura programming. Here's where to go next:

### Read More Documentation

- **[Language Guide](../user_docs/language_guide.md)** — Deep dive into all Aura features: enums, traits, specs, effects, and more.
- **[Method Reference](../user_docs/method_reference.md)** — Complete reference for 96+ built-in methods on String, List, Map, Option, and Result.
- **[Language Reference](../user_docs/language_reference.md)** — Formal reference for types, syntax, and the effect system.

### Features to Explore Next

Now that you know the basics, try learning these powerful Aura features:

1. **Enums** — Define types with multiple variants:
   ```aura
   pub enum Shape:
       Circle(Float)
       Rectangle(Float, Float)
       Triangle(Float, Float, Float)

   fn area(shape: Shape) -> Float:
       return match shape:
           Circle(r)         -> 3.14159 * r * r
           Rectangle(w, h)   -> w * h
           Triangle(a, b, c) -> # Heron's formula
   ```

2. **Result types** — Handle errors explicitly:
   ```aura
   fn divide(a: Float, b: Float) -> Result[Float, String]:
       if b == 0.0:
           return Err("Cannot divide by zero")
       return Ok(a / b)

   match divide(10.0, 3.0):
       Ok(v)  -> std.io.println("Result: {v}")
       Err(e) -> std.io.println("Error: {e}")
   ```

3. **Pipeline operator** — Chain operations elegantly:
   ```aura
   let result = "  Hello, World!  "
       |> |s| -> s.trim()
       |> |s| -> s.lower()
       |> |s| -> s.replace(",", "")
   # result = "hello world!"
   ```

4. **List comprehensions** — Create lists concisely:
   ```aura
   let squares = [x * x for x in range(10)]
   let evens = [x for x in range(20) if x % 2 == 0]
   ```

5. **Refinement types** — Attach constraints directly to types:
   ```aura
   type PositiveInt = Int where self > 0
   type ShortString = String where len >= 1 and len <= 100
   ```
   The compiler enforces these constraints at every assignment.

6. **The Effect System** — Aura's unique approach to managing side effects (file I/O, network, time) with full mockability for testing:
   ```aura
   pub fn save_file(path: String, content: String) -> Result with file:
       # `with file` declares this function uses the filesystem
       std.file.write(path, content)
   ```

### AI-First Development

Aura was designed for AI-human collaboration. Once you're comfortable with the basics, check out the [AI Mission Statement](../AI_MISSION.md) to learn about:
- **Spec blocks** — Structured contracts that AI uses to generate code
- **Effect annotations** — Tell AI exactly what side effects are allowed
- **`satisfies` clauses** — Automatic verification of AI-generated code

```aura
spec GreetUser:
    doc: "Greet a user by name"
    inputs:
        name: String - "The user's name"
    guarantees:
        - "Prints a greeting to the screen"
    effects: io

pub fn greet(name: String) -> None with io satisfies GreetUser:
    std.io.println("Hello, {name}!")
```

### Get Help

- Explore the **REPL** (`aura repl`) to experiment interactively
- Read the test files in the repository for working examples of every feature
- Check the [ROADMAP](../ROADMAP.md) to see what's coming next

---

*Happy coding with Aura!*
