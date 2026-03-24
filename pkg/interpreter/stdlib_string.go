package interpreter

import (
	"strings"
)

// createStdStringExports creates exports for the std.string module.
func createStdStringExports() map[string]Value {
	exports := make(map[string]Value)

	exports["join"] = &BuiltinFnVal{
		Name: "string.join",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "string.join() requires 2 arguments (list, separator)"})
			}
			list, ok := args[0].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "string.join() first argument must be a list"})
			}
			sep, ok := args[1].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "string.join() second argument must be a string"})
			}
			parts := make([]string, len(list.Elements))
			for i, el := range list.Elements {
				parts[i] = el.String()
			}
			return &StringVal{Val: joinStrings(parts, sep.Val)}
		},
	}

	exports["split"] = &BuiltinFnVal{
		Name: "string.split",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "string.split() requires 2 arguments (string, separator)"})
			}
			s, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "string.split() first argument must be a string"})
			}
			sep, ok := args[1].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "string.split() second argument must be a string"})
			}
			parts := strings.Split(s.Val, sep.Val)
			elems := make([]Value, len(parts))
			for i, p := range parts {
				elems[i] = &StringVal{Val: p}
			}
			return &ListVal{Elements: elems}
		},
	}

	exports["replace"] = &BuiltinFnVal{
		Name: "string.replace",
		Fn: func(args []Value) Value {
			if len(args) != 3 {
				panic(&RuntimeError{Message: "string.replace() requires 3 arguments (string, old, new)"})
			}
			s, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "string.replace() first argument must be a string"})
			}
			old, ok := args[1].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "string.replace() second argument must be a string"})
			}
			newStr, ok := args[2].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "string.replace() third argument must be a string"})
			}
			return &StringVal{Val: strings.ReplaceAll(s.Val, old.Val, newStr.Val)}
		},
	}

	exports["repeat"] = &BuiltinFnVal{
		Name: "string.repeat",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "string.repeat() requires 2 arguments (string, count)"})
			}
			s, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "string.repeat() first argument must be a string"})
			}
			count, ok := args[1].(*IntVal)
			if !ok {
				panic(&RuntimeError{Message: "string.repeat() second argument must be an integer"})
			}
			return &StringVal{Val: strings.Repeat(s.Val, int(count.Val))}
		},
	}

	return exports
}
