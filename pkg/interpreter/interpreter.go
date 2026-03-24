package interpreter

import (
        "fmt"
        "path/filepath"
        "strings"

        "github.com/unclebucklarson/aura/pkg/ast"
        "github.com/unclebucklarson/aura/pkg/module"
        "github.com/unclebucklarson/aura/pkg/token"
)

// Interpreter executes an Aura module.
type Interpreter struct {
        module   *ast.Module
        env      *Environment
        resolver *module.Resolver
        filePath string // path of the source file being interpreted
        effects  *EffectContext // effect system capabilities
}

// New creates a new interpreter for the given module.
func New(mod *ast.Module) *Interpreter {
        interp := &Interpreter{
                module:  mod,
                env:     NewEnvironment(),
                effects: NewEffectContext(),
        }
        interp.registerBuiltins()
        return interp
}

// NewWithResolver creates an interpreter with a module resolver for import support.
func NewWithResolver(mod *ast.Module, filePath string, resolver *module.Resolver) *Interpreter {
        interp := &Interpreter{
                module:   mod,
                env:      NewEnvironment(),
                resolver: resolver,
                filePath: filePath,
                effects:  NewEffectContext(),
        }
        interp.registerBuiltins()
        return interp
}

// NewWithEffects creates an interpreter with a custom effect context.
// This is useful for testing with mock effect providers.
func NewWithEffects(mod *ast.Module, effects *EffectContext) *Interpreter {
        interp := &Interpreter{
                module:  mod,
                env:     NewEnvironment(),
                effects: effects,
        }
        interp.registerBuiltins()
        return interp
}

// NewWithResolverAndEffects creates an interpreter with both resolver and custom effects.
func NewWithResolverAndEffects(mod *ast.Module, filePath string, resolver *module.Resolver, effects *EffectContext) *Interpreter {
        interp := &Interpreter{
                module:   mod,
                env:      NewEnvironment(),
                resolver: resolver,
                filePath: filePath,
                effects:  effects,
        }
        interp.registerBuiltins()
        return interp
}

// Effects returns the interpreter's effect context.
func (interp *Interpreter) Effects() *EffectContext {
        return interp.effects
}

