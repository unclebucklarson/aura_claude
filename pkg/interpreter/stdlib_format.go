package interpreter

import (
	"fmt"
	"math"
	"strings"
)

// createStdFormatExports creates exports for the std.format module.
func createStdFormatExports() map[string]Value {
	exports := make(map[string]Value)

	// pad_left(str, width, char?) - Left padding
	exports["pad_left"] = &BuiltinFnVal{
		Name: "format.pad_left",
		Fn: func(args []Value) Value {
			if len(args) < 2 || len(args) > 3 {
				panic(&RuntimeError{Message: "format.pad_left() requires 2-3 arguments (str, width, char?)"})
			}
			s, ok1 := args[0].(*StringVal)
			w, ok2 := args[1].(*IntVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "format.pad_left() requires (string, int) arguments"})
			}
			padChar := " "
			if len(args) == 3 {
				pc, ok := args[2].(*StringVal)
				if !ok {
					panic(&RuntimeError{Message: "format.pad_left() pad char must be a string"})
				}
				padChar = pc.Val
			}
			width := int(w.Val)
			if len(s.Val) >= width {
				return s
			}
			padding := strings.Repeat(padChar, width-len(s.Val))
			return &StringVal{Val: padding + s.Val}
		},
	}

	// pad_right(str, width, char?) - Right padding
	exports["pad_right"] = &BuiltinFnVal{
		Name: "format.pad_right",
		Fn: func(args []Value) Value {
			if len(args) < 2 || len(args) > 3 {
				panic(&RuntimeError{Message: "format.pad_right() requires 2-3 arguments (str, width, char?)"})
			}
			s, ok1 := args[0].(*StringVal)
			w, ok2 := args[1].(*IntVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "format.pad_right() requires (string, int) arguments"})
			}
			padChar := " "
			if len(args) == 3 {
				pc, ok := args[2].(*StringVal)
				if !ok {
					panic(&RuntimeError{Message: "format.pad_right() pad char must be a string"})
				}
				padChar = pc.Val
			}
			width := int(w.Val)
			if len(s.Val) >= width {
				return s
			}
			padding := strings.Repeat(padChar, width-len(s.Val))
			return &StringVal{Val: s.Val + padding}
		},
	}

	// center(str, width, char?) - Center text
	exports["center"] = &BuiltinFnVal{
		Name: "format.center",
		Fn: func(args []Value) Value {
			if len(args) < 2 || len(args) > 3 {
				panic(&RuntimeError{Message: "format.center() requires 2-3 arguments (str, width, char?)"})
			}
			s, ok1 := args[0].(*StringVal)
			w, ok2 := args[1].(*IntVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "format.center() requires (string, int) arguments"})
			}
			padChar := " "
			if len(args) == 3 {
				pc, ok := args[2].(*StringVal)
				if !ok {
					panic(&RuntimeError{Message: "format.center() pad char must be a string"})
				}
				padChar = pc.Val
			}
			width := int(w.Val)
			if len(s.Val) >= width {
				return s
			}
			totalPad := width - len(s.Val)
			leftPad := totalPad / 2
			rightPad := totalPad - leftPad
			return &StringVal{Val: strings.Repeat(padChar, leftPad) + s.Val + strings.Repeat(padChar, rightPad)}
		},
	}

	// truncate(str, max_len, suffix?) - Truncate with ellipsis
	exports["truncate"] = &BuiltinFnVal{
		Name: "format.truncate",
		Fn: func(args []Value) Value {
			if len(args) < 2 || len(args) > 3 {
				panic(&RuntimeError{Message: "format.truncate() requires 2-3 arguments (str, max_len, suffix?)"})
			}
			s, ok1 := args[0].(*StringVal)
			maxLen, ok2 := args[1].(*IntVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "format.truncate() requires (string, int) arguments"})
			}
			suffix := "..."
			if len(args) == 3 {
				sv, ok := args[2].(*StringVal)
				if !ok {
					panic(&RuntimeError{Message: "format.truncate() suffix must be a string"})
				}
				suffix = sv.Val
			}
			max := int(maxLen.Val)
			if len(s.Val) <= max {
				return s
			}
			if max <= len(suffix) {
				return &StringVal{Val: s.Val[:max]}
			}
			return &StringVal{Val: s.Val[:max-len(suffix)] + suffix}
		},
	}

	// wrap(str, width) - Word wrap
	exports["wrap"] = &BuiltinFnVal{
		Name: "format.wrap",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "format.wrap() requires exactly 2 arguments (str, width)"})
			}
			s, ok1 := args[0].(*StringVal)
			w, ok2 := args[1].(*IntVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "format.wrap() requires (string, int) arguments"})
			}
			width := int(w.Val)
			if width <= 0 {
				return s
			}
			words := strings.Fields(s.Val)
			if len(words) == 0 {
				return &StringVal{Val: ""}
			}
			var lines []string
			currentLine := words[0]
			for _, word := range words[1:] {
				if len(currentLine)+1+len(word) <= width {
					currentLine += " " + word
				} else {
					lines = append(lines, currentLine)
					currentLine = word
				}
			}
			lines = append(lines, currentLine)
			return &StringVal{Val: strings.Join(lines, "\n")}
		},
	}

	// indent(str, spaces) - Indent lines
	exports["indent"] = &BuiltinFnVal{
		Name: "format.indent",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "format.indent() requires exactly 2 arguments (str, spaces)"})
			}
			s, ok1 := args[0].(*StringVal)
			n, ok2 := args[1].(*IntVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "format.indent() requires (string, int) arguments"})
			}
			prefix := strings.Repeat(" ", int(n.Val))
			lines := strings.Split(s.Val, "\n")
			for i, line := range lines {
				if line != "" {
					lines[i] = prefix + line
				}
			}
			return &StringVal{Val: strings.Join(lines, "\n")}
		},
	}

	// dedent(str) - Remove common indentation
	exports["dedent"] = &BuiltinFnVal{
		Name: "format.dedent",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "format.dedent() requires exactly 1 argument (str)"})
			}
			s, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "format.dedent() requires a string argument"})
			}
			lines := strings.Split(s.Val, "\n")
			// Find minimum indentation of non-empty lines
			minIndent := math.MaxInt64
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}
				indent := 0
				for _, ch := range line {
					if ch == ' ' {
						indent++
					} else if ch == '\t' {
						indent += 4
					} else {
						break
					}
				}
				if indent < minIndent {
					minIndent = indent
				}
			}
			if minIndent == math.MaxInt64 || minIndent == 0 {
				return s
			}
			for i, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}
				// Remove minIndent chars from start
				removed := 0
				pos := 0
				for pos < len(line) && removed < minIndent {
					if line[pos] == ' ' {
						removed++
						pos++
					} else if line[pos] == '\t' {
						removed += 4
						pos++
					} else {
						break
					}
				}
				lines[i] = line[pos:]
			}
			return &StringVal{Val: strings.Join(lines, "\n")}
		},
	}

	return exports
}

// Ensure fmt import is used
var _ = fmt.Sprintf
