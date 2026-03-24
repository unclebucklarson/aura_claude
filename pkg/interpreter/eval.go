package interpreter

import (
        "fmt"
        "math"
        "strconv"
        "strings"

        "github.com/unclebucklarson/aura/pkg/ast"
        "github.com/unclebucklarson/aura/pkg/token"
)

// --- Control flow signals ---

type returnSignal struct{ val Value }
type breakSignal struct{}
type continueSignal struct{}

// RuntimeError represents a runtime error with location info.
type RuntimeError struct {
        Message string
        Span    token.Span
}

func (e *RuntimeError) Error() string {
        if e.Span.File != "" {
                return fmt.Sprintf("%s:%d:%d: runtime error: %s", e.Span.File, e.Span.Start.Line, e.Span.Start.Column, e.Message)
        }
        return fmt.Sprintf("runtime error: %s", e.Message)
}

// runtimePanic panics with a RuntimeError.
func runtimePanic(span token.Span, format string, args ...interface{}) {
        panic(&RuntimeError{Message: fmt.Sprintf(format, args...), Span: span})
}

// --- Evaluator ---

// EvalExpr evaluates an expression and returns a Value.
func EvalExpr(expr ast.Expr, env *Environment) Value {
        if expr == nil {
                return &NoneVal{}
        }
        switch e := expr.(type) {
        case *ast.IntLiteral:
                val, err := strconv.ParseInt(e.Value, 10, 64)
                if err != nil {
                        runtimePanic(e.Span, "invalid integer literal: %s", e.Value)
                }
                return &IntVal{Val: val}

        case *ast.FloatLiteral:
                val, err := strconv.ParseFloat(e.Value, 64)
                if err != nil {
                        runtimePanic(e.Span, "invalid float literal: %s", e.Value)
                }
                return &FloatVal{Val: val}

        case *ast.StringLiteral:
                if len(e.Parts) > 0 {
                        return evalStringInterp(e, env)
                }
                return &StringVal{Val: e.Value}

        case *ast.BoolLiteral:
                return &BoolVal{Val: e.Value}

        case *ast.NoneLiteral:
                return &NoneVal{}

        case *ast.Identifier:
                return evalIdentifier(e, env)

        case *ast.BinaryOp:
                return evalBinaryOp(e, env)

        case *ast.UnaryOp:
                return evalUnaryOp(e, env)

        case *ast.CallExpr:
                return evalCallExpr(e, env)

        case *ast.FieldAccess:
                return evalFieldAccess(e, env)

        case *ast.OptionalFieldAccess:
                return evalOptionalFieldAccess(e, env)

        case *ast.PipelineExpr:
                return evalPipelineExpr(e, env)

        case *ast.IndexExpr:
                return evalIndexExpr(e, env)

        case *ast.StructExpr:
                return evalStructExpr(e, env)

        case *ast.ListExpr:
                return evalListExpr(e, env)

        case *ast.MapExpr:
                return evalMapExpr(e, env)

        case *ast.IfExpr:
                return evalIfExpr(e, env)

        case *ast.ListComp:
                return evalListComp(e, env)

        case *ast.Lambda:
                return evalLambda(e, env)

        case *ast.TupleLiteral:
                return evalTupleLiteral(e, env)

        case *ast.OptionPropagate:
                return evalOptionPropagate(e, env)

        case *ast.MatchExpr:
                return evalMatchExpr(e, env)

        default:
                runtimePanic(expr.GetSpan(), "unsupported expression type: %T", expr)
        }
        return &NoneVal{}
}

func evalStringInterp(e *ast.StringLiteral, env *Environment) Value {
        var sb strings.Builder
        for _, part := range e.Parts {
                if part.IsExpr {
                        val := EvalExpr(part.Expr, env)
                        sb.WriteString(val.String())
                } else {
                        sb.WriteString(part.Text)
                }
        }
        return &StringVal{Val: sb.String()}
}

func evalIdentifier(e *ast.Identifier, env *Environment) Value {
        val, ok := env.Get(e.Name)
        if !ok {
                runtimePanic(e.Span, "undefined variable '%s'", e.Name)
        }
        return val
}

func evalBinaryOp(e *ast.BinaryOp, env *Environment) Value {
        // Short-circuit for logical operators
        if e.Op == "and" {
                left := EvalExpr(e.Left, env)
                if !IsTruthy(left) {
                        return left
                }
                return EvalExpr(e.Right, env)
        }
        if e.Op == "or" {
                left := EvalExpr(e.Left, env)
                if IsTruthy(left) {
                        return left
                }
                return EvalExpr(e.Right, env)
        }

        left := EvalExpr(e.Left, env)
        right := EvalExpr(e.Right, env)

        switch e.Op {
        case "==":
                return &BoolVal{Val: Equal(left, right)}
        case "!=":
                return &BoolVal{Val: !Equal(left, right)}
        case "+":
                return evalAdd(e, left, right)
        case "-":
                return evalArith(e, left, right, func(a, b int64) int64 { return a - b }, func(a, b float64) float64 { return a - b })
        case "*":
                return evalArith(e, left, right, func(a, b int64) int64 { return a * b }, func(a, b float64) float64 { return a * b })
        case "/":
                return evalDiv(e, left, right)
        case "%":
                return evalMod(e, left, right)
        case "**":
                return evalPow(e, left, right)
        case "<":
                return &BoolVal{Val: compareValues(e, left, right) < 0}
        case ">":
                return &BoolVal{Val: compareValues(e, left, right) > 0}
        case "<=":
                return &BoolVal{Val: compareValues(e, left, right) <= 0}
        case ">=":
                return &BoolVal{Val: compareValues(e, left, right) >= 0}
        case "|>":
                return evalPipeline(e, left, right, env)
        default:
                runtimePanic(e.Span, "unsupported binary operator: %s", e.Op)
        }
        return &NoneVal{}
}

