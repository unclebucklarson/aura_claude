# Aura Examples — Showcasing Built-in Methods

> Real-world examples demonstrating Aura's 108+ built-in methods across all types.
> For the complete method reference, see [method_reference.md](./method_reference.md).

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
fn clean_input(raw) {
  raw
    .trim()
    .lower()
    .replace("  ", " ")
}

let input = "  Hello   World  "
let clean = clean_input(input)  // => "hello world"
```

### Title Case Conversion
```aura
fn title_case(s) {
  s.lower()
   .split(" ")
   .map(|word| {
     if word.is_empty() {
       word
     } else {
       word.slice(0, 1).upper() + word.slice(1)
     }
   })
   .join(" ")
}

let title = title_case("the quick brown fox")
// => "The Quick Brown Fox"
```

### String Searching and Extraction
```aura
let url = "https://example.com/users/42/profile"

// Find a substring position
let idx = url.index_of("/users/")   // => Some(23)

// Extract segments
let parts = url.split("/")
// => ["https:", "", "example.com", "users", "42", "profile"]

let user_id = parts.get(4).unwrap_or("unknown")  // => "42"

// Check patterns
let is_https = url.starts_with("https")           // => true
let is_profile = url.ends_with("profile")          // => true
```

### Formatting and Padding
```aura
// Right-aligned numeric table
let items = [("Widget", 9.99), ("Gadget", 149.50), ("Doohickey", 3.00)]

items.for_each(|item| {
  let name = item[0].pad_right(12, ".")
  let price = "${item[1]}".pad_left(8)
  print("{name}{price}")
})
// Widget......    9.99
// Gadget......  149.50
// Doohickey...    3.00
```

### Building Strings from Characters
```aura
// Reverse words but keep word order
fn reverse_words(sentence) {
  sentence
    .split(" ")
    .map(|word| word.reverse())
    .join(" ")
}

let result = reverse_words("hello world")  // => "olleh dlrow"
```

### Repeating Patterns
```aura
// Build a visual separator
let separator = "=-".repeat(20)  // => "=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-="

// Build an indentation string
fn indent(level) {
  "  ".repeat(level)
}

let code = indent(3) + "print('hello')"
// => "      print('hello')"
```

---

## List Functional Programming

### Map / Filter / Reduce Chains
```aura
let transactions = [
  {"type": "credit", "amount": 100},
  {"type": "debit", "amount": 50},
  {"type": "credit", "amount": 200},
  {"type": "debit", "amount": 75},
  {"type": "credit", "amount": 150}
]

// Total credits
let total_credits = transactions
  .filter(|t| t.get("type").unwrap() == "credit")
  .map(|t| t.get("amount").unwrap())
  .reduce(0, |acc, x| acc + x)
// => 450

// Or more concisely with sum
let total_credits = transactions
  .filter(|t| t.get("type").unwrap() == "credit")
  .map(|t| t.get("amount").unwrap())
  .sum()
// => 450
```

### Flat Map for One-to-Many Transformations
```aura
let sentences = ["hello world", "foo bar baz"]

let words = sentences.flat_map(|s| s.split(" "))
// => ["hello", "world", "foo", "bar", "baz"]

// Generate coordinate pairs
let rows = [1, 2, 3]
let pairs = rows.flat_map(|r| [1, 2].map(|c| [r, c]))
// => [[1,1], [1,2], [2,1], [2,2], [3,1], [3,2]]
```

### Sorting and Deduplication
```aura
let tags = ["rust", "go", "rust", "python", "go", "aura", "python"]

let unique_sorted = tags.unique().sort()
// => ["aura", "go", "python", "rust"]
```

### Working with Indices
```aura
let names = ["Alice", "Bob", "Charlie"]

// Enumerate for indexed processing
names.enumerate().for_each(|pair| {
  let idx = pair[0]
  let name = pair[1]
  print("{idx + 1}. {name}")
})
// 1. Alice
// 2. Bob
// 3. Charlie
```

### Zipping Parallel Lists
```aura
let students = ["Alice", "Bob", "Charlie"]
let scores = [95, 87, 92]

