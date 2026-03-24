package interpreter

import (
	"fmt"
)

// createStdCollectionsExports creates exports for the std.collections module.
func createStdCollectionsExports() map[string]Value {
	exports := make(map[string]Value)

	// range(start, end, step?) - Generate number ranges
	exports["range"] = &BuiltinFnVal{
		Name: "collections.range",
		Fn: func(args []Value) Value {
			var start, end, step int64
			switch len(args) {
			case 1:
				e, ok := args[0].(*IntVal)
				if !ok {
					panic(&RuntimeError{Message: "collections.range() requires integer arguments"})
				}
				start, end, step = 0, e.Val, 1
			case 2:
				s, ok1 := args[0].(*IntVal)
				e, ok2 := args[1].(*IntVal)
				if !ok1 || !ok2 {
					panic(&RuntimeError{Message: "collections.range() requires integer arguments"})
				}
				start, end, step = s.Val, e.Val, 1
			case 3:
				s, ok1 := args[0].(*IntVal)
				e, ok2 := args[1].(*IntVal)
				st, ok3 := args[2].(*IntVal)
				if !ok1 || !ok2 || !ok3 {
					panic(&RuntimeError{Message: "collections.range() requires integer arguments"})
				}
				start, end, step = s.Val, e.Val, st.Val
			default:
				panic(&RuntimeError{Message: "collections.range() requires 1-3 arguments"})
			}
			if step == 0 {
				panic(&RuntimeError{Message: "collections.range() step cannot be zero"})
			}
			elems := make([]Value, 0)
			if step > 0 {
				for i := start; i < end; i += step {
					elems = append(elems, &IntVal{Val: i})
				}
			} else {
				for i := start; i > end; i += step {
					elems = append(elems, &IntVal{Val: i})
				}
			}
			return &ListVal{Elements: elems}
		},
	}

	// zip_with(fn, list1, list2) - Zip with custom function
	exports["zip_with"] = &BuiltinFnVal{
		Name: "collections.zip_with",
		Fn: func(args []Value) Value {
			if len(args) != 3 {
				panic(&RuntimeError{Message: "collections.zip_with() requires exactly 3 arguments (fn, list1, list2)"})
			}
			l1, ok1 := args[1].(*ListVal)
			l2, ok2 := args[2].(*ListVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "collections.zip_with() requires list arguments"})
			}
			minLen := len(l1.Elements)
			if len(l2.Elements) < minLen {
				minLen = len(l2.Elements)
			}
			elems := make([]Value, minLen)
			for i := 0; i < minLen; i++ {
				elems[i] = callValue(args[0], []Value{l1.Elements[i], l2.Elements[i]})
			}
			return &ListVal{Elements: elems}
		},
	}

	// partition(fn, list) - Split list by predicate
	exports["partition"] = &BuiltinFnVal{
		Name: "collections.partition",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "collections.partition() requires exactly 2 arguments (fn, list)"})
			}
			list, ok := args[1].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "collections.partition() requires a list argument"})
			}
			trueElems := make([]Value, 0)
			falseElems := make([]Value, 0)
			for _, elem := range list.Elements {
				result := callValue(args[0], []Value{elem})
				if IsTruthy(result) {
					trueElems = append(trueElems, elem)
				} else {
					falseElems = append(falseElems, elem)
				}
			}
			// Return a list of two lists: [matching, non-matching]
			return &ListVal{Elements: []Value{
				&ListVal{Elements: trueElems},
				&ListVal{Elements: falseElems},
			}}
		},
	}

	// group_by(fn, list) - Group elements by key function
	exports["group_by"] = &BuiltinFnVal{
		Name: "collections.group_by",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "collections.group_by() requires exactly 2 arguments (fn, list)"})
			}
			list, ok := args[1].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "collections.group_by() requires a list argument"})
			}
			// Use ordered map for deterministic output
			keys := make([]Value, 0)
			values := make([]Value, 0)
			keyIndex := make(map[string]int) // key.String() -> index
			for _, elem := range list.Elements {
				key := callValue(args[0], []Value{elem})
				keyStr := key.String()
				if idx, exists := keyIndex[keyStr]; exists {
					// Append to existing group
					group := values[idx].(*ListVal)
					group.Elements = append(group.Elements, elem)
				} else {
					// New group
					keyIndex[keyStr] = len(keys)
					keys = append(keys, key)
					values = append(values, &ListVal{Elements: []Value{elem}})
				}
			}
			return &MapVal{Keys: keys, Values: values}
		},
	}

	// chunk(n, list) - Split into chunks
	exports["chunk"] = &BuiltinFnVal{
		Name: "collections.chunk",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "collections.chunk() requires exactly 2 arguments (n, list)"})
			}
			n, ok1 := args[0].(*IntVal)
			list, ok2 := args[1].(*ListVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "collections.chunk() requires (int, list) arguments"})
			}
			if n.Val <= 0 {
				panic(&RuntimeError{Message: "collections.chunk() size must be positive"})
			}
			chunks := make([]Value, 0)
			size := int(n.Val)
			for i := 0; i < len(list.Elements); i += size {
				end := i + size
				if end > len(list.Elements) {
					end = len(list.Elements)
				}
				chunk := make([]Value, end-i)
				copy(chunk, list.Elements[i:end])
				chunks = append(chunks, &ListVal{Elements: chunk})
			}
			return &ListVal{Elements: chunks}
		},
	}

	// take(n, list) - Take first n elements
	exports["take"] = &BuiltinFnVal{
		Name: "collections.take",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "collections.take() requires exactly 2 arguments (n, list)"})
			}
			n, ok1 := args[0].(*IntVal)
			list, ok2 := args[1].(*ListVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "collections.take() requires (int, list) arguments"})
			}
			count := int(n.Val)
			if count > len(list.Elements) {
				count = len(list.Elements)
			}
			if count < 0 {
				count = 0
			}
			elems := make([]Value, count)
			copy(elems, list.Elements[:count])
			return &ListVal{Elements: elems}
		},
	}

	// drop(n, list) - Drop first n elements
	exports["drop"] = &BuiltinFnVal{
		Name: "collections.drop",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "collections.drop() requires exactly 2 arguments (n, list)"})
			}
			n, ok1 := args[0].(*IntVal)
			list, ok2 := args[1].(*ListVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "collections.drop() requires (int, list) arguments"})
			}
			start := int(n.Val)
			if start > len(list.Elements) {
				start = len(list.Elements)
			}
			if start < 0 {
				start = 0
			}
			elems := make([]Value, len(list.Elements)-start)
			copy(elems, list.Elements[start:])
			return &ListVal{Elements: elems}
		},
	}

	// take_while(fn, list) - Take while predicate is true
	exports["take_while"] = &BuiltinFnVal{
		Name: "collections.take_while",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "collections.take_while() requires exactly 2 arguments (fn, list)"})
			}
			list, ok := args[1].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "collections.take_while() requires a list argument"})
			}
			elems := make([]Value, 0)
			for _, elem := range list.Elements {
				result := callValue(args[0], []Value{elem})
				if !IsTruthy(result) {
					break
				}
				elems = append(elems, elem)
			}
			return &ListVal{Elements: elems}
		},
	}

	// drop_while(fn, list) - Drop while predicate is true
	exports["drop_while"] = &BuiltinFnVal{
		Name: "collections.drop_while",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "collections.drop_while() requires exactly 2 arguments (fn, list)"})
			}
			list, ok := args[1].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "collections.drop_while() requires a list argument"})
			}
			dropping := true
			elems := make([]Value, 0)
			for _, elem := range list.Elements {
				if dropping {
					result := callValue(args[0], []Value{elem})
					if !IsTruthy(result) {
						dropping = false
						elems = append(elems, elem)
					}
				} else {
					elems = append(elems, elem)
				}
			}
			return &ListVal{Elements: elems}
		},
	}

	return exports
}

// Ensure fmt import is used
var _ = fmt.Sprintf
