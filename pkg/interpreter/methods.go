package interpreter

import "fmt"

// MethodFunc is the signature for all built-in methods.
// receiver is the value the method is called on, args are the call arguments.
type MethodFunc func(receiver Value, args []Value) Value

// methodRegistry maps type → method name → implementation.
var methodRegistry = map[ValueType]map[string]MethodFunc{}

// RegisterMethod registers a built-in method for a value type.
func RegisterMethod(vt ValueType, name string, fn MethodFunc) {
	if methodRegistry[vt] == nil {
		methodRegistry[vt] = map[string]MethodFunc{}
	}
	methodRegistry[vt][name] = fn
}

// LookupMethod returns the method function for a given type and name, or nil.
func LookupMethod(vt ValueType, name string) MethodFunc {
	if methods, ok := methodRegistry[vt]; ok {
		return methods[name]
	}
	return nil
}

// MethodNames returns all registered method names for a given type.
func MethodNames(vt ValueType) []string {
	methods, ok := methodRegistry[vt]
	if !ok {
		return nil
	}
	names := make([]string, 0, len(methods))
	for name := range methods {
		names = append(names, name)
	}
	return names
}

// resolveMethod looks up a method on a value and returns a BuiltinFnVal closure.
// Returns nil if no method is found.
func resolveMethod(obj Value, methodName string) *BuiltinFnVal {
	method := LookupMethod(obj.Type(), methodName)
	if method == nil {
		return nil
	}
	captured := obj // capture for closure
	return &BuiltinFnVal{
		Name: fmt.Sprintf("%s.%s", valueTypeNames[obj.Type()], methodName),
		Fn:   func(args []Value) Value { return method(captured, args) },
	}
}