func evalAdd(e *ast.BinaryOp, left, right Value) Value {
        switch lv := left.(type) {
        case *IntVal:
                switch rv := right.(type) {
                case *IntVal:
                        return &IntVal{Val: lv.Val + rv.Val}
                case *FloatVal:
                        return &FloatVal{Val: float64(lv.Val) + rv.Val}
                }
        case *FloatVal:
                switch rv := right.(type) {
                case *FloatVal:
                        return &FloatVal{Val: lv.Val + rv.Val}
                case *IntVal:
                        return &FloatVal{Val: lv.Val + float64(rv.Val)}
                }
        case *StringVal:
                if rv, ok := right.(*StringVal); ok {
                        return &StringVal{Val: lv.Val + rv.Val}
                }
        case *ListVal:
                if rv, ok := right.(*ListVal); ok {
                        elems := make([]Value, 0, len(lv.Elements)+len(rv.Elements))
                        elems = append(elems, lv.Elements...)
                        elems = append(elems, rv.Elements...)
                        return &ListVal{Elements: elems}
                }
        }
        runtimePanic(e.Span, "cannot add %s and %s", valueTypeNames[left.Type()], valueTypeNames[right.Type()])
        return nil
}

func evalArith(e *ast.BinaryOp, left, right Value, intOp func(int64, int64) int64, floatOp func(float64, float64) float64) Value {
        switch lv := left.(type) {
        case *IntVal:
                switch rv := right.(type) {
                case *IntVal:
                        return &IntVal{Val: intOp(lv.Val, rv.Val)}
                case *FloatVal:
                        return &FloatVal{Val: floatOp(float64(lv.Val), rv.Val)}
                }
        case *FloatVal:
                switch rv := right.(type) {
                case *FloatVal:
                        return &FloatVal{Val: floatOp(lv.Val, rv.Val)}
                case *IntVal:
                        return &FloatVal{Val: floatOp(lv.Val, float64(rv.Val))}
                }
        }
        runtimePanic(e.Span, "cannot perform arithmetic on %s and %s", valueTypeNames[left.Type()], valueTypeNames[right.Type()])
        return nil
}

func evalDiv(e *ast.BinaryOp, left, right Value) Value {
        switch lv := left.(type) {
        case *IntVal:
                switch rv := right.(type) {
                case *IntVal:
                        if rv.Val == 0 {
                                runtimePanic(e.Span, "division by zero")
                        }
                        return &IntVal{Val: lv.Val / rv.Val}
                case *FloatVal:
                        if rv.Val == 0 {
                                runtimePanic(e.Span, "division by zero")
                        }
                        return &FloatVal{Val: float64(lv.Val) / rv.Val}
                }
        case *FloatVal:
                switch rv := right.(type) {
                case *FloatVal:
                        if rv.Val == 0 {
                                runtimePanic(e.Span, "division by zero")
                        }
                        return &FloatVal{Val: lv.Val / rv.Val}
                case *IntVal:
                        if rv.Val == 0 {
                                runtimePanic(e.Span, "division by zero")
                        }
                        return &FloatVal{Val: lv.Val / float64(rv.Val)}
                }
        }
        runtimePanic(e.Span, "cannot divide %s by %s", valueTypeNames[left.Type()], valueTypeNames[right.Type()])
        return nil
}

func evalMod(e *ast.BinaryOp, left, right Value) Value {
        lv, lok := left.(*IntVal)
        rv, rok := right.(*IntVal)
        if !lok || !rok {
                runtimePanic(e.Span, "modulo requires integer operands")
        }
        if rv.Val == 0 {
                runtimePanic(e.Span, "division by zero")
        }
        return &IntVal{Val: lv.Val % rv.Val}
}

func evalPow(e *ast.BinaryOp, left, right Value) Value {
        switch lv := left.(type) {
        case *IntVal:
                switch rv := right.(type) {
                case *IntVal:
                        return &IntVal{Val: intPow(lv.Val, rv.Val)}
                case *FloatVal:
                        return &FloatVal{Val: math.Pow(float64(lv.Val), rv.Val)}
                }
        case *FloatVal:
                switch rv := right.(type) {
                case *FloatVal:
                        return &FloatVal{Val: math.Pow(lv.Val, rv.Val)}
                case *IntVal:
                        return &FloatVal{Val: math.Pow(lv.Val, float64(rv.Val))}
                }
        }
        runtimePanic(e.Span, "cannot raise %s to power of %s", valueTypeNames[left.Type()], valueTypeNames[right.Type()])
        return nil
}