// registerBuiltins adds built-in functions and constructors to the environment.
func (interp *Interpreter) registerBuiltins() {
        env := interp.env

        // Ok constructor
        env.DefineConst("Ok", &BuiltinFnVal{
                Name: "Ok",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "Ok() requires exactly one argument"})
                        }
                        return &ResultVal{IsOk: true, Val: args[0]}
                },
        })

        // Err constructor
        env.DefineConst("Err", &BuiltinFnVal{
                Name: "Err",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "Err() requires exactly one argument"})
                        }
                        return &ResultVal{IsOk: false, Val: args[0]}
                },
        })

        // Some constructor
        env.DefineConst("Some", &BuiltinFnVal{
                Name: "Some",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "Some() requires exactly one argument"})
                        }
                        return &OptionVal{IsSome: true, Val: args[0]}
                },
        })

        // None value
        env.DefineConst("None", &OptionVal{IsSome: false})

        // print function
        env.DefineConst("print", &BuiltinFnVal{
                Name: "print",
                Fn: func(args []Value) Value {
                        parts := make([]string, len(args))
                        for i, a := range args {
                                parts[i] = a.String()
                        }
                        fmt.Println(joinStrings(parts, " "))
                        return &NoneVal{}
                },
        })

        // len function
        env.DefineConst("len", &BuiltinFnVal{
                Name: "len",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "len() requires exactly one argument"})
                        }
                        switch v := args[0].(type) {
                        case *ListVal:
                                return &IntVal{Val: int64(len(v.Elements))}
                        case *MapVal:
                                return &IntVal{Val: int64(len(v.Keys))}
                        case *SetVal:
                                return &IntVal{Val: int64(len(v.Elements))}
                        case *StringVal:
                                return &IntVal{Val: int64(len(v.Val))}
                        case *TupleVal:
                                return &IntVal{Val: int64(len(v.Elements))}
                        default:
                                panic(&RuntimeError{Message: fmt.Sprintf("len() not supported for %s", valueTypeNames[args[0].Type()])})
                        }
                },
        })

        // str function
        env.DefineConst("str", &BuiltinFnVal{
                Name: "str",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "str() requires exactly one argument"})
                        }
                        return &StringVal{Val: args[0].String()}
                },
        })

        // int function
        env.DefineConst("int", &BuiltinFnVal{
                Name: "int",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "int() requires exactly one argument"})
                        }
                        switch v := args[0].(type) {
                        case *IntVal:
                                return v
                        case *FloatVal:
                                return &IntVal{Val: int64(v.Val)}
                        case *StringVal:
                                n, err := fmt.Sscanf(v.Val, "%d", new(int64))
                                if err != nil || n != 1 {
                                        panic(&RuntimeError{Message: fmt.Sprintf("cannot convert '%s' to int", v.Val)})
                                }
                                var i int64
                                fmt.Sscanf(v.Val, "%d", &i)
                                return &IntVal{Val: i}
                        case *BoolVal:
                                if v.Val {
                                        return &IntVal{Val: 1}
                                }
                                return &IntVal{Val: 0}
                        default:
                                panic(&RuntimeError{Message: fmt.Sprintf("cannot convert %s to int", valueTypeNames[args[0].Type()])})
                        }
                },
        })

        // float function
        env.DefineConst("float", &BuiltinFnVal{
                Name: "float",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "float() requires exactly one argument"})
                        }
                        switch v := args[0].(type) {
                        case *FloatVal:
                                return v
                        case *IntVal:
                                return &FloatVal{Val: float64(v.Val)}
                        default:
                                panic(&RuntimeError{Message: fmt.Sprintf("cannot convert %s to float", valueTypeNames[args[0].Type()])})
                        }
                },
        })

        // range function for for loops
        env.DefineConst("range", &BuiltinFnVal{
                Name: "range",
                Fn: func(args []Value) Value {
                        var start, end, step int64
                        switch len(args) {
                        case 1:
                                e, ok := args[0].(*IntVal)
                                if !ok {
                                        panic(&RuntimeError{Message: "range() requires integer arguments"})
                                }
                                start, end, step = 0, e.Val, 1
                        case 2:
                                s, ok1 := args[0].(*IntVal)
                                e, ok2 := args[1].(*IntVal)
                                if !ok1 || !ok2 {
                                        panic(&RuntimeError{Message: "range() requires integer arguments"})
                                }
                                start, end, step = s.Val, e.Val, 1
                        case 3:
                                s, ok1 := args[0].(*IntVal)
                                e, ok2 := args[1].(*IntVal)
                                st, ok3 := args[2].(*IntVal)
                                if !ok1 || !ok2 || !ok3 {
                                        panic(&RuntimeError{Message: "range() requires integer arguments"})
                                }
                                start, end, step = s.Val, e.Val, st.Val
                        default:
                                panic(&RuntimeError{Message: "range() requires 1-3 arguments"})
                        }
                        if step == 0 {
                                panic(&RuntimeError{Message: "range() step cannot be zero"})
                        }
                        elems := make([]Value, 0)
                        if step > 0 {
                                for i := start; i < end; i += step {
                                        elems = append(elems, &IntVal{Val: i})
                                }
                        } else {
                                for i := start; i > end; i += step {
                                        elems = append(elems, &IntVal{Val: i})
                                }
                        }
                        return &ListVal{Elements: elems}
                },
        })

        // type function
        env.DefineConst("type_of", &BuiltinFnVal{
                Name: "type_of",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "type_of() requires exactly one argument"})
                        }
                        return &StringVal{Val: valueTypeNames[args[0].Type()]}
                },
        })

        // abs function
        env.DefineConst("abs", &BuiltinFnVal{
                Name: "abs",
                Fn: func(args []Value) Value {
                        if len(args) != 1 {
                                panic(&RuntimeError{Message: "abs() requires exactly one argument"})
                        }
                        switch v := args[0].(type) {
                        case *IntVal:
                                if v.Val < 0 {
                                        return &IntVal{Val: -v.Val}
                                }
                                return v
                        case *FloatVal:
                                if v.Val < 0 {
                                        return &FloatVal{Val: -v.Val}
                                }
                                return v
                        default:
                                panic(&RuntimeError{Message: fmt.Sprintf("abs() not supported for %s", valueTypeNames[args[0].Type()])})
                        }
                },
        })

        // min / max
        env.DefineConst("min", &BuiltinFnVal{
                Name: "min",
                Fn: func(args []Value) Value {
                        if len(args) < 2 {
                                panic(&RuntimeError{Message: "min() requires at least 2 arguments"})
                        }
                        result := args[0]
                        for _, a := range args[1:] {
                                if compareRaw(a, result) < 0 {
                                        result = a
                                }
                        }
                        return result
                },
        })

        env.DefineConst("max", &BuiltinFnVal{
                Name: "max",
                Fn: func(args []Value) Value {
                        if len(args) < 2 {
                                panic(&RuntimeError{Message: "max() requires at least 2 arguments"})
                        }
                        result := args[0]
                        for _, a := range args[1:] {
                                if compareRaw(a, result) > 0 {
                                        result = a
                                }
                        }
                        return result
                },
        })
}

