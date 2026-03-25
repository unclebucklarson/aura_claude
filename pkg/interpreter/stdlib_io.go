package interpreter

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// createStdIoExports creates exports for the std.io module.
func createStdIoExports() map[string]Value {
	exports := make(map[string]Value)

	exports["print"] = &BuiltinFnVal{
		Name: "io.print",
		Fn: func(args []Value) Value {
			parts := make([]string, len(args))
			for i, a := range args {
				parts[i] = a.String()
			}
			fmt.Println(joinStrings(parts, " "))
			return &NoneVal{}
		},
	}

	exports["println"] = &BuiltinFnVal{
		Name: "io.println",
		Fn: func(args []Value) Value {
			parts := make([]string, len(args))
			for i, a := range args {
				parts[i] = a.String()
			}
			fmt.Println(joinStrings(parts, " "))
			return &NoneVal{}
		},
	}

	exports["format"] = &BuiltinFnVal{
		Name: "io.format",
		Fn: func(args []Value) Value {
			if len(args) < 1 {
				panic(&RuntimeError{Message: "io.format() requires at least 1 argument"})
			}
			tmpl, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "io.format() first argument must be a string"})
			}
			// Simple {} placeholder replacement
			result := tmpl.Val
			for _, a := range args[1:] {
				idx := indexOf(result, "{}")
				if idx == -1 {
					break
				}
				result = result[:idx] + a.String() + result[idx+2:]
			}
			return &StringVal{Val: result}
		},
	}

	// read_line() -> Option[String] — Read a line from stdin (None on EOF).
	exports["read_line"] = &BuiltinFnVal{
		Name: "io.read_line",
		Fn: func(args []Value) Value {
			reader := bufio.NewReader(os.Stdin)
			line, err := reader.ReadString('\n')
			if err != nil && line == "" {
				return &OptionVal{IsSome: false}
			}
			return &OptionVal{IsSome: true, Val: &StringVal{Val: strings.TrimRight(line, "\r\n")}}
		},
	}

	// input(prompt?) -> String — Print optional prompt then read a line from stdin.
	// Returns empty string on EOF.
	exports["input"] = &BuiltinFnVal{
		Name: "io.input",
		Fn: func(args []Value) Value {
			if len(args) >= 1 {
				if prompt, ok := args[0].(*StringVal); ok {
					fmt.Print(prompt.Val)
				}
			}
			reader := bufio.NewReader(os.Stdin)
			line, _ := reader.ReadString('\n')
			return &StringVal{Val: strings.TrimRight(line, "\r\n")}
		},
	}

	return exports
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