func intPow(base, exp int64) int64 {
        if exp < 0 {
                return 0
        }
        result := int64(1)
        for exp > 0 {
                if exp%2 == 1 {
                        result *= base
                }
                base *= base
                exp /= 2
        }
        return result
}

func compareValues(e *ast.BinaryOp, left, right Value) int {
        switch lv := left.(type) {
        case *IntVal:
                switch rv := right.(type) {
                case *IntVal:
                        if lv.Val < rv.Val {
                                return -1
                        }
                        if lv.Val > rv.Val {
                                return 1
                        }
                        return 0
                case *FloatVal:
                        lf := float64(lv.Val)
                        if lf < rv.Val {
                                return -1
                        }
                        if lf > rv.Val {
                                return 1
                        }
                        return 0
                }
        case *FloatVal:
                var rf float64
                switch rv := right.(type) {
                case *FloatVal:
                        rf = rv.Val
                case *IntVal:
                        rf = float64(rv.Val)
                default:
                        runtimePanic(e.Span, "cannot compare %s and %s", valueTypeNames[left.Type()], valueTypeNames[right.Type()])
                }
                if lv.Val < rf {
                        return -1
                }
                if lv.Val > rf {
                        return 1
                }
                return 0
        case *StringVal:
                rv, ok := right.(*StringVal)
                if !ok {
                        runtimePanic(e.Span, "cannot compare %s and %s", valueTypeNames[left.Type()], valueTypeNames[right.Type()])
                }
                if lv.Val < rv.Val {
                        return -1
                }
                if lv.Val > rv.Val {
                        return 1
                }
                return 0
        }
        runtimePanic(e.Span, "cannot compare %s and %s", valueTypeNames[left.Type()], valueTypeNames[right.Type()])
        return 0
}

func evalPipeline(e *ast.BinaryOp, left Value, right Value, env *Environment) Value {
        // right should be a function; call it with left as argument
        return callFunction(e.Span, right, []Value{left}, env)
}

func evalUnaryOp(e *ast.UnaryOp, env *Environment) Value {
        operand := EvalExpr(e.Operand, env)
        switch e.Op {
        case "-":
                switch v := operand.(type) {
                case *IntVal:
                        return &IntVal{Val: -v.Val}
                case *FloatVal:
                        return &FloatVal{Val: -v.Val}
                default:
                        runtimePanic(e.Span, "cannot negate %s", valueTypeNames[operand.Type()])
                }
        case "not":
                return &BoolVal{Val: !IsTruthy(operand)}
        default:
                runtimePanic(e.Span, "unsupported unary operator: %s", e.Op)
        }
        return &NoneVal{}
}

func evalCallExpr(e *ast.CallExpr, env *Environment) Value {
        // Check for enum variant construction: EnumName.Variant(args)
        if fa, ok := e.Callee.(*ast.FieldAccess); ok {
                if ident, ok := fa.Object.(*ast.Identifier); ok {
                        if variants, found := env.GetEnumDef(ident.Name); found {
                                if _, variantExists := variants[fa.Field]; variantExists {
                                        args := evalArgs(e.Args, env)
                                        return &EnumVal{
                                                TypeName:    ident.Name,
                                                VariantName: fa.Field,
                                                Data:        args,
                                        }
                                }
                        }
                }
        }

        callee := EvalExpr(e.Callee, env)
        args := evalArgs(e.Args, env)
        return callFunction(e.Span, callee, args, env)
}

func evalArgs(args []*ast.Arg, env *Environment) []Value {
        vals := make([]Value, len(args))
        for i, arg := range args {
                vals[i] = EvalExpr(arg.Value, env)
        }
        return vals
}

func callFunction(span token.Span, callee Value, args []Value, env *Environment) Value {
        switch fn := callee.(type) {
        case *FunctionVal:
                return callUserFn(span, fn, args)
        case *LambdaVal:
                return callLambda(span, fn, args)
        case *BuiltinFnVal:
                return fn.Fn(args)
        case *StringVal:
                // Could be a struct constructor: callee is a string that is a type name
                runtimePanic(span, "'%s' is not callable", fn.Val)
        default:
                runtimePanic(span, "cannot call value of type %s", valueTypeNames[callee.Type()])
        }
        return &NoneVal{}
}

func callUserFn(span token.Span, fn *FunctionVal, args []Value) Value {
        fnEnv := NewEnclosedEnvironment(fn.Env)

        // Bind parameters
        for i, param := range fn.Params {
                if i < len(args) {
                        fnEnv.Define(param.Name, args[i])
                } else if param.Default != nil {
                        fnEnv.Define(param.Name, EvalExpr(param.Default, fn.Env))
                } else {
                        runtimePanic(span, "missing argument for parameter '%s' in function '%s'", param.Name, fn.Name)
                }
        }

        result := execBlock(fn.Body, fnEnv)
        return result
}

