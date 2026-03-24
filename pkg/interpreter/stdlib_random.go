package interpreter

import (
	"fmt"
	"math/rand"
)

// createStdRandomExports creates exports for the std.random module.
func createStdRandomExports() map[string]Value {
	exports := make(map[string]Value)

	// int(min, max) - Random integer in range [min, max]
	exports["int"] = &BuiltinFnVal{
		Name: "random.int",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "random.int() requires exactly 2 arguments (min, max)"})
			}
			minV, ok1 := args[0].(*IntVal)
			maxV, ok2 := args[1].(*IntVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "random.int() requires integer arguments"})
			}
			if minV.Val > maxV.Val {
				panic(&RuntimeError{Message: "random.int() min must be <= max"})
			}
			if minV.Val == maxV.Val {
				return &IntVal{Val: minV.Val}
			}
			result := minV.Val + rand.Int63n(maxV.Val-minV.Val+1)
			return &IntVal{Val: result}
		},
	}

	// float() - Random float [0.0, 1.0)
	exports["float"] = &BuiltinFnVal{
		Name: "random.float",
		Fn: func(args []Value) Value {
			if len(args) != 0 {
				panic(&RuntimeError{Message: "random.float() takes no arguments"})
			}
			return &FloatVal{Val: rand.Float64()}
		},
	}

	// choice(list) - Random element from list
	exports["choice"] = &BuiltinFnVal{
		Name: "random.choice",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "random.choice() requires exactly 1 argument (list)"})
			}
			list, ok := args[0].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "random.choice() requires a list argument"})
			}
			if len(list.Elements) == 0 {
				panic(&RuntimeError{Message: "random.choice() cannot choose from empty list"})
			}
			idx := rand.Intn(len(list.Elements))
			return list.Elements[idx]
		},
	}

	// shuffle(list) - Shuffle list randomly (returns new list)
	exports["shuffle"] = &BuiltinFnVal{
		Name: "random.shuffle",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "random.shuffle() requires exactly 1 argument (list)"})
			}
			list, ok := args[0].(*ListVal)
			if !ok {
				panic(&RuntimeError{Message: "random.shuffle() requires a list argument"})
			}
			// Create a copy
			newElems := make([]Value, len(list.Elements))
			copy(newElems, list.Elements)
			rand.Shuffle(len(newElems), func(i, j int) {
				newElems[i], newElems[j] = newElems[j], newElems[i]
			})
			return &ListVal{Elements: newElems}
		},
	}

	// sample(list, n) - Random sample of n elements (without replacement)
	exports["sample"] = &BuiltinFnVal{
		Name: "random.sample",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "random.sample() requires exactly 2 arguments (list, n)"})
			}
			list, ok1 := args[0].(*ListVal)
			n, ok2 := args[1].(*IntVal)
			if !ok1 || !ok2 {
				panic(&RuntimeError{Message: "random.sample() requires (list, int) arguments"})
			}
			count := int(n.Val)
			if count < 0 {
				panic(&RuntimeError{Message: "random.sample() n must be non-negative"})
			}
			if count > len(list.Elements) {
				panic(&RuntimeError{Message: fmt.Sprintf("random.sample() n (%d) > list length (%d)", count, len(list.Elements))})
			}
			// Fisher-Yates partial shuffle
			indices := make([]int, len(list.Elements))
			for i := range indices {
				indices[i] = i
			}
			for i := 0; i < count; i++ {
				j := i + rand.Intn(len(indices)-i)
				indices[i], indices[j] = indices[j], indices[i]
			}
			result := make([]Value, count)
			for i := 0; i < count; i++ {
				result[i] = list.Elements[indices[i]]
			}
			return &ListVal{Elements: result}
		},
	}

	// seed(value) - Set random seed
	exports["seed"] = &BuiltinFnVal{
		Name: "random.seed",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "random.seed() requires exactly 1 argument"})
			}
			v, ok := args[0].(*IntVal)
			if !ok {
				panic(&RuntimeError{Message: "random.seed() requires an integer argument"})
			}
			rand.Seed(v.Val)
			return &NoneVal{}
		},
	}

	return exports
}
