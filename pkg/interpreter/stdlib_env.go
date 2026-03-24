package interpreter

import "fmt"

// createStdEnvExports creates the exports for the std.env module.
// The env provider is captured via closure, enabling effect mocking.
func createStdEnvExports(ep EnvProvider) map[string]Value {
	exports := make(map[string]Value)

	// get(key) -> Option[String] - Get environment variable
	exports["get"] = &BuiltinFnVal{
		Name: "env.get",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "env.get() requires exactly 1 argument (key)"})
			}
			key, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("env.get() key must be a String, got %s", valueTypeNames[args[0].Type()])})
			}
			val, exists := ep.Get(key.Val)
			if !exists {
				return &OptionVal{IsSome: false}
			}
			return &OptionVal{IsSome: true, Val: &StringVal{Val: val}}
		},
	}

	// set(key, value) -> None - Set environment variable
	exports["set"] = &BuiltinFnVal{
		Name: "env.set",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "env.set() requires exactly 2 arguments (key, value)"})
			}
			key, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("env.set() key must be a String, got %s", valueTypeNames[args[0].Type()])})
			}
			val, ok := args[1].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("env.set() value must be a String, got %s", valueTypeNames[args[1].Type()])})
			}
			ep.Set(key.Val, val.Val)
			return &NoneVal{}
		},
	}

	// has(key) -> Bool - Check if variable exists
	exports["has"] = &BuiltinFnVal{
		Name: "env.has",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "env.has() requires exactly 1 argument (key)"})
			}
			key, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("env.has() key must be a String, got %s", valueTypeNames[args[0].Type()])})
			}
			return &BoolVal{Val: ep.Has(key.Val)}
		},
	}

	// list() -> Map[String, String] - List all environment variables
	exports["list"] = &BuiltinFnVal{
		Name: "env.list",
		Fn: func(args []Value) Value {
			if len(args) != 0 {
				panic(&RuntimeError{Message: "env.list() takes no arguments"})
			}
			vars := ep.List()
			keys := make([]Value, 0, len(vars))
			values := make([]Value, 0, len(vars))
			for k, v := range vars {
				keys = append(keys, &StringVal{Val: k})
				values = append(values, &StringVal{Val: v})
			}
			return &MapVal{Keys: keys, Values: values}
		},
	}

	// cwd() -> String - Current working directory
	exports["cwd"] = &BuiltinFnVal{
		Name: "env.cwd",
		Fn: func(args []Value) Value {
			if len(args) != 0 {
				panic(&RuntimeError{Message: "env.cwd() takes no arguments"})
			}
			cwd, err := ep.Cwd()
			if err != nil {
				panic(&RuntimeError{Message: fmt.Sprintf("env.cwd() failed: %s", err.Error())})
			}
			return &StringVal{Val: cwd}
		},
	}

	// args() -> List[String] - Command line arguments
	exports["args"] = &BuiltinFnVal{
		Name: "env.args",
		Fn: func(args []Value) Value {
			if len(args) != 0 {
				panic(&RuntimeError{Message: "env.args() takes no arguments"})
			}
			cliArgs := ep.Args()
			elements := make([]Value, len(cliArgs))
			for i, a := range cliArgs {
				elements[i] = &StringVal{Val: a}
			}
			return &ListVal{Elements: elements}
		},
	}

	return exports
}
