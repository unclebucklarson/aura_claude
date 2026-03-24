package interpreter

import (
	"strings"
	"unicode/utf8"
)

func init() {
	registerStringMethods()
}

func registerStringMethods() {
	// len() -> Int — Get string length
	RegisterMethod(TypeString, "len", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		return &IntVal{Val: int64(utf8.RuneCountInString(s))}
	})

	// length() -> Int — Alias for len
	RegisterMethod(TypeString, "length", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		return &IntVal{Val: int64(utf8.RuneCountInString(s))}
	})

	// upper() -> String — Convert to uppercase
	RegisterMethod(TypeString, "upper", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		return &StringVal{Val: strings.ToUpper(s)}
	})

	// to_upper() -> String — Alias for upper
	RegisterMethod(TypeString, "to_upper", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		return &StringVal{Val: strings.ToUpper(s)}
	})

	// lower() -> String — Convert to lowercase
	RegisterMethod(TypeString, "lower", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		return &StringVal{Val: strings.ToLower(s)}
	})

	// to_lower() -> String — Alias for lower
	RegisterMethod(TypeString, "to_lower", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		return &StringVal{Val: strings.ToLower(s)}
	})

	// contains(sub) -> Bool — Check if contains substring
	RegisterMethod(TypeString, "contains", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		if len(args) < 1 {
			return &BoolVal{Val: false}
		}
		sub, ok := args[0].(*StringVal)
		if !ok {
			panic(&RuntimeError{Message: "String.contains requires a string argument"})
		}
		return &BoolVal{Val: strings.Contains(s, sub.Val)}
	})

	// split(sep) -> List[String] — Split into list
	RegisterMethod(TypeString, "split", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		sep := " "
		if len(args) > 0 {
			if sv, ok := args[0].(*StringVal); ok {
				sep = sv.Val
			}
		}
		parts := strings.Split(s, sep)
		elems := make([]Value, len(parts))
		for i, p := range parts {
			elems[i] = &StringVal{Val: p}
		}
		return &ListVal{Elements: elems}
	})

	// trim() -> String — Remove leading/trailing whitespace
	RegisterMethod(TypeString, "trim", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		return &StringVal{Val: strings.TrimSpace(s)}
	})

	// trim_left() -> String — Remove leading whitespace
	RegisterMethod(TypeString, "trim_left", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		return &StringVal{Val: strings.TrimLeft(s, " \t\n\r")}
	})

	// trim_right() -> String — Remove trailing whitespace
	RegisterMethod(TypeString, "trim_right", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		return &StringVal{Val: strings.TrimRight(s, " \t\n\r")}
	})

	// starts_with(prefix) -> Bool — Check prefix
	RegisterMethod(TypeString, "starts_with", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		if len(args) < 1 {
			panic(&RuntimeError{Message: "String.starts_with requires a string argument"})
		}
		prefix, ok := args[0].(*StringVal)
		if !ok {
			panic(&RuntimeError{Message: "String.starts_with requires a string argument"})
		}
		return &BoolVal{Val: strings.HasPrefix(s, prefix.Val)}
	})

	// ends_with(suffix) -> Bool — Check suffix
	RegisterMethod(TypeString, "ends_with", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		if len(args) < 1 {
			panic(&RuntimeError{Message: "String.ends_with requires a string argument"})
		}
		suffix, ok := args[0].(*StringVal)
		if !ok {
			panic(&RuntimeError{Message: "String.ends_with requires a string argument"})
		}
		return &BoolVal{Val: strings.HasSuffix(s, suffix.Val)}
	})

	// replace(old, new) -> String — Replace all occurrences
	RegisterMethod(TypeString, "replace", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		if len(args) < 2 {
			panic(&RuntimeError{Message: "String.replace requires two string arguments"})
		}
		old, ok1 := args[0].(*StringVal)
		new_, ok2 := args[1].(*StringVal)
		if !ok1 || !ok2 {
			panic(&RuntimeError{Message: "String.replace requires string arguments"})
		}
		return &StringVal{Val: strings.ReplaceAll(s, old.Val, new_.Val)}
	})

	// replace_first(old, new) -> String — Replace first occurrence
	RegisterMethod(TypeString, "replace_first", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		if len(args) < 2 {
			panic(&RuntimeError{Message: "String.replace_first requires two string arguments"})
		}
		old, ok1 := args[0].(*StringVal)
		new_, ok2 := args[1].(*StringVal)
		if !ok1 || !ok2 {
			panic(&RuntimeError{Message: "String.replace_first requires string arguments"})
		}
		return &StringVal{Val: strings.Replace(s, old.Val, new_.Val, 1)}
	})

	// slice(start, end?) -> String — Get substring with bounds checking
	RegisterMethod(TypeString, "slice", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		runes := []rune(s)
		strLen := len(runes)

		if len(args) < 1 {
			panic(&RuntimeError{Message: "String.slice requires at least one argument"})
		}

		startVal, ok := args[0].(*IntVal)
		if !ok {
			panic(&RuntimeError{Message: "String.slice requires integer arguments"})
		}
		start := int(startVal.Val)

		// Handle negative indices
		if start < 0 {
			start = strLen + start
		}
		if start < 0 {
			start = 0
		}
		if start > strLen {
			start = strLen
		}

		end := strLen
		if len(args) >= 2 {
			endVal, ok := args[1].(*IntVal)
			if !ok {
				panic(&RuntimeError{Message: "String.slice requires integer arguments"})
			}
			end = int(endVal.Val)
			if end < 0 {
				end = strLen + end
			}
			if end < 0 {
				end = 0
			}
			if end > strLen {
				end = strLen
			}
		}

		if start > end {
			return &StringVal{Val: ""}
		}

		return &StringVal{Val: string(runes[start:end])}
	})

	// index_of(sub) -> Option[Int] — Find index of substring, returns Some(i) or None
	RegisterMethod(TypeString, "index_of", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		if len(args) < 1 {
			panic(&RuntimeError{Message: "String.index_of requires a string argument"})
		}
		sub, ok := args[0].(*StringVal)
		if !ok {
			panic(&RuntimeError{Message: "String.index_of requires a string argument"})
		}
		// Find byte index, then convert to rune index
		byteIdx := strings.Index(s, sub.Val)
		if byteIdx < 0 {
			return &OptionVal{IsSome: false}
		}
		runeIdx := utf8.RuneCountInString(s[:byteIdx])
		return &OptionVal{IsSome: true, Val: &IntVal{Val: int64(runeIdx)}}
	})

	// chars() -> List[String] — Split into individual characters
	RegisterMethod(TypeString, "chars", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		runes := []rune(s)
		elems := make([]Value, len(runes))
		for i, r := range runes {
			elems[i] = &StringVal{Val: string(r)}
		}
		return &ListVal{Elements: elems}
	})

	// join(list) -> String — Join a list of strings with this string as separator
	RegisterMethod(TypeString, "join", func(receiver Value, args []Value) Value {
		sep := receiver.(*StringVal).Val
		if len(args) < 1 {
			panic(&RuntimeError{Message: "String.join requires a list argument"})
		}
		list, ok := args[0].(*ListVal)
		if !ok {
			panic(&RuntimeError{Message: "String.join requires a list argument"})
		}
		parts := make([]string, len(list.Elements))
		for i, elem := range list.Elements {
			parts[i] = elem.String()
		}
		return &StringVal{Val: strings.Join(parts, sep)}
	})

	// repeat(n) -> String — Repeat string n times
	RegisterMethod(TypeString, "repeat", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		if len(args) < 1 {
			panic(&RuntimeError{Message: "String.repeat requires an integer argument"})
		}
		n, ok := args[0].(*IntVal)
		if !ok {
			panic(&RuntimeError{Message: "String.repeat requires an integer argument"})
		}
		if n.Val < 0 {
			panic(&RuntimeError{Message: "String.repeat count cannot be negative"})
		}
		return &StringVal{Val: strings.Repeat(s, int(n.Val))}
	})

	// is_empty() -> Bool — Check if string is empty
	RegisterMethod(TypeString, "is_empty", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		return &BoolVal{Val: len(s) == 0}
	})

	// reverse() -> String — Reverse the string (rune-aware)
	RegisterMethod(TypeString, "reverse", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return &StringVal{Val: string(runes)}
	})

	// pad_left(n, char?) -> String — Left-pad to length n
	RegisterMethod(TypeString, "pad_left", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		if len(args) < 1 {
			panic(&RuntimeError{Message: "String.pad_left requires at least one argument"})
		}
		n, ok := args[0].(*IntVal)
		if !ok {
			panic(&RuntimeError{Message: "String.pad_left requires an integer first argument"})
		}
		padChar := " "
		if len(args) >= 2 {
			if pc, ok := args[1].(*StringVal); ok {
				padChar = pc.Val
			}
		}
		targetLen := int(n.Val)
		currentLen := utf8.RuneCountInString(s)
		if currentLen >= targetLen {
			return &StringVal{Val: s}
		}
		padding := strings.Repeat(padChar, targetLen-currentLen)
		// Trim padding to exact length needed
		padRunes := []rune(padding)
		needed := targetLen - currentLen
		if len(padRunes) > needed {
			padRunes = padRunes[:needed]
		}
		return &StringVal{Val: string(padRunes) + s}
	})

	// pad_right(n, char?) -> String — Right-pad to length n
	RegisterMethod(TypeString, "pad_right", func(receiver Value, args []Value) Value {
		s := receiver.(*StringVal).Val
		if len(args) < 1 {
			panic(&RuntimeError{Message: "String.pad_right requires at least one argument"})
		}
		n, ok := args[0].(*IntVal)
		if !ok {
			panic(&RuntimeError{Message: "String.pad_right requires an integer first argument"})
		}
		padChar := " "
		if len(args) >= 2 {
			if pc, ok := args[1].(*StringVal); ok {
				padChar = pc.Val
			}
		}
		targetLen := int(n.Val)
		currentLen := utf8.RuneCountInString(s)
		if currentLen >= targetLen {
			return &StringVal{Val: s}
		}
		padding := strings.Repeat(padChar, targetLen-currentLen)
		padRunes := []rune(padding)
		needed := targetLen - currentLen
		if len(padRunes) > needed {
			padRunes = padRunes[:needed]
		}
		return &StringVal{Val: s + string(padRunes)}
	})
}
