// Package types defines the type system representation for Aura.
// It includes primitives, structs, enums, unions, generics, refinement types,
// and the rules for type equality and subtyping.
package types

import (
        "fmt"
        "strings"
)

// TypeKind represents the kind of a type.
type TypeKind int

const (
        KindPrimitive TypeKind = iota
        KindStruct
        KindEnum
        KindUnion
        KindIntersection
        KindFunction
        KindTuple
        KindList
        KindMap
        KindSet
        KindOption
        KindResult
        KindRefinement
        KindStringLit  // string literal type in unions like "pending" | "done"
        KindTypeParam  // generic type parameter
        KindNever      // bottom type
        KindAny        // top type
        KindNone       // the None/unit type
        KindAlias      // type alias
)

func (k TypeKind) String() string {
        names := [...]string{
                "primitive", "struct", "enum", "union", "intersection",
                "function", "tuple", "list", "map", "set", "option", "result",
                "refinement", "string_literal", "type_param", "never", "any",
                "none", "alias",
        }
        if int(k) < len(names) {
                return names[k]
        }
        return fmt.Sprintf("TypeKind(%d)", int(k))
}

// Type represents a type in the Aura type system.
type Type struct {
        Kind       TypeKind
        Name       string      // name for named types, primitives, type params
        Fields     []*Field    // for structs
        Variants   []*Variant  // for enums
        Members    []*Type     // for unions, intersections, tuples
        ElementT   *Type       // for list, set, option
        KeyT       *Type       // for map
        ValueT     *Type       // for map, result (Err type)
        ParamTypes []*Type     // for function param types
        ReturnT    *Type       // for function return type
        BaseT      *Type       // for refinement, alias
        Predicate  string      // for refinement types (stored as text)
        TypeParams []string    // for generic type definitions
        TypeArgs   []*Type     // for instantiated generic types
        StringVal  string      // for string literal types
        Effects    []string    // for function types
}

// Field represents a struct field with its type.
type Field struct {
        Name       string
        Type       *Type
        Optional   bool   // has a default value
        Public     bool
}

// Variant represents an enum variant.
type Variant struct {
        Name   string
        Fields []*Type // data carried by the variant
}

// --- Primitive type singletons ---

var (
        BuiltinInt    = &Type{Kind: KindPrimitive, Name: "Int"}
        BuiltinFloat  = &Type{Kind: KindPrimitive, Name: "Float"}
        BuiltinString = &Type{Kind: KindPrimitive, Name: "String"}
        BuiltinBool   = &Type{Kind: KindPrimitive, Name: "Bool"}
        BuiltinNone   = &Type{Kind: KindNone, Name: "None"}
        BuiltinNever  = &Type{Kind: KindNever, Name: "Never"}
        BuiltinAny    = &Type{Kind: KindAny, Name: "Any"}
)

// Builtins maps primitive type names to their type objects.
var Builtins = map[string]*Type{
        "Int":    BuiltinInt,
        "Float":  BuiltinFloat,
        "String": BuiltinString,
        "Bool":   BuiltinBool,
        "None":   BuiltinNone,
        "Never":  BuiltinNever,
        "Any":    BuiltinAny,
}

// --- Constructor functions ---

// NewStructType creates a new struct type.
func NewStructType(name string, fields []*Field, typeParams []string) *Type {
        return &Type{Kind: KindStruct, Name: name, Fields: fields, TypeParams: typeParams}
}

// NewEnumType creates a new enum type.
func NewEnumType(name string, variants []*Variant, typeParams []string) *Type {
        return &Type{Kind: KindEnum, Name: name, Variants: variants, TypeParams: typeParams}
}

// NewUnionType creates a union type from members.
func NewUnionType(members []*Type) *Type {
        return &Type{Kind: KindUnion, Members: members}
}

// NewFunctionType creates a function type.
func NewFunctionType(params []*Type, ret *Type, effects []string) *Type {
        return &Type{Kind: KindFunction, ParamTypes: params, ReturnT: ret, Effects: effects}
}

// NewListType creates a list type.
func NewListType(elem *Type) *Type {
        return &Type{Kind: KindList, ElementT: elem}
}

// NewMapType creates a map type.
func NewMapType(key, val *Type) *Type {
        return &Type{Kind: KindMap, KeyT: key, ValueT: val}
}

// NewSetType creates a set type.
func NewSetType(elem *Type) *Type {
        return &Type{Kind: KindSet, ElementT: elem}
}

// NewOptionType creates an Option[T] type.
func NewOptionType(inner *Type) *Type {
        return &Type{Kind: KindOption, ElementT: inner}
}

// NewResultType creates a Result[Ok, Err] type.
func NewResultType(okT, errT *Type) *Type {
        return &Type{Kind: KindResult, ElementT: okT, ValueT: errT}
}

