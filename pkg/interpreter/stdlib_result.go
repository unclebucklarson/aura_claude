package interpreter

// createStdResultExports creates exports for the std.result module.
func createStdResultExports() map[string]Value {
	exports := make(map[string]Value)

	// all_ok(results) - Check if all Results are Ok
	exports["all_ok"] = &BuiltinFnVal{
		Name: "result.all_ok",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "result.all_ok() requires exactly 1 argument (list)"})
			}
			list, ok := args[0].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "result.all_ok() requires a list argument"})
			}
			for _, elem := range list.Elements {
				r, ok := elem.(*ResultVal)
				if !ok {
					panic(&RuntimeError{Message: "result.all_ok() list must contain Result values"})
				}
				if !r.IsOk {
					return &BoolVal{Val: false}
				}
			}
			return &BoolVal{Val: true}
		},
	}

	// any_ok(results) - Check if any Result is Ok
	exports["any_ok"] = &BuiltinFnVal{
		Name: "result.any_ok",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "result.any_ok() requires exactly 1 argument (list)"})
			}
			list, ok := args[0].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "result.any_ok() requires a list argument"})
			}
			for _, elem := range list.Elements {
				r, ok := elem.(*ResultVal)
				if !ok {
					panic(&RuntimeError{Message: "result.any_ok() list must contain Result values"})
				}
				if r.IsOk {
					return &BoolVal{Val: true}
				}
			}
			return &BoolVal{Val: false}
		},
	}

	// collect(results) - Collect Ok values, fail on first Err
	exports["collect"] = &BuiltinFnVal{
		Name: "result.collect",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "result.collect() requires exactly 1 argument (list)"})
			}
			list, ok := args[0].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "result.collect() requires a list argument"})
			}
			okValues := make([]Value, 0, len(list.Elements))
			for _, elem := range list.Elements {
				r, ok := elem.(*ResultVal)
				if !ok {
					panic(&RuntimeError{Message: "result.collect() list must contain Result values"})
				}
				if !r.IsOk {
					// Return first Err
					return &ResultVal{IsOk: false, Val: r.Val}
				}
				okValues = append(okValues, r.Val)
			}
			return &ResultVal{IsOk: true, Val: &ListVal{Elements: okValues}}
		},
	}

	// partition_results(results) - Separate Oks and Errs
	exports["partition_results"] = &BuiltinFnVal{
		Name: "result.partition_results",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "result.partition_results() requires exactly 1 argument (list)"})
			}
			list, ok := args[0].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "result.partition_results() requires a list argument"})
			}
			oks := make([]Value, 0)
			errs := make([]Value, 0)
			for _, elem := range list.Elements {
				r, ok := elem.(*ResultVal)
				if !ok {
					panic(&RuntimeError{Message: "result.partition_results() list must contain Result values"})
				}
				if r.IsOk {
					oks = append(oks, r.Val)
				} else {
					errs = append(errs, r.Val)
				}
			}
			return &ListVal{Elements: []Value{
				&ListVal{Elements: oks},
				&ListVal{Elements: errs},
			}}
		},
	}

	// from_option(option, err) - Convert Option to Result
	exports["from_option"] = &BuiltinFnVal{
		Name: "result.from_option",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "result.from_option() requires exactly 2 arguments (option, err)"})
			}
			opt, ok := args[0].(*OptionVal)
			if !ok {
				panic(&RuntimeError{Message: "result.from_option() first argument must be an Option"})
			}
			if opt.IsSome {
				return &ResultVal{IsOk: true, Val: opt.Val}
			}
			return &ResultVal{IsOk: false, Val: args[1]}
		},
	}

	return exports
}