func joinStrings(parts []string, sep string) string {
        result := ""
        for i, p := range parts {
                if i > 0 {
                        result += sep
                }
                result += p
        }
        return result
}

func compareRaw(a, b Value) int {
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
                var bf float64
                switch bv := b.(type) {
                case *FloatVal:
                        bf = bv.Val
                case *IntVal:
                        bf = float64(bv.Val)
                }
                if av.Val < bf {
                        return -1
                }
                if av.Val > bf {
                        return 1
                }
                return 0
        }
        return 0
}

// Run executes the module's top-level items.
func (interp *Interpreter) Run() (result Value, err error) {
        defer func() {
                if r := recover(); r != nil {
                        switch e := r.(type) {
                        case *RuntimeError:
                                err = e
                        case returnSignal:
                                result = e.val
                        default:
                                err = fmt.Errorf("runtime panic: %v", r)
                        }
                }
        }()

        result = &NoneVal{}

        // Process imports first
        if err := interp.processImports(); err != nil {
                return nil, err
        }

        // First pass: register all type definitions, functions, constants, enums
        for _, item := range interp.module.Items {
                switch it := item.(type) {
                case *ast.StructDef:
                        fields := make([]string, len(it.Fields))
                        for i, f := range it.Fields {
                                fields[i] = f.Name
                        }
                        interp.env.DefineStruct(it.Name, fields)

                case *ast.EnumDef:
                        variants := make(map[string]int, len(it.Variants))
                        for _, v := range it.Variants {
                                variants[v.Name] = len(v.Fields)
                        }
                        interp.env.DefineEnum(it.Name, variants)

                case *ast.FnDef:
                        fn := &FunctionVal{
                                Name:   it.Name,
                                Params: it.Params,
                                Body:   it.Body,
                                Env:    interp.env,
                        }
                        interp.env.DefineConst(it.Name, fn)

                case *ast.ConstDef:
                        val := EvalExpr(it.Value, interp.env)
                        interp.env.DefineConst(it.Name, val)

                case *ast.TypeDef:
                        // Type aliases: no runtime effect
                case *ast.TraitDef:
                        // Traits: no runtime effect for now
                case *ast.ImplBlock:
                        // Impl blocks: no runtime effect for now
                case *ast.SpecBlock:
                        // Specs: no runtime effect
                case *ast.TestBlock:
                        // Tests: handled separately
                }
        }

        return result, nil
}

// RunFunction calls a named function with the given arguments.
func (interp *Interpreter) RunFunction(name string, args []Value) (result Value, err error) {
        defer func() {
                if r := recover(); r != nil {
                        switch e := r.(type) {
                        case *RuntimeError:
                                err = e
                        case returnSignal:
                                result = e.val
                        default:
                                err = fmt.Errorf("runtime panic: %v", r)
                        }
                }
        }()

        fnVal, ok := interp.env.Get(name)
        if !ok {
                return nil, fmt.Errorf("function '%s' not found", name)
        }

        fn, ok := fnVal.(*FunctionVal)
        if !ok {
                return nil, fmt.Errorf("'%s' is not a function", name)
        }

        var span ast.Node
        if len(fn.Params) > 0 {
                span = fn.Params[0]
        }
        var spanVal token.Span
        if span != nil {
                spanVal = span.GetSpan()
        }
        result = callUserFn(spanVal, fn, args)
        return result, nil
}

// Env returns the interpreter's environment (for testing/REPL).
func (interp *Interpreter) Env() *Environment {
        return interp.env
}

