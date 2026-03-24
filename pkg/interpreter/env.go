package interpreter

import "fmt"

// Environment implements lexical scoping with a parent chain.
type Environment struct {
	parent *Environment
	values map[string]Value
	// consts tracks which names are immutable (let without mut)
	consts map[string]bool
	// structDefs stores struct definitions (name -> field names)
	structDefs map[string][]string
	// enumDefs stores enum definitions (name -> variant definitions)
	enumDefs map[string]map[string]int // enumName -> variantName -> arity
	// effects tracks available effect capabilities
	effects map[string]bool
}

// NewEnvironment creates a new top-level environment.
func NewEnvironment() *Environment {
	return &Environment{
		values:     make(map[string]Value),
		consts:     make(map[string]bool),
		structDefs: make(map[string][]string),
		enumDefs:   make(map[string]map[string]int),
		effects:    make(map[string]bool),
	}
}

// NewEnclosedEnvironment creates a child environment.
func NewEnclosedEnvironment(parent *Environment) *Environment {
	env := NewEnvironment()
	env.parent = parent
	return env
}

// Define defines a new variable in the current scope.
func (e *Environment) Define(name string, val Value) {
	e.values[name] = val
}

// DefineConst defines an immutable variable.
func (e *Environment) DefineConst(name string, val Value) {
	e.values[name] = val
	e.consts[name] = true
}

// Get looks up a variable by name, walking the parent chain.
func (e *Environment) Get(name string) (Value, bool) {
	val, ok := e.values[name]
	if ok {
		return val, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

// Set updates an existing variable's value, walking the parent chain.
func (e *Environment) Set(name string, val Value) error {
	if _, ok := e.values[name]; ok {
		if e.consts[name] {
			return fmt.Errorf("cannot assign to immutable variable '%s'", name)
		}
		e.values[name] = val
		return nil
	}
	if e.parent != nil {
		return e.parent.Set(name, val)
	}
	return fmt.Errorf("undefined variable '%s'", name)
}

// DefineStruct registers a struct definition.
func (e *Environment) DefineStruct(name string, fields []string) {
	e.structDefs[name] = fields
}

// GetStructDef looks up a struct definition.
func (e *Environment) GetStructDef(name string) ([]string, bool) {
	fields, ok := e.structDefs[name]
	if ok {
		return fields, true
	}
	if e.parent != nil {
		return e.parent.GetStructDef(name)
	}
	return nil, false
}

// DefineEnum registers an enum definition.
func (e *Environment) DefineEnum(name string, variants map[string]int) {
	e.enumDefs[name] = variants
}

// GetEnumDef looks up an enum definition.
func (e *Environment) GetEnumDef(name string) (map[string]int, bool) {
	variants, ok := e.enumDefs[name]
	if ok {
		return variants, true
	}
	if e.parent != nil {
		return e.parent.GetEnumDef(name)
	}
	return nil, false
}

// AddEffect adds an effect capability to this scope.
func (e *Environment) AddEffect(name string) {
	e.effects[name] = true
}

// HasEffect checks if an effect capability is available.
func (e *Environment) HasEffect(name string) bool {
	if e.effects[name] {
		return true
	}
	if e.parent != nil {
		return e.parent.HasEffect(name)
	}
	return false
}
