package interpreter

import (
        "fmt"
        "strings"
)

// TestRegistration holds a registered test from std.testing.
type TestRegistration struct {
        Name string
        Fn   Value
}

// Global test registry for std.testing
var registeredTests []TestRegistration

// createStdTestingExports creates exports for the std.testing module.
func createStdTestingExports() map[string]Value {
        exports := make(map[string]Value)

        // assert(condition, message?) - Assertion function
        exports["assert"] = &BuiltinFnVal{
                Name: "testing.assert",
                Fn: func(args []Value) Value {
                        if len(args) < 1 || len(args) > 2 {
                                panic(&RuntimeError{Message: "testing.assert() requires 1-2 arguments (condition, message?)"})
                        }
                        if !IsTruthy(args[0]) {
                                msg := "assertion failed"
                                if len(args) >= 2 {
                                        if s, ok := args[1].(*StringVal); ok {
                                                msg = s.Val
                                        } else {
                                                msg = "assertion failed: " + args[1].String()
                                        }
                                }
                                panic(&RuntimeError{Message: msg})
                        }
                        return &BoolVal{Val: true}
                },
        }

        // assert_eq(actual, expected) - Equality assertion
        exports["assert_eq"] = &BuiltinFnVal{
                Name: "testing.assert_eq",
                Fn: func(args []Value) Value {
                        if len(args) < 2 || len(args) > 3 {
                                panic(&RuntimeError{Message: "testing.assert_eq() requires 2-3 arguments (actual, expected, message?)"})
                        }
                        actual := args[0]
                        expected := args[1]
                        if !Equal(actual, expected) {
                                msg := fmt.Sprintf("assertion failed: expected %s, got %s", valueRepr(expected), valueRepr(actual))
                                if len(args) >= 3 {
                                        if s, ok := args[2].(*StringVal); ok {
                                                msg = fmt.Sprintf("%s: expected %s, got %s", s.Val, valueRepr(expected), valueRepr(actual))
                                        }
                                }
                                panic(&RuntimeError{Message: msg})
                        }
                        return &BoolVal{Val: true}
                },
        }

        // assert_ne(actual, expected) - Inequality assertion
        exports["assert_ne"] = &BuiltinFnVal{
                Name: "testing.assert_ne",
                Fn: func(args []Value) Value {
                        if len(args) < 2 || len(args) > 3 {
                                panic(&RuntimeError{Message: "testing.assert_ne() requires 2-3 arguments (actual, expected, message?)"})
                        }
                        actual := args[0]
                        unexpected := args[1]
                        if Equal(actual, unexpected) {
                                msg := fmt.Sprintf("assertion failed: values should not be equal, got %s", valueRepr(actual))
                                if len(args) >= 3 {
                                        if s, ok := args[2].(*StringVal); ok {
                                                msg = fmt.Sprintf("%s: values should not be equal, got %s", s.Val, valueRepr(actual))
                                        }
                                }
                                panic(&RuntimeError{Message: msg})
                        }
                        return &BoolVal{Val: true}
                },
        }

        // assert_true(value) - Assert value is truthy
        exports["assert_true"] = &BuiltinFnVal{
                Name: "testing.assert_true",
                Fn: func(args []Value) Value {
                        if len(args) < 1 || len(args) > 2 {
                                panic(&RuntimeError{Message: "testing.assert_true() requires 1-2 arguments"})
                        }
                        if !IsTruthy(args[0]) {
                                msg := fmt.Sprintf("expected truthy value, got %s", valueRepr(args[0]))
                                if len(args) >= 2 {
                                        if s, ok := args[1].(*StringVal); ok {
                                                msg = s.Val
                                        }
                                }
                                panic(&RuntimeError{Message: msg})
                        }
                        return &BoolVal{Val: true}
                },
        }

        // assert_false(value) - Assert value is falsy
        exports["assert_false"] = &BuiltinFnVal{
                Name: "testing.assert_false",
                Fn: func(args []Value) Value {
                        if len(args) < 1 || len(args) > 2 {
                                panic(&RuntimeError{Message: "testing.assert_false() requires 1-2 arguments"})
                        }
                        if IsTruthy(args[0]) {
                                msg := fmt.Sprintf("expected falsy value, got %s", valueRepr(args[0]))
                                if len(args) >= 2 {
                                        if s, ok := args[1].(*StringVal); ok {
                                                msg = s.Val
                                        }
                                }
                                panic(&RuntimeError{Message: msg})
                        }
                        return &BoolVal{Val: true}
                },
        }

        // assert_none(value) - Assert value is None
        exports["assert_none"] = &BuiltinFnVal{
                Name: "testing.assert_none",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "testing.assert_none() requires exactly 1 argument"})
                        }
                        isNone := false
                        switch v := args[0].(type) {
                        case *NoneVal:
                                isNone = true
                        case *OptionVal:
                                isNone = !v.IsSome
                        }
                        if !isNone {
                                panic(&RuntimeError{Message: fmt.Sprintf("expected None, got %s", valueRepr(args[0]))})
                        }
                        return &BoolVal{Val: true}
                },
        }

        // assert_some(value) - Assert value is Some
        exports["assert_some"] = &BuiltinFnVal{
                Name: "testing.assert_some",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "testing.assert_some() requires exactly 1 argument"})
                        }
                        opt, ok := args[0].(*OptionVal)
                        if !ok || !opt.IsSome {
                                panic(&RuntimeError{Message: fmt.Sprintf("expected Some, got %s", valueRepr(args[0]))})
                        }
                        return opt.Val
                },
        }

        // assert_ok(value) - Assert value is Ok result
        exports["assert_ok"] = &BuiltinFnVal{
                Name: "testing.assert_ok",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "testing.assert_ok() requires exactly 1 argument"})
                        }
                        res, ok := args[0].(*ResultVal)
                        if !ok || !res.IsOk {
                                panic(&RuntimeError{Message: fmt.Sprintf("expected Ok, got %s", valueRepr(args[0]))})
                        }
                        return res.Val
                },
        }

        // assert_err(value) - Assert value is Err result
        exports["assert_err"] = &BuiltinFnVal{
                Name: "testing.assert_err",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "testing.assert_err() requires exactly 1 argument"})
                        }
                        res, ok := args[0].(*ResultVal)
                        if !ok || res.IsOk {
                                panic(&RuntimeError{Message: fmt.Sprintf("expected Err, got %s", valueRepr(args[0]))})
                        }
                        return res.Val
                },
        }

        // test(name, fn) - Test registration
        exports["test"] = &BuiltinFnVal{
                Name: "testing.test",
                Fn: func(args []Value) Value {
                        if len(args) != 2 {
                                panic(&RuntimeError{Message: "testing.test() requires 2 arguments (name, fn)"})
                        }
                        name, ok := args[0].(*StringVal)
                        if !ok {
                                panic(&RuntimeError{Message: "testing.test() first argument must be a string"})
                        }
                        registeredTests = append(registeredTests, TestRegistration{
                                Name: name.Val,
                                Fn:   args[1],
                        })
                        return &NoneVal{}
                },
        }

        // run_tests() - Run all registered tests and return results
        exports["run_tests"] = &BuiltinFnVal{
                Name: "testing.run_tests",
                Fn: func(args []Value) Value {
                        results := make([]Value, 0, len(registeredTests))
                        for _, test := range registeredTests {
                                passed := true
                                errMsg := ""

                                func() {
                                        defer func() {
                                                if r := recover(); r != nil {
                                                        passed = false
                                                        switch e := r.(type) {
                                                        case *RuntimeError:
                                                                errMsg = e.Message
                                                        default:
                                                                errMsg = fmt.Sprintf("%v", r)
                                                        }
                                                }
                                        }()
                                        // Call the test function
                                        callValue(test.Fn, nil)
                                }()

                                result := &MapVal{
                                        Keys: []Value{
                                                &StringVal{Val: "name"},
                                                &StringVal{Val: "passed"},
                                                &StringVal{Val: "error"},
                                        },
                                        Values: []Value{
                                                &StringVal{Val: test.Name},
                                                &BoolVal{Val: passed},
                                                &StringVal{Val: errMsg},
                                        },
                                }
                                results = append(results, result)
                        }
                        // Clear registered tests after running
                        registeredTests = nil
                        return &ListVal{Elements: results}
                },
        }

        return exports
}