let report = students.zip(scores)
  .filter(|pair| pair[1] >= 90)
  .map(|pair| "{pair[0]}: {pair[1]}")
  .join(", ")
// => "Alice: 95, Charlie: 92"
```

### Predicates: Any and All
```aura
let numbers = [2, 4, 6, 8, 10]

let all_even = numbers.all(|n| n % 2 == 0)     // => true
let any_big = numbers.any(|n| n > 100)          // => false
let count_gt5 = numbers.count(|n| n > 5)        // => 3
```

### Aggregation: Sum, Min, Max
```aura
let scores = [72, 95, 88, 63, 91]

let total = scores.sum()                          // => 409
let best = scores.max().unwrap_or(0)              // => 95
let worst = scores.min().unwrap_or(0)             // => 63
let average = total / scores.len()                // => 81
```

### Nested List Flattening
```aura
let matrix = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
let flat = matrix.flatten()
// => [1, 2, 3, 4, 5, 6, 7, 8, 9]

let total = matrix.flatten().sum()   // => 45
```

### Safe Element Access
```aura
let items = [10, 20, 30]

// Safe access — never panics
let first = items.first()              // => Some(10)
let last = items.last()                // => Some(30)
let third = items.get(2)               // => Some(30)
let missing = items.get(99)            // => None

// Empty list safety
let empty = []
empty.first()                          // => None
empty.pop()                            // => None
empty.min()                            // => None
```

---

## Map Operations & Transformations

### Building and Querying Maps
```aura
let config = {
  "host": "localhost",
  "port": 8080,
  "debug": true
}

// Safe access with defaults
let host = config.get("host").unwrap_or("0.0.0.0")
let timeout = config.get_or("timeout", 30)

// Check existence
let has_port = config.has("port")           // => true
let has_ssl = config.contains_key("ssl")    // => false
```

### Transforming Map Values
```aura
let prices = {"apple": 1.50, "banana": 0.75, "cherry": 2.00}

// Apply a discount
let discounted = prices.map(|name, price| price * 0.9)
// => {"apple": 1.35, "banana": 0.675, "cherry": 1.80}

// Filter expensive items
let premium = prices.filter(|name, price| price > 1.00)
// => {"apple": 1.50, "cherry": 2.00}
```

### Aggregating Map Data
```aura
let inventory = {"apples": 50, "bananas": 120, "cherries": 30}

// Total items
let total = inventory.reduce(0, |acc, name, count| acc + count)
// => 200

// Check stock
let all_in_stock = inventory.all(|name, count| count > 0)   // => true
let any_low = inventory.any(|name, count| count < 40)        // => true

// Count categories
let low_stock = inventory.count(|name, count| count < 50)    // => 1
```

### Finding Entries
```aura
let users = {
  "alice": {"age": 30, "role": "admin"},
  "bob": {"age": 25, "role": "user"},
  "carol": {"age": 35, "role": "admin"}
}

// Find first admin
let admin = users.find(|name, info| info.get("role").unwrap() == "admin")
// => Some(["alice", {"age": 30, "role": "admin"}])
```

### Merging Maps
```aura
let defaults = {"color": "blue", "size": "medium", "visible": true}
let overrides = {"color": "red", "size": "large"}

let config = {"color": "blue", "size": "medium", "visible": true}
config.merge(overrides)
// config => {"color": "red", "size": "large", "visible": true}
```

### Map ↔ List Conversion
```aura
let scores = {"alice": 95, "bob": 87, "carol": 92}

// Convert to sorted list of entries
let ranked = scores.entries()
  .sort()     // sorts by first element (name) alphabetically