// NewTupleType creates a tuple type.
func NewTupleType(elements []*Type) *Type {
        return &Type{Kind: KindTuple, Members: elements}
}

// NewRefinementType creates a refinement type (T where predicate).
func NewRefinementType(base *Type, predicate string) *Type {
        return &Type{Kind: KindRefinement, BaseT: base, Predicate: predicate}
}

// NewStringLitType creates a string literal type.
func NewStringLitType(val string) *Type {
        return &Type{Kind: KindStringLit, StringVal: val}
}

// NewTypeParam creates a type parameter placeholder.
func NewTypeParam(name string) *Type {
        return &Type{Kind: KindTypeParam, Name: name}
}

// NewAliasType creates a type alias.
func NewAliasType(name string, base *Type) *Type {
        return &Type{Kind: KindAlias, Name: name, BaseT: base}
}

// --- Type display ---

// String returns a human-readable representation of the type.
func (t *Type) String() string {
        if t == nil {
                return "<nil>"
        }
        switch t.Kind {
        case KindPrimitive, KindTypeParam, KindNever, KindAny, KindNone:
                return t.Name
        case KindStruct:
                if len(t.TypeArgs) > 0 {
                        args := make([]string, len(t.TypeArgs))
                        for i, a := range t.TypeArgs {
                                args[i] = a.String()
                        }
                        return fmt.Sprintf("%s[%s]", t.Name, strings.Join(args, ", "))
                }
                return t.Name
        case KindEnum:
                if len(t.TypeArgs) > 0 {
                        args := make([]string, len(t.TypeArgs))
                        for i, a := range t.TypeArgs {
                                args[i] = a.String()
                        }
                        return fmt.Sprintf("%s[%s]", t.Name, strings.Join(args, ", "))
                }
                return t.Name
        case KindUnion:
                parts := make([]string, len(t.Members))
                for i, m := range t.Members {
                        parts[i] = m.String()
                }
                return strings.Join(parts, " | ")
        case KindIntersection:
                parts := make([]string, len(t.Members))
                for i, m := range t.Members {
                        parts[i] = m.String()
                }
                return strings.Join(parts, " & ")
        case KindFunction:
                params := make([]string, len(t.ParamTypes))
                for i, p := range t.ParamTypes {
                        params[i] = p.String()
                }
                ret := "None"
                if t.ReturnT != nil {
                        ret = t.ReturnT.String()
                }
                return fmt.Sprintf("fn(%s) -> %s", strings.Join(params, ", "), ret)
        case KindTuple:
                parts := make([]string, len(t.Members))
                for i, m := range t.Members {
                        parts[i] = m.String()
                }
                return fmt.Sprintf("(%s)", strings.Join(parts, ", "))
        case KindList:
                return fmt.Sprintf("[%s]", t.ElementT.String())
        case KindMap:
                return fmt.Sprintf("{%s: %s}", t.KeyT.String(), t.ValueT.String())
        case KindSet:
                return fmt.Sprintf("{%s}", t.ElementT.String())
        case KindOption:
                return fmt.Sprintf("%s?", t.ElementT.String())
        case KindResult:
                return fmt.Sprintf("Result[%s, %s]", t.ElementT.String(), t.ValueT.String())
        case KindRefinement:
                return fmt.Sprintf("%s where %s", t.BaseT.String(), t.Predicate)
        case KindStringLit:
                return fmt.Sprintf("%q", t.StringVal)
        case KindAlias:
                return t.Name
        default:
                return fmt.Sprintf("<unknown:%s>", t.Kind)
        }
}

// --- Type equality and subtyping ---

// Equal returns true if two types are structurally equal.
func Equal(a, b *Type) bool {
        if a == b {
                return true
        }
        if a == nil || b == nil {
                return false
        }
        // Unwrap aliases
        if a.Kind == KindAlias {
                return Equal(a.BaseT, b)
        }
        if b.Kind == KindAlias {
                return Equal(a, b.BaseT)
        }

        if a.Kind != b.Kind {
                return false
        }

        switch a.Kind {
        case KindPrimitive, KindTypeParam, KindNone, KindNever, KindAny:
                return a.Name == b.Name
        case KindStruct:
                return a.Name == b.Name && typeArgsEqual(a.TypeArgs, b.TypeArgs)
        case KindEnum:
                return a.Name == b.Name && typeArgsEqual(a.TypeArgs, b.TypeArgs)
        case KindUnion, KindIntersection, KindTuple:
                if len(a.Members) != len(b.Members) {
                        return false
                }
                for i := range a.Members {
                        if !Equal(a.Members[i], b.Members[i]) {
                                return false
                        }
                }
                return true
        case KindList, KindSet, KindOption:
                return Equal(a.ElementT, b.ElementT)
        case KindMap:
                return Equal(a.KeyT, b.KeyT) && Equal(a.ValueT, b.ValueT)
        case KindResult:
                return Equal(a.ElementT, b.ElementT) && Equal(a.ValueT, b.ValueT)
        case KindFunction:
                if len(a.ParamTypes) != len(b.ParamTypes) {
                        return false
                }
                for i := range a.ParamTypes {
                        if !Equal(a.ParamTypes[i], b.ParamTypes[i]) {
                                return false
                        }
                }
                return Equal(a.ReturnT, b.ReturnT)
        case KindRefinement:
                return Equal(a.BaseT, b.BaseT) && a.Predicate == b.Predicate
        case KindStringLit:
                return a.StringVal == b.StringVal
        }
        return false
}

