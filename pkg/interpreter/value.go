// Package interpreter implements a tree-walk interpreter for the Aura language.
package interpreter

import (
        "fmt"
        "strings"

        "github.com/unclebucklarson/aura/pkg/ast"
)

// ValueType identifies the kind of runtime value.
type ValueType int

const (
        TypeInt ValueType = iota
        TypeFloat
        TypeString
        TypeBool
        TypeNone
        TypeList
        TypeMap
        TypeSet
        TypeTuple
        TypeStruct
        TypeEnum
        TypeFunction
        TypeOption
        TypeResult
        TypeBuiltinFn
        TypeModule
)

var valueTypeNames = map[ValueType]string{
        TypeInt:       "Int",
        TypeFloat:     "Float",
        TypeString:    "String",
        TypeBool:      "Bool",
        TypeNone:      "None",
        TypeList:      "List",
        TypeMap:       "Map",
        TypeSet:       "Set",
        TypeTuple:     "Tuple",
        TypeStruct:    "Struct",
        TypeEnum:      "Enum",
        TypeFunction:  "Function",
        TypeOption:    "Option",
        TypeResult:    "Result",
        TypeBuiltinFn: "BuiltinFn",
        TypeModule:    "Module",
}

// Value is the interface for all runtime values.
type Value interface {
        Type() ValueType
        String() string
}

// --- Primitive Values ---

// IntVal represents an integer value.
type IntVal struct{ Val int64 }

func (v *IntVal) Type() ValueType { return TypeInt }
func (v *IntVal) String() string  { return fmt.Sprintf("%d", v.Val) }

// FloatVal represents a float value.
type FloatVal struct{ Val float64 }

func (v *FloatVal) Type() ValueType { return TypeFloat }
func (v *FloatVal) String() string  { return fmt.Sprintf("%g", v.Val) }

// StringVal represents a string value.
type StringVal struct{ Val string }

func (v *StringVal) Type() ValueType { return TypeString }
func (v *StringVal) String() string  { return v.Val }

// BoolVal represents a boolean value.
type BoolVal struct{ Val bool }

func (v *BoolVal) Type() ValueType { return TypeBool }
func (v *BoolVal) String() string {
        if v.Val {
                return "true"
        }
        return "false"
}

// NoneVal represents the none value.
type NoneVal struct{}

func (v *NoneVal) Type() ValueType { return TypeNone }
func (v *NoneVal) String() string  { return "none" }

// --- Collection Values ---

// ListVal represents a list value.
type ListVal struct{ Elements []Value }

func (v *ListVal) Type() ValueType { return TypeList }
func (v *ListVal) String() string {
        parts := make([]string, len(v.Elements))
        for i, e := range v.Elements {
                parts[i] = valueRepr(e)
        }
        return "[" + strings.Join(parts, ", ") + "]"
}

// MapVal represents a map value (ordered by insertion).
type MapVal struct {
        Keys   []Value
        Values []Value
}

func (v *MapVal) Type() ValueType { return TypeMap }
func (v *MapVal) String() string {
        parts := make([]string, len(v.Keys))
        for i := range v.Keys {
                parts[i] = valueRepr(v.Keys[i]) + ": " + valueRepr(v.Values[i])
        }
        return "{" + strings.Join(parts, ", ") + "}"
}

// SetVal represents a set value.
type SetVal struct{ Elements []Value }

func (v *SetVal) Type() ValueType { return TypeSet }
func (v *SetVal) String() string {
        parts := make([]string, len(v.Elements))
        for i, e := range v.Elements {
                parts[i] = valueRepr(e)
        }
        return "{" + strings.Join(parts, ", ") + "}"
}

// TupleVal represents a tuple value.
type TupleVal struct{ Elements []Value }

func (v *TupleVal) Type() ValueType { return TypeTuple }
func (v *TupleVal) String() string {
        parts := make([]string, len(v.Elements))
        for i, e := range v.Elements {
                parts[i] = valueRepr(e)
        }
        return "(" + strings.Join(parts, ", ") + ")"
}

// --- Composite Values ---

// StructVal represents a struct instance.
type StructVal struct {
        TypeName string
        Fields   map[string]Value
}

func (v *StructVal) Type() ValueType { return TypeStruct }
func (v *StructVal) String() string {
        parts := make([]string, 0, len(v.Fields))
        for k, val := range v.Fields {
                parts = append(parts, k+": "+valueRepr(val))
        }
        return v.TypeName + "(" + strings.Join(parts, ", ") + ")"
}

// EnumVal represents an enum variant value.
type EnumVal struct {
        TypeName    string
        VariantName string
        Data        []Value
}

func (v *EnumVal) Type() ValueType { return TypeEnum }
func (v *EnumVal) String() string {
        if len(v.Data) == 0 {
                return v.TypeName + "." + v.VariantName
        }
        parts := make([]string, len(v.Data))
        for i, d := range v.Data {
                parts[i] = valueRepr(d)
        }
        return v.TypeName + "." + v.VariantName + "(" + strings.Join(parts, ", ") + ")"
}

// FunctionVal represents a function closure.
type FunctionVal struct {
        Name   string
        Params []*ast.Param
        Body   []ast.Statement
        Env    *Environment // captured environment for closures
}

func (v *FunctionVal) Type() ValueType { return TypeFunction }
func (v *FunctionVal) String() string {
        if v.Name != "" {
                return "<fn " + v.Name + ">"
        }
        return "<fn>"
}

// LambdaVal represents a lambda closure.
type LambdaVal struct {
        Params []*ast.Param
        Body   ast.Expr        // single-expression lambda
        Block  []ast.Statement // block lambda
        Env    *Environment
}

func (v *LambdaVal) Type() ValueType { return TypeFunction }
func (v *LambdaVal) String() string  { return "<lambda>" }