// Get just the keys and values
let names = scores.keys()       // => ["alice", "bob", "carol"]
let values = scores.values()    // => [95, 87, 92]
```

---

## Option Handling Patterns

### Safe Navigation Pattern
Avoid nulls and panics by chaining Option methods:

```aura
let user = {
  "name": "Alice",
  "address": {
    "city": "Paris",
    "zip": "75001"
  }
}

// Safe deep access
let city = user.get("address")
  .and_then(|addr| addr.get("city"))
  .map(|c| c.upper())
  .unwrap_or("UNKNOWN")
// => "PARIS"

// When the path doesn't exist
let phone = user.get("contact")
  .and_then(|c| c.get("phone"))
  .unwrap_or("N/A")
// => "N/A"
```

### Option as a Null-Safe Container
```aura
fn find_user(id) {
  if id == 1 {
    Some({"name": "Alice", "email": "alice@example.com"})
  } else {
    None
  }
}

// Pattern: Transform if found, default if not
let greeting = find_user(1)
  .map(|u| "Hello, " + u.get("name").unwrap())
  .unwrap_or("Hello, stranger")
// => "Hello, Alice"

let greeting2 = find_user(99)
  .map(|u| "Hello, " + u.get("name").unwrap())
  .unwrap_or("Hello, stranger")
// => "Hello, stranger"
```

### Fallback Chains with `or`
```aura
fn from_cache(key) { None }
fn from_db(key) { Some(42) }
fn default_value() { Some(0) }

let value = from_cache("x")
  .or(from_db("x"))
  .or(default_value())
  .unwrap()
// => 42
```

### Filtering Options
```aura
fn parse_port(s) {
  // Simulated: returns Some(port) if valid
  Some(8080)
}

let port = parse_port("8080")
  .filter(|p| p > 0)
  .filter(|p| p < 65536)
  .unwrap_or(3000)
// => 8080
```

### Zipping Options Together
```aura
let first_name = Some("Alice")
let last_name = Some("Smith")

let full_name = first_name.zip(last_name)
  .map(|pair| pair.get(0).unwrap() + " " + pair.get(1).unwrap())
  .unwrap_or("Anonymous")
// => "Alice Smith"

// If either is None, result is None
let partial = Some("Alice").zip(None)
// => None
```

### Converting Between Option and Result
```aura
// Option → Result: attach an error message
let opt = Some(42)
let res = opt.to_result("value was missing")   // => Ok(42)

let none_opt = None
let res2 = none_opt.to_result("value was missing")  // => Err("value was missing")

// Result → Option: discard the error
let ok_result = Ok(42)
let opt2 = ok_result.ok()     // => Some(42)
let opt3 = ok_result.to_option()  // => Some(42)

let err_result = Err("oops")
let opt4 = err_result.ok()    // => None
```

---

## Result Error Handling

### Railway-Oriented Programming
Chain operations that might fail — errors short-circuit automatically:

```aura
fn parse_int(s) {
  if s == "42" { Ok(42) }
  else if s == "10" { Ok(10) }
  else { Err("not a valid number: " + s) }
}

fn validate_positive(n) {
  if n > 0 { Ok(n) }
  else { Err("must be positive") }
}

fn double(n) {
  Ok(n * 2)
}

// Happy path — all steps succeed
let result = parse_int("42")
  .and_then(|n| validate_positive(n))
  .and_then(|n| double(n))
// => Ok(84)

// Error path — first failure short-circuits
let result2 = parse_int("abc")
  .and_then(|n| validate_positive(n))
  .and_then(|n| double(n))
// => Err("not a valid number: abc")
```

### Transforming Success and Error Values
```aura
let result = Ok(5)

// Transform the success value
let doubled = result.map(|x| x * 2)           // => Ok(10)

// Transform the error value
let err = Err("fail")
let detailed = err.map_err(|e| "ERROR: " + e) // => Err("ERROR: fail")

// Map doesn't affect errors
Err("fail").map(|x| x * 2)                    // => Err("fail")
```

### Providing Fallbacks
```aura
fn fetch_from_primary() {
  Err("primary unavailable")
}

