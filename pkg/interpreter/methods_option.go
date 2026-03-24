package interpreter

import "fmt"

func init() {
	// =========================================================================
	// Option Methods
	// =========================================================================

	// is_some() -> Bool
	RegisterMethod(TypeOption, "is_some", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		return &BoolVal{Val: o.IsSome}
	})

	// is_none() -> Bool
	RegisterMethod(TypeOption, "is_none", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		return &BoolVal{Val: !o.IsSome}
	})

	// unwrap() -> T  (panics if None)
	RegisterMethod(TypeOption, "unwrap", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		if !o.IsSome {
			panic(&RuntimeError{Message: "called unwrap() on a None value"})
		}
		return o.Val
	})

	// expect(msg) -> T  (panics with custom message if None)
	RegisterMethod(TypeOption, "expect", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "expect() requires a message argument"})
		}
		if !o.IsSome {
			msg := args[0].String()
			panic(&RuntimeError{Message: msg})
		}
		return o.Val
	})

	// unwrap_or(default) -> T
	RegisterMethod(TypeOption, "unwrap_or", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "unwrap_or() requires a default argument"})
		}
		if o.IsSome {
			return o.Val
		}
		return args[0]
	})

	// unwrap_or_else(fn) -> T
	RegisterMethod(TypeOption, "unwrap_or_else", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "unwrap_or_else() requires a function argument"})
		}
		if o.IsSome {
			return o.Val
		}
		return callValue(args[0], []Value{})
	})

	// map(fn) -> Option[U]
	RegisterMethod(TypeOption, "map", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "map() requires a function argument"})
		}
		if !o.IsSome {
			return &OptionVal{IsSome: false}
		}
		result := callValue(args[0], []Value{o.Val})
		return &OptionVal{IsSome: true, Val: result}
	})

	// flat_map(fn) -> Option[U]  (monadic bind)
	RegisterMethod(TypeOption, "flat_map", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "flat_map() requires a function argument"})
		}
		if !o.IsSome {
			return &OptionVal{IsSome: false}
		}
		result := callValue(args[0], []Value{o.Val})
		// The function must return an Option
		if opt, ok := result.(*OptionVal); ok {
			return opt
		}
		panic(&RuntimeError{Message: fmt.Sprintf("flat_map() callback must return an Option, got %s", valueTypeNames[result.Type()])})
	})

	// and_then(fn) -> Option[U]  (alias for flat_map)
	RegisterMethod(TypeOption, "and_then", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "and_then() requires a function argument"})
		}
		if !o.IsSome {
			return &OptionVal{IsSome: false}
		}
		result := callValue(args[0], []Value{o.Val})
		if opt, ok := result.(*OptionVal); ok {
			return opt
		}
		panic(&RuntimeError{Message: fmt.Sprintf("and_then() callback must return an Option, got %s", valueTypeNames[result.Type()])})
	})

	// filter(fn) -> Option[T]
	RegisterMethod(TypeOption, "filter", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "filter() requires a predicate function argument"})
		}
		if !o.IsSome {
			return &OptionVal{IsSome: false}
		}
		result := callValue(args[0], []Value{o.Val})
		if IsTruthy(result) {
			return o
		}
		return &OptionVal{IsSome: false}
	})

	// or_else(fn) -> Option[T]
	RegisterMethod(TypeOption, "or_else", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "or_else() requires a function argument"})
		}
		if o.IsSome {
			return o
		}
		result := callValue(args[0], []Value{})
		if opt, ok := result.(*OptionVal); ok {
			return opt
		}
		panic(&RuntimeError{Message: fmt.Sprintf("or_else() callback must return an Option, got %s", valueTypeNames[result.Type()])})
	})

	// or(alternative) -> Option[T]
	RegisterMethod(TypeOption, "or", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "or() requires an alternative Option argument"})
		}
		if o.IsSome {
			return o
		}
		return args[0]
	})

	// and(other) -> Option[U]  (returns other if Some, None otherwise)
	RegisterMethod(TypeOption, "and", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "and() requires an Option argument"})
		}
		if !o.IsSome {
			return &OptionVal{IsSome: false}
		}
		return args[0]
	})

	// contains(value) -> Bool
	RegisterMethod(TypeOption, "contains", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "contains() requires a value argument"})
		}
		if !o.IsSome {
			return &BoolVal{Val: false}
		}
		return &BoolVal{Val: Equal(o.Val, args[0])}
	})

	// zip(other) -> Option[[T, U]]
	RegisterMethod(TypeOption, "zip", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "zip() requires an Option argument"})
		}
		if !o.IsSome {
			return &OptionVal{IsSome: false}
		}
		other, ok := args[0].(*OptionVal)
		if !ok {
			panic(&RuntimeError{Message: fmt.Sprintf("zip() requires an Option argument, got %s", valueTypeNames[args[0].Type()])})
		}
		if !other.IsSome {
			return &OptionVal{IsSome: false}
		}
		pair := &ListVal{Elements: []Value{o.Val, other.Val}}
		return &OptionVal{IsSome: true, Val: pair}
	})

	// flatten() -> Option[T]  (unwraps nested Option)
	RegisterMethod(TypeOption, "flatten", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		if !o.IsSome {
			return &OptionVal{IsSome: false}
		}
		if inner, ok := o.Val.(*OptionVal); ok {
			return inner
		}
		// If not nested, return as-is
		return o
	})

	// to_result(err_val) -> Result[T, E]
	RegisterMethod(TypeOption, "to_result", func(receiver Value, args []Value) Value {
		o := receiver.(*OptionVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "to_result() requires an error value argument"})
		}
		if o.IsSome {
			return &ResultVal{IsOk: true, Val: o.Val}
		}
		return &ResultVal{IsOk: false, Val: args[0]}
	})

	// =========================================================================
	// Result Methods
	// =========================================================================

	// is_ok() -> Bool
	RegisterMethod(TypeResult, "is_ok", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		return &BoolVal{Val: r.IsOk}
	})

	// is_err() -> Bool
	RegisterMethod(TypeResult, "is_err", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		return &BoolVal{Val: !r.IsOk}
	})

	// unwrap() -> T  (panics if Err)
	RegisterMethod(TypeResult, "unwrap", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if !r.IsOk {
			panic(&RuntimeError{Message: fmt.Sprintf("called unwrap() on an Err value: %s", r.Val.String())})
		}
		return r.Val
	})

	// unwrap_err() -> E  (panics if Ok)
	RegisterMethod(TypeResult, "unwrap_err", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if r.IsOk {
			panic(&RuntimeError{Message: fmt.Sprintf("called unwrap_err() on an Ok value: %s", r.Val.String())})
		}
		return r.Val
	})

	// expect(msg) -> T  (panics with custom message if Err)
	RegisterMethod(TypeResult, "expect", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "expect() requires a message argument"})
		}
		if !r.IsOk {
			msg := args[0].String()
			panic(&RuntimeError{Message: msg})
		}
		return r.Val
	})

	// unwrap_or(default) -> T
	RegisterMethod(TypeResult, "unwrap_or", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "unwrap_or() requires a default argument"})
		}
		if r.IsOk {
			return r.Val
		}
		return args[0]
	})

	// unwrap_or_else(fn) -> T  (fn receives the error value)
	RegisterMethod(TypeResult, "unwrap_or_else", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "unwrap_or_else() requires a function argument"})
		}
		if r.IsOk {
			return r.Val
		}
		return callValue(args[0], []Value{r.Val})
	})

	// map(fn) -> Result[U, E]
	RegisterMethod(TypeResult, "map", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "map() requires a function argument"})
		}
		if !r.IsOk {
			return r
		}
		result := callValue(args[0], []Value{r.Val})
		return &ResultVal{IsOk: true, Val: result}
	})

	// map_err(fn) -> Result[T, F]
	RegisterMethod(TypeResult, "map_err", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "map_err() requires a function argument"})
		}
		if r.IsOk {
			return r
		}
		result := callValue(args[0], []Value{r.Val})
		return &ResultVal{IsOk: false, Val: result}
	})

	// and_then(fn) -> Result[U, E]  (monadic bind)
	RegisterMethod(TypeResult, "and_then", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "and_then() requires a function argument"})
		}
		if !r.IsOk {
			return r
		}
		result := callValue(args[0], []Value{r.Val})
		if res, ok := result.(*ResultVal); ok {
			return res
		}
		panic(&RuntimeError{Message: fmt.Sprintf("and_then() callback must return a Result, got %s", valueTypeNames[result.Type()])})
	})

	// or_else(fn) -> Result[T, F]
	RegisterMethod(TypeResult, "or_else", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "or_else() requires a function argument"})
		}
		if r.IsOk {
			return r
		}
		result := callValue(args[0], []Value{r.Val})
		if res, ok := result.(*ResultVal); ok {
			return res
		}
		panic(&RuntimeError{Message: fmt.Sprintf("or_else() callback must return a Result, got %s", valueTypeNames[result.Type()])})
	})

	// ok() -> Option[T]  (converts Ok to Some, Err to None)
	RegisterMethod(TypeResult, "ok", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if r.IsOk {
			return &OptionVal{IsSome: true, Val: r.Val}
		}
		return &OptionVal{IsSome: false}
	})

	// err() -> Option[E]  (converts Err to Some, Ok to None)
	RegisterMethod(TypeResult, "err", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if !r.IsOk {
			return &OptionVal{IsSome: true, Val: r.Val}
		}
		return &OptionVal{IsSome: false}
	})

	// contains(value) -> Bool  (checks if Ok contains value)
	RegisterMethod(TypeResult, "contains", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "contains() requires a value argument"})
		}
		if !r.IsOk {
			return &BoolVal{Val: false}
		}
		return &BoolVal{Val: Equal(r.Val, args[0])}
	})

	// contains_err(value) -> Bool  (checks if Err contains value)
	RegisterMethod(TypeResult, "contains_err", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "contains_err() requires a value argument"})
		}
		if r.IsOk {
			return &BoolVal{Val: false}
		}
		return &BoolVal{Val: Equal(r.Val, args[0])}
	})

	// or(alternative) -> Result[T, F]  (returns self if Ok, alternative if Err)
	RegisterMethod(TypeResult, "or", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "or() requires a Result argument"})
		}
		if r.IsOk {
			return r
		}
		return args[0]
	})

	// and(other) -> Result[U, E]  (returns other if Ok, self if Err)
	RegisterMethod(TypeResult, "and", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if len(args) < 1 {
			panic(&RuntimeError{Message: "and() requires a Result argument"})
		}
		if !r.IsOk {
			return r
		}
		return args[0]
	})

	// flatten() -> Result[T, E]  (unwraps nested Result)
	RegisterMethod(TypeResult, "flatten", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if !r.IsOk {
			return r
		}
		if inner, ok := r.Val.(*ResultVal); ok {
			return inner
		}
		return r
	})

	// to_option() -> Option[T]  (alias for ok())
	RegisterMethod(TypeResult, "to_option", func(receiver Value, args []Value) Value {
		r := receiver.(*ResultVal)
		if r.IsOk {
			return &OptionVal{IsSome: true, Val: r.Val}
		}
		return &OptionVal{IsSome: false}
	})
}
