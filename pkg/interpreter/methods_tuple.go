// Package interpreter implements tuple methods for the Aura language.
package interpreter

func init() {
	// len() - returns the number of elements in the tuple
	RegisterMethod(TypeTuple, "len", func(receiver Value, args []Value) Value {
		t := receiver.(*TupleVal)
		return &IntVal{Val: int64(len(t.Elements))}
	})

	// length() - alias for len()
	RegisterMethod(TypeTuple, "length", func(receiver Value, args []Value) Value {
		t := receiver.(*TupleVal)
		return &IntVal{Val: int64(len(t.Elements))}
	})

	// get(index) - returns Option[Value] at the given index
	RegisterMethod(TypeTuple, "get", func(receiver Value, args []Value) Value {
		t := receiver.(*TupleVal)
		if len(args) != 1 {
			panic(&RuntimeError{Message: "get() takes exactly 1 argument"})
		}
		idx, ok := args[0].(*IntVal)
		if !ok {
			panic(&RuntimeError{Message: "get() index must be an integer"})
		}
		i := idx.Val
		if i < 0 || i >= int64(len(t.Elements)) {
			return &OptionVal{IsSome: false}
		}
		return &OptionVal{IsSome: true, Val: t.Elements[i]}
	})

	// to_list() - converts the tuple to a list
	RegisterMethod(TypeTuple, "to_list", func(receiver Value, args []Value) Value {
		t := receiver.(*TupleVal)
		elems := make([]Value, len(t.Elements))
		copy(elems, t.Elements)
		return &ListVal{Elements: elems}
	})

	// is_empty() - returns true if the tuple has no elements
	RegisterMethod(TypeTuple, "is_empty", func(receiver Value, args []Value) Value {
		t := receiver.(*TupleVal)
		return &BoolVal{Val: len(t.Elements) == 0}
	})

	// contains(value) - returns true if the tuple contains the given value
	RegisterMethod(TypeTuple, "contains", func(receiver Value, args []Value) Value {
		t := receiver.(*TupleVal)
		if len(args) != 1 {
			panic(&RuntimeError{Message: "contains() takes exactly 1 argument"})
		}
		for _, elem := range t.Elements {
			if Equal(elem, args[0]) {
				return &BoolVal{Val: true}
			}
		}
		return &BoolVal{Val: false}
	})

	// first() - returns Option of the first element
	RegisterMethod(TypeTuple, "first", func(receiver Value, args []Value) Value {
		t := receiver.(*TupleVal)
		if len(t.Elements) == 0 {
			return &OptionVal{IsSome: false}
		}
		return &OptionVal{IsSome: true, Val: t.Elements[0]}
	})

	// last() - returns Option of the last element
	RegisterMethod(TypeTuple, "last", func(receiver Value, args []Value) Value {
		t := receiver.(*TupleVal)
		if len(t.Elements) == 0 {
			return &OptionVal{IsSome: false}
		}
		return &OptionVal{IsSome: true, Val: t.Elements[len(t.Elements)-1]}
	})

	// reverse() - returns a new tuple with elements in reverse order
	RegisterMethod(TypeTuple, "reverse", func(receiver Value, args []Value) Value {
		t := receiver.(*TupleVal)
		elems := make([]Value, len(t.Elements))
		for i, e := range t.Elements {
			elems[len(t.Elements)-1-i] = e
		}
		return &TupleVal{Elements: elems}
	})

	// enumerate() - returns a list of (index, value) tuples
	RegisterMethod(TypeTuple, "enumerate", func(receiver Value, args []Value) Value {
		t := receiver.(*TupleVal)
		elems := make([]Value, len(t.Elements))
		for i, e := range t.Elements {
			elems[i] = &TupleVal{Elements: []Value{&IntVal{Val: int64(i)}, e}}
		}
		return &ListVal{Elements: elems}
	})

	// map(fn) - applies fn to each element and returns a new tuple
	RegisterMethod(TypeTuple, "map", func(receiver Value, args []Value) Value {
		t := receiver.(*TupleVal)
		if len(args) != 1 {
			panic(&RuntimeError{Message: "map() takes exactly 1 argument"})
		}
		elems := make([]Value, len(t.Elements))
		for i, e := range t.Elements {
			elems[i] = callValue(args[0], []Value{e})
		}
		return &TupleVal{Elements: elems}
	})

	// for_each(fn) - applies fn to each element for side effects
	RegisterMethod(TypeTuple, "for_each", func(receiver Value, args []Value) Value {
		t := receiver.(*TupleVal)
		if len(args) != 1 {
			panic(&RuntimeError{Message: "for_each() takes exactly 1 argument"})
		}
		for _, e := range t.Elements {
			callValue(args[0], []Value{e})
		}
		return &NoneVal{}
	})

	// zip(other) - zips with another tuple/list, returns list of tuples
	RegisterMethod(TypeTuple, "zip", func(receiver Value, args []Value) Value {
		t := receiver.(*TupleVal)
		if len(args) != 1 {
			panic(&RuntimeError{Message: "zip() takes exactly 1 argument"})
		}
		var other []Value
		switch o := args[0].(type) {
		case *TupleVal:
			other = o.Elements
		case *ListVal:
			other = o.Elements
		default:
			panic(&RuntimeError{Message: "zip() argument must be a tuple or list"})
		}
		minLen := len(t.Elements)
		if len(other) < minLen {
			minLen = len(other)
		}
		result := make([]Value, minLen)
		for i := 0; i < minLen; i++ {
			result[i] = &TupleVal{Elements: []Value{t.Elements[i], other[i]}}
		}
		return &ListVal{Elements: result}
	})
}
