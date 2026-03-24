package interpreter

import (
        "fmt"
        "sort"

        "github.com/unclebucklarson/aura/pkg/token"
)

// callValue invokes a callable value (FunctionVal, LambdaVal, BuiltinFnVal)
// with the given arguments. Panics with RuntimeError if not callable.
func callValue(fn Value, args []Value) Value {
        return callFunction(token.Span{}, fn, args, nil)
}

func init() {
        registerListMethods()
}

func registerListMethods() {
        // len() -> Int — Get list length
        RegisterMethod(TypeList, "len", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                return &IntVal{Val: int64(len(list.Elements))}
        })

        // length() -> Int — Alias for len
        RegisterMethod(TypeList, "length", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                return &IntVal{Val: int64(len(list.Elements))}
        })

        // append(item) -> None — Append item to list (mutates)
        RegisterMethod(TypeList, "append", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) < 1 {
                        panic(&RuntimeError{Message: "List.append requires at least one argument"})
                }
                list.Elements = append(list.Elements, args[0])
                return &NoneVal{}
        })

        // push(item) -> None — Alias for append (mutates)
        RegisterMethod(TypeList, "push", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) < 1 {
                        panic(&RuntimeError{Message: "List.push requires at least one argument"})
                }
                list.Elements = append(list.Elements, args[0])
                return &NoneVal{}
        })

        // contains(item) -> Bool — Check if list contains item
        RegisterMethod(TypeList, "contains", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) < 1 {
                        return &BoolVal{Val: false}
                }
                for _, elem := range list.Elements {
                        if Equal(elem, args[0]) {
                                return &BoolVal{Val: true}
                        }
                }
                return &BoolVal{Val: false}
        })

        // is_empty() -> Bool — Check if list is empty
        RegisterMethod(TypeList, "is_empty", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                return &BoolVal{Val: len(list.Elements) == 0}
        })

        // first() -> Option[T] — Get first element
        RegisterMethod(TypeList, "first", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(list.Elements) == 0 {
                        return &OptionVal{IsSome: false}
                }
                return &OptionVal{IsSome: true, Val: list.Elements[0]}
        })

        // last() -> Option[T] — Get last element
        RegisterMethod(TypeList, "last", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(list.Elements) == 0 {
                        return &OptionVal{IsSome: false}
                }
                return &OptionVal{IsSome: true, Val: list.Elements[len(list.Elements)-1]}
        })

        // get(index) -> Option[T] — Safe index access
        RegisterMethod(TypeList, "get", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) < 1 {
                        panic(&RuntimeError{Message: "List.get requires an index argument"})
                }
                idx, ok := args[0].(*IntVal)
                if !ok {
                        panic(&RuntimeError{Message: "List.get index must be an Int"})
                }
                i := int(idx.Val)
                if i < 0 {
                        i = len(list.Elements) + i
                }
                if i < 0 || i >= len(list.Elements) {
                        return &OptionVal{IsSome: false}
                }
                return &OptionVal{IsSome: true, Val: list.Elements[i]}
        })

        // pop() -> Option[T] — Remove and return last element (mutates)
        RegisterMethod(TypeList, "pop", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(list.Elements) == 0 {
                        return &OptionVal{IsSome: false}
                }
                last := list.Elements[len(list.Elements)-1]
                list.Elements = list.Elements[:len(list.Elements)-1]
                return &OptionVal{IsSome: true, Val: last}
        })

        // remove(index) -> T — Remove element at index (mutates)
        RegisterMethod(TypeList, "remove", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) < 1 {
                        panic(&RuntimeError{Message: "List.remove requires an index argument"})
                }
                idx, ok := args[0].(*IntVal)
                if !ok {
                        panic(&RuntimeError{Message: "List.remove index must be an Int"})
                }
                i := int(idx.Val)
                if i < 0 {
                        i = len(list.Elements) + i
                }
                if i < 0 || i >= len(list.Elements) {
                        panic(&RuntimeError{Message: "List.remove index out of bounds"})
                }
                elem := list.Elements[i]
                list.Elements = append(list.Elements[:i], list.Elements[i+1:]...)
                return elem
        })

        // reverse() -> List — Reverse list (returns new list)
        RegisterMethod(TypeList, "reverse", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                n := len(list.Elements)
                elems := make([]Value, n)
                for i, e := range list.Elements {
                        elems[n-1-i] = e
                }
                return &ListVal{Elements: elems}
        })

        // slice(start, end?) -> List — Extract sublist with negative index support
        RegisterMethod(TypeList, "slice", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) < 1 {
                        panic(&RuntimeError{Message: "List.slice requires at least a start argument"})
                }
                startArg, ok := args[0].(*IntVal)
                if !ok {
                        panic(&RuntimeError{Message: "List.slice start must be an Int"})
                }
                n := len(list.Elements)
                start := int(startArg.Val)
                if start < 0 {
                        start = n + start
                }
                if start < 0 {
                        start = 0
                }
                if start > n {
                        start = n
                }

                end := n
                if len(args) >= 2 {
                        endArg, ok := args[1].(*IntVal)
                        if !ok {
                                panic(&RuntimeError{Message: "List.slice end must be an Int"})
                        }
                        end = int(endArg.Val)
                        if end < 0 {
                                end = n + end
                        }
                        if end < 0 {
                                end = 0
                        }
                        if end > n {
                                end = n
                        }
                }
                if start > end {
                        start = end
                }

                elems := make([]Value, end-start)
                copy(elems, list.Elements[start:end])
                return &ListVal{Elements: elems}
        })

        // join(separator) -> String — Join elements into string
        RegisterMethod(TypeList, "join", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                sep := ""
                if len(args) >= 1 {
                        sepArg, ok := args[0].(*StringVal)
                        if !ok {
                                panic(&RuntimeError{Message: "List.join separator must be a String"})
                        }
                        sep = sepArg.Val
                }
                parts := make([]string, len(list.Elements))
                for i, e := range list.Elements {
                        parts[i] = e.String()
                }
                result := ""
                for i, p := range parts {
                        if i > 0 {
                                result += sep
                        }
                        result += p
                }
                return &StringVal{Val: result}
        })

        // index_of(item) -> Option[Int] — Find index of first matching element
        RegisterMethod(TypeList, "index_of", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) < 1 {
                        panic(&RuntimeError{Message: "List.index_of requires an argument"})
                }
                for i, elem := range list.Elements {
                        if Equal(elem, args[0]) {
                                return &OptionVal{IsSome: true, Val: &IntVal{Val: int64(i)}}
                        }
                }
                return &OptionVal{IsSome: false}
        })

        // map(fn) -> List — Apply function to each element
        RegisterMethod(TypeList, "map", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) < 1 {
                        panic(&RuntimeError{Message: "List.map requires a function argument"})
                }
                fn := args[0]
                elems := make([]Value, len(list.Elements))
                for i, e := range list.Elements {
                        elems[i] = callValue(fn, []Value{e})
                }
                return &ListVal{Elements: elems}
        })

        // filter(fn) -> List — Keep elements where fn returns true
        RegisterMethod(TypeList, "filter", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) < 1 {
                        panic(&RuntimeError{Message: "List.filter requires a function argument"})
                }
                fn := args[0]
                var elems []Value
                for _, e := range list.Elements {
                        result := callValue(fn, []Value{e})
                        if IsTruthy(result) {
                                elems = append(elems, e)
                        }
                }
                if elems == nil {
                        elems = []Value{}
                }
                return &ListVal{Elements: elems}
        })

        // reduce(init, fn) -> T — Fold list with accumulator
        RegisterMethod(TypeList, "reduce", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) < 2 {
                        panic(&RuntimeError{Message: "List.reduce requires an initial value and a function"})
                }
                acc := args[0]
                fn := args[1]
                for _, e := range list.Elements {
                        acc = callValue(fn, []Value{acc, e})
                }
                return acc
        })

        // for_each(fn) -> None — Execute fn for each element (side effects)
        RegisterMethod(TypeList, "for_each", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) < 1 {
                        panic(&RuntimeError{Message: "List.for_each requires a function argument"})
                }
                fn := args[0]
                for _, e := range list.Elements {
                        callValue(fn, []Value{e})
                }
                return &NoneVal{}
        })

        // flat_map(fn) -> List — Map + flatten one level
        RegisterMethod(TypeList, "flat_map", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) < 1 {
                        panic(&RuntimeError{Message: "List.flat_map requires a function argument"})
                }
                fn := args[0]
                var elems []Value
                for _, e := range list.Elements {
                        result := callValue(fn, []Value{e})
                        if inner, ok := result.(*ListVal); ok {
                                elems = append(elems, inner.Elements...)
                        } else {
                                elems = append(elems, result)
                        }
                }
                if elems == nil {
                        elems = []Value{}
                }
                return &ListVal{Elements: elems}
        })

        // flatten() -> List — Flatten one level of nesting
        RegisterMethod(TypeList, "flatten", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                var elems []Value
                for _, e := range list.Elements {
                        if inner, ok := e.(*ListVal); ok {
                                elems = append(elems, inner.Elements...)
                        } else {
                                elems = append(elems, e)
                        }
                }
                if elems == nil {
                        elems = []Value{}
                }
                return &ListVal{Elements: elems}
        })

        // any(fn) -> Bool — True if any element satisfies predicate
        RegisterMethod(TypeList, "any", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) < 1 {
                        panic(&RuntimeError{Message: "List.any requires a function argument"})
                }
                fn := args[0]
                for _, e := range list.Elements {
                        if IsTruthy(callValue(fn, []Value{e})) {
                                return &BoolVal{Val: true}
                        }
                }
                return &BoolVal{Val: false}
        })

        // all(fn) -> Bool — True if all elements satisfy predicate
        RegisterMethod(TypeList, "all", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) < 1 {
                        panic(&RuntimeError{Message: "List.all requires a function argument"})
                }
                fn := args[0]
                for _, e := range list.Elements {
                        if !IsTruthy(callValue(fn, []Value{e})) {
                                return &BoolVal{Val: false}
                        }
                }
                return &BoolVal{Val: true}
        })

        // count(fn?) -> Int — Count elements (optionally matching predicate)
        RegisterMethod(TypeList, "count", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) == 0 {
                        return &IntVal{Val: int64(len(list.Elements))}
                }
                fn := args[0]
                count := int64(0)
                for _, e := range list.Elements {
                        if IsTruthy(callValue(fn, []Value{e})) {
                                count++
                        }
                }
                return &IntVal{Val: count}
        })

        // unique() -> List — Remove duplicates (preserves order)
        RegisterMethod(TypeList, "unique", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                var elems []Value
                for _, e := range list.Elements {
                        found := false
                        for _, existing := range elems {
                                if Equal(e, existing) {
                                        found = true
                                        break
                                }
                        }
                        if !found {
                                elems = append(elems, e)
                        }
                }
                if elems == nil {
                        elems = []Value{}
                }
                return &ListVal{Elements: elems}
        })

        // sum() -> Int|Float — Sum numeric elements
        RegisterMethod(TypeList, "sum", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(list.Elements) == 0 {
                        return &IntVal{Val: 0}
                }
                hasFloat := false
                floatSum := 0.0
                intSum := int64(0)
                for _, e := range list.Elements {
                        switch v := e.(type) {
                        case *IntVal:
                                intSum += v.Val
                                floatSum += float64(v.Val)
                        case *FloatVal:
                                hasFloat = true
                                floatSum += v.Val
                        default:
                                panic(&RuntimeError{Message: "List.sum requires all elements to be numeric"})
                        }
                }
                if hasFloat {
                        return &FloatVal{Val: floatSum}
                }
                return &IntVal{Val: intSum}
        })

        // min() -> Option[T] — Find minimum element
        RegisterMethod(TypeList, "min", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(list.Elements) == 0 {
                        return &OptionVal{IsSome: false}
                }
                minVal := list.Elements[0]
                for _, e := range list.Elements[1:] {
                        if cmpValues(e, minVal) < 0 {
                                minVal = e
                        }
                }
                return &OptionVal{IsSome: true, Val: minVal}
        })

        // max() -> Option[T] — Find maximum element
        RegisterMethod(TypeList, "max", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(list.Elements) == 0 {
                        return &OptionVal{IsSome: false}
                }
                maxVal := list.Elements[0]
                for _, e := range list.Elements[1:] {
                        if cmpValues(e, maxVal) > 0 {
                                maxVal = e
                        }
                }
                return &OptionVal{IsSome: true, Val: maxVal}
        })

        // sort() -> List — Sort elements (returns new list)
        RegisterMethod(TypeList, "sort", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                elems := make([]Value, len(list.Elements))
                copy(elems, list.Elements)
                sort.SliceStable(elems, func(i, j int) bool {
                        return cmpValues(elems[i], elems[j]) < 0
                })
                return &ListVal{Elements: elems}
        })

        // zip(other) -> List — Zip two lists into list of tuples
        RegisterMethod(TypeList, "zip", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                if len(args) < 1 {
                        panic(&RuntimeError{Message: "List.zip requires a list argument"})
                }
                other, ok := args[0].(*ListVal)
                if !ok {
                        panic(&RuntimeError{Message: "List.zip argument must be a List"})
                }
                n := len(list.Elements)
                if len(other.Elements) < n {
                        n = len(other.Elements)
                }
                elems := make([]Value, n)
                for i := 0; i < n; i++ {
                        elems[i] = &TupleVal{Elements: []Value{list.Elements[i], other.Elements[i]}}
                }
                return &ListVal{Elements: elems}
        })

        // enumerate() -> List — List of (index, element) tuples
        RegisterMethod(TypeList, "enumerate", func(receiver Value, args []Value) Value {
                list := receiver.(*ListVal)
                elems := make([]Value, len(list.Elements))
                for i, e := range list.Elements {
                        elems[i] = &TupleVal{Elements: []Value{&IntVal{Val: int64(i)}, e}}
                }
                return &ListVal{Elements: elems}
        })
}

