# Aura Method Reference

> **Version:** 0.8.1 (Phase 3.1.1 — Tuple Literal Syntax COMPLETE)
> **Total Methods:** 120+ built-in methods + 117 stdlib functions  
> **Types Covered:** String (22) · List (27) · Map (24) · Tuple (12) · Option (17) · Result (18)  
> **Stdlib Modules:** math · string · io · testing · json · regex · collections · random · format · result · option · iter · file · time · env · net · log  
> **Tests:** 905 total across all packages

---

## Table of Contents

- [String Methods](#string-methods)
  - [Size & Emptiness](#string--size--emptiness)
  - [Case Conversion](#string--case-conversion)
  - [Search & Test](#string--search--test)
  - [Manipulation](#string--manipulation)
  - [Splitting & Joining](#string--splitting--joining)
  - [Slicing & Indexing](#string--slicing--indexing)
  - [Padding](#string--padding)
- [List Methods](#list-methods)
  - [Size & Emptiness](#list--size--emptiness)
  - [Element Access](#list--element-access)
  - [Mutation](#list--mutation)
  - [Slicing & Reordering](#list--slicing--reordering)
  - [Joining & Searching](#list--joining--searching)
  - [Higher-Order Functions](#list--higher-order-functions)
  - [Predicates & Counting](#list--predicates--counting)
  - [Aggregation](#list--aggregation)
  - [Pairing](#list--pairing)
- [Map Methods](#map-methods)
  - [Size & Emptiness](#map--size--emptiness)
  - [Key / Value / Entry Access](#map--key--value--entry-access)
  - [Lookup](#map--lookup)
  - [Mutation](#map--mutation)
  - [Higher-Order Functions](#map--higher-order-functions)
  - [Utility](#map--utility)
- [Tuple Methods](#tuple-methods)
  - [Creation](#tuple--creation)
  - [Size & Emptiness](#tuple--size--emptiness)
  - [Access](#tuple--access)
  - [Search](#tuple--search)
  - [Transformation](#tuple--transformation)
  - [Iteration](#tuple--iteration)
  - [Destructuring](#tuple--destructuring)
- [Option Methods](#option-methods)
  - [Predicates](#option--predicates)
  - [Extraction](#option--extraction)
  - [Transformation](#option--transformation)
  - [Combinators](#option--combinators)
  - [Querying & Conversion](#option--querying--conversion)
- [Result Methods](#result-methods)
  - [Predicates](#result--predicates)
  - [Extraction](#result--extraction)
  - [Transformation](#result--transformation)
  - [Combinators](#result--combinators)
  - [Querying & Conversion](#result--querying--conversion)
- [Best Practices](#best-practices)
- [Common Patterns](#common-patterns)

---

## String Methods

Aura strings are immutable and Unicode-aware. All methods that produce a new string return a **new** `String` — the original is never modified.

### String · Size & Emptiness

#### `len() -> Int`
Returns the number of Unicode code points (runes) in the string.

```aura
"hello".len()       // => 5
"café".len()        // => 4  (Unicode-aware)
"".len()            // => 0
```

#### `length() -> Int`
Alias for `len()`.

```aura
"hello".length()    // => 5
```

#### `is_empty() -> Bool`
Returns `true` if the string has zero length.

```aura
"".is_empty()       // => true
"hi".is_empty()     // => false
```

---

### String · Case Conversion

#### `upper() -> String`
Returns a new string with all characters converted to uppercase.

```aura
"hello".upper()     // => "HELLO"
```

#### `to_upper() -> String`
Alias for `upper()`.

#### `lower() -> String`
Returns a new string with all characters converted to lowercase.

```aura
"HELLO".lower()     // => "hello"
```

#### `to_lower() -> String`
Alias for `lower()`.

---

### String · Search & Test

#### `contains(sub: String) -> Bool`
Returns `true` if the string contains the given substring.

```aura
"hello world".contains("world")   // => true
"hello world".contains("xyz")     // => false
```

#### `starts_with(prefix: String) -> Bool`
Returns `true` if the string begins with the given prefix.

```aura
"hello".starts_with("he")    // => true
"hello".starts_with("lo")    // => false
```

#### `ends_with(suffix: String) -> Bool`
Returns `true` if the string ends with the given suffix.

```aura
"hello".ends_with("lo")      // => true
"hello".ends_with("he")      // => false
```

#### `index_of(sub: String) -> Option[Int]`
Returns `Some(index)` of the first occurrence of `sub`, or `None` if not found. The index is rune-based (Unicode-aware).

```aura
"hello".index_of("ll")       // => Some(2)
"hello".index_of("xyz")      // => None
```

---

### String · Manipulation

#### `trim() -> String`
Removes leading and trailing whitespace.

```aura
"  hello  ".trim()            // => "hello"
```

#### `trim_left() -> String`
Removes leading whitespace (spaces, tabs, newlines, carriage returns).

```aura
"  hello  ".trim_left()       // => "hello  "
```

#### `trim_right() -> String`
Removes trailing whitespace.

```aura
"  hello  ".trim_right()      // => "  hello"
```

#### `replace(old: String, new: String) -> String`
Replaces **all** occurrences of `old` with `new`.

```aura
"aabaa".replace("a", "x")    // => "xxbxx"
```

#### `replace_first(old: String, new: String) -> String`
Replaces only the **first** occurrence of `old` with `new`.

```aura
"aabaa".replace_first("a", "x")  // => "xabaa"
```

#### `repeat(n: Int) -> String`
Repeats the string `n` times. Panics if `n` is negative.

```aura
"ab".repeat(3)                // => "ababab"
"hi".repeat(0)                // => ""
```

#### `reverse() -> String`
Returns a new string with runes in reverse order (Unicode-aware).

```aura
"hello".reverse()             // => "olleh"
"café".reverse()              // => "éfac"
```

#### `chars() -> List[String]`
Splits the string into a list of individual characters (each as a single-character string).

```aura
"hi!".chars()                 // => ["h", "i", "!"]
```

---

### String · Splitting & Joining

#### `split(sep: String = " ") -> List[String]`
Splits the string by the given separator. Defaults to splitting on spaces.

```aura
"a,b,c".split(",")           // => ["a", "b", "c"]
"hello world".split()         // => ["hello", "world"]
```

#### `join(list: List) -> String`
Uses the receiver string as a separator to join elements of the given list.

```aura
", ".join(["a", "b", "c"])    // => "a, b, c"
"-".join([1, 2, 3])           // => "1-2-3"
```

---

### String · Slicing & Indexing

#### `slice(start: Int, end?: Int) -> String`
Extracts a substring from `start` (inclusive) to `end` (exclusive). Supports negative indices (counting from end). If `end` is omitted, slices to the end of the string.

| Parameter | Description |
|-----------|-------------|
| `start`   | Start index (0-based, supports negative) |
| `end`     | End index (exclusive, optional, supports negative) |

```aura
"hello".slice(1, 4)          // => "ell"
"hello".slice(2)             // => "llo"
"hello".slice(-3)            // => "llo"
"hello".slice(0, -1)         // => "hell"
```

---

### String · Padding

#### `pad_left(n: Int, char?: String = " ") -> String`
Left-pads the string to length `n` with the given character. If the string is already longer than `n`, returns it unchanged.

```aura
"42".pad_left(5)              // => "   42"
"42".pad_left(5, "0")         // => "00042"
```

#### `pad_right(n: Int, char?: String = " ") -> String`
Right-pads the string to length `n` with the given character.

```aura
"hi".pad_right(5)             // => "hi   "
"hi".pad_right(5, ".")        // => "hi..."
```

---

## List Methods

Lists in Aura are ordered, mutable collections. Methods that produce new lists leave the original unchanged unless explicitly noted as **mutating**.

### List · Size & Emptiness

#### `len() -> Int`
Returns the number of elements.

```aura
[1, 2, 3].len()              // => 3
```

#### `length() -> Int`
Alias for `len()`.

#### `is_empty() -> Bool`
Returns `true` if the list has zero elements.

```aura
[].is_empty()                 // => true
[1].is_empty()                // => false
```

---

### List · Element Access

#### `first() -> Option[T]`
Returns `Some(element)` for the first element, or `None` if empty.

```aura
[10, 20, 30].first()          // => Some(10)
[].first()                    // => None
```

#### `last() -> Option[T]`
Returns `Some(element)` for the last element, or `None` if empty.

```aura
[10, 20, 30].last()           // => Some(30)
[].last()                     // => None
```

#### `get(index: Int) -> Option[T]`
Safe index access. Returns `Some(element)` or `None` if out of bounds. Supports negative indices.

```aura
[10, 20, 30].get(1)           // => Some(20)
[10, 20, 30].get(-1)          // => Some(30)
[10, 20, 30].get(99)          // => None
```

---

### List · Mutation

> ⚠️ These methods **modify the list in place**.

#### `append(item: T) -> None`
Appends an item to the end of the list.

```aura
let items = [1, 2]
items.append(3)               // items is now [1, 2, 3]
```

#### `push(item: T) -> None`
Alias for `append()`.

#### `pop() -> Option[T]`
Removes and returns the last element, or `None` if the list is empty.

```aura
let items = [1, 2, 3]
items.pop()                   // => Some(3), items is now [1, 2]
[].pop()                      // => None
```

#### `remove(index: Int) -> T`
Removes and returns the element at the given index. Supports negative indices. Panics if index is out of bounds.

```aura
let items = [10, 20, 30]
items.remove(1)               // => 20, items is now [10, 30]
```

---

### List · Slicing & Reordering

#### `reverse() -> List`
Returns a **new** reversed list. Does not mutate the original.

```aura
[1, 2, 3].reverse()           // => [3, 2, 1]
```

#### `slice(start: Int, end?: Int) -> List`
Extracts a sublist. Supports negative indices. Returns a new list.

```aura
[10, 20, 30, 40].slice(1, 3)  // => [20, 30]
[10, 20, 30, 40].slice(-2)    // => [30, 40]
```

#### `sort() -> List`
Returns a **new** sorted list. Supports `Int`, `Float`, and `String` comparisons.

```aura
[3, 1, 2].sort()              // => [1, 2, 3]
["c", "a", "b"].sort()        // => ["a", "b", "c"]
```

#### `unique() -> List`
Returns a **new** list with duplicates removed (preserves first-occurrence order).

```aura
[1, 2, 2, 3, 1].unique()     // => [1, 2, 3]
```

---

### List · Joining & Searching

#### `join(separator: String = "") -> String`
Joins all elements into a string separated by the given separator.

```aura
["a", "b", "c"].join(", ")    // => "a, b, c"
[1, 2, 3].join("-")           // => "1-2-3"
```

#### `contains(item: T) -> Bool`
Returns `true` if the list contains the given item (by structural equality).

```aura
[1, 2, 3].contains(2)         // => true
[1, 2, 3].contains(4)         // => false
```

#### `index_of(item: T) -> Option[Int]`
Returns `Some(index)` of the first matching element, or `None`.

```aura
[10, 20, 30].index_of(20)     // => Some(1)
[10, 20, 30].index_of(99)     // => None
```

---

### List · Higher-Order Functions

#### `map(fn: (T) -> U) -> List[U]`
Applies `fn` to each element, returning a new list of results.

```aura
[1, 2, 3].map(|x| x * 2)     // => [2, 4, 6]
```

#### `filter(fn: (T) -> Bool) -> List[T]`
Returns a new list containing only elements where `fn` returns truthy.

```aura
[1, 2, 3, 4].filter(|x| x > 2)   // => [3, 4]
```

#### `reduce(init: U, fn: (U, T) -> U) -> U`
Folds the list into a single value using an accumulator.

```aura
[1, 2, 3].reduce(0, |acc, x| acc + x)   // => 6
```

#### `for_each(fn: (T) -> None) -> None`
Executes `fn` for each element (for side effects). Returns `None`.

```aura
[1, 2, 3].for_each(|x| print(x))
```

#### `flat_map(fn: (T) -> List[U]) -> List[U]`
Maps each element to a list, then flattens one level.

```aura
[1, 2, 3].flat_map(|x| [x, x * 10])
// => [1, 10, 2, 20, 3, 30]
```

#### `flatten() -> List`
Flattens one level of nesting. Non-list elements are preserved as-is.

```aura
[[1, 2], [3, 4], [5]].flatten()   // => [1, 2, 3, 4, 5]
[1, [2, 3], 4].flatten()          // => [1, 2, 3, 4]
```

---

### List · Predicates & Counting

#### `any(fn: (T) -> Bool) -> Bool`
Returns `true` if **any** element satisfies the predicate.

```aura
[1, 2, 3].any(|x| x > 2)         // => true
[1, 2, 3].any(|x| x > 5)         // => false
```

#### `all(fn: (T) -> Bool) -> Bool`
Returns `true` if **all** elements satisfy the predicate.

```aura
[2, 4, 6].all(|x| x % 2 == 0)    // => true
[2, 3, 6].all(|x| x % 2 == 0)    // => false
```

#### `count(fn?: (T) -> Bool) -> Int`
Without arguments: returns the list length. With a predicate: counts matching elements.

```aura
[1, 2, 3, 4].count()              // => 4
[1, 2, 3, 4].count(|x| x > 2)    // => 2
```

---

### List · Aggregation

#### `sum() -> Int | Float`
Sums all numeric elements. Returns `Int` if all elements are `Int`, `Float` if any is `Float`. Panics on non-numeric elements.

```aura
[1, 2, 3].sum()                   // => 6
[1.5, 2.5].sum()                  // => 4.0
```

#### `min() -> Option[T]`
Returns `Some(minimum)` or `None` for empty lists. Supports `Int`, `Float`, and `String`.

```aura
[3, 1, 2].min()                   // => Some(1)
[].min()                          // => None
```

#### `max() -> Option[T]`
Returns `Some(maximum)` or `None` for empty lists.

```aura
[3, 1, 2].max()                   // => Some(3)
[].max()                          // => None
```

---

### List · Pairing

#### `zip(other: List) -> List[Tuple[T, U]]`
Pairs elements from two lists into tuples. Stops at the shorter list.

```aura
[1, 2, 3].zip(["a", "b", "c"])
// => [(1, "a"), (2, "b"), (3, "c")]

[1, 2].zip(["a", "b", "c"])
// => [(1, "a"), (2, "b")]
```

#### `enumerate() -> List[Tuple[Int, T]]`
Returns a list of `(index, element)` tuples.

```aura
["a", "b", "c"].enumerate()
// => [(0, "a"), (1, "b"), (2, "c")]
```

---

## Map Methods

Maps in Aura are ordered key-value collections (insertion order is preserved). Keys can be any value type. Methods marked as **mutating** modify the map in place.

### Map · Size & Emptiness

#### `len() -> Int`
Returns the number of key-value pairs.

```aura
{"a": 1, "b": 2}.len()        // => 2
```

#### `length() -> Int`
Alias for `len()`.

#### `size() -> Int`
Alias for `len()`.

#### `is_empty() -> Bool`
Returns `true` if the map has no entries.

```aura
{}.is_empty()                  // => true
{"a": 1}.is_empty()            // => false
```

---

### Map · Key / Value / Entry Access

#### `keys() -> List`
Returns a list of all keys (in insertion order).

```aura
{"a": 1, "b": 2}.keys()       // => ["a", "b"]
```

#### `values() -> List`
Returns a list of all values (in insertion order).

```aura
{"a": 1, "b": 2}.values()     // => [1, 2]
```

#### `entries() -> List[List[Key, Value]]`
Returns a list of `[key, value]` pairs.

```aura
{"a": 1, "b": 2}.entries()    // => [["a", 1], ["b", 2]]
```

---

### Map · Lookup

#### `has(key: K) -> Bool`
Returns `true` if the key exists in the map.

```aura
{"a": 1}.has("a")              // => true
{"a": 1}.has("b")              // => false
```

#### `contains_key(key: K) -> Bool`
Alias for `has()`.

#### `contains_value(value: V) -> Bool`
Returns `true` if the value exists in the map (by structural equality).

```aura
{"a": 1, "b": 2}.contains_value(2)   // => true
{"a": 1, "b": 2}.contains_value(3)   // => false
```

#### `get(key: K) -> Option[V]`
Safe key access. Returns `Some(value)` or `None`.

```aura
{"a": 1}.get("a")             // => Some(1)
{"a": 1}.get("z")             // => None
```

#### `get_or(key: K, default: V) -> V`
Returns the value for the key, or the default if the key doesn't exist.

```aura
{"a": 1}.get_or("a", 0)       // => 1
{"a": 1}.get_or("z", 0)       // => 0
```

---

### Map · Mutation

> ⚠️ These methods **modify the map in place**.

#### `set(key: K, value: V) -> None`
Adds a new key-value pair or updates an existing key.

```aura
let m = {"a": 1}
m.set("b", 2)                 // m is now {"a": 1, "b": 2}
m.set("a", 10)                // m is now {"a": 10, "b": 2}
```

#### `remove(key: K) -> Option[V]`
Removes an entry and returns `Some(removed_value)` or `None` if key didn't exist.

```aura
let m = {"a": 1, "b": 2}
m.remove("a")                 // => Some(1), m is now {"b": 2}
m.remove("z")                 // => None
```

#### `delete(key: K) -> Bool`
Removes an entry and returns `true` if the key existed, `false` otherwise.

```aura
let m = {"a": 1, "b": 2}
m.delete("a")                 // => true, m is now {"b": 2}
m.delete("z")                 // => false
```

#### `clear() -> None`
Removes all entries from the map.

```aura
let m = {"a": 1, "b": 2}
m.clear()                     // m is now {}
```

#### `merge(other: Map) -> None`
Merges another map into this one. Keys from `other` overwrite existing keys.

```aura
let m = {"a": 1, "b": 2}
m.merge({"b": 20, "c": 30})   // m is now {"a": 1, "b": 20, "c": 30}
```

---

### Map · Higher-Order Functions

All higher-order map methods receive **both key and value** as arguments to the callback.

#### `filter(fn: (K, V) -> Bool) -> Map`
Returns a **new** map with only entries where the predicate is true.

```aura
{"a": 1, "b": 2, "c": 3}.filter(|k, v| v > 1)
// => {"b": 2, "c": 3}
```

#### `map(fn: (K, V) -> NewV) -> Map`
Returns a **new** map with the same keys but transformed values.

```aura
{"a": 1, "b": 2}.map(|k, v| v * 10)
// => {"a": 10, "b": 20}
```

#### `for_each(fn: (K, V) -> None) -> None`
Executes `fn` for each entry (for side effects).

```aura
{"a": 1}.for_each(|k, v| print("{k}: {v}"))
```

#### `reduce(init: U, fn: (U, K, V) -> U) -> U`
Reduces all entries into a single value. The callback receives `(accumulator, key, value)`.

```aura
{"a": 1, "b": 2}.reduce(0, |acc, k, v| acc + v)   // => 3
```

#### `any(fn: (K, V) -> Bool) -> Bool`
Returns `true` if any entry satisfies the predicate.

```aura
{"a": 1, "b": 5}.any(|k, v| v > 3)    // => true
```

#### `all(fn: (K, V) -> Bool) -> Bool`
Returns `true` if all entries satisfy the predicate.

```aura
{"a": 2, "b": 4}.all(|k, v| v % 2 == 0)   // => true
```

#### `count(fn?: (K, V) -> Bool) -> Int`
Without arguments: returns the number of entries. With predicate: counts matching entries.

```aura
{"a": 1, "b": 2, "c": 3}.count(|k, v| v > 1)   // => 2
```

---

### Map · Utility

#### `to_list() -> List[List[K, V]]`
Alias for `entries()`. Converts the map to a list of `[key, value]` pairs.

#### `find(fn: (K, V) -> Bool) -> Option[List[K, V]]`
Returns `Some([key, value])` for the first matching entry, or `None`.

```aura
{"a": 1, "b": 5, "c": 3}.find(|k, v| v > 4)
// => Some(["b", 5])

{"a": 1}.find(|k, v| v > 10)
// => None
```

---

## Tuple Methods

Tuples are immutable, ordered collections of values. They support mixed types and are ideal for returning multiple values from functions.

### Tuple — Creation

```aura
# Empty tuple
let empty = ()

# Single-element tuple (trailing comma required)
let single = (42,)

# Multi-element tuple
let point = (10, 20)
let record = ("Alice", 30, true)

# Nested tuples
let nested = ((1, 2), (3, 4))
```

### Tuple — Size & Emptiness

| Method | Returns | Description |
|--------|---------|-------------|
| `len()` | `Int` | Number of elements |
| `length()` | `Int` | Alias for `len()` |
| `is_empty()` | `Bool` | True if tuple has no elements |

```aura
let t = (1, 2, 3)
t.len()       # 3
t.is_empty()  # false
().is_empty()  # true
```

### Tuple — Access

| Method | Returns | Description |
|--------|---------|-------------|
| `t[i]` | `Value` | Direct index access (panics on out of bounds) |
| `get(index)` | `Option[Value]` | Safe index access |
| `first()` | `Option[Value]` | First element |
| `last()` | `Option[Value]` | Last element |

```aura
let t = (10, 20, 30)
t[0]           # 10
t.get(1)       # Some(20)
t.get(99)      # None
t.first()      # Some(10)
t.last()       # Some(30)
```

### Tuple — Search

| Method | Returns | Description |
|--------|---------|-------------|
| `contains(value)` | `Bool` | True if tuple contains the value |

```aura
let t = (1, 2, 3)
t.contains(2)   # true
t.contains(99)  # false
```

### Tuple — Transformation

| Method | Returns | Description |
|--------|---------|-------------|
| `to_list()` | `List` | Convert tuple to list |
| `reverse()` | `Tuple` | New tuple with reversed elements |
| `map(fn)` | `Tuple` | Apply function to each element |
| `zip(other)` | `List[Tuple]` | Pair elements with another tuple/list |
| `enumerate()` | `List[Tuple]` | List of (index, value) pairs |

```aura
let t = (1, 2, 3)
t.to_list()           # [1, 2, 3]
t.reverse()           # (3, 2, 1)
t.map(|x| -> x * 2)  # (2, 4, 6)
t.zip((10, 20, 30))   # [(1, 10), (2, 20), (3, 30)]
t.enumerate()          # [(0, 1), (1, 2), (2, 3)]
```

### Tuple — Iteration

Tuples are iterable and can be used in for loops:

```aura
let t = (1, 2, 3)
for x in t:
    print(x)

# With for_each method
t.for_each(|x| -> print(x))
```

### Tuple — Destructuring

Tuples support destructuring in `let` bindings:

```aura
# Basic destructuring
let (x, y) = (10, 20)
# x = 10, y = 20

# Three elements
let (name, age, active) = ("Alice", 30, true)

# With wildcard (ignore elements)
let (first, _, last) = (1, 2, 3)
# first = 1, last = 3

# Mutable destructuring
let mut (a, b) = (1, 2)
a = 100

# Destructuring from lists
let (x, y) = [5, 10]

# Function return value destructuring
fn divide(a: Int, b: Int):
    return (a / b, a % b)

let (quotient, remainder) = divide(17, 5)
# quotient = 3, remainder = 2
```

---

## Option Methods

`Option` represents a value that may or may not exist: `Some(value)` or `None`. Options are created by many methods (e.g., `List.first()`, `Map.get()`) and are central to Aura's safe-by-default philosophy.

### Option · Predicates

#### `is_some() -> Bool`
Returns `true` if the Option contains a value.

```aura
Some(42).is_some()             // => true
None.is_some()                 // => false
```

#### `is_none() -> Bool`
Returns `true` if the Option is `None`.

```aura
None.is_none()                 // => true
Some(42).is_none()             // => false
```

---

### Option · Extraction

#### `unwrap() -> T`
Returns the contained value. **Panics** if `None`.

```aura
Some(42).unwrap()              // => 42
// None.unwrap()               // PANIC: "called unwrap() on a None value"
```

#### `expect(msg: String) -> T`
Returns the contained value. **Panics** with a custom message if `None`.

```aura
Some(42).expect("missing!")    // => 42
// None.expect("missing!")     // PANIC: "missing!"
```

#### `unwrap_or(default: T) -> T`
Returns the contained value, or the default if `None`.

```aura
Some(42).unwrap_or(0)          // => 42
None.unwrap_or(0)              // => 0
```

#### `unwrap_or_else(fn: () -> T) -> T`
Returns the contained value, or calls `fn` to compute a default if `None`.

```aura
None.unwrap_or_else(|| compute_default())
```

---

### Option · Transformation

#### `map(fn: (T) -> U) -> Option[U]`
Applies `fn` to the contained value if `Some`, wrapping the result in `Some`. Returns `None` if `None`.

```aura
Some(5).map(|x| x * 2)        // => Some(10)
None.map(|x| x * 2)           // => None
```

#### `flat_map(fn: (T) -> Option[U]) -> Option[U]`
Monadic bind. Applies `fn` (which must return an `Option`) to the contained value. The callback **must** return an `Option`.

```aura
Some(5).flat_map(|x| Some(x * 2))   // => Some(10)
Some(5).flat_map(|x| None)          // => None
None.flat_map(|x| Some(x * 2))     // => None
```

#### `and_then(fn: (T) -> Option[U]) -> Option[U]`
Alias for `flat_map()`. The callback **must** return an `Option`.

#### `filter(fn: (T) -> Bool) -> Option[T]`
Returns `Some(value)` if the predicate returns `true`, otherwise `None`.

```aura
Some(4).filter(|x| x > 3)     // => Some(4)
Some(2).filter(|x| x > 3)     // => None
None.filter(|x| x > 3)        // => None
```

#### `flatten() -> Option[T]`
Unwraps one level of nesting. `Some(Some(x))` becomes `Some(x)`. Non-nested Options are returned as-is.

```aura
Some(Some(42)).flatten()       // => Some(42)
Some(None).flatten()           // => None
None.flatten()                 // => None
Some(42).flatten()             // => Some(42) — no-op for non-nested
```

---

### Option · Combinators

#### `or(alternative: Option[T]) -> Option[T]`
Returns `self` if `Some`, otherwise returns `alternative`.

```aura
Some(1).or(Some(2))            // => Some(1)
None.or(Some(2))               // => Some(2)
```

#### `or_else(fn: () -> Option[T]) -> Option[T]`
Returns `self` if `Some`, otherwise calls `fn` to produce an alternative. The callback **must** return an `Option`.

```aura
None.or_else(|| Some(99))      // => Some(99)
Some(1).or_else(|| Some(99))   // => Some(1)
```

#### `and(other: Option[U]) -> Option[U]`
Returns `other` if `self` is `Some`, otherwise `None`.

```aura
Some(1).and(Some("a"))         // => Some("a")
None.and(Some("a"))            // => None
```

#### `zip(other: Option[U]) -> Option[List[T, U]]`
Combines two Options into `Some([a, b])` if both are `Some`, otherwise `None`.

```aura
Some(1).zip(Some("a"))         // => Some([1, "a"])
Some(1).zip(None)              // => None
None.zip(Some("a"))            // => None
```

---

### Option · Querying & Conversion

#### `contains(value: T) -> Bool`
Returns `true` if the Option is `Some` and the contained value equals `value`.

```aura
Some(42).contains(42)          // => true
Some(42).contains(99)          // => false
None.contains(42)              // => false
```

#### `to_result(err: E) -> Result[T, E]`
Converts `Some(value)` to `Ok(value)` and `None` to `Err(err)`.

```aura
Some(42).to_result("not found")    // => Ok(42)
None.to_result("not found")       // => Err("not found")
```

---

## Result Methods

`Result` represents an operation that can succeed (`Ok(value)`) or fail (`Err(error)`). Results enable railway-oriented programming — chaining operations that might fail.

### Result · Predicates

#### `is_ok() -> Bool`
Returns `true` if the Result is `Ok`.

```aura
Ok(42).is_ok()                 // => true
Err("fail").is_ok()            // => false
```

#### `is_err() -> Bool`
Returns `true` if the Result is `Err`.

```aura
Err("fail").is_err()           // => true
Ok(42).is_err()                // => false
```

---

### Result · Extraction

#### `unwrap() -> T`
Returns the `Ok` value. **Panics** if `Err`, showing the error.

```aura
Ok(42).unwrap()                // => 42
// Err("fail").unwrap()        // PANIC: "called unwrap() on an Err value: fail"
```

#### `unwrap_err() -> E`
Returns the `Err` value. **Panics** if `Ok`.

```aura
Err("fail").unwrap_err()       // => "fail"
// Ok(42).unwrap_err()         // PANIC: "called unwrap_err() on an Ok value: 42"
```

#### `expect(msg: String) -> T`
Returns the `Ok` value. **Panics** with a custom message if `Err`.

```aura
Ok(42).expect("failed!")       // => 42
// Err("x").expect("failed!")  // PANIC: "failed!"
```

#### `unwrap_or(default: T) -> T`
Returns the `Ok` value, or the default if `Err`.

```aura
Ok(42).unwrap_or(0)            // => 42
Err("fail").unwrap_or(0)       // => 0
```

#### `unwrap_or_else(fn: (E) -> T) -> T`
Returns the `Ok` value, or calls `fn` with the error to compute a default.

```aura
Err("fail").unwrap_or_else(|e| 0)   // => 0
```

---

### Result · Transformation

#### `map(fn: (T) -> U) -> Result[U, E]`
Transforms the `Ok` value, leaving `Err` untouched.

```aura
Ok(5).map(|x| x * 2)          // => Ok(10)
Err("fail").map(|x| x * 2)    // => Err("fail")
```

#### `map_err(fn: (E) -> F) -> Result[T, F]`
Transforms the `Err` value, leaving `Ok` untouched.

```aura
Err("fail").map_err(|e| "ERROR: " + e)   // => Err("ERROR: fail")
Ok(42).map_err(|e| "ERROR: " + e)        // => Ok(42)
```

#### `and_then(fn: (T) -> Result[U, E]) -> Result[U, E]`
Monadic bind. Chains fallible operations. The callback **must** return a `Result`.

```aura
Ok(5).and_then(|x| Ok(x * 2))        // => Ok(10)
Ok(5).and_then(|x| Err("nope"))      // => Err("nope")
Err("fail").and_then(|x| Ok(x * 2))  // => Err("fail")
```

#### `or_else(fn: (E) -> Result[T, F]) -> Result[T, F]`
If `Err`, calls `fn` with the error to try an alternative. The callback **must** return a `Result`.

```aura
Err("fail").or_else(|e| Ok(0))        // => Ok(0)
Ok(42).or_else(|e| Ok(0))            // => Ok(42)
```

#### `flatten() -> Result[T, E]`
Unwraps one level of nesting. `Ok(Ok(x))` becomes `Ok(x)`. Non-nested Results are returned as-is.

```aura
Ok(Ok(42)).flatten()           // => Ok(42)
Ok(Err("inner")).flatten()     // => Err("inner")
Err("outer").flatten()         // => Err("outer")
```

---

### Result · Combinators

#### `or(alternative: Result[T, F]) -> Result[T, F]`
Returns `self` if `Ok`, otherwise returns `alternative`.

```aura
Ok(1).or(Ok(2))                // => Ok(1)
Err("fail").or(Ok(2))          // => Ok(2)
```

#### `and(other: Result[U, E]) -> Result[U, E]`
Returns `other` if `self` is `Ok`, otherwise returns `self` (the `Err`).

```aura
Ok(1).and(Ok("a"))             // => Ok("a")
Err("fail").and(Ok("a"))       // => Err("fail")
```

---

### Result · Querying & Conversion

#### `contains(value: T) -> Bool`
Returns `true` if the Result is `Ok` and the contained value equals `value`.

```aura
Ok(42).contains(42)            // => true
Ok(42).contains(99)            // => false
Err("fail").contains(42)       // => false
```

#### `contains_err(value: E) -> Bool`
Returns `true` if the Result is `Err` and the error equals `value`.

```aura
Err("fail").contains_err("fail")     // => true
Err("fail").contains_err("other")    // => false
Ok(42).contains_err("fail")          // => false
```

#### `ok() -> Option[T]`
Converts `Ok(value)` to `Some(value)`, `Err` to `None`.

```aura
Ok(42).ok()                    // => Some(42)
Err("fail").ok()               // => None
```

#### `err() -> Option[E]`
Converts `Err(error)` to `Some(error)`, `Ok` to `None`.

```aura
Err("fail").err()              // => Some("fail")
Ok(42).err()                   // => None
```

#### `to_option() -> Option[T]`
Alias for `ok()`. Converts `Ok(value)` to `Some(value)`, `Err` to `None`.

---

## Best Practices

### 1. Prefer Safe Access Over Direct Indexing
Use methods that return `Option` instead of panicking on missing data:

```aura
// ✅ Good — safe access
let name = users.get(0).unwrap_or("anonymous")

// ❌ Avoid — panics if list is empty
let name = users[0]
```

### 2. Use Method Chaining for Readable Data Pipelines
Chain transformations to express intent clearly:

```aura
let result = items
  .filter(|x| x.is_active)
  .map(|x| x.name)
  .sort()
  .join(", ")
```

### 3. Handle Errors with `Result` Instead of Panicking
Use `and_then` chains for railway-oriented programming:

```aura
let output = parse_input(raw)
  .and_then(|data| validate(data))
  .and_then(|data| process(data))
  .unwrap_or(default_value)
```

### 4. Use `unwrap_or` / `unwrap_or_else` Over Bare `unwrap`
Bare `unwrap()` panics on `None` / `Err`. Prefer providing defaults:

```aura
// ✅ Good
let port = config.get("port").unwrap_or(8080)

// ⚠️ Use only when you're certain the value exists
let port = config.get("port").unwrap()
```

### 5. Mutating vs. Non-Mutating Methods
Be aware which methods mutate:

| Mutating | Non-Mutating |
|----------|-------------|
| `List.append()`, `push()`, `pop()`, `remove()` | `List.map()`, `filter()`, `sort()`, `reverse()` |
| `Map.set()`, `remove()`, `delete()`, `clear()`, `merge()` | `Map.map()`, `filter()`, `find()` |

### 6. Convert Between Option and Result
Use `to_result()` and `to_option()` / `ok()` / `err()` to bridge between the two types:

```aura
// Option → Result
let result = maybe_value.to_result("value was missing")

// Result → Option
let maybe = some_result.ok()
```

---

## Common Patterns

### Pattern: Safe Division
```aura
fn safe_div(a, b) {
  if b == 0 {
    None
  } else {
    Some(a / b)
  }
}

let result = safe_div(10, 3).unwrap_or(0)
```

### Pattern: Transform-and-Collect
```aura
let names = users
  .filter(|u| u.age >= 18)
  .map(|u| u.name.upper())
  .sort()
  .unique()
```

### Pattern: Map Inversion
```aura
let inverted = original
  .entries()
  .reduce({}, |acc, entry| {
    acc.set(entry.get(1).unwrap(), entry.get(0).unwrap())
    acc
  })
```

### Pattern: Option Chaining (Monadic)
```aura
let city = user
  .get("address")           // Option[Map]
  .and_then(|a| a.get("city"))  // Option[String]
  .map(|c| c.upper())      // Option[String]
  .unwrap_or("UNKNOWN")    // String
```

### Pattern: Result Chaining (Railway-Oriented)
```aura
fn parse_and_double(input) {
  parse_int(input)
    .map(|n| n * 2)
    .map_err(|e| "Parse failed: " + e)
}

let value = parse_and_double("21")
  .unwrap_or(0)             // => 42
```

### Pattern: Aggregating with Reduce
```aura
// Word frequency counter
let words = "the cat sat on the mat".split(" ")
let freq = words.reduce({}, |acc, word| {
  let count = acc.get_or(word, 0)
  acc.set(word, count + 1)
  acc
})
// freq => {"the": 2, "cat": 1, "sat": 1, "on": 1, "mat": 1}
```

### Pattern: Zip for Parallel Iteration
```aura
let keys = ["name", "age", "city"]
let vals = ["Alice", 30, "Paris"]

let record = keys.zip(vals).reduce({}, |acc, pair| {
  acc.set(pair[0], pair[1])
  acc
})
// record => {"name": "Alice", "age": 30, "city": "Paris"}
```

---

> *This reference was generated for Aura v0.4.0. For language syntax, see the main [README](../README.md). For the development roadmap, see [ROADMAP.md](../ROADMAP.md).*



---

## Standard Library Modules (Phase 4.2)

The following modules are available via `import std.<module>`.

---

### std.regex

Pattern matching and regular expression operations using Go's regexp syntax.

```aura
import std.regex as re

re.match(pattern, text)           // Bool - test if pattern matches text
re.find(pattern, text)            // Option<String> - find first match
re.find_all(pattern, text)        // List<String> - find all matches
re.replace(pattern, text, repl)   // String - replace all matches
re.split(pattern, text)           // List<String> - split by pattern
re.compile(pattern)               // Result<String, String> - validate & compile pattern
```

| Function | Args | Returns | Description |
|----------|------|---------|-------------|
| `match` | `(pattern: String, text: String)` | `Bool` | Test if pattern matches anywhere in text |
| `find` | `(pattern: String, text: String)` | `Option<String>` | Find first match, None if not found |
| `find_all` | `(pattern: String, text: String)` | `List<String>` | Find all non-overlapping matches |
| `replace` | `(pattern: String, text: String, replacement: String)` | `String` | Replace all matches |
| `split` | `(pattern: String, text: String)` | `List<String>` | Split text by pattern |
| `compile` | `(pattern: String)` | `Result<String, String>` | Validate pattern, Ok(pattern) or Err(message) |

---

### std.collections

Collection utilities for advanced list manipulation.

```aura
import std.collections as col

col.range(5)                      // [0, 1, 2, 3, 4]
col.range(2, 5)                   // [2, 3, 4]
col.range(0, 10, 3)               // [0, 3, 6, 9]
col.zip_with(fn, list1, list2)    // Zip with custom combiner
col.partition(fn, list)           // [matching, non_matching]
col.group_by(fn, list)            // Map<Key, List>
col.chunk(2, [1,2,3,4,5])        // [[1,2], [3,4], [5]]
col.take(3, list)                 // First 3 elements
col.drop(3, list)                 // All but first 3
col.take_while(fn, list)          // Take while predicate true
col.drop_while(fn, list)          // Drop while predicate true
```

| Function | Args | Returns | Description |
|----------|------|---------|-------------|
| `range` | `(end)` or `(start, end)` or `(start, end, step)` | `List<Int>` | Generate number range |
| `zip_with` | `(fn, list1, list2)` | `List` | Combine two lists element-wise with fn |
| `partition` | `(fn, list)` | `List<List>` | Split into [true, false] by predicate |
| `group_by` | `(fn, list)` | `Map` | Group elements by key function |
| `chunk` | `(n: Int, list)` | `List<List>` | Split into chunks of size n |
| `take` | `(n: Int, list)` | `List` | First n elements |
| `drop` | `(n: Int, list)` | `List` | All but first n elements |
| `take_while` | `(fn, list)` | `List` | Take while predicate is true |
| `drop_while` | `(fn, list)` | `List` | Drop while predicate is true |

---

### std.random

Random number generation and sampling.

```aura
import std.random as rand

rand.seed(42)                     // Set seed for reproducibility
rand.int(1, 100)                  // Random int in [1, 100]
rand.float()                      // Random float in [0.0, 1.0)
rand.choice(list)                 // Random element from list
rand.shuffle(list)                // New shuffled list (non-mutating)
rand.sample(list, 3)              // 3 random elements without replacement
```

| Function | Args | Returns | Description |
|----------|------|---------|-------------|
| `int` | `(min: Int, max: Int)` | `Int` | Random integer in [min, max] |
| `float` | `()` | `Float` | Random float in [0.0, 1.0) |
| `choice` | `(list)` | `Value` | Random element from list |
| `shuffle` | `(list)` | `List` | Shuffled copy (original unchanged) |
| `sample` | `(list, n: Int)` | `List` | n random elements without replacement |
| `seed` | `(value: Int)` | `None` | Set random seed for reproducibility |

---

### std.format

Advanced string formatting utilities.

```aura
import std.format as fmt

fmt.pad_left("42", 5, "0")       // "00042"
fmt.pad_right("hi", 10)          // "hi        "
fmt.center("title", 20, "=")     // "=======title========"
fmt.truncate("long text", 7)     // "long..."
fmt.wrap("long text here", 10)   // Word-wrapped string
fmt.indent("code\nhere", 4)      // "    code\n    here"
fmt.dedent("  a\n  b")           // "a\nb"
```

| Function | Args | Returns | Description |
|----------|------|---------|-------------|
| `pad_left` | `(str, width, char?)` | `String` | Left-pad to width (default: space) |
| `pad_right` | `(str, width, char?)` | `String` | Right-pad to width (default: space) |
| `center` | `(str, width, char?)` | `String` | Center in width (default: space) |
| `truncate` | `(str, max_len, suffix?)` | `String` | Truncate with suffix (default: "...") |
| `wrap` | `(str, width)` | `String` | Word-wrap at width |
| `indent` | `(str, spaces)` | `String` | Indent each non-empty line |
| `dedent` | `(str)` | `String` | Remove common leading whitespace |

---

### std.result

Utilities for working with collections of Result values.

```aura
import std.result as res

res.all_ok(results)               // Bool - all Ok?
res.any_ok(results)               // Bool - any Ok?
res.collect(results)              // Result<List, Err> - collect or first Err
res.partition_results(results)    // [ok_values, err_values]
res.from_option(opt, "error")     // Option -> Result
```

| Function | Args | Returns | Description |
|----------|------|---------|-------------|
| `all_ok` | `(results: List<Result>)` | `Bool` | True if all are Ok |
| `any_ok` | `(results: List<Result>)` | `Bool` | True if any is Ok |
| `collect` | `(results: List<Result>)` | `Result<List, Err>` | Collect Ok values, fail on first Err |
| `partition_results` | `(results: List<Result>)` | `List<List>` | Separate [oks, errs] |
| `from_option` | `(option, err_value)` | `Result` | Some→Ok, None→Err(err_value) |

---

### std.option

Utilities for working with collections of Option values.

```aura
import std.option as opt

opt.all_some(options)             // Bool - all Some?
opt.any_some(options)             // Bool - any Some?
opt.collect(options)              // Option<List> - collect or None
opt.first_some(options)           // Option - first Some value
opt.from_result(result)           // Result -> Option
```

| Function | Args | Returns | Description |
|----------|------|---------|-------------|
| `all_some` | `(options: List<Option>)` | `Bool` | True if all are Some |
| `any_some` | `(options: List<Option>)` | `Bool` | True if any is Some |
| `collect` | `(options: List<Option>)` | `Option<List>` | Collect Some values, None if any None |
| `first_some` | `(options: List<Option>)` | `Option` | First Some value, or None |
| `from_result` | `(result)` | `Option` | Ok→Some, Err→None |

---

### std.iter

Iterator utilities for list manipulation and generation.

```aura
import std.iter as iter

iter.cycle([1, 2], 3)            // [1, 2, 1, 2, 1, 2]
iter.repeat("x", 4)              // ["x", "x", "x", "x"]
iter.chain([[1,2], [3], [4,5]])  // [1, 2, 3, 4, 5]
iter.interleave([1,3,5], [2,4,6]) // [1, 2, 3, 4, 5, 6]
iter.pairwise([1, 2, 3, 4])     // [[1,2], [2,3], [3,4]]
```

| Function | Args | Returns | Description |
|----------|------|---------|-------------|
| `cycle` | `(list, n: Int)` | `List` | Repeat list n times |
| `repeat` | `(value, n: Int)` | `List` | List of n copies of value |
| `chain` | `(lists: List<List>)` | `List` | Concatenate multiple lists |
| `interleave` | `(list1, list2)` | `List` | Interleave two lists |
| `pairwise` | `(list)` | `List<List>` | Adjacent element pairs |



---

## std.file — File System Operations (Effect-Based)

> **Effect:** `file` — This module uses the file effect provider, which can be mocked for testing.

The `std.file` module provides file system operations through Aura's effect system. All I/O operations return `Result` types for safe error handling. In tests, the file provider can be replaced with a mock for deterministic, fast testing without actual filesystem access.

```aura
import std.file

# Read a file
let result = file.read("/path/to/file.txt")
match result:
    case Ok(content):
        print(content)
    case Err(msg):
        print("Error: {msg}")

# Write a file
file.write("/path/to/output.txt", "Hello, Aura!")

# Check existence
if file.exists("/path/to/file.txt"):
    print("File exists!")

# List directory contents
let entries = file.list_dir("/path/to/dir")
match entries:
    case Ok(list):
        list.for_each(|name| print(name))
    case Err(msg):
        print("Error: {msg}")
```

| Function | Args | Returns | Description |
|----------|------|---------|-------------|
| `read` | `(path: String)` | `Result[String, String]` | Read entire file contents |
| `write` | `(path: String, content: String)` | `Result[None, String]` | Write content to file (creates/overwrites) |
| `append` | `(path: String, content: String)` | `Result[None, String]` | Append content to file |
| `exists` | `(path: String)` | `Bool` | Check if path exists |
| `delete` | `(path: String)` | `Result[None, String]` | Delete a file or empty directory |
| `list_dir` | `(path: String)` | `Result[List[String], String]` | List directory entry names |
| `create_dir` | `(path: String)` | `Result[None, String]` | Create directory (with parents) |
| `is_file` | `(path: String)` | `Bool` | Check if path is a regular file |
| `is_dir` | `(path: String)` | `Bool` | Check if path is a directory |


---

## std.time — Time Operations (Effect-Based)

> **Effect:** `time` — Requires TimeProvider capability. Mockable for deterministic tests.

The `std.time` module provides time-related operations through Aura's effect system. The time provider can be replaced with a mock for deterministic, reproducible testing.

```aura
import std.time
```

#### Example Usage

```aura
import std.time

let now = time.now()           // Current Unix timestamp (seconds)
let ms = time.millis()          // Current time in milliseconds
time.sleep(100)                 // Sleep for 100ms

let formatted = time.format(now, "%Y-%m-%d %H:%M:%S")
let parsed = time.parse("2023-11-14 22:13:20", "%Y-%m-%d %H:%M:%S")

let future = time.add(now, 3600)   // Add 1 hour
let elapsed = time.diff(future, now)  // Difference in seconds
```

#### Format Tokens

| Token | Meaning | Example |
|-------|---------|---------|
| `%Y` | 4-digit year | `2023` |
| `%m` | 2-digit month | `01`–`12` |
| `%d` | 2-digit day | `01`–`31` |
| `%H` | 2-digit hour (24h) | `00`–`23` |
| `%M` | 2-digit minute | `00`–`59` |
| `%S` | 2-digit second | `00`–`59` |
| `%Z` | Timezone abbreviation | `UTC` |

#### Functions

| Function | Args | Returns | Description |
|----------|------|---------|-------------|
| `now` | `()` | `Int` | Current Unix timestamp in seconds |
| `unix` | `()` | `Int` | Alias for `now()` |
| `millis` | `()` | `Int` | Current time in milliseconds |
| `sleep` | `(ms: Int)` | `None` | Sleep for milliseconds |
| `format` | `(timestamp: Int, format: String)` | `String` | Format timestamp to string |
| `parse` | `(str: String, format: String)` | `Result[Int, String]` | Parse string to timestamp |
| `add` | `(timestamp: Int, seconds: Int)` | `Int` | Add seconds to timestamp |
| `diff` | `(ts1: Int, ts2: Int)` | `Int` | Difference in seconds (ts1 - ts2) |

---

## std.env — Environment Variables (Effect-Based)

> **Effect:** `env` — Requires EnvProvider capability. Mockable for isolated tests.

The `std.env` module provides access to environment variables, working directory, and command-line arguments through Aura's effect system. The env provider can be replaced with a mock for deterministic testing.

```aura
import std.env
```

#### Example Usage

```aura
import std.env

let home = env.get("HOME")      // Option[String]
env.set("APP_MODE", "production")
let exists = env.has("PATH")    // Bool

let vars = env.list()           // Map[String, String]
let cwd = env.cwd()             // String
let args = env.args()           // List[String]
```

#### Functions

| Function | Args | Returns | Description |
|----------|------|---------|-------------|
| `get` | `(key: String)` | `Option[String]` | Get environment variable |
| `set` | `(key: String, value: String)` | `None` | Set environment variable |
| `has` | `(key: String)` | `Bool` | Check if variable exists |
| `list` | `()` | `Map[String, String]` | List all environment variables |
| `cwd` | `()` | `String` | Current working directory |
| `args` | `()` | `List[String]` | Command-line arguments |

---

## std.testing — Effect-Aware Testing Helpers

> Effect-aware testing utilities for testing code that uses `std.file`, `std.time`, and `std.env`.
> These functions enable AI-driven TDD with minimal friction by providing mock effect contexts.

#### Usage

```aura
import std.testing

// Run a test with fresh mock effects (empty filesystem, default time, no env vars)
testing.with_mock_effects(fn() {
    // Inside here, all effects are mocked
    file.write("/test.txt", "hello")
    testing.assert_file_exists("/test.txt")
    testing.assert_file_content("/test.txt", "hello")
})

// Run with custom effect configuration
testing.with_effects({
    "time": 1700000000,
    "files": { "/config.json": '{"key": "val"}' },
    "env": { "MODE": "test" },
    "cwd": "/project",
    "args": ["aura", "--test"]
}, fn() {
    testing.assert_env_var("MODE", "test")
    let t = testing.get_mock_time()
    testing.assert_eq(t, 1700000000)
})

// Set and advance mock time
testing.with_mock_effects(fn() {
    testing.mock_time(1000)
    testing.advance_time(60)
    testing.assert_eq(testing.get_mock_time(), 1060)
})

// Reset effects to clean state
testing.reset_effects()
```

#### Effect Testing Functions

| Function | Args | Returns | Description |
|----------|------|---------|-------------|
| `with_mock_effects` | `(fn)` | `Any` | Run function with fresh mock effects |
| `with_effects` | `(config: Map, fn)` | `Any` | Run function with custom mock effects |
| `assert_file_exists` | `(path: String)` | `Bool` | Assert file exists in mock filesystem |
| `assert_file_content` | `(path: String, expected: String)` | `Bool` | Assert file has expected content |
| `assert_file_contains` | `(path: String, substr: String)` | `Bool` | Assert file contains substring |
| `assert_no_file` | `(path: String)` | `Bool` | Assert file does NOT exist |
| `assert_env_var` | `(key: String, expected: String)` | `Bool` | Assert environment variable value |
| `mock_time` | `(timestamp: Int)` | `None` | Set mock time to specific Unix timestamp |
| `advance_time` | `(seconds: Int)` | `None` | Advance mock time by N seconds |
| `reset_effects` | `()` | `None` | Reset all mock effects to clean state |
| `get_mock_time` | `()` | `Int` | Get current mock time |
| `get_env` | `(key: String)` | `Option[String]` | Get environment variable value |

#### `with_effects` Config Map Keys

| Key | Type | Description |
|-----|------|-------------|
| `"time"` | `Int` | Unix timestamp for mock time |
| `"files"` | `Map[String, String]` | Files to pre-populate (path → content) |
| `"env"` | `Map[String, String]` | Environment variables to set |
| `"cwd"` | `String` | Current working directory |
| `"args"` | `List[String]` | Command-line arguments |



---

## std.net — HTTP Network Operations

The `std.net` module provides HTTP client operations using the effect system's `NetProvider`. All functions return `Result[Response, String]` for safe error handling. Network operations are fully mockable for testing.

### Usage

```aura
import std.net

// Simple GET request
let result = net.get("https://api.example.com/users")
match result {
    Ok(response) => {
        print(response.status)  // 200
        print(response.body)    // Response body string
    }
    Err(msg) => print("Error: " + msg)
}

// POST with body and headers
let headers = {"Content-Type": "application/json", "Authorization": "Bearer token"}
let result = net.post("https://api.example.com/users", '{"name":"Alice"}', headers)

// Custom request with config map
let result = net.request({
    "method": "PATCH",
    "url": "https://api.example.com/users/1",
    "body": '{"name":"Bob"}',
    "headers": {"Content-Type": "application/json"},
    "timeout": 5000
})
```

### Response Map Structure

All successful responses return a map with these fields:

| Field | Type | Description |
|-------|------|-------------|
| `status` | `Int` | HTTP status code (e.g., 200, 404) |
| `status_text` | `String` | HTTP status text (e.g., "200 OK") |
| `body` | `String` | Response body as string |
| `headers` | `Map[String, String]` | Response headers |

### Functions

| Function | Arguments | Return | Description |
|----------|-----------|--------|-------------|
| `get` | `(url: String, headers?: Map)` | `Result[Map, String]` | HTTP GET request |
| `post` | `(url: String, body: String, headers?: Map)` | `Result[Map, String]` | HTTP POST request |
| `put` | `(url: String, body: String, headers?: Map)` | `Result[Map, String]` | HTTP PUT request |
| `delete` | `(url: String, headers?: Map)` | `Result[Map, String]` | HTTP DELETE request |
| `request` | `(config: Map)` | `Result[Map, String]` | Custom HTTP request |

### `request` Config Map Keys

| Key | Type | Required | Description |
|-----|------|----------|-------------|
| `method` | `String` | ✅ | HTTP method (GET, POST, PUT, DELETE, PATCH, etc.) |
| `url` | `String` | ✅ | Request URL |
| `body` | `String` | ❌ | Request body |
| `headers` | `Map[String, String]` | ❌ | Request headers |
| `timeout` | `Int` | ❌ | Timeout in milliseconds (default: 30000) |

### Testing with Mock NetProvider

```aura
import std.testing
import std.net

// Set up mock responses for testing
testing.with_mock_effects(fn() {
    let result = net.get("http://mock.api/data")
    // Uses MockNetProvider - returns default 200 OK
})
```

---

## std.log — Structured Logging

The `std.log` module provides structured logging using the effect system's `LogProvider`. All log functions accept an optional context map for structured data. Logging is fully mockable for verification in tests.

### Usage

```aura
import std.log

// Basic logging at different levels
log.info("Server started")
log.warn("Disk space low")
log.error("Connection failed")
log.debug("Processing item 42")

// Logging with structured context
log.info("User action", {"user_id": 42, "action": "login", "ip": "192.168.1.1"})
log.error("Request failed", {"status": 500, "url": "/api/data"})

// Retrieve logs (useful in tests with mock provider)
let logs = log.get_logs()
for entry in logs {
    print(entry.level + ": " + entry.message)
}
```

### Log Entry Structure

Each log entry (returned by `get_logs()`) is a map with:

| Field | Type | Description |
|-------|------|-------------|
| `level` | `String` | Log level: "INFO", "WARN", "ERROR", "DEBUG" |
| `message` | `String` | Log message |
| `context` | `Map` | Structured context data |
| `timestamp` | `Int` | Unix timestamp when logged |

### Functions

| Function | Arguments | Return | Description |
|----------|-----------|--------|-------------|
| `info` | `(message: String, context?: Map)` | `None` | Log info message |
| `warn` | `(message: String, context?: Map)` | `None` | Log warning message |
| `error` | `(message: String, context?: Map)` | `None` | Log error message |
| `debug` | `(message: String, context?: Map)` | `None` | Log debug message |
| `with_context` | `(context: Map, fn: Function)` | `Any` | Execute function with context |
| `get_logs` | `()` | `List[Map]` | Get all logged entries |

### Testing with Mock LogProvider

```aura
import std.testing
import std.log

testing.with_mock_effects(fn() {
    log.info("test message", {"key": "value"})
    
    let logs = log.get_logs()
    testing.assert_eq(logs.len(), 1)
    testing.assert_eq(logs.first().unwrap().level, "INFO")
})
```