func callLambda(span token.Span, fn *LambdaVal, args []Value) Value {
        fnEnv := NewEnclosedEnvironment(fn.Env)

        for i, param := range fn.Params {
                if i < len(args) {
                        fnEnv.Define(param.Name, args[i])
                } else if param.Default != nil {
                        fnEnv.Define(param.Name, EvalExpr(param.Default, fn.Env))
                } else {
                        runtimePanic(span, "missing argument for lambda parameter '%s'", param.Name)
                }
        }

        if fn.Body != nil {
                return EvalExpr(fn.Body, fnEnv)
        }
        return execBlock(fn.Block, fnEnv)
}

// execBlock executes a list of statements and returns the result.
// It handles return signals by catching them with recover.
func execBlock(stmts []ast.Statement, env *Environment) (result Value) {
        result = &NoneVal{}
        defer func() {
                if r := recover(); r != nil {
                        switch sig := r.(type) {
                        case returnSignal:
                                result = sig.val
                        default:
                                panic(r) // re-panic for break/continue/errors
                        }
                }
        }()
        for _, stmt := range stmts {
                result = ExecStmt(stmt, env)
        }
        return result
}

func evalFieldAccess(e *ast.FieldAccess, env *Environment) Value {
        // Check if it's an enum type name accessing a variant FIRST
        // (before evaluating the object, which would fail for type names)
        if ident, ok := e.Object.(*ast.Identifier); ok {
                if variants, found := env.GetEnumDef(ident.Name); found {
                        arity, variantExists := variants[e.Field]
                        if !variantExists {
                                runtimePanic(e.Span, "enum '%s' has no variant '%s'", ident.Name, e.Field)
                        }
                        if arity == 0 {
                                // Unit variant
                                return &EnumVal{TypeName: ident.Name, VariantName: e.Field}
                        }
                        // Return a constructor function for variants with data
                        enumName := ident.Name
                        fieldName := e.Field
                        return &BuiltinFnVal{
                                Name: enumName + "." + fieldName,
                                Fn: func(args []Value) Value {
                                        return &EnumVal{TypeName: enumName, VariantName: fieldName, Data: args}
                                },
                        }
                }
        }

        obj := EvalExpr(e.Object, env)

        switch v := obj.(type) {
        case *ModuleVal:
                val, ok := v.GetExport(e.Field)
                if !ok {
                        runtimePanic(e.Span, "module '%s' has no export '%s'", v.Name, e.Field)
                }
                return val
        case *StructVal:
                val, ok := v.Fields[e.Field]
                if !ok {
                        runtimePanic(e.Span, "struct '%s' has no field '%s'", v.TypeName, e.Field)
                }
                return val
        case *MapVal:
                // Map field access (dot notation for string keys)
                key := &StringVal{Val: e.Field}
                for i, k := range v.Keys {
                        if Equal(k, key) {
                                return v.Values[i]
                        }
                }
                // Fall through to method registry for map methods
                if m := resolveMethod(obj, e.Field); m != nil {
                        return m
                }
                return &NoneVal{}
        default:
                // Use the method registry for all other types
                if m := resolveMethod(obj, e.Field); m != nil {
                        return m
                }
                runtimePanic(e.Span, "cannot access field '%s' on value of type %s", e.Field, valueTypeNames[obj.Type()])
        }
        return &NoneVal{}
}



func evalOptionalFieldAccess(e *ast.OptionalFieldAccess, env *Environment) Value {
        obj := EvalExpr(e.Object, env)

        // Short-circuit on None/OptionVal(None)
        switch v := obj.(type) {
        case *NoneVal:
                return &NoneVal{}
        case *OptionVal:
                if !v.IsSome {
                        return &NoneVal{}
                }
                // Unwrap and access field on inner value
                obj = v.Val
        }

        // Now access the field on the (possibly unwrapped) value
        switch v := obj.(type) {
        case *StructVal:
                val, ok := v.Fields[e.Field]
                if !ok {
                        return &NoneVal{}
                }
                return val
        case *MapVal:
                key := &StringVal{Val: e.Field}
                for i, k := range v.Keys {
                        if Equal(k, key) {
                                return v.Values[i]
                        }
                }
                return &NoneVal{}
        case *NoneVal:
                return &NoneVal{}
        default:
                runtimePanic(e.Span, "cannot access field '%s' on value of type %s", e.Field, valueTypeNames[obj.Type()])
        }
        return &NoneVal{}
}

func evalPipelineExpr(e *ast.PipelineExpr, env *Environment) Value {
        left := EvalExpr(e.Left, env)
        right := EvalExpr(e.Right, env)
        return callFunction(e.Span, right, []Value{left}, env)
}