fn fetch_from_backup() {
  Ok("backup data")
}

// Try primary, fall back to backup
let data = fetch_from_primary()
  .or_else(|e| fetch_from_backup())
  .unwrap_or("no data")
// => "backup data"

// Simple alternative with `or`
let data2 = Err("fail").or(Ok("default"))
// => Ok("default")
```

### Extracting from Results Safely
```aura
let ok = Ok(42)
let err = Err("problem")

// Safe extraction with defaults
ok.unwrap_or(0)                    // => 42
err.unwrap_or(0)                   // => 0

// Lazy defaults
err.unwrap_or_else(|e| {
  print("Error occurred: {e}")
  0
})

// Check contents without extracting
ok.contains(42)                    // => true
err.contains_err("problem")       // => true
```

### Converting Results to Options
```aura
let ok = Ok(42)
let err = Err("fail")

// Extract the success value as an Option
ok.ok()                            // => Some(42)
err.ok()                           // => None

// Extract the error as an Option
ok.err()                           // => None
err.err()                          // => Some("fail")
```

---

## Method Chaining Across Types

### String → List → String
```aura
// CSV processing pipeline
let csv_line = "  Alice, 30, Paris  "

let fields = csv_line
  .trim()
  .split(",")
  .map(|f| f.trim())
// => ["Alice", "30", "Paris"]

let formatted = fields.join(" | ")
// => "Alice | 30 | Paris"
```

### List → Map → List
```aura
// Word frequency analysis
let text = "the cat sat on the mat the cat"
let words = text.split(" ")

let freq = words.reduce({}, |acc, word| {
  let count = acc.get_or(word, 0)
  acc.set(word, count + 1)
  acc
})
// freq => {"the": 3, "cat": 2, "sat": 1, "on": 1, "mat": 1}

// Find words appearing more than once
let common = freq
  .filter(|word, count| count > 1)
  .keys()
  .sort()
// => ["cat", "the"]
```

### Map → Option → Result
```aura
let config = {"database_url": "postgres://localhost/mydb"}

let db_result = config
  .get("database_url")               // Option[String]
  .filter(|url| url.starts_with("postgres"))  // Option[String]
  .to_result("Missing or invalid database_url")  // Result[String, String]

// => Ok("postgres://localhost/mydb")
```

### Full Pipeline: Parse, Validate, Transform
```aura
let raw_data = "  42, 17, 99, 3, 55, 88  "

let result = raw_data
  .trim()
  .split(",")
  .map(|s| s.trim())
  .filter(|s| s.is_empty() == false)
  .map(|s| s.len())    // using len as a proxy transform
  .filter(|n| n > 1)
  .sort()
  .reverse()
// Pipeline of string → list → filtered list → sorted list
```

---

## Advanced Patterns

### Monadic Option Composition
Chain operations that each might fail, using `and_then`:

```aura
fn safe_div(a, b) {
  if b == 0 { None } else { Some(a / b) }
}

fn safe_sqrt(n) {
  if n < 0 { None } else { Some(n) }  // simplified
}

// Chain: divide then square root
let result = safe_div(100, 4)
  .and_then(|x| safe_sqrt(x))
  .map(|x| x * 2)
  .unwrap_or(0)
// => 50
```

### Result Monadic Chains (Railway Pattern)
Build complex pipelines where any step can fail:

```aura
fn fetch_user(id) {
  if id > 0 { Ok({"name": "Alice", "age": 30}) }
  else { Err("invalid user id") }
}

fn check_age(user) {
  let age = user.get("age").unwrap_or(0)
  if age >= 18 { Ok(user) }
  else { Err("user is under 18") }
}

fn format_greeting(user) {
  let name = user.get("name").unwrap_or("Unknown")
  Ok("Welcome, {name}!")
}

let greeting = fetch_user(1)
  .and_then(|u| check_age(u))
  .and_then(|u| format_greeting(u))
  .unwrap_or("Access denied")