func typeArgsEqual(a, b []*Type) bool {
        if len(a) != len(b) {
                return false
        }
        for i := range a {
                if !Equal(a[i], b[i]) {
                        return false
                }
        }
        return true
}

// IsAssignableTo checks if type `from` can be assigned to type `to`.
// This implements the subtyping rules from the Aura spec.
func IsAssignableTo(from, to *Type) bool {
        if from == nil || to == nil {
                return false
        }

        // Unwrap aliases
        if from.Kind == KindAlias {
                return IsAssignableTo(from.BaseT, to)
        }
        if to.Kind == KindAlias {
                return IsAssignableTo(from, to.BaseT)
        }

        // Equal types are always assignable
        if Equal(from, to) {
                return true
        }

        // Never is subtype of everything
        if from.Kind == KindNever {
                return true
        }

        // Everything is subtype of Any
        if to.Kind == KindAny {
                return true
        }

        // Any (unknown) is assignable to anything (for inference gaps)
        if from.Kind == KindAny {
                return true
        }

        // None is assignable to Option[T]
        if from.Kind == KindNone && to.Kind == KindOption {
                return true
        }

        // T is assignable to Option[T]
        if to.Kind == KindOption {
                return IsAssignableTo(from, to.ElementT)
        }

        // Refinement type is subtype of its base
        if from.Kind == KindRefinement {
                return IsAssignableTo(from.BaseT, to)
        }

        // String literal type is subtype of String
        if from.Kind == KindStringLit && to.Kind == KindPrimitive && to.Name == "String" {
                return true
        }

        // String literal is assignable to union containing that literal
        if from.Kind == KindStringLit && to.Kind == KindUnion {
                for _, m := range to.Members {
                        if IsAssignableTo(from, m) {
                                return true
                        }
                }
                return false
        }

        // Union type: from is subtype of to if every member of from is subtype of to
        if from.Kind == KindUnion {
                for _, m := range from.Members {
                        if !IsAssignableTo(m, to) {
                                return false
                        }
                }
                return true
        }

        // Assigning to union: from needs to be assignable to at least one member
        if to.Kind == KindUnion {
                for _, m := range to.Members {
                        if IsAssignableTo(from, m) {
                                return true
                        }
                }
                return false
        }

        // Struct width subtyping: struct with more fields is subtype of one with fewer
        if from.Kind == KindStruct && to.Kind == KindStruct {
                return structSubtype(from, to)
        }

        // Int -> Float widening
        if from.Kind == KindPrimitive && from.Name == "Int" &&
                to.Kind == KindPrimitive && to.Name == "Float" {
                return true
        }

        return false
}

// structSubtype checks width subtyping: from has at least all fields of to with compatible types.
func structSubtype(from, to *Type) bool {
        for _, toField := range to.Fields {
                found := false
                for _, fromField := range from.Fields {
                        if fromField.Name == toField.Name {
                                if !IsAssignableTo(fromField.Type, toField.Type) {
                                        return false
                                }
                                found = true
                                break
                        }
                }
                if !found {
                        return false
                }
        }
        return true
}

// Underlying returns the underlying type, unwrapping aliases and refinements.
func Underlying(t *Type) *Type {
        for t != nil {
                switch t.Kind {
                case KindAlias:
                        t = t.BaseT
                case KindRefinement:
                        t = t.BaseT
                default:
                        return t
                }
        }
        return t
}

// --- Type Registry ---

// Registry stores all types defined in a module and provides type resolution.
type Registry struct {
        types map[string]*Type
}

// NewRegistry creates a new type registry pre-populated with builtins.
func NewRegistry() *Registry {
        r := &Registry{types: make(map[string]*Type)}
        for name, t := range Builtins {
                r.types[name] = t
        }
        return r
}

// Register adds a type to the registry.
func (r *Registry) Register(name string, t *Type) error {
        if _, exists := r.types[name]; exists {
                if _, isBuiltin := Builtins[name]; isBuiltin {
                        return fmt.Errorf("cannot redefine builtin type %q", name)
                }
                return fmt.Errorf("type %q already defined", name)
        }
        r.types[name] = t
        return nil
}

// Lookup looks up a type by name.
func (r *Registry) Lookup(name string) (*Type, bool) {
        t, ok := r.types[name]
        return t, ok
}