func evalIndexExpr(e *ast.IndexExpr, env *Environment) Value {
        obj := EvalExpr(e.Object, env)
        index := EvalExpr(e.Index, env)

        switch v := obj.(type) {
        case *ListVal:
                idx, ok := index.(*IntVal)
                if !ok {
                        runtimePanic(e.Span, "list index must be an integer, got %s", valueTypeNames[index.Type()])
                }
                i := idx.Val
                if i < 0 {
                        i = int64(len(v.Elements)) + i
                }
                if i < 0 || i >= int64(len(v.Elements)) {
                        runtimePanic(e.Span, "index %d out of bounds for list of length %d", idx.Val, len(v.Elements))
                }
                return v.Elements[i]
        case *MapVal:
                for i, k := range v.Keys {
                        if Equal(k, index) {
                                return v.Values[i]
                        }
                }
                return &NoneVal{}
        case *TupleVal:
                idx, ok := index.(*IntVal)
                if !ok {
                        runtimePanic(e.Span, "tuple index must be an integer")
                }
                i := idx.Val
                if i < 0 || i >= int64(len(v.Elements)) {
                        runtimePanic(e.Span, "index %d out of bounds for tuple of length %d", i, len(v.Elements))
                }
                return v.Elements[i]
        case *StringVal:
                idx, ok := index.(*IntVal)
                if !ok {
                        runtimePanic(e.Span, "string index must be an integer")
                }
                i := idx.Val
                if i < 0 {
                        i = int64(len(v.Val)) + i
                }
                if i < 0 || i >= int64(len(v.Val)) {
                        runtimePanic(e.Span, "index %d out of bounds for string of length %d", idx.Val, len(v.Val))
                }
                return &StringVal{Val: string(v.Val[i])}
        default:
                runtimePanic(e.Span, "cannot index into %s", valueTypeNames[obj.Type()])
        }
        return &NoneVal{}
}

func evalStructExpr(e *ast.StructExpr, env *Environment) Value {
        fields := make(map[string]Value, len(e.Fields))
        for _, f := range e.Fields {
                fields[f.Name] = EvalExpr(f.Value, env)
        }
        return &StructVal{TypeName: e.TypeName, Fields: fields}
}

func evalListExpr(e *ast.ListExpr, env *Environment) Value {
        elems := make([]Value, len(e.Elements))
        for i, elem := range e.Elements {
                elems[i] = EvalExpr(elem, env)
        }
        return &ListVal{Elements: elems}
}

func evalTupleLiteral(e *ast.TupleLiteral, env *Environment) Value {
        elems := make([]Value, len(e.Elements))
        for i, elem := range e.Elements {
                elems[i] = EvalExpr(elem, env)
        }
        return &TupleVal{Elements: elems}
}

func execLetTupleDestructure(s *ast.LetTupleDestructure, env *Environment) Value {
        val := EvalExpr(s.Value, env)
        tuple, ok := val.(*TupleVal)
        if !ok {
                // Also support destructuring lists
                if list, ok2 := val.(*ListVal); ok2 {
                        if len(list.Elements) != len(s.Names) {
                                runtimePanic(s.Span, "cannot destructure list of length %d into %d variables", len(list.Elements), len(s.Names))
                        }
                        for i, name := range s.Names {
                                if name == "_" {
                                        continue
                                }
                                if s.Mutable {
                                        env.Define(name, list.Elements[i])
                                } else {
                                        env.DefineConst(name, list.Elements[i])
                                }
                        }
                        return &NoneVal{}
                }
                runtimePanic(s.Span, "cannot destructure %s, expected tuple or list", valueTypeNames[val.Type()])
        }
        if len(tuple.Elements) != len(s.Names) {
                runtimePanic(s.Span, "cannot destructure tuple of length %d into %d variables", len(tuple.Elements), len(s.Names))
        }
        for i, name := range s.Names {
                if name == "_" {
                        continue
                }
                if s.Mutable {
                        env.Define(name, tuple.Elements[i])
                } else {
                        env.DefineConst(name, tuple.Elements[i])
                }
        }
        return &NoneVal{}
}

func evalMapExpr(e *ast.MapExpr, env *Environment) Value {
        keys := make([]Value, 0, len(e.Entries))
        values := make([]Value, 0, len(e.Entries))
        for _, entry := range e.Entries {
                keys = append(keys, EvalExpr(entry.Key, env))
                values = append(values, EvalExpr(entry.Value, env))
        }
        return &MapVal{Keys: keys, Values: values}
}

func evalIfExpr(e *ast.IfExpr, env *Environment) Value {
        cond := EvalExpr(e.Condition, env)
        if IsTruthy(cond) {
                return EvalExpr(e.ThenExpr, env)
        }
        if e.ElseExpr != nil {
                return EvalExpr(e.ElseExpr, env)
        }
        return &NoneVal{}
}

func evalListComp(e *ast.ListComp, env *Environment) Value {
        iterable := EvalExpr(e.Iterable, env)
        list, ok := iterable.(*ListVal)
        if !ok {
                runtimePanic(e.Span, "list comprehension requires a list, got %s", valueTypeNames[iterable.Type()])
        }

        result := make([]Value, 0)
        for _, elem := range list.Elements {
                compEnv := NewEnclosedEnvironment(env)
                compEnv.Define(e.Variable, elem)

                if e.Filter != nil {
                        filterVal := EvalExpr(e.Filter, compEnv)
                        if !IsTruthy(filterVal) {
                                continue
                        }
                }

                result = append(result, EvalExpr(e.Element, compEnv))
        }
        return &ListVal{Elements: result}
}