// processImports handles all import statements in the module.
func (interp *Interpreter) processImports() error {
        if interp.module.Imports == nil || len(interp.module.Imports) == 0 {
                return nil
        }

        if interp.resolver == nil {
                return fmt.Errorf("import statements require a module resolver (use NewWithResolver)")
        }

        fromDir := "."
        if interp.filePath != "" {
                fromDir = filepath.Dir(interp.filePath)
        }

        for _, imp := range interp.module.Imports {
                importPath := imp.Path.String()

                // Check for standard library imports
                if module.IsStdLib(importPath) {
                        if err := interp.processStdImport(imp, importPath); err != nil {
                                return err
                        }
                        continue
                }

                // Resolve the module
                cached, err := interp.resolver.Resolve(importPath, fromDir)
                if err != nil {
                        return err
                }

                // Load and execute the module to get its exports
                modVal, err := interp.loadModuleValue(cached, importPath)
                if err != nil {
                        return err
                }

                // Bind the module into the current environment
                if err := interp.bindImport(imp, modVal); err != nil {
                        return err
                }
        }

        return nil
}

// processStdImport handles standard library imports.
func (interp *Interpreter) processStdImport(imp *ast.ImportNode, importPath string) error {
        modVal := interp.createStdModule(importPath)
        if modVal == nil {
                return fmt.Errorf("unknown standard library module: '%s'", importPath)
        }
        return interp.bindImport(imp, modVal)
}

// createStdModule creates a ModuleVal for a standard library module.
func (interp *Interpreter) createStdModule(importPath string) *ModuleVal {
        parts := strings.Split(importPath, ".")
        if len(parts) < 2 || parts[0] != "std" {
                return nil
        }

        modName := parts[len(parts)-1]
        exports := make(map[string]Value)

        switch importPath {
        case "std.math":
                exports = createStdMathExports()

        case "std.string":
                exports = createStdStringExports()

        case "std.io":
                exports = createStdIoExports()

        case "std.testing":
                exports = createStdTestingExports()
                // Merge in effect-aware testing helpers
                effectExports := createStdTestingEffectExports(interp)
                for k, v := range effectExports {
                        exports[k] = v
                }

        case "std.json":
                exports = createStdJsonExports()

        case "std.regex":
                exports = createStdRegexExports()

        case "std.collections":
                exports = createStdCollectionsExports()

        case "std.random":
                exports = createStdRandomExports()

        case "std.format":
                exports = createStdFormatExports()

        case "std.result":
                exports = createStdResultExports()

        case "std.option":
                exports = createStdOptionExports()

        case "std.iter":
                exports = createStdIterExports()

        case "std.file":
                exports = createStdFileExports(interp.effects.File())

        case "std.time":
                exports = createStdTimeExports(interp.effects.Time())

        case "std.env":
                exports = createStdEnvExports(interp.effects.Env())

        case "std.net":
                exports = createStdNetExports(interp.effects.Net())

        case "std.log":
                exports = createStdLogExports(interp.effects.Log())

        default:
                return nil
        }

        return &ModuleVal{
                Name:    modName,
                Path:    importPath,
                Exports: exports,
        }
}