// cmpValues compares two values for ordering.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
// Panics with RuntimeError if values are not comparable.
func cmpValues(a, b Value) int {
        switch av := a.(type) {
        case *IntVal:
                switch bv := b.(type) {
                case *IntVal:
                        if av.Val < bv.Val {
                                return -1
                        }
                        if av.Val > bv.Val {
                                return 1
                        }
                        return 0
                case *FloatVal:
                        af := float64(av.Val)
                        if af < bv.Val {
                                return -1
                        }
                        if af > bv.Val {
                                return 1
                        }
                        return 0
                }
        case *FloatVal:
                switch bv := b.(type) {
                case *FloatVal:
                        if av.Val < bv.Val {
                                return -1
                        }
                        if av.Val > bv.Val {
                                return 1
                        }
                        return 0
                case *IntVal:
                        bf := float64(bv.Val)
                        if av.Val < bf {
                                return -1
                        }
                        if av.Val > bf {
                                return 1
                        }
                        return 0
                }
        case *StringVal:
                if bv, ok := b.(*StringVal); ok {
                        if av.Val < bv.Val {
                                return -1
                        }
                        if av.Val > bv.Val {
                                return 1
                        }
                        return 0
                }
        }
        panic(&RuntimeError{Message: fmt.Sprintf("cannot compare %s and %s", valueTypeNames[a.Type()], valueTypeNames[b.Type()])})
}