func evalLambda(e *ast.Lambda, env *Environment) Value {
        return &LambdaVal{
                Params: e.Params,
                Body:   e.Body,
                Block:  e.Block,
                Env:    env,
        }
}

func evalOptionPropagate(e *ast.OptionPropagate, env *Environment) Value {
        val := EvalExpr(e.Expr, env)
        switch v := val.(type) {
        case *OptionVal:
                if !v.IsSome {
                        // Propagate None by returning it
                        panic(returnSignal{val: &OptionVal{IsSome: false}})
                }
                return v.Val
        case *NoneVal:
                panic(returnSignal{val: &OptionVal{IsSome: false}})
        default:
                return val
        }
}

// --- Statement Execution ---

// ExecStmt executes a statement and returns its result value.
func ExecStmt(stmt ast.Statement, env *Environment) Value {
        switch s := stmt.(type) {
        case *ast.LetStmt:
                return execLetStmt(s, env)
        case *ast.LetTupleDestructure:
                return execLetTupleDestructure(s, env)
        case *ast.AssignStmt:
                return execAssignStmt(s, env)
        case *ast.ReturnStmt:
                return execReturnStmt(s, env)
        case *ast.IfStmt:
                return execIfStmt(s, env)
        case *ast.MatchStmt:
                return execMatchStmt(s, env)
        case *ast.ForStmt:
                return execForStmt(s, env)
        case *ast.WhileStmt:
                return execWhileStmt(s, env)
        case *ast.BreakStmt:
                panic(breakSignal{})
        case *ast.ContinueStmt:
                panic(continueSignal{})
        case *ast.WithStmt:
                return execWithStmt(s, env)
        case *ast.AssertStmt:
                return execAssertStmt(s, env)
        case *ast.ExprStmt:
                return EvalExpr(s.Expr, env)
        default:
                runtimePanic(stmt.GetSpan(), "unsupported statement type: %T", stmt)
        }
        return &NoneVal{}
}

func execLetStmt(s *ast.LetStmt, env *Environment) Value {
        val := EvalExpr(s.Value, env)
        if s.Mutable {
                env.Define(s.Name, val)
        } else {
                env.DefineConst(s.Name, val)
        }
        return &NoneVal{}
}

func execAssignStmt(s *ast.AssignStmt, env *Environment) Value {
        val := EvalExpr(s.Value, env)

        switch target := s.Target.(type) {
        case *ast.Identifier:
                if err := env.Set(target.Name, val); err != nil {
                        runtimePanic(s.Span, "%s", err.Error())
                }
        case *ast.FieldAccess:
                obj := EvalExpr(target.Object, env)
                if sv, ok := obj.(*StructVal); ok {
                        sv.Fields[target.Field] = val
                } else {
                        runtimePanic(s.Span, "cannot assign field on %s", valueTypeNames[obj.Type()])
                }
        case *ast.IndexExpr:
                obj := EvalExpr(target.Object, env)
                idx := EvalExpr(target.Index, env)
                switch v := obj.(type) {
                case *ListVal:
                        i, ok := idx.(*IntVal)
                        if !ok {
                                runtimePanic(s.Span, "list index must be an integer")
                        }
                        if i.Val < 0 || i.Val >= int64(len(v.Elements)) {
                                runtimePanic(s.Span, "index %d out of bounds", i.Val)
                        }
                        v.Elements[i.Val] = val
                case *MapVal:
                        for i, k := range v.Keys {
                                if Equal(k, idx) {
                                        v.Values[i] = val
                                        return &NoneVal{}
                                }
                        }
                        v.Keys = append(v.Keys, idx)
                        v.Values = append(v.Values, val)
                default:
                        runtimePanic(s.Span, "cannot index-assign on %s", valueTypeNames[obj.Type()])
                }
        default:
                runtimePanic(s.Span, "invalid assignment target")
        }
        return &NoneVal{}
}

func execReturnStmt(s *ast.ReturnStmt, env *Environment) Value {
        var val Value = &NoneVal{}
        if s.Value != nil {
                val = EvalExpr(s.Value, env)
        }
        panic(returnSignal{val: val})
}

func execIfStmt(s *ast.IfStmt, env *Environment) Value {
        cond := EvalExpr(s.Condition, env)
        if IsTruthy(cond) {
                return execStmtBlock(s.ThenBody, env)
        }

        for _, elif := range s.ElifClauses {
                elifCond := EvalExpr(elif.Condition, env)
                if IsTruthy(elifCond) {
                        return execStmtBlock(elif.Body, env)
                }
        }

        if len(s.ElseBody) > 0 {
                return execStmtBlock(s.ElseBody, env)
        }
        return &NoneVal{}
}

// execStmtBlock executes a block of statements in a new scope (for if/for/while bodies).
func execStmtBlock(stmts []ast.Statement, env *Environment) Value {
        blockEnv := NewEnclosedEnvironment(env)
        var result Value = &NoneVal{}
        for _, stmt := range stmts {
                result = ExecStmt(stmt, blockEnv)
        }
        return result
}

