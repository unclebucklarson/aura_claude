package interpreter

// stdlib_log.go implements the std.log module for structured logging.
// All functions use the LogProvider from the effect system, enabling mockable logging.

// createStdLogExports creates the exports for the std.log module.
func createStdLogExports(provider LogProvider) map[string]Value {
	exports := make(map[string]Value)

	// Helper to convert an optional Aura MapVal to Go map[string]interface{}
	extractContext := func(args []Value, idx int) map[string]interface{} {
		ctx := make(map[string]interface{})
		if idx < len(args) {
			if m, ok := args[idx].(*MapVal); ok {
				for i, k := range m.Keys {
					if ks, ok := k.(*StringVal); ok {
						ctx[ks.Val] = auraValueToGo(m.Values[i])
					}
				}
			}
		}
		return ctx
	}

	// info(message) -> None
	// info(message, context) -> None
	exports["info"] = &BuiltinFnVal{
		Name: "log.info",
		Fn: func(args []Value) Value {
			if len(args) < 1 || len(args) > 2 {
				panic(&RuntimeError{Message: "log.info() requires 1-2 arguments (message, context?)"})
			}
			msg, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "log.info() first argument must be a string"})
			}
			ctx := extractContext(args, 1)
			provider.Info(msg.Val, ctx)
			return &NoneVal{}
		},
	}

	// warn(message) -> None
	// warn(message, context) -> None
	exports["warn"] = &BuiltinFnVal{
		Name: "log.warn",
		Fn: func(args []Value) Value {
			if len(args) < 1 || len(args) > 2 {
				panic(&RuntimeError{Message: "log.warn() requires 1-2 arguments (message, context?)"})
			}
			msg, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "log.warn() first argument must be a string"})
			}
			ctx := extractContext(args, 1)
			provider.Warn(msg.Val, ctx)
			return &NoneVal{}
		},
	}

	// error(message) -> None
	// error(message, context) -> None
	exports["error"] = &BuiltinFnVal{
		Name: "log.error",
		Fn: func(args []Value) Value {
			if len(args) < 1 || len(args) > 2 {
				panic(&RuntimeError{Message: "log.error() requires 1-2 arguments (message, context?)"})
			}
			msg, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "log.error() first argument must be a string"})
			}
			ctx := extractContext(args, 1)
			provider.Error(msg.Val, ctx)
			return &NoneVal{}
		},
	}

	// debug(message) -> None
	// debug(message, context) -> None
	exports["debug"] = &BuiltinFnVal{
		Name: "log.debug",
		Fn: func(args []Value) Value {
			if len(args) < 1 || len(args) > 2 {
				panic(&RuntimeError{Message: "log.debug() requires 1-2 arguments (message, context?)"})
			}
			msg, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "log.debug() first argument must be a string"})
			}
			ctx := extractContext(args, 1)
			provider.Debug(msg.Val, ctx)
			return &NoneVal{}
		},
	}

	// with_context(context_map, fn) -> Any
	// Adds context to all log calls within the function scope.
	// Note: This is a convenience wrapper; actual context threading is manual.
	exports["with_context"] = &BuiltinFnVal{
		Name: "log.with_context",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "log.with_context() requires 2 arguments (context_map, fn)"})
			}
			// For now, just call the function - context is added by the caller
			// The context_map is available for logging within the function
			return callValue(args[1], []Value{args[0]})
		},
	}

	// get_logs() -> List[Map]
	// Returns all logged entries (mainly for testing with mock provider).
	exports["get_logs"] = &BuiltinFnVal{
		Name: "log.get_logs",
		Fn: func(args []Value) Value {
			if len(args) != 0 {
				panic(&RuntimeError{Message: "log.get_logs() takes no arguments"})
			}
			logs := provider.GetLogs()
			elements := make([]Value, len(logs))
			for i, entry := range logs {
				ctxMap := &MapVal{
					Keys:   make([]Value, 0),
					Values: make([]Value, 0),
				}
				for k, v := range entry.Context {
					ctxMap.Keys = append(ctxMap.Keys, &StringVal{Val: k})
					ctxMap.Values = append(ctxMap.Values, goValueToAura(v))
				}

				elements[i] = &MapVal{
					Keys: []Value{
						&StringVal{Val: "level"},
						&StringVal{Val: "message"},
						&StringVal{Val: "context"},
						&StringVal{Val: "timestamp"},
					},
					Values: []Value{
						&StringVal{Val: entry.Level},
						&StringVal{Val: entry.Message},
						ctxMap,
						&IntVal{Val: entry.Timestamp},
					},
				}
			}
			return &ListVal{Elements: elements}
		},
	}

	return exports
}

// auraValueToGo converts an Aura Value to a Go interface{} for log context.
func auraValueToGo(v Value) interface{} {
	switch val := v.(type) {
	case *IntVal:
		return val.Val
	case *FloatVal:
		return val.Val
	case *StringVal:
		return val.Val
	case *BoolVal:
		return val.Val
	case *NoneVal:
		return nil
	default:
		return v.String()
	}
}

// goValueToAura converts a Go interface{} back to an Aura Value.
func goValueToAura(v interface{}) Value {
	switch val := v.(type) {
	case int64:
		return &IntVal{Val: val}
	case float64:
		return &FloatVal{Val: val}
	case string:
		return &StringVal{Val: val}
	case bool:
		if val {
			return &BoolVal{Val: true}
		}
		return &BoolVal{Val: false}
	case nil:
		return &NoneVal{}
	default:
		return &StringVal{Val: "<unknown>"}
	}
}