// BuiltinFnVal represents a built-in function.
type BuiltinFnVal struct {
        Name string
        Fn   func(args []Value) Value
}

func (v *BuiltinFnVal) Type() ValueType { return TypeBuiltinFn }
func (v *BuiltinFnVal) String() string  { return "<builtin " + v.Name + ">" }

// --- Option/Result Values ---

// OptionVal represents Some(value) or None.
type OptionVal struct {
        IsSome bool
        Val    Value
}

func (v *OptionVal) Type() ValueType { return TypeOption }
func (v *OptionVal) String() string {
        if v.IsSome {
                return "Some(" + valueRepr(v.Val) + ")"
        }
        return "None"
}

// ResultVal represents Ok(value) or Err(value).
type ResultVal struct {
        IsOk bool
        Val  Value
}

func (v *ResultVal) Type() ValueType { return TypeResult }
func (v *ResultVal) String() string {
        if v.IsOk {
                return "Ok(" + valueRepr(v.Val) + ")"
        }
        return "Err(" + valueRepr(v.Val) + ")"
}

// ModuleVal represents a loaded module as a runtime value.
// It acts as a namespace: module.function_name accesses exports.
type ModuleVal struct {
        Name    string            // module name (e.g., "math", "helpers")
        Path    string            // resolved file path
        Exports map[string]Value  // exported symbols
}

func (v *ModuleVal) Type() ValueType { return TypeModule }
func (v *ModuleVal) String() string {
        return fmt.Sprintf("<module '%s'>", v.Name)
}

// GetExport retrieves an exported symbol from the module.
func (v *ModuleVal) GetExport(name string) (Value, bool) {
        val, ok := v.Exports[name]
        return val, ok
}

// --- Helper Functions ---

// valueRepr returns a string representation suitable for display.
func valueRepr(v Value) string {
        if v == nil {
                return "none"
        }
        if s, ok := v.(*StringVal); ok {
                return fmt.Sprintf("%q", s.Val)
        }
        return v.String()
}

// IsTruthy returns whether a value is considered truthy.
func IsTruthy(v Value) bool {
        if v == nil {
                return false
        }
        switch val := v.(type) {
        case *BoolVal:
                return val.Val
        case *NoneVal:
                return false
        case *IntVal:
                return val.Val != 0
        case *FloatVal:
                return val.Val != 0
        case *StringVal:
                return val.Val != ""
        case *ListVal:
                return len(val.Elements) > 0
        case *MapVal:
                return len(val.Keys) > 0
        case *OptionVal:
                return val.IsSome
        default:
                return true
        }
}

// Equal returns whether two values are equal.
func Equal(a, b Value) bool {
        if a == nil && b == nil {
                return true
        }
        if a == nil || b == nil {
                // none == NoneVal
                if a == nil {
                        _, ok := b.(*NoneVal)
                        return ok
                }
                _, ok := a.(*NoneVal)
                return ok
        }
        switch av := a.(type) {
        case *IntVal:
                switch bv := b.(type) {
                case *IntVal:
                        return av.Val == bv.Val
                case *FloatVal:
                        return float64(av.Val) == bv.Val
                }
        case *FloatVal:
                switch bv := b.(type) {
                case *FloatVal:
                        return av.Val == bv.Val
                case *IntVal:
                        return av.Val == float64(bv.Val)
                }
        case *StringVal:
                if bv, ok := b.(*StringVal); ok {
                        return av.Val == bv.Val
                }
        case *BoolVal:
                if bv, ok := b.(*BoolVal); ok {
                        return av.Val == bv.Val
                }
        case *NoneVal:
                _, ok := b.(*NoneVal)
                return ok
        case *OptionVal:
                if bv, ok := b.(*OptionVal); ok {
                        if !av.IsSome && !bv.IsSome {
                                return true
                        }
                        if av.IsSome && bv.IsSome {
                                return Equal(av.Val, bv.Val)
                        }
                }
                // None option == NoneVal
                if !av.IsSome {
                        _, ok := b.(*NoneVal)
                        return ok
                }
        case *ResultVal:
                if bv, ok := b.(*ResultVal); ok {
                        if av.IsOk == bv.IsOk {
                                return Equal(av.Val, bv.Val)
                        }
                }
        case *ListVal:
                if bv, ok := b.(*ListVal); ok {
                        if len(av.Elements) != len(bv.Elements) {
                                return false
                        }
                        for i := range av.Elements {
                                if !Equal(av.Elements[i], bv.Elements[i]) {
                                        return false
                                }
                        }
                        return true
                }
        case *TupleVal:
                if bv, ok := b.(*TupleVal); ok {
                        if len(av.Elements) != len(bv.Elements) {
                                return false
                        }
                        for i := range av.Elements {
                                if !Equal(av.Elements[i], bv.Elements[i]) {
                                        return false
                                }
                        }
                        return true
                }
        case *EnumVal:
                if bv, ok := b.(*EnumVal); ok {
                        if av.TypeName == bv.TypeName && av.VariantName == bv.VariantName {
                                if len(av.Data) != len(bv.Data) {
                                        return false
                                }
                                for i := range av.Data {
                                        if !Equal(av.Data[i], bv.Data[i]) {
                                                return false
                                        }
                                }
                                return true
                        }
                }
        case *StructVal:
                if bv, ok := b.(*StructVal); ok {
                        if av.TypeName != bv.TypeName {
                                return false
                        }
                        if len(av.Fields) != len(bv.Fields) {
                                return false
                        }
                        for k, v1 := range av.Fields {
                                v2, exists := bv.Fields[k]
                                if !exists || !Equal(v1, v2) {
                                        return false
                                }
                        }
                        return true
                }
        }
        return false
}
