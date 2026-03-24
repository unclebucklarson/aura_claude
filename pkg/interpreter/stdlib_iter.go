package interpreter

import (
	"fmt"
)

// createStdIterExports creates exports for the std.iter module.
func createStdIterExports() map[string]Value {
	exports := make(map[string]Value)

	// cycle(list, n) - Cycle through list n times
	exports["cycle"] = &BuiltinFnVal{
		Name: "iter.cycle",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "iter.cycle() requires exactly 2 arguments (list, n)"})
			}
			list, ok1 := args[0].(*ListVal)
			n, ok2 := args[1].(*IntVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "iter.cycle() requires (list, int) arguments"})
			}
			if n.Val < 0 {
				panic(&RuntimeError{Message: "iter.cycle() n must be non-negative"})
			}
			elems := make([]Value, 0, len(list.Elements)*int(n.Val))
			for i := int64(0); i < n.Val; i++ {
				elems = append(elems, list.Elements...)
			}
			return &ListVal{Elements: elems}
		},
	}

	// repeat(value, n) - Repeat value n times
	exports["repeat"] = &BuiltinFnVal{
		Name: "iter.repeat",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "iter.repeat() requires exactly 2 arguments (value, n)"})
			}
			n, ok := args[1].(*IntVal)
			if !ok {
				panic(&RuntimeError{Message: "iter.repeat() second argument must be an integer"})
			}
			if n.Val < 0 {
				panic(&RuntimeError{Message: "iter.repeat() n must be non-negative"})
			}
			elems := make([]Value, n.Val)
			for i := int64(0); i < n.Val; i++ {
				elems[i] = args[0]
			}
			return &ListVal{Elements: elems}
		},
	}

	// chain(lists) - Chain multiple lists
	exports["chain"] = &BuiltinFnVal{
		Name: "iter.chain",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "iter.chain() requires exactly 1 argument (list of lists)"})
			}
			outer, ok := args[0].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "iter.chain() requires a list argument"})
			}
			elems := make([]Value, 0)
			for _, item := range outer.Elements {
				inner, ok := item.(*ListVal)
				if !ok {
					panic(&RuntimeError{Message: "iter.chain() all elements must be lists"})
				}
				elems = append(elems, inner.Elements...)
			}
			return &ListVal{Elements: elems}
		},
	}

	// interleave(list1, list2) - Interleave two lists
	exports["interleave"] = &BuiltinFnVal{
		Name: "iter.interleave",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "iter.interleave() requires exactly 2 arguments (list1, list2)"})
			}
			l1, ok1 := args[0].(*ListVal)
			l2, ok2 := args[1].(*ListVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "iter.interleave() requires list arguments"})
			}
			maxLen := len(l1.Elements)
			if len(l2.Elements) > maxLen {
				maxLen = len(l2.Elements)
			}
			elems := make([]Value, 0, len(l1.Elements)+len(l2.Elements))
			for i := 0; i < maxLen; i++ {
				if i < len(l1.Elements) {
					elems = append(elems, l1.Elements[i])
				}
				if i < len(l2.Elements) {
					elems = append(elems, l2.Elements[i])
				}
			}
			return &ListVal{Elements: elems}
		},
	}

	// pairwise(list) - Create pairs of adjacent elements
	exports["pairwise"] = &BuiltinFnVal{
		Name: "iter.pairwise",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "iter.pairwise() requires exactly 1 argument (list)"})
			}
			list, ok := args[0].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "iter.pairwise() requires a list argument"})
			}
			if len(list.Elements) < 2 {
				return &ListVal{Elements: []Value{}}
			}
			pairs := make([]Value, len(list.Elements)-1)
			for i := 0; i < len(list.Elements)-1; i++ {
				pairs[i] = &ListVal{Elements: []Value{list.Elements[i], list.Elements[i+1]}}
			}
			return &ListVal{Elements: pairs}
		},
	}

	return exports
}

// Ensure fmt import is used
var _ = fmt.Sprintf