func execMatchStmt(s *ast.MatchStmt, env *Environment) Value {
        subject := EvalExpr(s.Subject, env)

        for _, c := range s.Cases {
                caseEnv := NewEnclosedEnvironment(env)
                if matchPattern(c.Pattern, subject, caseEnv) {
                        // Check guard
                        if c.Guard != nil {
                                guardVal := EvalExpr(c.Guard, caseEnv)
                                if !IsTruthy(guardVal) {
                                        continue
                                }
                        }
                        var result Value = &NoneVal{}
                        for _, stmt := range c.Body {
                                result = ExecStmt(stmt, caseEnv)
                        }
                        return result
                }
        }
        return &NoneVal{}
}

// evalMatchExpr evaluates a match expression (pattern -> expr form).
// Unlike execMatchStmt, this uses arrow syntax and each arm is an expression.
// Panics with RuntimeError if no pattern matches.
func evalMatchExpr(m *ast.MatchExpr, env *Environment) Value {
        subject := EvalExpr(m.Subject, env)

        for _, arm := range m.Arms {
                armEnv := NewEnclosedEnvironment(env)
                if matchPattern(arm.Pattern, subject, armEnv) {
                        return EvalExpr(arm.Body, armEnv)
                }
        }

        runtimePanic(m.Span, "no pattern matched in match expression for value: %s", subject.String())
        return &NoneVal{}
}

func matchPattern(pattern ast.Pattern, value Value, env *Environment) bool {
        switch p := pattern.(type) {
        case *ast.WildcardPattern:
                return true
        case *ast.BindingPattern:
                env.Define(p.Name, value)
                return true
        case *ast.LiteralPattern:
                return matchLiteralPattern(p, value)
        case *ast.ConstructorPattern:
                return matchConstructorPattern(p, value, env)
        case *ast.ListPattern:
                if list, ok := value.(*ListVal); ok {
                        return matchListPattern(p, list, env)
                }
                return false
        case *ast.TuplePattern:
                if tuple, ok := value.(*TupleVal); ok {
                        if len(p.Elements) != len(tuple.Elements) {
                                return false
                        }
                        for i, elem := range p.Elements {
                                if !matchPattern(elem, tuple.Elements[i], env) {
                                        return false
                                }
                        }
                        return true
                }
                return false
        default:
                return false
        }
}

func matchLiteralPattern(p *ast.LiteralPattern, value Value) bool {
        switch p.Kind {
        case token.INT_LIT:
                iv, ok := value.(*IntVal)
                if !ok {
                        return false
                }
                pv, err := strconv.ParseInt(p.Value, 10, 64)
                if err != nil {
                        return false
                }
                return iv.Val == pv
        case token.FLOAT_LIT:
                fv, ok := value.(*FloatVal)
                if !ok {
                        return false
                }
                pv, err := strconv.ParseFloat(p.Value, 64)
                if err != nil {
                        return false
                }
                return fv.Val == pv
        case token.STRING_LIT:
                sv, ok := value.(*StringVal)
                if !ok {
                        return false
                }
                return sv.Val == p.Value
        case token.BOOL_LIT:
                bv, ok := value.(*BoolVal)
                if !ok {
                        return false
                }
                return (p.Value == "true") == bv.Val
        case token.NONE_VAL, token.NONE_LIT:
                _, ok := value.(*NoneVal)
                return ok
        default:
                return false
        }
}

func matchConstructorPattern(p *ast.ConstructorPattern, value Value, env *Environment) bool {
        // Handle dotted names like "Color.Red" or just "Red"
        typeName := ""
        variantName := p.TypeName
        if idx := strings.LastIndex(p.TypeName, "."); idx >= 0 {
                typeName = p.TypeName[:idx]
                variantName = p.TypeName[idx+1:]
        }

        // Handle Some/None/Ok/Err constructors
        switch variantName {
        case "Some":
                if opt, ok := value.(*OptionVal); ok {
                        if !opt.IsSome {
                                return false
                        }
                        if len(p.Fields) == 1 {
                                return matchPattern(p.Fields[0], opt.Val, env)
                        }
                        return true
                }
                return false
        case "None":
                if opt, ok := value.(*OptionVal); ok {
                        return !opt.IsSome
                }
                _, isNone := value.(*NoneVal)
                return isNone
        case "Ok":
                if res, ok := value.(*ResultVal); ok {
                        if !res.IsOk {
                                return false
                        }
                        if len(p.Fields) == 1 {
                                return matchPattern(p.Fields[0], res.Val, env)
                        }
                        return true
                }
                return false
        case "Err":
                if res, ok := value.(*ResultVal); ok {
                        if res.IsOk {
                                return false
                        }
                        if len(p.Fields) == 1 {
                                return matchPattern(p.Fields[0], res.Val, env)
                        }
                        return true
                }
                return false
        }

        // Enum variant matching
        ev, ok := value.(*EnumVal)
        if !ok {
                return false
        }

        if typeName != "" && ev.TypeName != typeName {
                return false
        }
        if ev.VariantName != variantName {
                return false
        }

        // Match data fields
        for i, fp := range p.Fields {
                if i < len(ev.Data) {
                        if !matchPattern(fp, ev.Data[i], env) {
                                return false
                        }
                }
        }
        return true
}

