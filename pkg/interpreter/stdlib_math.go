package interpreter

import (
	"fmt"
	"math"
)

// createStdMathExports creates exports for the std.math module.
func createStdMathExports() map[string]Value {
	exports := make(map[string]Value)

	exports["pi"] = &FloatVal{Val: math.Pi}
	exports["e"] = &FloatVal{Val: math.E}
	exports["inf"] = &FloatVal{Val: math.Inf(1)}
	exports["nan"] = &FloatVal{Val: math.NaN()}

	exports["abs"] = &BuiltinFnVal{
		Name: "math.abs",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "math.abs() requires exactly one argument"})
			}
			switch v := args[0].(type) {
			case *IntVal:
				if v.Val < 0 {
					return &IntVal{Val: -v.Val}
				}
				return v
			case *FloatVal:
				return &FloatVal{Val: math.Abs(v.Val)}
			default:
				panic(&RuntimeError{Message: fmt.Sprintf("math.abs() not supported for %s", valueTypeNames[args[0].Type()])})
			}
		},
	}

	exports["max"] = &BuiltinFnVal{
		Name: "math.max",
		Fn: func(args []Value) Value {
			if len(args) < 2 {
				panic(&RuntimeError{Message: "math.max() requires at least 2 arguments"})
			}
			result := args[0]
			for _, a := range args[1:] {
				if compareRaw(a, result) > 0 {
					result = a
				}
			}
			return result
		},
	}

	exports["min"] = &BuiltinFnVal{
		Name: "math.min",
		Fn: func(args []Value) Value {
			if len(args) < 2 {
				panic(&RuntimeError{Message: "math.min() requires at least 2 arguments"})
			}
			result := args[0]
			for _, a := range args[1:] {
				if compareRaw(a, result) < 0 {
					result = a
				}
			}
			return result
		},
	}

	exports["floor"] = &BuiltinFnVal{
		Name: "math.floor",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "math.floor() requires exactly one argument"})
			}
			switch v := args[0].(type) {
			case *FloatVal:
				return &IntVal{Val: int64(math.Floor(v.Val))}
			case *IntVal:
				return v
			default:
				panic(&RuntimeError{Message: fmt.Sprintf("math.floor() not supported for %s", valueTypeNames[args[0].Type()])})
			}
		},
	}

	exports["ceil"] = &BuiltinFnVal{
		Name: "math.ceil",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "math.ceil() requires exactly one argument"})
			}
			switch v := args[0].(type) {
			case *FloatVal:
				return &IntVal{Val: int64(math.Ceil(v.Val))}
			case *IntVal:
				return v
			default:
				panic(&RuntimeError{Message: fmt.Sprintf("math.ceil() not supported for %s", valueTypeNames[args[0].Type()])})
			}
		},
	}

	exports["round"] = &BuiltinFnVal{
		Name: "math.round",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "math.round() requires exactly one argument"})
			}
			switch v := args[0].(type) {
			case *FloatVal:
				return &IntVal{Val: int64(math.Round(v.Val))}
			case *IntVal:
				return v
			default:
				panic(&RuntimeError{Message: fmt.Sprintf("math.round() not supported for %s", valueTypeNames[args[0].Type()])})
			}
		},
	}

	exports["sqrt"] = &BuiltinFnVal{
		Name: "math.sqrt",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "math.sqrt() requires exactly one argument"})
			}
			switch v := args[0].(type) {
			case *FloatVal:
				return &FloatVal{Val: math.Sqrt(v.Val)}
			case *IntVal:
				return &FloatVal{Val: math.Sqrt(float64(v.Val))}
			default:
				panic(&RuntimeError{Message: fmt.Sprintf("math.sqrt() not supported for %s", valueTypeNames[args[0].Type()])})
			}
		},
	}

	exports["pow"] = &BuiltinFnVal{
		Name: "math.pow",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "math.pow() requires exactly 2 arguments"})
			}
			base := toFloat(args[0], "math.pow()")
			exp := toFloat(args[1], "math.pow()")
			return &FloatVal{Val: math.Pow(base, exp)}
		},
	}

	return exports
}

// toFloat converts a numeric value to float64.
func toFloat(v Value, context string) float64 {
	switch val := v.(type) {
	case *FloatVal:
		return val.Val
	case *IntVal:
		return float64(val.Val)
	default:
		panic(&RuntimeError{Message: fmt.Sprintf("%s requires numeric argument, got %s", context, valueTypeNames[v.Type()])})
	}
}