// => "Welcome, Alice!"

let denied = fetch_user(-1)
  .and_then(|u| check_age(u))
  .and_then(|u| format_greeting(u))
  .unwrap_or("Access denied")
// => "Access denied"
```

### Functional Data Processing Pipeline
```aura
// Process a list of user records
let users = [
  {"name": "Alice", "age": 30, "active": true},
  {"name": "Bob", "age": 17, "active": true},
  {"name": "Carol", "age": 25, "active": false},
  {"name": "Dave", "age": 45, "active": true},
  {"name": "Eve", "age": 22, "active": true}
]

// Find active adults, sorted by name
let active_adults = users
  .filter(|u| u.get("active").unwrap_or(false))
  .filter(|u| u.get("age").unwrap_or(0) >= 18)
  .map(|u| u.get("name").unwrap_or("Unknown"))
  .sort()
// => ["Alice", "Dave", "Eve"]

// Summary statistics
let ages = users.map(|u| u.get("age").unwrap_or(0))
let avg_age = ages.sum() / ages.len()           // => 27
let youngest = ages.min().unwrap_or(0)           // => 17
let oldest = ages.max().unwrap_or(0)             // => 45
```

### Building a Simple Key-Value Store
```aura
let store = {}

// Set values
store.set("users:1", {"name": "Alice"})
store.set("users:2", {"name": "Bob"})
store.set("config:theme", "dark")

// Query by prefix
let user_keys = store.keys()
  .filter(|k| k.starts_with("users:"))
// => ["users:1", "users:2"]

// Count entries by category
let user_count = store.count(|k, v| k.starts_with("users:"))
let config_count = store.count(|k, v| k.starts_with("config:"))
// user_count => 2, config_count => 1
```

### Option/Result Round-Trip
```aura
// Start with Option, convert to Result, process, convert back
let maybe_value = Some(42)

let processed = maybe_value
  .to_result("no value")           // => Ok(42)
  .map(|x| x * 2)                 // => Ok(84)
  .and_then(|x| {
    if x > 100 { Err("too large") }
    else { Ok(x) }
  })                               // => Ok(84)
  .ok()                            // => Some(84)
  .unwrap_or(0)                    // => 84
```

### Nested Flatten Operations
```aura
// Flatten nested Options
let nested = Some(Some(Some(42)))
let flat = nested.flatten()         // => Some(Some(42))
let flatter = nested.flatten().flatten()  // => Some(42)

// Flatten nested Results
let nested_result = Ok(Ok(42))
let flat_result = nested_result.flatten()  // => Ok(42)
```

### Combining Multiple Safe Lookups
```aura
let data = {
  "server": {
    "host": "example.com",
    "port": 443
  },
  "auth": {
    "token": "abc123"
  }
}

// Safely extract nested values and combine them
let host = data.get("server").and_then(|s| s.get("host"))
let port = data.get("server").and_then(|s| s.get("port"))

let url = host.zip(port)
  .map(|pair| "https://{pair.get(0).unwrap()}:{pair.get(1).unwrap()}")
  .unwrap_or("https://localhost:8080")
// => "https://example.com:443"
```

---

## Quick Reference: Method Counts by Type

| Type | Methods | Key Highlights |
|------|---------|----------------|
| **String** | 22 | Unicode-aware, immutable, pad/slice/reverse |
| **List** | 27 | map/filter/reduce, sort, zip, flat_map |
| **Map** | 24 | Ordered, safe get → Option, higher-order ops |
| **Option** | 17 | Monadic chaining, safe defaults, combinators |
| **Result** | 18 | Railway-oriented, map/map_err, and_then chains |
| **Total** | **108+** | |

> For the full method reference with signatures and parameters, see [method_reference.md](./method_reference.md).
> For the development roadmap, see [ROADMAP.md](../ROADMAP.md).

---

*Built for Aura v0.4.0 — An AI-first programming language designed for vibe coding.*