func matchListPattern(p *ast.ListPattern, list *ListVal, env *Environment) bool {
        // Find spread pattern index (-1 if none)
        spreadIdx := -1
        for i, elem := range p.Elements {
                if _, ok := elem.(*ast.SpreadPattern); ok {
                        spreadIdx = i
                        break
                }
        }

        if spreadIdx == -1 {
                // No spread: exact length match required
                if len(p.Elements) != len(list.Elements) {
                        return false
                }
                for i, elem := range p.Elements {
                        if !matchPattern(elem, list.Elements[i], env) {
                                return false
                        }
                }
                return true
        }

        // With spread pattern: minimum length is len(patterns) - 1
        nonSpreadCount := len(p.Elements) - 1
        if len(list.Elements) < nonSpreadCount {
                return false
        }

        // Match elements before the spread
        for i := 0; i < spreadIdx; i++ {
                if !matchPattern(p.Elements[i], list.Elements[i], env) {
                        return false
                }
        }

        // Match elements after the spread
        afterSpread := len(p.Elements) - spreadIdx - 1
        for i := 0; i < afterSpread; i++ {
                patIdx := spreadIdx + 1 + i
                listIdx := len(list.Elements) - afterSpread + i
                if !matchPattern(p.Elements[patIdx], list.Elements[listIdx], env) {
                        return false
                }
        }

        // Bind the spread variable to the remaining elements
        sp := p.Elements[spreadIdx].(*ast.SpreadPattern)
        restStart := spreadIdx
        restEnd := len(list.Elements) - afterSpread
        restElements := make([]Value, restEnd-restStart)
        copy(restElements, list.Elements[restStart:restEnd])
        env.Define(sp.Name, &ListVal{Elements: restElements})

        return true
}

func execForStmt(s *ast.ForStmt, env *Environment) Value {
        iterable := EvalExpr(s.Iterable, env)

        var items []Value
        switch v := iterable.(type) {
        case *ListVal:
                items = v.Elements
        case *MapVal:
                items = v.Keys
        case *SetVal:
                items = v.Elements
        case *TupleVal:
                items = v.Elements
        default:
                runtimePanic(s.Span, "cannot iterate over %s", valueTypeNames[iterable.Type()])
        }

        var result Value = &NoneVal{}
        for _, item := range items {
                loopEnv := NewEnclosedEnvironment(env)
                loopEnv.Define(s.Variable, item)

                shouldBreak := false
                shouldContinue := false

                for _, stmt := range s.Body {
                        func() {
                                defer func() {
                                        if r := recover(); r != nil {
                                                switch r.(type) {
                                                case breakSignal:
                                                        shouldBreak = true
                                                case continueSignal:
                                                        shouldContinue = true
                                                default:
                                                        panic(r)
                                                }
                                        }
                                }()
                                result = ExecStmt(stmt, loopEnv)
                        }()
                        if shouldBreak || shouldContinue {
                                break
                        }
                }
                if shouldBreak {
                        break
                }
        }
        return result
}

func execWhileStmt(s *ast.WhileStmt, env *Environment) Value {
        var result Value = &NoneVal{}
        for {
                cond := EvalExpr(s.Condition, env)
                if !IsTruthy(cond) {
                        break
                }

                loopEnv := NewEnclosedEnvironment(env)
                shouldBreak := false

                for _, stmt := range s.Body {
                        func() {
                                defer func() {
                                        if r := recover(); r != nil {
                                                switch r.(type) {
                                                case breakSignal:
                                                        shouldBreak = true
                                                case continueSignal:
                                                        // continue
                                                default:
                                                        panic(r)
                                                }
                                        }
                                }()
                                result = ExecStmt(stmt, loopEnv)
                        }()
                        if shouldBreak {
                                break
                        }
                }
                if shouldBreak {
                        break
                }
        }
        return result
}

func execWithStmt(s *ast.WithStmt, env *Environment) Value {
        withEnv := NewEnclosedEnvironment(env)

        for _, binding := range s.Bindings {
                val := EvalExpr(binding.Expr, env)
                if binding.Alias != "" {
                        withEnv.Define(binding.Alias, val)
                }
                // If the binding expr is an identifier, treat it as an effect capability
                if ident, ok := binding.Expr.(*ast.Identifier); ok {
                        withEnv.AddEffect(ident.Name)
                        if binding.Alias == "" {
                                withEnv.Define(ident.Name, val)
                        }
                }
        }

        var result Value = &NoneVal{}
        for _, stmt := range s.Body {
                result = ExecStmt(stmt, withEnv)
        }
        return result
}

func execAssertStmt(s *ast.AssertStmt, env *Environment) Value {
        cond := EvalExpr(s.Condition, env)
        if !IsTruthy(cond) {
                msg := s.Message
                if msg == "" {
                        msg = "assertion failed"
                }
                runtimePanic(s.Span, "%s", msg)
        }
        return &NoneVal{}
}
