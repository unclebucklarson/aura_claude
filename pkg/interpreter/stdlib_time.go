package interpreter

import (
	"fmt"
	"time"
)

// createStdTimeExports creates the exports for the std.time module.
// The time provider is captured via closure, enabling effect mocking.
func createStdTimeExports(tp TimeProvider) map[string]Value {
	exports := make(map[string]Value)

	// now() -> Int - Current Unix timestamp in seconds
	exports["now"] = &BuiltinFnVal{
		Name: "time.now",
		Fn: func(args []Value) Value {
			if len(args) != 0 {
				panic(&RuntimeError{Message: "time.now() takes no arguments"})
			}
			return &IntVal{Val: tp.Now()}
		},
	}

	// unix() -> Int - Alias for now()
	exports["unix"] = &BuiltinFnVal{
		Name: "time.unix",
		Fn: func(args []Value) Value {
			if len(args) != 0 {
				panic(&RuntimeError{Message: "time.unix() takes no arguments"})
			}
			return &IntVal{Val: tp.Now()}
		},
	}

	// sleep(ms) -> None - Sleep for milliseconds
	exports["sleep"] = &BuiltinFnVal{
		Name: "time.sleep",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "time.sleep() requires exactly 1 argument (milliseconds)"})
			}
			ms, ok := args[0].(*IntVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("time.sleep() milliseconds must be an Int, got %s", valueTypeNames[args[0].Type()])})
			}
			if ms.Val < 0 {
				panic(&RuntimeError{Message: "time.sleep() milliseconds must be non-negative"})
			}
			tp.Sleep(int(ms.Val))
			return &NoneVal{}
		},
	}

	// millis() -> Int - Current time in milliseconds
	exports["millis"] = &BuiltinFnVal{
		Name: "time.millis",
		Fn: func(args []Value) Value {
			if len(args) != 0 {
				panic(&RuntimeError{Message: "time.millis() takes no arguments"})
			}
			return &IntVal{Val: tp.NowNano() / 1e6}
		},
	}

	// format(timestamp, format) -> String - Format timestamp
	exports["format"] = &BuiltinFnVal{
		Name: "time.format",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "time.format() requires exactly 2 arguments (timestamp, format)"})
			}
			ts, ok := args[0].(*IntVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("time.format() timestamp must be an Int, got %s", valueTypeNames[args[0].Type()])})
			}
			fmtStr, ok := args[1].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("time.format() format must be a String, got %s", valueTypeNames[args[1].Type()])})
			}
			t := time.Unix(ts.Val, 0).UTC()
			goFmt := auraToGoTimeFormat(fmtStr.Val)
			return &StringVal{Val: t.Format(goFmt)}
		},
	}

	// parse(str, format) -> Result[Int, String] - Parse timestamp string
	exports["parse"] = &BuiltinFnVal{
		Name: "time.parse",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "time.parse() requires exactly 2 arguments (str, format)"})
			}
			str, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("time.parse() str must be a String, got %s", valueTypeNames[args[0].Type()])})
			}
			fmtStr, ok := args[1].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("time.parse() format must be a String, got %s", valueTypeNames[args[1].Type()])})
			}
			goFmt := auraToGoTimeFormat(fmtStr.Val)
			t, err := time.Parse(goFmt, str.Val)
			if err != nil {
				return &ResultVal{IsOk: false, Val: &StringVal{Val: err.Error()}}
			}
			return &ResultVal{IsOk: true, Val: &IntVal{Val: t.Unix()}}
		},
	}

	// add(timestamp, seconds) -> Int - Add seconds to timestamp
	exports["add"] = &BuiltinFnVal{
		Name: "time.add",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "time.add() requires exactly 2 arguments (timestamp, seconds)"})
			}
			ts, ok := args[0].(*IntVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("time.add() timestamp must be an Int, got %s", valueTypeNames[args[0].Type()])})
			}
			secs, ok := args[1].(*IntVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("time.add() seconds must be an Int, got %s", valueTypeNames[args[1].Type()])})
			}
			return &IntVal{Val: ts.Val + secs.Val}
		},
	}

	// diff(ts1, ts2) -> Int - Difference in seconds (ts1 - ts2)
	exports["diff"] = &BuiltinFnVal{
		Name: "time.diff",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "time.diff() requires exactly 2 arguments (ts1, ts2)"})
			}
			ts1, ok := args[0].(*IntVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("time.diff() ts1 must be an Int, got %s", valueTypeNames[args[0].Type()])})
			}
			ts2, ok := args[1].(*IntVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("time.diff() ts2 must be an Int, got %s", valueTypeNames[args[1].Type()])})
			}
			return &IntVal{Val: ts1.Val - ts2.Val}
		},
	}

	return exports
}

// auraToGoTimeFormat converts Aura-style time format strings to Go reference time format.
// Aura uses common format tokens:
//   %Y - 4-digit year
//   %m - 2-digit month
//   %d - 2-digit day
//   %H - 2-digit hour (24h)
//   %M - 2-digit minute
//   %S - 2-digit second
//   %Z - timezone abbreviation
func auraToGoTimeFormat(format string) string {
	replacements := []struct{ from, to string }{
		{"%Y", "2006"},
		{"%m", "01"},
		{"%d", "02"},
		{"%H", "15"},
		{"%M", "04"},
		{"%S", "05"},
		{"%Z", "MST"},
	}
	result := format
	for _, r := range replacements {
		result = goStringReplace(result, r.from, r.to)
	}
	return result
}

// goStringReplace replaces all occurrences of old with new in s.
func goStringReplace(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += new
			i += len(old)
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}