// loadModuleValue executes a cached module and returns a ModuleVal with its exports.
func (interp *Interpreter) loadModuleValue(cached *module.CachedModule, importPath string) (*ModuleVal, error) {
        if cached.AST == nil {
                return nil, fmt.Errorf("module '%s' has no AST (internal error)", importPath)
        }

        // Check if module is already initialized (prevent re-initialization)
        if interp.resolver != nil {
                state := interp.resolver.GetInitState(cached.Path)
                if state == module.InitComplete {
                        // Module already initialized - return cached exports
                        modName := module.GetModuleName(importPath)
                        exports := make(map[string]Value)
                        // Re-create a child interpreter just to get the env, but skip execution
                        childInterp := NewWithResolver(cached.AST, cached.Path, interp.resolver)
                        // Register definitions without executing
                        for _, item := range cached.AST.Items {
                                switch it := item.(type) {
                                case *ast.FnDef:
                                        fn := &FunctionVal{
                                                Name:   it.Name,
                                                Params: it.Params,
                                                Body:   it.Body,
                                                Env:    childInterp.env,
                                        }
                                        childInterp.env.DefineConst(it.Name, fn)
                                case *ast.ConstDef:
                                        val := EvalExpr(it.Value, childInterp.env)
                                        childInterp.env.DefineConst(it.Name, val)
                                case *ast.StructDef:
                                        fields := make([]string, len(it.Fields))
                                        for i, f := range it.Fields {
                                                fields[i] = f.Name
                                        }
                                        childInterp.env.DefineStruct(it.Name, fields)
                                case *ast.EnumDef:
                                        variants := make(map[string]int, len(it.Variants))
                                        for _, v := range it.Variants {
                                                variants[v.Name] = len(v.Fields)
                                        }
                                        childInterp.env.DefineEnum(it.Name, variants)
                                }
                        }
                        for name := range cached.Exports {
                                if val, ok := childInterp.env.Get(name); ok {
                                        exports[name] = val
                                }
                        }
                        return &ModuleVal{
                                Name:    modName,
                                Path:    cached.Path,
                                Exports: exports,
                        }, nil
                }
                if state == module.InitInProgress {
                        return nil, fmt.Errorf("circular initialization detected for module '%s'", importPath)
                }
                // Mark as in-progress
                interp.resolver.SetInitState(cached.Path, module.InitInProgress)
        }

        // Create a child interpreter for the imported module
        childInterp := NewWithResolver(cached.AST, cached.Path, interp.resolver)

        // Execute the module's top-level items (this triggers package-level initialization)
        if _, err := childInterp.Run(); err != nil {
                if interp.resolver != nil {
                        interp.resolver.SetInitState(cached.Path, module.InitError)
                }
                return nil, fmt.Errorf("error initializing module '%s': %v", importPath, err)
        }

        // Mark as initialized
        if interp.resolver != nil {
                interp.resolver.SetInitState(cached.Path, module.InitComplete)
        }

        // Collect exports
        modName := module.GetModuleName(importPath)
        exports := make(map[string]Value)

        for name := range cached.Exports {
                if val, ok := childInterp.env.Get(name); ok {
                        exports[name] = val
                }
        }

        return &ModuleVal{
                Name:    modName,
                Path:    cached.Path,
                Exports: exports,
        }, nil
}

// bindImport binds a module's exports into the current environment.
func (interp *Interpreter) bindImport(imp *ast.ImportNode, modVal *ModuleVal) error {
        importPath := imp.Path.String()

        if imp.Names != nil {
                // "from X import a, b" or "from X import *"
                if len(imp.Names) == 1 && imp.Names[0] == "*" {
                        // Wildcard import: bind all exports directly
                        for name, val := range modVal.Exports {
                                interp.env.DefineConst(name, val)
                        }
                } else {
                        // Named imports: bind specific symbols
                        for _, name := range imp.Names {
                                val, ok := modVal.Exports[name]
                                if !ok {
                                        availableExports := make([]string, 0, len(modVal.Exports))
                                        for k := range modVal.Exports {
                                                availableExports = append(availableExports, k)
                                        }
                                        return fmt.Errorf("module '%s' does not export '%s' (available: %s)",
                                                importPath, name, strings.Join(availableExports, ", "))
                                }
                                interp.env.DefineConst(name, val)
                        }
                }
        } else if imp.Alias != "" {
                // "import X as Y": bind module under alias
                interp.env.DefineConst(imp.Alias, modVal)
        } else {
                // "import X": bind module under its short name (e.g., "io" for "std.io")
                modName := module.GetModuleName(importPath)
                interp.env.DefineConst(modName, modVal)

                // For std library imports, also bind under the "std" namespace
                // so that both io.println() and std.io.println() work.
                if strings.HasPrefix(importPath, "std.") {
                        parts := strings.Split(importPath, ".")
                        if len(parts) == 2 {
                                // Create or extend the "std" namespace module
                                var stdMod *ModuleVal
                                if existing, ok := interp.env.Get("std"); ok {
                                        if m, ok := existing.(*ModuleVal); ok {
                                                stdMod = m
                                        }
                                }
                                if stdMod == nil {
                                        stdMod = &ModuleVal{
                                                Name:    "std",
                                                Path:    "std",
                                                Exports: make(map[string]Value),
                                        }
                                        interp.env.DefineConst("std", stdMod)
                                }
                                stdMod.Exports[parts[1]] = modVal
                        }
                }
        }

        return nil
}
