package interpreter

// createStdOptionExports creates exports for the std.option module.
func createStdOptionExports() map[string]Value {
	exports := make(map[string]Value)

	// all_some(options) - Check if all Options are Some
	exports["all_some"] = &BuiltinFnVal{
		Name: "option.all_some",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "option.all_some() requires exactly 1 argument (list)"})
			}
			list, ok := args[0].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "option.all_some() requires a list argument"})
			}
			for _, elem := range list.Elements {
				opt, ok := elem.(*OptionVal)
				if !ok {
					panic(&RuntimeError{Message: "option.all_some() list must contain Option values"})
				}
				if !opt.IsSome {
					return &BoolVal{Val: false}
				}
			}
			return &BoolVal{Val: true}
		},
	}

	// any_some(options) - Check if any Option is Some
	exports["any_some"] = &BuiltinFnVal{
		Name: "option.any_some",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "option.any_some() requires exactly 1 argument (list)"})
			}
			list, ok := args[0].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "option.any_some() requires a list argument"})
			}
			for _, elem := range list.Elements {
				opt, ok := elem.(*OptionVal)
				if !ok {
					panic(&RuntimeError{Message: "option.any_some() list must contain Option values"})
				}
				if opt.IsSome {
					return &BoolVal{Val: true}
				}
			}
			return &BoolVal{Val: false}
		},
	}

	// collect(options) - Collect Some values, None if any None
	exports["collect"] = &BuiltinFnVal{
		Name: "option.collect",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "option.collect() requires exactly 1 argument (list)"})
			}
			list, ok := args[0].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "option.collect() requires a list argument"})
			}
			someValues := make([]Value, 0, len(list.Elements))
			for _, elem := range list.Elements {
				opt, ok := elem.(*OptionVal)
				if !ok {
					panic(&RuntimeError{Message: "option.collect() list must contain Option values"})
				}
				if !opt.IsSome {
					return &OptionVal{IsSome: false}
				}
				someValues = append(someValues, opt.Val)
			}
			return &OptionVal{IsSome: true, Val: &ListVal{Elements: someValues}}
		},
	}

	// first_some(options) - Get first Some value
	exports["first_some"] = &BuiltinFnVal{
		Name: "option.first_some",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "option.first_some() requires exactly 1 argument (list)"})
			}
			list, ok := args[0].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "option.first_some() requires a list argument"})
			}
			for _, elem := range list.Elements {
				opt, ok := elem.(*OptionVal)
				if !ok {
					panic(&RuntimeError{Message: "option.first_some() list must contain Option values"})
				}
				if opt.IsSome {
					return opt
				}
			}
			return &OptionVal{IsSome: false}
		},
	}

	// from_result(result) - Convert Result to Option
	exports["from_result"] = &BuiltinFnVal{
		Name: "option.from_result",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "option.from_result() requires exactly 1 argument (result)"})
			}
			result, ok := args[0].(*ResultVal)
			if !ok {
				panic(&RuntimeError{Message: "option.from_result() argument must be a Result"})
			}
			if result.IsOk {
				return &OptionVal{IsSome: true, Val: result.Val}
			}
			return &OptionVal{IsSome: false}
		},
	}

	return exports
}