// createStdTestingEffectExports creates effect-aware testing exports.
// These are added to std.testing when an interpreter has an EffectContext.
func createStdTestingEffectExports(interp *Interpreter) map[string]Value {
        exports := make(map[string]Value)

        // with_mock_effects(fn) - Run function with a fresh mock effect context
        exports["with_mock_effects"] = &BuiltinFnVal{
                Name: "testing.with_mock_effects",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "testing.with_mock_effects() requires exactly 1 argument (fn)"})
                        }
                        // Save current effects
                        oldEffects := interp.effects
                        // Swap to mock effects
                        interp.effects = NewMockEffectContext()
                        defer func() {
                                interp.effects = oldEffects
                        }()
                        // Call the function
                        return callValue(args[0], nil)
                },
        }

        // with_effects(effects_map, fn) - Run function with custom effects configuration
        // effects_map can have keys: "time" (int), "files" (map), "env" (map), "cwd" (string), "args" (list)
        exports["with_effects"] = &BuiltinFnVal{
                Name: "testing.with_effects",
                Fn: func(args []Value) Value {
                        if len(args) != 2 {
                                panic(&RuntimeError{Message: "testing.with_effects() requires 2 arguments (config_map, fn)"})
                        }
                        configMap, ok := args[0].(*MapVal)
                        if !ok {
                                panic(&RuntimeError{Message: "testing.with_effects() first argument must be a map"})
                        }

                        mb := NewMockBuilder()
                        for i, key := range configMap.Keys {
                                keyStr, ok := key.(*StringVal)
                                if !ok {
                                        continue
                                }
                                val := configMap.Values[i]
                                switch keyStr.Val {
                                case "time":
                                        if t, ok := val.(*IntVal); ok {
                                                mb.WithTime(t.Val)
                                        }
                                case "files":
                                        if fm, ok := val.(*MapVal); ok {
                                                for j, fk := range fm.Keys {
                                                        if fks, ok := fk.(*StringVal); ok {
                                                                if fvs, ok := fm.Values[j].(*StringVal); ok {
                                                                        mb.WithFile(fks.Val, fvs.Val)
                                                                }
                                                        }
                                                }
                                        }
                                case "env":
                                        if em, ok := val.(*MapVal); ok {
                                                for j, ek := range em.Keys {
                                                        if eks, ok := ek.(*StringVal); ok {
                                                                if evs, ok := em.Values[j].(*StringVal); ok {
                                                                        mb.WithEnvVar(eks.Val, evs.Val)
                                                                }
                                                        }
                                                }
                                        }
                                case "cwd":
                                        if s, ok := val.(*StringVal); ok {
                                                mb.WithCwd(s.Val)
                                        }
                                case "args":
                                        if l, ok := val.(*ListVal); ok {
                                                strs := make([]string, 0, len(l.Elements))
                                                for _, el := range l.Elements {
                                                        if s, ok := el.(*StringVal); ok {
                                                                strs = append(strs, s.Val)
                                                        }
                                                }
                                                mb.WithArgs(strs)
                                        }
                                }
                        }

                        oldEffects := interp.effects
                        interp.effects = mb.Build()
                        defer func() {
                                interp.effects = oldEffects
                        }()
                        return callValue(args[1], nil)
                },
        }

        // assert_file_exists(path) - Assert file exists in mock filesystem
        exports["assert_file_exists"] = &BuiltinFnVal{
                Name: "testing.assert_file_exists",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "testing.assert_file_exists() requires exactly 1 argument (path)"})
                        }
                        path, ok := args[0].(*StringVal)
                        if !ok {
                                panic(&RuntimeError{Message: "testing.assert_file_exists() argument must be a string"})
                        }
                        if !interp.effects.File().Exists(path.Val) {
                                panic(&RuntimeError{Message: fmt.Sprintf("assertion failed: file '%s' does not exist", path.Val)})
                        }
                        return &BoolVal{Val: true}
                },
        }

        // assert_file_content(path, expected) - Assert file has expected content
        exports["assert_file_content"] = &BuiltinFnVal{
                Name: "testing.assert_file_content",
                Fn: func(args []Value) Value {
                        if len(args) != 2 {
                                panic(&RuntimeError{Message: "testing.assert_file_content() requires 2 arguments (path, expected)"})
                        }
                        path, ok := args[0].(*StringVal)
                        if !ok {
                                panic(&RuntimeError{Message: "testing.assert_file_content() first argument must be a string"})
                        }
                        expected, ok := args[1].(*StringVal)
                        if !ok {
                                panic(&RuntimeError{Message: "testing.assert_file_content() second argument must be a string"})
                        }
                        content, err := interp.effects.File().ReadFile(path.Val)
                        if err != nil {
                                panic(&RuntimeError{Message: fmt.Sprintf("assertion failed: cannot read file '%s': %v", path.Val, err)})
                        }
                        if content != expected.Val {
                                panic(&RuntimeError{Message: fmt.Sprintf("assertion failed: file '%s' content mismatch\nexpected: %q\ngot:      %q", path.Val, expected.Val, content)})
                        }
                        return &BoolVal{Val: true}
                },
        }

        // assert_env_var(key, expected) - Assert environment variable value
        exports["assert_env_var"] = &BuiltinFnVal{
                Name: "testing.assert_env_var",
                Fn: func(args []Value) Value {
                        if len(args) != 2 {
                                panic(&RuntimeError{Message: "testing.assert_env_var() requires 2 arguments (key, expected)"})
                        }
                        key, ok := args[0].(*StringVal)
                        if !ok {
                                panic(&RuntimeError{Message: "testing.assert_env_var() first argument must be a string"})
                        }
                        expected, ok := args[1].(*StringVal)
                        if !ok {
                                panic(&RuntimeError{Message: "testing.assert_env_var() second argument must be a string"})
                        }
                        val, exists := interp.effects.Env().Get(key.Val)
                        if !exists {
                                panic(&RuntimeError{Message: fmt.Sprintf("assertion failed: env var '%s' does not exist", key.Val)})
                        }
                        if val != expected.Val {
                                panic(&RuntimeError{Message: fmt.Sprintf("assertion failed: env var '%s' expected %q, got %q", key.Val, expected.Val, val)})
                        }
                        return &BoolVal{Val: true}
                },
        }

        // mock_time(timestamp) - Set mock time to specific value
        exports["mock_time"] = &BuiltinFnVal{
                Name: "testing.mock_time",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "testing.mock_time() requires exactly 1 argument (timestamp)"})
                        }
                        ts, ok := args[0].(*IntVal)
                        if !ok {
                                panic(&RuntimeError{Message: "testing.mock_time() argument must be an integer"})
                        }
                        if tp, ok := interp.effects.Time().(*MockTimeProvider); ok {
                                tp.SetTime(ts.Val)
                        } else {
                                panic(&RuntimeError{Message: "testing.mock_time() requires mock time provider"})
                        }
                        return &NoneVal{}
                },
        }

        // advance_time(seconds) - Advance mock time by seconds
        exports["advance_time"] = &BuiltinFnVal{
                Name: "testing.advance_time",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "testing.advance_time() requires exactly 1 argument (seconds)"})
                        }
                        secs, ok := args[0].(*IntVal)
                        if !ok {
                                panic(&RuntimeError{Message: "testing.advance_time() argument must be an integer"})
                        }
                        if tp, ok := interp.effects.Time().(*MockTimeProvider); ok {
                                current := tp.Now()
                                tp.SetTime(current + secs.Val)
                        } else {
                                panic(&RuntimeError{Message: "testing.advance_time() requires mock time provider"})
                        }
                        return &NoneVal{}
                },
        }

        // reset_effects() - Reset all mock effects to clean state
        exports["reset_effects"] = &BuiltinFnVal{
                Name: "testing.reset_effects",
                Fn: func(args []Value) Value {
                        interp.effects = NewMockEffectContext()
                        return &NoneVal{}
                },
        }

        // assert_file_contains(path, substring) - Assert file contains a substring
        exports["assert_file_contains"] = &BuiltinFnVal{
                Name: "testing.assert_file_contains",
                Fn: func(args []Value) Value {
                        if len(args) != 2 {
                                panic(&RuntimeError{Message: "testing.assert_file_contains() requires 2 arguments (path, substring)"})
                        }
                        path, ok := args[0].(*StringVal)
                        if !ok {
                                panic(&RuntimeError{Message: "testing.assert_file_contains() first argument must be a string"})
                        }
                        substr, ok := args[1].(*StringVal)
                        if !ok {
                                panic(&RuntimeError{Message: "testing.assert_file_contains() second argument must be a string"})
                        }
                        content, err := interp.effects.File().ReadFile(path.Val)
                        if err != nil {
                                panic(&RuntimeError{Message: fmt.Sprintf("assertion failed: cannot read file '%s': %v", path.Val, err)})
                        }
                        if !strings.Contains(content, substr.Val) {
                                panic(&RuntimeError{Message: fmt.Sprintf("assertion failed: file '%s' does not contain %q", path.Val, substr.Val)})
                        }
                        return &BoolVal{Val: true}
                },
        }

        // assert_no_file(path) - Assert file does NOT exist
        exports["assert_no_file"] = &BuiltinFnVal{
                Name: "testing.assert_no_file",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "testing.assert_no_file() requires exactly 1 argument (path)"})
                        }
                        path, ok := args[0].(*StringVal)
                        if !ok {
                                panic(&RuntimeError{Message: "testing.assert_no_file() argument must be a string"})
                        }
                        if interp.effects.File().Exists(path.Val) {
                                panic(&RuntimeError{Message: fmt.Sprintf("assertion failed: file '%s' should not exist", path.Val)})
                        }
                        return &BoolVal{Val: true}
                },
        }

        // get_mock_time() - Get the current mock time as an integer
        exports["get_mock_time"] = &BuiltinFnVal{
                Name: "testing.get_mock_time",
                Fn: func(args []Value) Value {
                        return &IntVal{Val: interp.effects.Time().Now()}
                },
        }

        // get_env(key) - Get env var value (returns Option)
        exports["get_env"] = &BuiltinFnVal{
                Name: "testing.get_env",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "testing.get_env() requires exactly 1 argument (key)"})
                        }
                        key, ok := args[0].(*StringVal)
                        if !ok {
                                panic(&RuntimeError{Message: "testing.get_env() argument must be a string"})
                        }
                        val, exists := interp.effects.Env().Get(key.Val)
                        if !exists {
                                return &OptionVal{IsSome: false}
                        }
                        return &OptionVal{IsSome: true, Val: &StringVal{Val: val}}
                },
        }

        return exports
}

// Suppress unused import warning
var _ = strings.Contains
