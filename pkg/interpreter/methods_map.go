package interpreter

import "fmt"

func init() {
	registerMapMethods()
}

// mapFindKey returns the index of a key in a MapVal, or -1 if not found.
func mapFindKey(m *MapVal, key Value) int {
	for i, k := range m.Keys {
		if Equal(k, key) {
			return i
		}
	}
	return -1
}

func registerMapMethods() {
	// =========================================================================
	// Size / emptiness
	// =========================================================================

	// len() -> Int — Get number of entries
	RegisterMethod(TypeMap, "len", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		return &IntVal{Val: int64(len(m.Keys))}
	})

	// length() -> Int — Alias for len
	RegisterMethod(TypeMap, "length", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		return &IntVal{Val: int64(len(m.Keys))}
	})

	// size() -> Int — Alias for len
	RegisterMethod(TypeMap, "size", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		return &IntVal{Val: int64(len(m.Keys))}
	})

	// is_empty() -> Bool — Check if map is empty
	RegisterMethod(TypeMap, "is_empty", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		return &BoolVal{Val: len(m.Keys) == 0}
	})

	// =========================================================================
	// Key / value / entry accessors
	// =========================================================================

	// keys() -> List — Get all keys as a list
	RegisterMethod(TypeMap, "keys", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		elems := make([]Value, len(m.Keys))
		copy(elems, m.Keys)
		return &ListVal{Elements: elems}
	})

	// values() -> List — Get all values as a list
	RegisterMethod(TypeMap, "values", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		elems := make([]Value, len(m.Values))
		copy(elems, m.Values)
		return &ListVal{Elements: elems}
	})

	// entries() -> List[List[key, value]] — Get key-value pairs as list of [key, value] lists
	RegisterMethod(TypeMap, "entries", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		pairs := make([]Value, len(m.Keys))
		for i := range m.Keys {
			pairs[i] = &ListVal{Elements: []Value{m.Keys[i], m.Values[i]}}
		}
		return &ListVal{Elements: pairs}
	})

	// =========================================================================
	// Lookup
	// =========================================================================

	// has(key) -> Bool — Check if key exists
	RegisterMethod(TypeMap, "has", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "Map.has requires one argument"})
		}
		return &BoolVal{Val: mapFindKey(m, args[0]) >= 0}
	})

	// contains_key(key) -> Bool — Alias for has
	RegisterMethod(TypeMap, "contains_key", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "Map.contains_key requires one argument"})
		}
		return &BoolVal{Val: mapFindKey(m, args[0]) >= 0}
	})

	// contains_value(value) -> Bool — Check if value exists
	RegisterMethod(TypeMap, "contains_value", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "Map.contains_value requires one argument"})
		}
		for _, v := range m.Values {
			if Equal(v, args[0]) {
				return &BoolVal{Val: true}
			}
		}
		return &BoolVal{Val: false}
	})

	// get(key) -> Option — Safe access, returns Some(value) or None
	RegisterMethod(TypeMap, "get", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "Map.get requires one argument"})
		}
		idx := mapFindKey(m, args[0])
		if idx < 0 {
			return &OptionVal{IsSome: false}
		}
		return &OptionVal{IsSome: true, Val: m.Values[idx]}
	})

	// get_or(key, default) -> Value — Get value or return default
	RegisterMethod(TypeMap, "get_or", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 2 {
			panic(&RuntimeError{Message: "Map.get_or requires two arguments (key, default)"})
		}
		idx := mapFindKey(m, args[0])
		if idx < 0 {
			return args[1]
		}
		return m.Values[idx]
	})

	// =========================================================================
	// Mutation
	// =========================================================================

	// set(key, value) -> None — Add or update entry (mutates)
	RegisterMethod(TypeMap, "set", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 2 {
			panic(&RuntimeError{Message: "Map.set requires two arguments (key, value)"})
		}
		idx := mapFindKey(m, args[0])
		if idx >= 0 {
			m.Values[idx] = args[1]
		} else {
			m.Keys = append(m.Keys, args[0])
			m.Values = append(m.Values, args[1])
		}
		return &NoneVal{}
	})

	// remove(key) -> Option — Remove entry and return Option of removed value (mutates)
	RegisterMethod(TypeMap, "remove", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "Map.remove requires one argument"})
		}
		idx := mapFindKey(m, args[0])
		if idx < 0 {
			return &OptionVal{IsSome: false}
		}
		removed := m.Values[idx]
		// Remove by shifting (preserves insertion order)
		m.Keys = append(m.Keys[:idx], m.Keys[idx+1:]...)
		m.Values = append(m.Values[:idx], m.Values[idx+1:]...)
		return &OptionVal{IsSome: true, Val: removed}
	})

	// delete(key) -> Bool — Remove entry, returns whether key existed (mutates)
	RegisterMethod(TypeMap, "delete", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "Map.delete requires one argument"})
		}
		idx := mapFindKey(m, args[0])
		if idx < 0 {
			return &BoolVal{Val: false}
		}
		m.Keys = append(m.Keys[:idx], m.Keys[idx+1:]...)
		m.Values = append(m.Values[:idx], m.Values[idx+1:]...)
		return &BoolVal{Val: true}
	})

	// clear() -> None — Remove all entries (mutates)
	RegisterMethod(TypeMap, "clear", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		m.Keys = nil
		m.Values = nil
		return &NoneVal{}
	})

	// merge(other) -> None — Merge another map into this one (mutates).
	// Keys in other overwrite existing keys.
	RegisterMethod(TypeMap, "merge", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "Map.merge requires one argument (another map)"})
		}
		other, ok := args[0].(*MapVal)
		if !ok {
			panic(&RuntimeError{Message: fmt.Sprintf("Map.merge argument must be a Map, got %s", valueTypeNames[args[0].Type()])})
		}
		for i, k := range other.Keys {
			idx := mapFindKey(m, k)
			if idx >= 0 {
				m.Values[idx] = other.Values[i]
			} else {
				m.Keys = append(m.Keys, k)
				m.Values = append(m.Values, other.Values[i])
			}
		}
		return &NoneVal{}
	})

	// =========================================================================
	// Higher-order methods (non-mutating)
	// =========================================================================

	// filter(fn) -> Map — Filter entries by predicate fn(key, value) -> Bool (returns new map)
	RegisterMethod(TypeMap, "filter", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "Map.filter requires a function argument"})
		}
		fn := args[0]
		newKeys := []Value{}
		newVals := []Value{}
		for i := range m.Keys {
			result := callValue(fn, []Value{m.Keys[i], m.Values[i]})
			if IsTruthy(result) {
				newKeys = append(newKeys, m.Keys[i])
				newVals = append(newVals, m.Values[i])
			}
		}
		return &MapVal{Keys: newKeys, Values: newVals}
	})

	// map(fn) -> Map — Transform values with fn(key, value) -> new_value (returns new map)
	RegisterMethod(TypeMap, "map", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "Map.map requires a function argument"})
		}
		fn := args[0]
		newKeys := make([]Value, len(m.Keys))
		newVals := make([]Value, len(m.Keys))
		copy(newKeys, m.Keys)
		for i := range m.Keys {
			newVals[i] = callValue(fn, []Value{m.Keys[i], m.Values[i]})
		}
		return &MapVal{Keys: newKeys, Values: newVals}
	})

	// for_each(fn) -> None — Execute fn(key, value) for each entry
	RegisterMethod(TypeMap, "for_each", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "Map.for_each requires a function argument"})
		}
		fn := args[0]
		for i := range m.Keys {
			callValue(fn, []Value{m.Keys[i], m.Values[i]})
		}
		return &NoneVal{}
	})

	// reduce(init, fn) -> Value — Reduce entries with fn(acc, key, value) -> acc
	RegisterMethod(TypeMap, "reduce", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 2 {
			panic(&RuntimeError{Message: "Map.reduce requires two arguments (initial, function)"})
		}
		acc := args[0]
		fn := args[1]
		for i := range m.Keys {
			acc = callValue(fn, []Value{acc, m.Keys[i], m.Values[i]})
		}
		return acc
	})

	// any(fn) -> Bool — Check if any entry satisfies fn(key, value) -> Bool
	RegisterMethod(TypeMap, "any", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "Map.any requires a function argument"})
		}
		fn := args[0]
		for i := range m.Keys {
			if IsTruthy(callValue(fn, []Value{m.Keys[i], m.Values[i]})) {
				return &BoolVal{Val: true}
			}
		}
		return &BoolVal{Val: false}
	})

	// all(fn) -> Bool — Check if all entries satisfy fn(key, value) -> Bool
	RegisterMethod(TypeMap, "all", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "Map.all requires a function argument"})
		}
		fn := args[0]
		for i := range m.Keys {
			if !IsTruthy(callValue(fn, []Value{m.Keys[i], m.Values[i]})) {
				return &BoolVal{Val: false}
			}
		}
		return &BoolVal{Val: true}
	})

	// count(fn?) -> Int — Count entries; if fn provided, count matching fn(key, value) -> Bool
	RegisterMethod(TypeMap, "count", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) == 0 {
			return &IntVal{Val: int64(len(m.Keys))}
		}
		fn := args[0]
		count := int64(0)
		for i := range m.Keys {
			if IsTruthy(callValue(fn, []Value{m.Keys[i], m.Values[i]})) {
				count++
			}
		}
		return &IntVal{Val: count}
	})

	// =========================================================================
	// Utility
	// =========================================================================

	// to_list() -> List[List[key, value]] — Convert to list of [key, value] pairs (alias for entries)
	RegisterMethod(TypeMap, "to_list", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		pairs := make([]Value, len(m.Keys))
		for i := range m.Keys {
			pairs[i] = &ListVal{Elements: []Value{m.Keys[i], m.Values[i]}}
		}
		return &ListVal{Elements: pairs}
	})

	// from_entries — static-like helper: takes list of [k,v] pairs and builds a map
	// (Registered on Map, called as: {}.from_entries(list) — or more practically used after to_list)

	// find(fn) -> Option[List[key, value]] — Find first entry matching fn(key, value) -> Bool
	RegisterMethod(TypeMap, "find", func(receiver Value, args []Value) Value {
		m := receiver.(*MapVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "Map.find requires a function argument"})
		}
		fn := args[0]
		for i := range m.Keys {
			if IsTruthy(callValue(fn, []Value{m.Keys[i], m.Values[i]})) {
				pair := &ListVal{Elements: []Value{m.Keys[i], m.Values[i]}}
				return &OptionVal{IsSome: true, Val: pair}
			}
		}
		return &OptionVal{IsSome: false}
	})
}
