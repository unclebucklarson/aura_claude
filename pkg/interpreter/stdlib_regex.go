package interpreter

import (
	"fmt"
	"regexp"
)

// createStdRegexExports creates exports for the std.regex module.
func createStdRegexExports() map[string]Value {
	exports := make(map[string]Value)

	// match(pattern, text) - Test if pattern matches text
	exports["match"] = &BuiltinFnVal{
		Name: "regex.match",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "regex.match() requires exactly 2 arguments (pattern, text)"})
			}
			pattern, ok1 := args[0].(*StringVal)
			text, ok2 := args[1].(*StringVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "regex.match() requires string arguments"})
			}
			re, err := regexp.Compile(pattern.Val)
			if err != nil {
				panic(&RuntimeError{Message: fmt.Sprintf("regex.match(): invalid pattern: %v", err)})
			}
			return &BoolVal{Val: re.MatchString(text.Val)}
		},
	}

	// find(pattern, text) - Find first match (returns Option)
	exports["find"] = &BuiltinFnVal{
		Name: "regex.find",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "regex.find() requires exactly 2 arguments (pattern, text)"})
			}
			pattern, ok1 := args[0].(*StringVal)
			text, ok2 := args[1].(*StringVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "regex.find() requires string arguments"})
			}
			re, err := regexp.Compile(pattern.Val)
			if err != nil {
				panic(&RuntimeError{Message: fmt.Sprintf("regex.find(): invalid pattern: %v", err)})
			}
			match := re.FindString(text.Val)
			if match == "" {
				return &OptionVal{IsSome: false}
			}
			return &OptionVal{IsSome: true, Val: &StringVal{Val: match}}
		},
	}

	// find_all(pattern, text) - Find all matches (returns List)
	exports["find_all"] = &BuiltinFnVal{
		Name: "regex.find_all",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "regex.find_all() requires exactly 2 arguments (pattern, text)"})
			}
			pattern, ok1 := args[0].(*StringVal)
			text, ok2 := args[1].(*StringVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "regex.find_all() requires string arguments"})
			}
			re, err := regexp.Compile(pattern.Val)
			if err != nil {
				panic(&RuntimeError{Message: fmt.Sprintf("regex.find_all(): invalid pattern: %v", err)})
			}
			matches := re.FindAllString(text.Val, -1)
			elems := make([]Value, len(matches))
			for i, m := range matches {
				elems[i] = &StringVal{Val: m}
			}
			return &ListVal{Elements: elems}
		},
	}

	// replace(pattern, text, replacement) - Replace matches
	exports["replace"] = &BuiltinFnVal{
		Name: "regex.replace",
		Fn: func(args []Value) Value {
			if len(args) != 3 {
				panic(&RuntimeError{Message: "regex.replace() requires exactly 3 arguments (pattern, text, replacement)"})
			}
			pattern, ok1 := args[0].(*StringVal)
			text, ok2 := args[1].(*StringVal)
			replacement, ok3 := args[2].(*StringVal)
			if !ok1 || !ok2 || !ok3 {
				panic(&RuntimeError{Message: "regex.replace() requires string arguments"})
			}
			re, err := regexp.Compile(pattern.Val)
			if err != nil {
				panic(&RuntimeError{Message: fmt.Sprintf("regex.replace(): invalid pattern: %v", err)})
			}
			return &StringVal{Val: re.ReplaceAllString(text.Val, replacement.Val)}
		},
	}

	// split(pattern, text) - Split by pattern
	exports["split"] = &BuiltinFnVal{
		Name: "regex.split",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "regex.split() requires exactly 2 arguments (pattern, text)"})
			}
			pattern, ok1 := args[0].(*StringVal)
			text, ok2 := args[1].(*StringVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "regex.split() requires string arguments"})
			}
			re, err := regexp.Compile(pattern.Val)
			if err != nil {
				panic(&RuntimeError{Message: fmt.Sprintf("regex.split(): invalid pattern: %v", err)})
			}
			parts := re.Split(text.Val, -1)
			elems := make([]Value, len(parts))
			for i, p := range parts {
				elems[i] = &StringVal{Val: p}
			}
			return &ListVal{Elements: elems}
		},
	}

	// compile(pattern) - Compile regex (returns compiled regex as opaque string for reuse)
	exports["compile"] = &BuiltinFnVal{
		Name: "regex.compile",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "regex.compile() requires exactly 1 argument (pattern)"})
			}
			pattern, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "regex.compile() requires a string argument"})
			}
			_, err := regexp.Compile(pattern.Val)
			if err != nil {
				return &ResultVal{IsOk: false, Val: &StringVal{Val: err.Error()}}
			}
			// Return Ok with the pattern string (validated)
			return &ResultVal{IsOk: true, Val: &StringVal{Val: pattern.Val}}
		},
	}

	return exports
}
