# Aura Examples — Showcasing Built-in Methods

> Real-world examples demonstrating Aura's 96+ built-in methods across all types.
> For the complete method reference, see [method_reference.md](./method_reference.md).

**Note:** Examples that print output use `io.println`. Add `import std.io as io` at the top of any file that uses it.

---

## Table of Contents

- [String Manipulation](#string-manipulation)
- [List Functional Programming](#list-functional-programming)
- [Map Operations & Transformations](#map-operations--transformations)
- [Option Handling Patterns](#option-handling-patterns)
- [Result Error Handling](#result-error-handling)
- [Method Chaining Across Types](#method-chaining-across-types)
- [Advanced Patterns](#advanced-patterns)

---

## String Manipulation

### Cleaning User Input

```aura
fn clean_input(raw):
    return raw.trim().lower().replace("  ", " ")

let input = "  Hello   World  "
let clean = clean_input(input)
# => "hello world"
```

### Title Case Conversion

```aura
fn capitalize(word):
    if word.is_empty():
        return word
    return word.slice(0, 1).upper() + word.slice(1)

fn title_case(s):
    return s.lower().split(" ").map(|w| -> capitalize(w)).join(" ")

let title = title_case("the quick brown fox")
# => "The Quick Brown Fox"
```

### String Searching and Extraction

```aura
let url = "https://example.com/users/42/profile"

# Find a substring position
let idx = url.index_of("/users/")
# => Some(23)

# Extract segments
let parts = url.split("/")
# => ["https:", "", "example.com", "users", "42", "profile"]

let user_id = parts.get(4).unwrap_or("unknown")
# => "42"

# Check patterns
let is_https = url.starts_with("https")
# => true
let is_profile = url.ends_with("profile")
# => true
```

### Formatting and Padding

```aura
import std.io as io

let items = [("Widget", 9.99), ("Gadget", 149.50), ("Doohickey", 3.00)]

for item in items:
    let name = item[0].pad_right(12, ".")
    let price = "{item[1]}".pad_left(8)
    io.println("{name}{price}")

# Widget......    9.99
# Gadget......  149.50
# Doohickey...    3.00
```

### Reversing Words

```aura
fn reverse_words(sentence):
    return sentence.split(" ").map(|word| -> word.reverse()).join(" ")

let result = reverse_words("hello world")
# => "olleh dlrow"
```

### Repeating Patterns

```aura
# Build a visual separator
let separator = "=-".repeat(20)
# => "=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-="

fn make_indent(level):
    return "  ".repeat(level)

let code = make_indent(3) + "return x"
# => "      return x"
```

---

## List Functional Programming

### Map / Filter / Reduce Chains

```aura
let numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

# Square all even numbers
let even_squares = (numbers
    .filter(|n| -> n % 2 == 0)
    .map(|n| -> n * n))
# => [4, 16, 36, 64, 100]

# Sum of odd numbers
let odd_sum = (numbers
    .filter(|n| -> n % 2 != 0)
    .reduce(0, |acc, n| -> acc + n))
# => 25
```

### Flat Map for One-to-Many Transformations

```aura
let sentences = ["hello world", "foo bar baz"]

let words = sentences.flat_map(|s| -> s.split(" "))
# => ["hello", "world", "foo", "bar", "baz"]

# Generate all (row, col) pairs
let rows = [1, 2, 3]
let cols = [1, 2]
let pairs = rows.flat_map(|r| -> cols.map(|c| -> [r, c]))
# => [[1,1], [1,2], [2,1], [2,2], [3,1], [3,2]]
```

### Sorting and Deduplication

```aura
let tags = ["rust", "go", "rust", "python", "go", "aura", "python"]

let unique_sorted = tags.unique().sort()
# => ["aura", "go", "python", "rust"]
```

### Working with Indices

```aura
import std.io as io

let names = ["Alice", "Bob", "Charlie"]

for pair in names.enumerate():
    let idx = pair[0]
    let name = pair[1]
    io.println("{idx + 1}. {name}")

# 1. Alice
# 2. Bob
# 3. Charlie
```

### Zipping Parallel Lists

```aura
let students = ["Alice", "Bob", "Charlie"]
let scores = [95, 87, 92]

let report = (students
    .zip(scores)
    .filter(|pair| -> pair[1] >= 90)
    .map(|pair| -> "{pair[0]}: {pair[1]}")
    .join(", "))
# => "Alice: 95, Charlie: 92"
```

### Predicates: Any and All

```aura
let numbers = [2, 4, 6, 8, 10]

let all_even = numbers.all(|n| -> n % 2 == 0)
# => true
let any_big = numbers.any(|n| -> n > 100)
# => false
let count_gt5 = numbers.count(|n| -> n > 5)
# => 3
```

### Aggregation: Sum, Min, Max

```aura
let scores = [72, 95, 88, 63, 91]

let total = scores.sum()
# => 409
let best = scores.max().unwrap_or(0)
# => 95
let worst = scores.min().unwrap_or(0)
# => 63
let average = total / scores.len()
# => 81
```

### Nested List Flattening

```aura
let matrix = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
let flat = matrix.flatten()
# => [1, 2, 3, 4, 5, 6, 7, 8, 9]

let total = matrix.flatten().sum()
# => 45
```

### Safe Element Access

```aura
let items = [10, 20, 30]

# Safe access — never panics
let first = items.first()
# => Some(10)
let last = items.last()
# => Some(30)
let third = items.get(2)
# => Some(30)
let missing = items.get(99)
# => None

# Empty list safety
let empty = []
let none_first = empty.first()
# => None
let none_pop = empty.pop()
# => None
let none_min = empty.min()
# => None
```

---

## Map Operations & Transformations

### Building and Querying Maps

```aura
let config = {"host": "localhost", "port": "8080", "debug": "true"}

# Safe access with defaults
let host = config.get("host").unwrap_or("0.0.0.0")
# => "localhost"
let timeout = config.get_or("timeout", "30")
# => "30"

# Check existence
let has_port = config.has("port")
# => true
let has_ssl = config.contains_key("ssl")
# => false
```

### Transforming Map Values

```aura
let word_lengths = {"apple": 5, "banana": 6, "cherry": 6}

# Double all values
let doubled = word_lengths.map(|name, n| -> n * 2)
# => {"apple": 10, "banana": 12, "cherry": 12}

# Keep only longer words
let long_words = word_lengths.filter(|name, n| -> n > 5)
# => {"banana": 6, "cherry": 6}
```

### Aggregating Map Data

```aura
let inventory = {"apples": 50, "bananas": 120, "cherries": 30}

# Total items
let total = inventory.reduce(0, |acc, name, count| -> acc + count)
# => 200

# Check stock levels
let all_in_stock = inventory.all(|name, count| -> count > 0)
# => true
let any_low = inventory.any(|name, count| -> count < 40)
# => true

# Count low-stock items
let low_stock = inventory.count(|name, count| -> count < 50)
# => 2
```

### Finding Entries

```aura
let scores = {"alice": 95, "bob": 72, "carol": 88}

# Find first high scorer
let top = scores.find(|name, score| -> score >= 90)
# => Some(["alice", 95])  (first match)

# Check if anyone passed
let any_passed = scores.any(|name, score| -> score >= 70)
# => true
```

### Merging Maps

```aura
let defaults = {"color": "blue", "size": "medium", "visible": "true"}
let overrides = {"color": "red", "size": "large"}

let config = {"color": "blue", "size": "medium", "visible": "true"}
config.merge(overrides)
# config => {"color": "red", "size": "large", "visible": "true"}
```

### Map ↔ List Conversion

```aura
let scores = {"alice": 95, "bob": 87, "carol": 92}

# Get keys and values as lists
let names = scores.keys()
# => ["alice", "bob", "carol"]
let values = scores.values()
# => [95, 87, 92]

# Convert to list of entries (each entry is [key, value])
let entries = scores.entries()
# => [["alice", 95], ["bob", 87], ["carol", 92]]
```

### Building a Map Incrementally

```aura
# Word frequency count
let text = "the cat sat on the mat the cat"
let words = text.split(" ")

let mut freq = {}
for word in words:
    let count = freq.get_or(word, 0)
    freq.set(word, count + 1)
# freq => {"the": 3, "cat": 2, "sat": 1, "on": 1, "mat": 1}

# Find words appearing more than once
let common = (freq
    .filter(|word, count| -> count > 1)
    .keys()
    .sort())
# => ["cat", "the"]
```

---

## Option Handling Patterns

### Safe Navigation Pattern

Avoid panics by chaining Option methods:

```aura
# Simulate a lookup that might fail
fn get_city(user_data):
    return (user_data
        .get("city")
        .map(|c| -> c.upper())
        .unwrap_or("UNKNOWN"))

let data1 = {"city": "Paris", "zip": "75001"}
let data2 = {"zip": "00000"}

let city1 = get_city(data1)
# => "PARIS"
let city2 = get_city(data2)
# => "UNKNOWN"
```

### Option as a Null-Safe Container

```aura
fn find_user(id):
    if id == 1:
        return Some({"name": "Alice", "email": "alice@example.com"})
    return none

# Transform if found, default if not
let greeting = (find_user(1)
    .map(|u| -> "Hello, " + u.get("name").unwrap())
    .unwrap_or("Hello, stranger"))
# => "Hello, Alice"

let greeting2 = (find_user(99)
    .map(|u| -> "Hello, " + u.get("name").unwrap())
    .unwrap_or("Hello, stranger"))
# => "Hello, stranger"
```

### Fallback Chains with `or`

```aura
fn from_cache(key):
    return none

fn from_db(key):
    return Some(42)

fn default_value():
    return Some(0)

let value = (from_cache("x")
    .or(from_db("x"))
    .or(default_value())
    .unwrap())
# => 42
```

### Filtering Options

```aura
fn parse_port(s):
    # Simulated: returns Some(port) if valid
    return Some(8080)

let port = (parse_port("8080")
    .filter(|p| -> p > 0)
    .filter(|p| -> p < 65536)
    .unwrap_or(3000))
# => 8080
```

### Zipping Options Together

```aura
let first_name = Some("Alice")
let last_name = Some("Smith")

let full_name = (first_name
    .zip(last_name)
    .map(|pair| -> pair[0] + " " + pair[1])
    .unwrap_or("Anonymous"))
# => "Alice Smith"

# If either is none, result is none
let partial = Some("Alice").zip(none)
# => None
```

### Converting Between Option and Result

```aura
# Option → Result: attach an error message
let opt = Some(42)
let res = opt.to_result("value was missing")
# => Ok(42)

let res2 = none.to_result("value was missing")
# => Err("value was missing")

# Result → Option: discard the error
let ok_result = Ok(42)
let opt2 = ok_result.ok()
# => Some(42)

let err_result = Err("oops")
let opt3 = err_result.ok()
# => None
```

---

## Result Error Handling

### Railway-Oriented Programming

Chain operations that might fail — the first error short-circuits the chain:

```aura
fn parse_positive(s):
    if s == "42":
        return Ok(42)
    elif s == "10":
        return Ok(10)
    else:
        return Err("not a valid number: " + s)

fn validate_positive(n):
    if n > 0:
        return Ok(n)
    return Err("must be positive")

fn double(n):
    return Ok(n * 2)

# Happy path — all steps succeed
let result = (parse_positive("42")
    .and_then(|n| -> validate_positive(n))
    .and_then(|n| -> double(n)))
# => Ok(84)

# Error path — first failure short-circuits
let result2 = (parse_positive("abc")
    .and_then(|n| -> validate_positive(n))
    .and_then(|n| -> double(n)))
# => Err("not a valid number: abc")
```

### Transforming Success and Error Values

```aura
let result = Ok(5)

# Transform the success value
let doubled = result.map(|x| -> x * 2)
# => Ok(10)

# Transform the error value
let err = Err("fail")
let detailed = err.map_err(|e| -> "ERROR: " + e)
# => Err("ERROR: fail")

# map does not affect errors
let still_err = Err("fail").map(|x| -> x * 2)
# => Err("fail")
```

### Providing Fallbacks

```aura
fn fetch_from_primary():
    return Err("primary unavailable")

fn fetch_from_backup():
    return Ok("backup data")

# Try primary, fall back to backup
let data = (fetch_from_primary()
    .or_else(|e| -> fetch_from_backup())
    .unwrap_or("no data"))
# => "backup data"

# Simple fallback with `or`
let data2 = Err("fail").or(Ok("default"))
# => Ok("default")
```

### Extracting from Results Safely

```aura
let ok = Ok(42)
let err = Err("problem")

# Safe extraction with defaults
let v1 = ok.unwrap_or(0)
# => 42
let v2 = err.unwrap_or(0)
# => 0

# Check contents without extracting
let has42 = ok.contains(42)
# => true
let has_prob = err.contains_err("problem")
# => true
```

### Converting Results to Options

```aura
let ok = Ok(42)
let err = Err("fail")

# Extract the success value as an Option
let some42 = ok.ok()
# => Some(42)
let none_from_err = err.ok()
# => None

# Extract the error as an Option
let none_from_ok = ok.err()
# => None
let some_err = err.err()
# => Some("fail")
```

---

## Method Chaining Across Types

### String → List → String

```aura
# CSV processing pipeline
let csv_line = "  Alice, 30, Paris  "

let fields = (csv_line
    .trim()
    .split(",")
    .map(|f| -> f.trim()))
# => ["Alice", "30", "Paris"]

let formatted = fields.join(" | ")
# => "Alice | 30 | Paris"
```

### List → Map → List

```aura
# Group items by first character
let words = ["apple", "avocado", "banana", "blueberry", "cherry"]

let mut groups = {}
for word in words:
    let key = word.slice(0, 1)
    let existing = groups.get_or(key, [])
    groups.set(key, existing.append(word))
# groups => {"a": ["apple", "avocado"], "b": ["banana", "blueberry"], "c": ["cherry"]}

# Get the groups that have more than one item
let multi = (groups
    .filter(|k, v| -> v.len() > 1)
    .keys()
    .sort())
# => ["a", "b"]
```

### Map → Option → Result

```aura
let config = {"database_url": "postgres://localhost/mydb"}

let db_result = (config
    .get("database_url")
    .filter(|url| -> url.starts_with("postgres"))
    .to_result("Missing or invalid database_url"))
# => Ok("postgres://localhost/mydb")

let bad_config = {"database_url": "mysql://localhost/mydb"}
let bad_result = (bad_config
    .get("database_url")
    .filter(|url| -> url.starts_with("postgres"))
    .to_result("Missing or invalid database_url"))
# => Err("Missing or invalid database_url")
```

### Full Pipeline: Parse, Validate, Transform

```aura
let raw_data = "  hello, world, foo, hi, aura  "

let result = (raw_data
    .trim()
    .split(",")
    .map(|s| -> s.trim())
    .filter(|s| -> s.is_empty() == false)
    .map(|s| -> s.upper())
    .sort())
# => ["AURA", "FOO", "HI", "HELLO", "WORLD"]
```

---

## Advanced Patterns

### Monadic Option Composition

Chain operations that each might return `none`, using `and_then`:

```aura
fn safe_div(a, b):
    if b == 0:
        return none
    return Some(a / b)

fn safe_head(items):
    return items.first()

# Chain: divide, then get the head of a list
let result = (safe_div(100, 4)
    .and_then(|x| -> safe_head([x, x * 2, x * 3]))
    .map(|x| -> x + 1)
    .unwrap_or(0))
# => 26  (100/4=25, head=25, +1=26)
```

### Result Monadic Chains (Railway Pattern)

Build complex pipelines where any step can fail:

```aura
fn fetch_user(id):
    if id > 0:
        return Ok({"name": "Alice", "age": "30"})
    return Err("invalid user id")

fn check_age(user):
    let age_str = user.get("age").unwrap_or("0")
    if age_str != "0":
        return Ok(user)
    return Err("user age unavailable")

fn format_greeting(user):
    let name = user.get("name").unwrap_or("Unknown")
    return Ok("Welcome, {name}!")

let greeting = (fetch_user(1)
    .and_then(|u| -> check_age(u))
    .and_then(|u| -> format_greeting(u))
    .unwrap_or("Access denied"))
# => "Welcome, Alice!"

let denied = (fetch_user(-1)
    .and_then(|u| -> check_age(u))
    .and_then(|u| -> format_greeting(u))
    .unwrap_or("Access denied"))
# => "Access denied"
```

### Functional Data Processing Pipeline

```aura
pub struct Person:
    pub name: String
    pub age: Int
    pub active: Bool

let people = [
    Person(name: "Alice", age: 30, active: true),
    Person(name: "Bob", age: 17, active: true),
    Person(name: "Carol", age: 25, active: false),
    Person(name: "Dave", age: 45, active: true),
    Person(name: "Eve", age: 22, active: true),
]

# Active adults, sorted by name
let active_adults = (people
    .filter(|p| -> p.active)
    .filter(|p| -> p.age >= 18)
    .map(|p| -> p.name)
    .sort())
# => ["Alice", "Dave", "Eve"]

# Age statistics
let ages = people.map(|p| -> p.age)
let total = ages.sum()
let avg = total / ages.len()
# => 27
let youngest = ages.min().unwrap_or(0)
# => 17
let oldest = ages.max().unwrap_or(0)
# => 45
```

### Building a Simple Key-Value Store

```aura
let mut store = {}

# Set values
store.set("users:1", "Alice")
store.set("users:2", "Bob")
store.set("config:theme", "dark")

# Query by prefix
let user_keys = (store
    .keys()
    .filter(|k| -> k.starts_with("users:")))
# => ["users:1", "users:2"]

# Count entries by category
let user_count = store.count(|k, v| -> k.starts_with("users:"))
let config_count = store.count(|k, v| -> k.starts_with("config:"))
# user_count => 2, config_count => 1
```

### Option/Result Round-Trip

```aura
fn validate(x):
    if x > 100:
        return Err("too large")
    return Ok(x)

let maybe_value = Some(42)

let processed = (maybe_value
    .to_result("no value")
    .map(|x| -> x * 2)
    .and_then(|x| -> validate(x))
    .ok()
    .unwrap_or(0))
# => 84
```

### Nested Flatten Operations

```aura
# Flatten nested Options
let nested = Some(Some(Some(42)))
let flat = nested.flatten()
# => Some(Some(42))
let flatter = nested.flatten().flatten()
# => Some(42)

# Flatten nested Results
let nested_result = Ok(Ok(42))
let flat_result = nested_result.flatten()
# => Ok(42)
```

### Combining Multiple Safe Lookups

```aura
fn get_connection_string(server_config):
    let host = server_config.get("host")
    let port = server_config.get("port")
    return (host
        .zip(port)
        .map(|pair| -> "https://{pair[0]}:{pair[1]}")
        .unwrap_or("https://localhost:8080"))

let server = {"host": "example.com", "port": "443"}
let url = get_connection_string(server)
# => "https://example.com:443"

let empty_server = {}
let default_url = get_connection_string(empty_server)
# => "https://localhost:8080"
```

---

## Quick Reference: Method Counts by Type

| Type | Methods | Key Highlights |
|------|---------|----------------|
| **String** | 19 | `trim`, `split`, `upper`/`lower`, `slice`, `pad_left`/`pad_right`, `reverse` |
| **List** | 26 | `map`, `filter`, `reduce`, `sort`, `zip`, `flat_map`, `enumerate` |
| **Map** | 22 | `get` → Option, `filter`, `map`, `reduce`, higher-order ops |
| **Option** | 17 | Monadic chaining, safe defaults, `and_then`, `zip`, combinators |
| **Result** | 18 | Railway-oriented, `map`/`map_err`, `and_then`, `or_else` |
| **Total** | **96+** | |

> For the full method reference with signatures and parameters, see [method_reference.md](./method_reference.md).
> For the development roadmap, see [ROADMAP.md](../ROADMAP.md).

---

*Built for Aura v2.0.1 — An AI-first programming language designed for vibe coding.*
