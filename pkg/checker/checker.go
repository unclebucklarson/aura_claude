package checker

import (
        "fmt"
        "strings"

        "github.com/unclebucklarson/aura/pkg/ast"
        "github.com/unclebucklarson/aura/pkg/symbols"
        "github.com/unclebucklarson/aura/pkg/token"
        "github.com/unclebucklarson/aura/pkg/types"
)

// Checker performs type checking and semantic analysis on an Aura AST.
type Checker struct {
        module   *ast.Module
        symTable *symbols.Table
        typeReg  *types.Registry
        errors   []*CheckError

        // Current function context for return type checking and effect tracking
        currentFn       *ast.FnDef
        currentFnReturn *types.Type
        currentEffects  map[string]bool // effects declared by current function

        // Spec tracking
        specs map[string]*ast.SpecBlock // spec name -> spec block
        // Function declared effects (for effect checking across calls)
        fnEffects map[string][]string // fn name -> declared effects
        // Function types
        fnTypes map[string]*types.Type // fn name -> function type
        // Variable type tracking (symbol name in scope -> resolved type)
        varTypes map[string]*types.Type
}

// New creates a new Checker for the given module.
func New(module *ast.Module) *Checker {
        moduleName := ""
        if module.Name != nil {
                moduleName = module.Name.String()
        }
        return &Checker{
                module:   module,
                symTable: symbols.NewTable(moduleName),
                typeReg:  types.NewRegistry(),
                specs:    make(map[string]*ast.SpecBlock),
                fnEffects: make(map[string][]string),
                fnTypes:  make(map[string]*types.Type),
                varTypes: make(map[string]*types.Type),
        }
}

// Check runs the type checker and returns all errors.
func (c *Checker) Check() []*CheckError {
        c.errors = nil

        // Pass 1: Register all top-level type definitions
        c.registerTypes()

        // Pass 2: Register all specs
        c.registerSpecs()

        // Pass 3: Register all function signatures
        c.registerFunctions()

        // Pass 4: Register constants
        c.registerConstants()

        // Pass 5: Check function bodies
        c.checkFunctionBodies()

        // Pass 6: Validate spec contracts
        c.validateSpecs()

        // Pass 7: Check test blocks
        c.checkTestBlocks()

        return c.errors
}

// Errors returns accumulated errors.
func (c *Checker) Errors() []*CheckError {
        return c.errors
}

// addError appends a check error.
func (c *Checker) addError(err *CheckError) {
        c.errors = append(c.errors, err)
}

// errorf creates and adds a simple error.
func (c *Checker) errorf(code ErrorCode, span token.Span, format string, args ...interface{}) {
        c.addError(newError(code, span, fmt.Sprintf(format, args...)))
}

// --- Pass 1: Register Types ---

func (c *Checker) registerTypes() {
        for _, item := range c.module.Items {
                switch n := item.(type) {
                case *ast.TypeDef:
                        c.registerTypeDef(n)
                case *ast.StructDef:
                        c.registerStructDef(n)
                case *ast.EnumDef:
                        c.registerEnumDef(n)
                case *ast.TraitDef:
                        c.registerTraitDef(n)
                }
        }
}

func (c *Checker) registerTypeDef(td *ast.TypeDef) {
        bodyType := c.resolveTypeExpr(td.Body)
        aliasType := types.NewAliasType(td.Name, bodyType)
        aliasType.TypeParams = td.TypeParams

        if err := c.typeReg.Register(td.Name, aliasType); err != nil {
                c.addError(newError(ErrRedefinedType, td.Span,
                        fmt.Sprintf("type %q is already defined", td.Name)))
                return
        }

        c.symTable.Define(&symbols.Symbol{
                Name:       td.Name,
                Kind:       symbols.SymType,
                Span:       td.Span,
                Public:     td.Visibility == ast.Public,
                TypeParams: td.TypeParams,
        })
}

func (c *Checker) registerStructDef(sd *ast.StructDef) {
        fields := make([]*types.Field, len(sd.Fields))
        for i, f := range sd.Fields {
                ft := c.resolveTypeExpr(f.TypeExpr)
                fields[i] = &types.Field{
                        Name:     f.Name,
                        Type:     ft,
                        Optional: f.Default != nil,
                        Public:   f.Visibility == ast.Public,
                }
        }

        structType := types.NewStructType(sd.Name, fields, sd.TypeParams)
        if err := c.typeReg.Register(sd.Name, structType); err != nil {
                c.addError(newError(ErrRedefinedType, sd.Span,
                        fmt.Sprintf("type %q is already defined", sd.Name)))
                return
        }

        c.symTable.Define(&symbols.Symbol{
                Name:       sd.Name,
                Kind:       symbols.SymStruct,
                Span:       sd.Span,
                Public:     sd.Visibility == ast.Public,
                TypeParams: sd.TypeParams,
        })
}

func (c *Checker) registerEnumDef(ed *ast.EnumDef) {
        variants := make([]*types.Variant, len(ed.Variants))
        for i, v := range ed.Variants {
                vfields := make([]*types.Type, len(v.Fields))
                for j, f := range v.Fields {
                        vfields[j] = c.resolveTypeExpr(f)
                }
                variants[i] = &types.Variant{
                        Name:   v.Name,
                        Fields: vfields,
                }
        }

        enumType := types.NewEnumType(ed.Name, variants, ed.TypeParams)
        if err := c.typeReg.Register(ed.Name, enumType); err != nil {
                c.addError(newError(ErrRedefinedType, ed.Span,
                        fmt.Sprintf("type %q is already defined", ed.Name)))
                return
        }

        c.symTable.Define(&symbols.Symbol{
                Name:       ed.Name,
                Kind:       symbols.SymEnum,
                Span:       ed.Span,
                Public:     ed.Visibility == ast.Public,
                TypeParams: ed.TypeParams,
        })

        // Register enum variants as symbols too
        for _, v := range ed.Variants {
                qualName := ed.Name + "." + v.Name
                c.symTable.Define(&symbols.Symbol{
                        Name: qualName,
                        Kind: symbols.SymEnumVariant,
                        Span: v.Span,
                })
        }
}

func (c *Checker) registerTraitDef(td *ast.TraitDef) {
        c.symTable.Define(&symbols.Symbol{
                Name:       td.Name,
                Kind:       symbols.SymTrait,
                Span:       td.Span,
                Public:     td.Visibility == ast.Public,
                TypeParams: td.TypeParams,
        })
}

// --- Pass 2: Register Specs ---

func (c *Checker) registerSpecs() {
        for _, item := range c.module.Items {
                if spec, ok := item.(*ast.SpecBlock); ok {
                        if _, exists := c.specs[spec.Name]; exists {
                                c.addError(newError(ErrSpecDuplicate, spec.Span,
                                        fmt.Sprintf("spec %q is already defined", spec.Name)))
                                continue
                        }
                        c.specs[spec.Name] = spec
                        c.symTable.Define(&symbols.Symbol{
                                Name:   spec.Name,
                                Kind:   symbols.SymSpec,
                                Span:   spec.Span,
                                Effects: spec.Effects,
                        })
                }
        }
}

// --- Pass 3: Register Functions ---

func (c *Checker) registerFunctions() {
        for _, item := range c.module.Items {
                switch n := item.(type) {
                case *ast.FnDef:
                        c.registerFnDef(n)
                case *ast.ImplBlock:
                        for _, method := range n.Methods {
                                c.registerFnDef(method)
                        }
                }
        }
}

func (c *Checker) registerFnDef(fn *ast.FnDef) {
        paramTypes := make([]*types.Type, len(fn.Params))
        for i, p := range fn.Params {
                paramTypes[i] = c.resolveTypeExpr(p.TypeExpr)
        }

        retType := types.BuiltinNone
        if fn.ReturnType != nil {
                retType = c.resolveTypeExpr(fn.ReturnType)
        }

        fnType := types.NewFunctionType(paramTypes, retType, fn.Effects)
        fnType.TypeParams = fn.TypeParams

        c.fnTypes[fn.Name] = fnType
        c.fnEffects[fn.Name] = fn.Effects

        err := c.symTable.Define(&symbols.Symbol{
                Name:       fn.Name,
                Kind:       symbols.SymFunction,
                Span:       fn.Span,
                Public:     fn.Visibility == ast.Public,
                SpecName:   fn.Satisfies,
                Effects:    fn.Effects,
                TypeParams: fn.TypeParams,
        })
        if err != nil {
                c.addError(newError(ErrRedefinedName, fn.Span,
                        fmt.Sprintf("function %q is already defined", fn.Name)))
        }
}

// --- Pass 4: Register Constants ---

func (c *Checker) registerConstants() {
        for _, item := range c.module.Items {
                if cd, ok := item.(*ast.ConstDef); ok {
                        constType := c.resolveTypeExpr(cd.TypeExpr)
                        if cd.Value != nil {
                                valType := c.inferExpr(cd.Value)
                                if !types.IsAssignableTo(valType, constType) {
                                        c.addError(newError(ErrTypeMismatch, cd.Span,
                                                fmt.Sprintf("constant %q: cannot assign %s to %s",
                                                        cd.Name, valType, constType)).
                                                withExpectedGot(constType.String(), valType.String()))
                                }
                        }
                        c.symTable.Define(&symbols.Symbol{
                                Name:   cd.Name,
                                Kind:   symbols.SymConst,
                                Span:   cd.Span,
                                Public: cd.Visibility == ast.Public,
                        })
                }
        }
}

// --- Pass 5: Check Function Bodies ---

func (c *Checker) checkFunctionBodies() {
        for _, item := range c.module.Items {
                switch n := item.(type) {
                case *ast.FnDef:
                        c.checkFnBody(n)
                case *ast.ImplBlock:
                        for _, method := range n.Methods {
                                c.checkFnBody(method)
                        }
                }
        }
}

func (c *Checker) checkFnBody(fn *ast.FnDef) {
        c.currentFn = fn
        c.currentFnReturn = types.BuiltinNone
        if fn.ReturnType != nil {
                c.currentFnReturn = c.resolveTypeExpr(fn.ReturnType)
        }
        c.currentEffects = make(map[string]bool)
        for _, e := range fn.Effects {
                c.currentEffects[e] = true
        }

        // Push function scope
        c.symTable.PushScope(symbols.ScopeFunction, fn.Name)

        // Register parameters
        for _, p := range fn.Params {
                paramType := c.resolveTypeExpr(p.TypeExpr)
                c.symTable.Define(&symbols.Symbol{
                        Name: p.Name,
                        Kind: symbols.SymParam,
                        Span: p.Span,
                })
                c.varTypes[p.Name] = paramType

                // Check default value type
                if p.Default != nil {
                        defType := c.inferExpr(p.Default)
                        expected := c.resolveTypeExpr(p.TypeExpr)
                        if !types.IsAssignableTo(defType, expected) {
                                c.addError(newError(ErrTypeMismatch, p.Span,
                                        fmt.Sprintf("default value for parameter %q: cannot assign %s to %s",
                                                p.Name, defType, expected)).
                                        withExpectedGot(expected.String(), defType.String()))
                        }
                }
        }

        // Check body statements
        for _, stmt := range fn.Body {
                c.checkStatement(stmt)
        }

        c.symTable.PopScope()
        c.currentFn = nil
        c.currentFnReturn = nil
        c.currentEffects = nil
}

// --- Pass 6: Validate Specs ---

func (c *Checker) validateSpecs() {
        for _, item := range c.module.Items {
                fn, ok := item.(*ast.FnDef)
                if !ok || fn.Satisfies == "" {
                        continue
                }

                spec, exists := c.specs[fn.Satisfies]
                if !exists {
                        c.addError(newError(ErrSpecNotFound, fn.Span,
                                fmt.Sprintf("function %q satisfies spec %q, but spec is not defined",
                                        fn.Name, fn.Satisfies)).
                                withFix(fmt.Sprintf("Define a spec block: spec %s:", fn.Satisfies)))
                        continue
                }

                // Check inputs match
                c.validateSpecInputs(fn, spec)

                // Check effects match
                c.validateSpecEffects(fn, spec)
        }
}

func (c *Checker) validateSpecInputs(fn *ast.FnDef, spec *ast.SpecBlock) {
        for _, specInput := range spec.Inputs {
                found := false
                for _, param := range fn.Params {
                        if param.Name == specInput.Name {
                                found = true
                                // Check type compatibility
                                specType := c.resolveTypeExpr(specInput.TypeExpr)
                                paramType := c.resolveTypeExpr(param.TypeExpr)
                                if !types.IsAssignableTo(paramType, specType) && !types.IsAssignableTo(specType, paramType) {
                                        c.addError(newError(ErrSpecInputMismatch, param.Span,
                                                fmt.Sprintf("spec %q expects input %q of type %s, but function has %s",
                                                        spec.Name, specInput.Name, specType, paramType)).
                                                withExpectedGot(specType.String(), paramType.String()))
                                }
                                break
                        }
                }
                if !found {
                        c.addError(newError(ErrSpecInputMismatch, fn.Span,
                                fmt.Sprintf("spec %q expects input %q, but function %q has no such parameter",
                                        spec.Name, specInput.Name, fn.Name)).
                                withFix(fmt.Sprintf("Add parameter: %s: %s", specInput.Name,
                                        c.typeExprString(specInput.TypeExpr))))
                }
        }
}

func (c *Checker) validateSpecEffects(fn *ast.FnDef, spec *ast.SpecBlock) {
        fnSet := make(map[string]bool)
        for _, e := range fn.Effects {
                fnSet[e] = true
        }
        specSet := make(map[string]bool)
        for _, e := range spec.Effects {
                specSet[e] = true
        }

        // Check for missing effects in function
        for _, e := range spec.Effects {
                if !fnSet[e] {
                        c.addError(newError(ErrSpecEffectMismatch, fn.Span,
                                fmt.Sprintf("spec %q requires effect %q, but function %q does not declare it",
                                        spec.Name, e, fn.Name)).
                                withFix(fmt.Sprintf("Add 'with %s' to function signature", e)))
                }
        }

        // Check for extra effects not in spec
        for _, e := range fn.Effects {
                if !specSet[e] {
                        c.addError(newError(ErrSpecEffectMismatch, fn.Span,
                                fmt.Sprintf("function %q declares effect %q, but spec %q does not include it",
                                        fn.Name, e, spec.Name)).asWarning())
                }
        }
}

// --- Pass 7: Check Test Blocks ---

func (c *Checker) checkTestBlocks() {
        for _, item := range c.module.Items {
                if tb, ok := item.(*ast.TestBlock); ok {
                        c.symTable.PushScope(symbols.ScopeTest, tb.Name)
                        for _, stmt := range tb.Body {
                                c.checkStatement(stmt)
                        }
                        c.symTable.PopScope()
                }
        }
}

// --- Statement Checking ---

func (c *Checker) checkStatement(stmt ast.Statement) {
        switch s := stmt.(type) {
        case *ast.LetStmt:
                c.checkLetStmt(s)
        case *ast.AssignStmt:
                c.checkAssignStmt(s)
        case *ast.ReturnStmt:
                c.checkReturnStmt(s)
        case *ast.IfStmt:
                c.checkIfStmt(s)
        case *ast.MatchStmt:
                c.checkMatchStmt(s)
        case *ast.ForStmt:
                c.checkForStmt(s)
        case *ast.WhileStmt:
                c.checkWhileStmt(s)
        case *ast.ExprStmt:
                c.inferExpr(s.Expr)
        case *ast.AssertStmt:
                c.checkAssertStmt(s)
        case *ast.BreakStmt:
                if !c.symTable.Current.IsInsideLoop() {
                        c.addError(newError(ErrBreakOutside, s.Span,
                                "'break' can only be used inside a loop"))
                }
        case *ast.ContinueStmt:
                if !c.symTable.Current.IsInsideLoop() {
                        c.addError(newError(ErrContinueOutside, s.Span,
                                "'continue' can only be used inside a loop"))
                }
        case *ast.WithStmt:
                c.checkWithStmt(s)
        }
}

func (c *Checker) checkLetStmt(s *ast.LetStmt) {
        var valType *types.Type
        if s.Value != nil {
                valType = c.inferExpr(s.Value)
        }

        var resolvedType *types.Type
        if s.TypeHint != nil {
                resolvedType = c.resolveTypeExpr(s.TypeHint)
                if valType != nil && !types.IsAssignableTo(valType, resolvedType) {
                        c.addError(newError(ErrTypeMismatch, s.Span,
                                fmt.Sprintf("cannot assign %s to variable %q of type %s",
                                        valType, s.Name, resolvedType)).
                                withExpectedGot(resolvedType.String(), valType.String()))
                }
        } else if valType != nil {
                resolvedType = valType
        } else {
                resolvedType = types.BuiltinAny
        }

        err := c.symTable.Define(&symbols.Symbol{
                Name:    s.Name,
                Kind:    symbols.SymVariable,
                Span:    s.Span,
                Mutable: s.Mutable,
        })
        if err != nil {
                c.addError(newError(ErrRedefinedName, s.Span,
                        fmt.Sprintf("variable %q is already defined in this scope", s.Name)))
        }

        // Track the variable's type
        c.varTypes[s.Name] = resolvedType
}

func (c *Checker) checkAssignStmt(s *ast.AssignStmt) {
        // Check target is assignable (must be mutable variable or field)
        if ident, ok := s.Target.(*ast.Identifier); ok {
                sym, found := c.symTable.Lookup(ident.Name)
                if !found {
                        c.addError(newError(ErrUndefinedName, s.Span,
                                fmt.Sprintf("undefined variable %q", ident.Name)))
                        return
                }
                if sym.Kind == symbols.SymVariable && !sym.Mutable {
                        c.addError(newError(ErrImmutableAssign, s.Span,
                                fmt.Sprintf("cannot assign to immutable variable %q", ident.Name)).
                                withFix(fmt.Sprintf("Change declaration to: let mut %s", ident.Name)))
                }
        }

        c.inferExpr(s.Target)
        c.inferExpr(s.Value)
}

func (c *Checker) checkReturnStmt(s *ast.ReturnStmt) {
        if c.symTable.Current.EnclosingFunction() == nil {
                c.addError(newError(ErrReturnOutside, s.Span,
                        "'return' can only be used inside a function"))
                return
        }

        if s.Value != nil && c.currentFnReturn != nil {
                valType := c.inferExpr(s.Value)
                // Skip check if inferred type is Any (unknown) - we can't validate
                if valType != types.BuiltinAny && !types.IsAssignableTo(valType, c.currentFnReturn) {
                        c.addError(newError(ErrTypeMismatch, s.Span,
                                fmt.Sprintf("return type mismatch: expected %s, got %s",
                                        c.currentFnReturn, valType)).
                                withExpectedGot(c.currentFnReturn.String(), valType.String()))
                }
        }
}

func (c *Checker) checkIfStmt(s *ast.IfStmt) {
        condType := c.inferExpr(s.Condition)
        if condType != nil && !types.IsAssignableTo(condType, types.BuiltinBool) {
                c.addError(newError(ErrTypeMismatch, s.Condition.GetSpan(),
                        fmt.Sprintf("if condition must be Bool, got %s", condType)).
                        withExpectedGot("Bool", condType.String()))
        }

        c.symTable.PushScope(symbols.ScopeBlock, "if")
        for _, stmt := range s.ThenBody {
                c.checkStatement(stmt)
        }
        c.symTable.PopScope()

        for _, elif := range s.ElifClauses {
                elifCondType := c.inferExpr(elif.Condition)
                if elifCondType != nil && !types.IsAssignableTo(elifCondType, types.BuiltinBool) {
                        c.addError(newError(ErrTypeMismatch, elif.Condition.GetSpan(),
                                fmt.Sprintf("elif condition must be Bool, got %s", elifCondType)).
                                withExpectedGot("Bool", elifCondType.String()))
                }
                c.symTable.PushScope(symbols.ScopeBlock, "elif")
                for _, stmt := range elif.Body {
                        c.checkStatement(stmt)
                }
                c.symTable.PopScope()
        }

        if len(s.ElseBody) > 0 {
                c.symTable.PushScope(symbols.ScopeBlock, "else")
                for _, stmt := range s.ElseBody {
                        c.checkStatement(stmt)
                }
                c.symTable.PopScope()
        }
}

func (c *Checker) checkMatchStmt(s *ast.MatchStmt) {
        subjectType := c.inferExpr(s.Subject)

        for _, cs := range s.Cases {
                c.symTable.PushScope(symbols.ScopeBlock, "case")
                c.checkPattern(cs.Pattern, subjectType)
                if cs.Guard != nil {
                        guardType := c.inferExpr(cs.Guard)
                        if guardType != nil && !types.IsAssignableTo(guardType, types.BuiltinBool) {
                                c.addError(newError(ErrTypeMismatch, cs.Guard.GetSpan(),
                                        "match guard must be Bool"))
                        }
                }
                for _, stmt := range cs.Body {
                        c.checkStatement(stmt)
                }
                c.symTable.PopScope()
        }

        // Basic exhaustiveness check for enum types
        if subjectType != nil {
                underlying := types.Underlying(subjectType)
                if underlying != nil && underlying.Kind == types.KindEnum {
                        c.checkEnumExhaustiveness(s, underlying)
                }
                // Also check if it's a named type pointing to an enum
                if underlying != nil && underlying.Kind == types.KindAlias && underlying.BaseT != nil {
                        base := types.Underlying(underlying.BaseT)
                        if base != nil && base.Kind == types.KindEnum {
                                c.checkEnumExhaustiveness(s, base)
                        }
                }
        }
}

func (c *Checker) checkEnumExhaustiveness(s *ast.MatchStmt, enumType *types.Type) {
        covered := make(map[string]bool)
        hasWildcard := false

        for _, cs := range s.Cases {
                switch p := cs.Pattern.(type) {
                case *ast.WildcardPattern:
                        hasWildcard = true
                case *ast.ConstructorPattern:
                        // Handle dotted names like "TaskError.NotFound"
                        name := p.TypeName
                        parts := strings.Split(name, ".")
                        if len(parts) == 2 {
                                name = parts[1]
                        }
                        covered[name] = true
                case *ast.BindingPattern:
                        // A binding pattern acts as wildcard
                        hasWildcard = true
                }
        }

        if hasWildcard {
                return
        }

        missing := []string{}
        for _, v := range enumType.Variants {
                if !covered[v.Name] {
                        missing = append(missing, v.Name)
                }
        }

        if len(missing) > 0 {
                c.addError(newError(ErrNonExhaustive, s.Span,
                        fmt.Sprintf("non-exhaustive match: missing variants: %s",
                                strings.Join(missing, ", "))).
                        withFix(fmt.Sprintf("Add case clauses for: %s, or add a wildcard '_' case",
                                strings.Join(missing, ", "))))
        }
}

func (c *Checker) checkForStmt(s *ast.ForStmt) {
        iterType := c.inferExpr(s.Iterable)
        if iterType != nil && iterType != types.BuiltinAny {
                underlying := types.Underlying(iterType)
                if underlying != nil && underlying.Kind != types.KindList &&
                        underlying.Kind != types.KindSet &&
                        underlying.Kind != types.KindMap &&
                        underlying.Kind != types.KindAny {
                        // Allow String iteration and collection types
                        if !(underlying.Kind == types.KindPrimitive && underlying.Name == "String") {
                                c.addError(newError(ErrNotIterable, s.Iterable.GetSpan(),
                                        fmt.Sprintf("cannot iterate over %s", iterType)))
                        }
                }
        }

        c.symTable.PushScope(symbols.ScopeLoop, "for")
        c.symTable.Define(&symbols.Symbol{
                Name: s.Variable,
                Kind: symbols.SymVariable,
                Span: s.Span,
        })
        // Infer loop variable type from iterable
        if iterType != nil {
                underlying := types.Underlying(iterType)
                if underlying != nil && underlying.Kind == types.KindList && underlying.ElementT != nil {
                        c.varTypes[s.Variable] = underlying.ElementT
                } else {
                        c.varTypes[s.Variable] = types.BuiltinAny
                }
        }
        for _, stmt := range s.Body {
                c.checkStatement(stmt)
        }
        c.symTable.PopScope()
}

func (c *Checker) checkWhileStmt(s *ast.WhileStmt) {
        condType := c.inferExpr(s.Condition)
        if condType != nil && !types.IsAssignableTo(condType, types.BuiltinBool) {
                c.addError(newError(ErrTypeMismatch, s.Condition.GetSpan(),
                        fmt.Sprintf("while condition must be Bool, got %s", condType)).
                        withExpectedGot("Bool", condType.String()))
        }

        c.symTable.PushScope(symbols.ScopeLoop, "while")
        for _, stmt := range s.Body {
                c.checkStatement(stmt)
        }
        c.symTable.PopScope()
}

func (c *Checker) checkAssertStmt(s *ast.AssertStmt) {
        condType := c.inferExpr(s.Condition)
        if condType != nil && !types.IsAssignableTo(condType, types.BuiltinBool) {
                c.addError(newError(ErrTypeMismatch, s.Condition.GetSpan(),
                        fmt.Sprintf("assert condition must be Bool, got %s", condType)).
                        withExpectedGot("Bool", condType.String()))
        }
}

func (c *Checker) checkWithStmt(s *ast.WithStmt) {
        for _, b := range s.Bindings {
                c.inferExpr(b.Expr)
        }
        c.symTable.PushScope(symbols.ScopeBlock, "with")
        for _, b := range s.Bindings {
                if b.Alias != "" {
                        c.symTable.Define(&symbols.Symbol{
                                Name: b.Alias,
                                Kind: symbols.SymVariable,
                                Span: s.Span,
                        })
                }
        }
        for _, stmt := range s.Body {
                c.checkStatement(stmt)
        }
        c.symTable.PopScope()
}

// --- Pattern Checking ---

func (c *Checker) checkPattern(pat ast.Pattern, expectedType *types.Type) {
        if pat == nil {
                return
        }
        switch p := pat.(type) {
        case *ast.WildcardPattern:
                // always ok
        case *ast.BindingPattern:
                c.symTable.Define(&symbols.Symbol{
                        Name: p.Name,
                        Kind: symbols.SymVariable,
                        Span: p.Span,
                })
        case *ast.LiteralPattern:
                // Check that literal is compatible with expected type
                litType := c.literalPatternType(p)
                if expectedType != nil && litType != nil && !types.IsAssignableTo(litType, expectedType) {
                        c.addError(newError(ErrTypeMismatch, p.Span,
                                fmt.Sprintf("pattern type %s is not compatible with match subject type %s",
                                        litType, expectedType)))
                }
        case *ast.ConstructorPattern:
                for _, sub := range p.Fields {
                        c.checkPattern(sub, nil) // sub-pattern types need more context
                }
        case *ast.ListPattern:
                var elemType *types.Type
                if expectedType != nil {
                        underlying := types.Underlying(expectedType)
                        if underlying != nil && underlying.Kind == types.KindList {
                                elemType = underlying.ElementT
                        }
                }
                for _, sub := range p.Elements {
                        c.checkPattern(sub, elemType)
                }
        case *ast.TuplePattern:
                for i, sub := range p.Elements {
                        var elemType *types.Type
                        if expectedType != nil {
                                underlying := types.Underlying(expectedType)
                                if underlying != nil && underlying.Kind == types.KindTuple && i < len(underlying.Members) {
                                        elemType = underlying.Members[i]
                                }
                        }
                        c.checkPattern(sub, elemType)
                }
        }
}

func (c *Checker) literalPatternType(p *ast.LiteralPattern) *types.Type {
        switch p.Kind {
        case token.INT_LIT:
                return types.BuiltinInt
        case token.FLOAT_LIT:
                return types.BuiltinFloat
        case token.STRING_LIT:
                return types.BuiltinString
        case token.BOOL_LIT, token.TRUE, token.FALSE:
                return types.BuiltinBool
        case token.NONE_LIT, token.NONE_VAL:
                return types.BuiltinNone
        }
        return nil
}

// --- Expression Type Inference ---

func (c *Checker) inferExpr(expr ast.Expr) *types.Type {
        if expr == nil {
                return types.BuiltinNone
        }
        switch e := expr.(type) {
        case *ast.Identifier:
                return c.inferIdentifier(e)
        case *ast.IntLiteral:
                return types.BuiltinInt
        case *ast.FloatLiteral:
                return types.BuiltinFloat
        case *ast.StringLiteral:
                return types.BuiltinString
        case *ast.BoolLiteral:
                return types.BuiltinBool
        case *ast.NoneLiteral:
                return types.BuiltinNone
        case *ast.BinaryOp:
                return c.inferBinaryOp(e)
        case *ast.UnaryOp:
                return c.inferUnaryOp(e)
        case *ast.CallExpr:
                return c.inferCallExpr(e)
        case *ast.FieldAccess:
                return c.inferFieldAccess(e)
        case *ast.IndexExpr:
                return c.inferIndexExpr(e)
        case *ast.ListExpr:
                return c.inferListExpr(e)
        case *ast.ListComp:
                return c.inferListComp(e)
        case *ast.MapExpr:
                return c.inferMapExpr(e)
        case *ast.StructExpr:
                return c.inferStructExpr(e)
        case *ast.IfExpr:
                return c.inferIfExpr(e)
        case *ast.Lambda:
                return c.inferLambda(e)
        case *ast.OptionPropagate:
                return c.inferOptionPropagate(e)
        case *ast.OptionalFieldAccess:
                // Option chaining: result is always optional (could be None)
                c.inferExpr(e.Object)
                return types.BuiltinAny
        case *ast.PipelineExpr:
                // Pipeline: infer the return type of the right-hand function
                c.inferExpr(e.Left)
                return c.inferExpr(e.Right)
        default:
                return types.BuiltinAny
        }
}

func (c *Checker) inferIdentifier(e *ast.Identifier) *types.Type {
        // Check if it's an effect capability (db, time, net, fs, random, auth, log)
        knownEffects := map[string]bool{
                "db": true, "net": true, "fs": true, "time": true,
                "random": true, "auth": true, "log": true,
        }
        if knownEffects[e.Name] {
                // Effect capabilities are valid identifiers when declared
                if c.currentEffects != nil && c.currentEffects[e.Name] {
                        return types.BuiltinAny // effect capability object
                }
                // Still return Any but will get an effect error elsewhere
                return types.BuiltinAny
        }

        // Check built-in constructors
        switch e.Name {
        case "Ok", "Err", "Some":
                return types.BuiltinAny // constructor, type depends on context
        }

        sym, found := c.symTable.Lookup(e.Name)
        if !found {
                // Check if it's a type name used as constructor (enum variant, etc.)
                if t, ok := c.typeReg.Lookup(e.Name); ok {
                        return t
                }
                c.addError(newError(ErrUndefinedName, e.Span,
                        fmt.Sprintf("undefined name %q", e.Name)))
                return types.BuiltinAny
        }

        switch sym.Kind {
        case symbols.SymFunction:
                if fnType, ok := c.fnTypes[e.Name]; ok {
                        return fnType
                }
                return types.BuiltinAny
        case symbols.SymVariable, symbols.SymParam, symbols.SymConst:
                if t, ok := c.varTypes[e.Name]; ok {
                        return t
                }
                return types.BuiltinAny
        default:
                return types.BuiltinAny
        }
}

func (c *Checker) inferBinaryOp(e *ast.BinaryOp) *types.Type {
        left := c.inferExpr(e.Left)
        right := c.inferExpr(e.Right)

        switch e.Op {
        case "+", "-", "*", "/", "%", "**":
                // Arithmetic operations
                if left == types.BuiltinInt && right == types.BuiltinInt {
                        if e.Op == "/" {
                                return types.BuiltinFloat
                        }
                        return types.BuiltinInt
                }
                if (left == types.BuiltinFloat || left == types.BuiltinInt) &&
                        (right == types.BuiltinFloat || right == types.BuiltinInt) {
                        return types.BuiltinFloat
                }
                if e.Op == "+" && left == types.BuiltinString && right == types.BuiltinString {
                        return types.BuiltinString
                }
                return types.BuiltinAny

        case "==", "!=", "<", ">", "<=", ">=":
                return types.BuiltinBool

        case "and", "or":
                return types.BuiltinBool

        case "is":
                return types.BuiltinBool

        case "in":
                return types.BuiltinBool
        }

        return types.BuiltinAny
}

func (c *Checker) inferUnaryOp(e *ast.UnaryOp) *types.Type {
        operand := c.inferExpr(e.Operand)
        switch e.Op {
        case "-":
                if operand == types.BuiltinInt {
                        return types.BuiltinInt
                }
                if operand == types.BuiltinFloat {
                        return types.BuiltinFloat
                }
                return types.BuiltinAny
        case "not":
                return types.BuiltinBool
        }
        return types.BuiltinAny
}

func (c *Checker) inferCallExpr(e *ast.CallExpr) *types.Type {
        calleeType := c.inferExpr(e.Callee)

        // Check if callee is a function name for effect tracking
        if ident, ok := e.Callee.(*ast.Identifier); ok {
                c.checkCallEffects(ident.Name, e.GetSpan())
        }
        // Also check method calls like db.insert
        if fa, ok := e.Callee.(*ast.FieldAccess); ok {
                if ident, ok := fa.Object.(*ast.Identifier); ok {
                        c.checkCallEffects(ident.Name+"."+fa.Field, e.GetSpan())
                        // Check if the object name is an effect capability
                        c.checkEffectCapabilityUsage(ident.Name, e.GetSpan())
                }
        }

        if calleeType != nil && calleeType.Kind == types.KindFunction {
                // Check argument count
                expectedArgs := len(calleeType.ParamTypes)
                gotArgs := len(e.Args)
                if gotArgs > expectedArgs {
                        c.addError(newError(ErrArgCount, e.Span,
                                fmt.Sprintf("too many arguments: expected %d, got %d",
                                        expectedArgs, gotArgs)).
                                withExpectedGot(fmt.Sprintf("%d", expectedArgs), fmt.Sprintf("%d", gotArgs)))
                }

                return calleeType.ReturnT
        }

        // Check built-in constructors and struct/enum constructors
        if ident, ok := e.Callee.(*ast.Identifier); ok {
                // Built-in constructors: Ok, Err, Some
                switch ident.Name {
                case "Ok", "Err", "Some":
                        for _, arg := range e.Args {
                                c.inferExpr(arg.Value)
                        }
                        return types.BuiltinAny // contextual type
                }

                if t, ok := c.typeReg.Lookup(ident.Name); ok {
                        if t.Kind == types.KindStruct {
                                c.checkStructConstruction(e, t)
                                return t
                        }
                        if t.Kind == types.KindEnum {
                                for _, arg := range e.Args {
                                        c.inferExpr(arg.Value)
                                }
                                return t
                        }
                }
        }

        // Check for inferred types
        for _, arg := range e.Args {
                c.inferExpr(arg.Value)
        }

        return types.BuiltinAny
}

func (c *Checker) checkStructConstruction(e *ast.CallExpr, structType *types.Type) {
        provided := make(map[string]bool)
        for _, arg := range e.Args {
                if arg.Name != "" {
                        provided[arg.Name] = true
                        // Check field exists
                        found := false
                        for _, f := range structType.Fields {
                                if f.Name == arg.Name {
                                        found = true
                                        break
                                }
                        }
                        if !found {
                                c.addError(newError(ErrFieldNotFound, arg.Span,
                                        fmt.Sprintf("struct %q has no field %q", structType.Name, arg.Name)))
                        }
                }
                c.inferExpr(arg.Value)
        }

        // Check required fields are provided
        for _, f := range structType.Fields {
                if !f.Optional && !provided[f.Name] {
                        // Only report if named args are being used (positional is ok)
                        if len(e.Args) > 0 && e.Args[0].Name != "" {
                                c.addError(newError(ErrFieldNotFound, e.Span,
                                        fmt.Sprintf("missing required field %q in struct %q construction",
                                                f.Name, structType.Name)).
                                        withFix(fmt.Sprintf("Add field: %s: <value>", f.Name)))
                        }
                }
        }
}

func (c *Checker) checkCallEffects(calleeName string, span token.Span) {
        if c.currentEffects == nil {
                return // not inside a function
        }

        if effects, ok := c.fnEffects[calleeName]; ok {
                for _, e := range effects {
                        if !c.currentEffects[e] {
                                fnName := ""
                                if c.currentFn != nil {
                                        fnName = c.currentFn.Name
                                }
                                c.addError(newError(ErrMissingEffect, span,
                                        fmt.Sprintf("calling %q requires effect %q, but function %q does not declare it",
                                                calleeName, e, fnName)).
                                        withFix(fmt.Sprintf("Add 'with %s' to the function signature", e)))
                        }
                }
        }
}

func (c *Checker) checkEffectCapabilityUsage(name string, span token.Span) {
        if c.currentEffects == nil {
                return
        }
        // Known effect capabilities
        knownEffects := map[string]bool{
                "db": true, "net": true, "fs": true, "time": true,
                "random": true, "auth": true, "log": true,
        }
        if knownEffects[name] && !c.currentEffects[name] {
                fnName := ""
                if c.currentFn != nil {
                        fnName = c.currentFn.Name
                }
                c.addError(newError(ErrMissingEffect, span,
                        fmt.Sprintf("using capability %q requires declaring 'with %s' on function %q",
                                name, name, fnName)).
                        withFix(fmt.Sprintf("Add 'with %s' to the function signature", name)))
        }
}

func (c *Checker) inferFieldAccess(e *ast.FieldAccess) *types.Type {
        objType := c.inferExpr(e.Object)
        if objType == nil {
                return types.BuiltinAny
        }

        underlying := types.Underlying(objType)
        if underlying == nil {
                return types.BuiltinAny
        }

        // Struct field access
        if underlying.Kind == types.KindStruct {
                for _, f := range underlying.Fields {
                        if f.Name == e.Field {
                                return f.Type
                        }
                }
                c.addError(newError(ErrFieldNotFound, e.Span,
                        fmt.Sprintf("type %q has no field %q", objType, e.Field)))
                return types.BuiltinAny
        }

        // Built-in properties
        if underlying.Kind == types.KindPrimitive && underlying.Name == "String" {
                if e.Field == "len" {
                        return types.BuiltinInt
                }
        }
        if underlying.Kind == types.KindList {
                if e.Field == "len" {
                        return types.BuiltinInt
                }
        }

        // Enum variant access (e.g., TaskError.NotFound)
        if ident, ok := e.Object.(*ast.Identifier); ok {
                qualName := ident.Name + "." + e.Field
                if _, found := c.symTable.Lookup(qualName); found {
                        if t, ok := c.typeReg.Lookup(ident.Name); ok {
                                return t
                        }
                }
        }

        return types.BuiltinAny
}

func (c *Checker) inferIndexExpr(e *ast.IndexExpr) *types.Type {
        objType := c.inferExpr(e.Object)
        c.inferExpr(e.Index)

        if objType == nil {
                return types.BuiltinAny
        }

        underlying := types.Underlying(objType)
        if underlying == nil {
                return types.BuiltinAny
        }

        switch underlying.Kind {
        case types.KindList:
                return underlying.ElementT
        case types.KindMap:
                return underlying.ValueT
        case types.KindPrimitive:
                if underlying.Name == "String" {
                        return types.BuiltinString
                }
        }

        return types.BuiltinAny
}

func (c *Checker) inferListExpr(e *ast.ListExpr) *types.Type {
        if len(e.Elements) == 0 {
                return types.NewListType(types.BuiltinAny)
        }

        elemType := c.inferExpr(e.Elements[0])
        for i := 1; i < len(e.Elements); i++ {
                c.inferExpr(e.Elements[i])
        }
        return types.NewListType(elemType)
}

func (c *Checker) inferListComp(e *ast.ListComp) *types.Type {
        iterType := c.inferExpr(e.Iterable)

        c.symTable.PushScope(symbols.ScopeBlock, "listcomp")
        c.symTable.Define(&symbols.Symbol{
                Name: e.Variable,
                Kind: symbols.SymVariable,
                Span: e.GetSpan(),
        })
        // Infer loop variable type from iterable
        if iterType != nil {
                underlying := types.Underlying(iterType)
                if underlying != nil && underlying.Kind == types.KindList && underlying.ElementT != nil {
                        c.varTypes[e.Variable] = underlying.ElementT
                } else {
                        c.varTypes[e.Variable] = types.BuiltinAny
                }
        }

        elemType := c.inferExpr(e.Element)
        if e.Filter != nil {
                c.inferExpr(e.Filter)
        }

        c.symTable.PopScope()
        return types.NewListType(elemType)
}

func (c *Checker) inferMapExpr(e *ast.MapExpr) *types.Type {
        if len(e.Entries) == 0 {
                return types.NewMapType(types.BuiltinAny, types.BuiltinAny)
        }

        keyType := c.inferExpr(e.Entries[0].Key)
        valType := c.inferExpr(e.Entries[0].Value)
        for i := 1; i < len(e.Entries); i++ {
                c.inferExpr(e.Entries[i].Key)
                c.inferExpr(e.Entries[i].Value)
        }
        return types.NewMapType(keyType, valType)
}

func (c *Checker) inferStructExpr(e *ast.StructExpr) *types.Type {
        t, ok := c.typeReg.Lookup(e.TypeName)
        if !ok {
                c.addError(newError(ErrUndefinedType, e.Span,
                        fmt.Sprintf("undefined type %q", e.TypeName)))
                return types.BuiltinAny
        }

        if t.Kind == types.KindStruct || (t.Kind == types.KindAlias && t.BaseT != nil) {
                structType := types.Underlying(t)
                if structType != nil && structType.Kind == types.KindStruct {
                        provided := make(map[string]bool)
                        for _, fi := range e.Fields {
                                provided[fi.Name] = true
                                found := false
                                for _, sf := range structType.Fields {
                                        if sf.Name == fi.Name {
                                                found = true
                                                break
                                        }
                                }
                                if !found {
                                        c.addError(newError(ErrFieldNotFound, e.Span,
                                                fmt.Sprintf("struct %q has no field %q", e.TypeName, fi.Name)))
                                }
                                c.inferExpr(fi.Value)
                        }

                        for _, sf := range structType.Fields {
                                if !sf.Optional && !provided[sf.Name] {
                                        c.addError(newError(ErrFieldNotFound, e.Span,
                                                fmt.Sprintf("missing required field %q in struct %q",
                                                        sf.Name, e.TypeName)).
                                                withFix(fmt.Sprintf("Add: %s: <value>", sf.Name)))
                                }
                        }
                }
        }

        return t
}

func (c *Checker) inferIfExpr(e *ast.IfExpr) *types.Type {
        condType := c.inferExpr(e.Condition)
        if condType != nil && !types.IsAssignableTo(condType, types.BuiltinBool) {
                c.addError(newError(ErrTypeMismatch, e.Condition.GetSpan(),
                        fmt.Sprintf("if expression condition must be Bool, got %s", condType)))
        }

        thenType := c.inferExpr(e.ThenExpr)
        c.inferExpr(e.ElseExpr)
        return thenType // simplified: should unify then/else types
}

func (c *Checker) inferLambda(e *ast.Lambda) *types.Type {
        c.symTable.PushScope(symbols.ScopeFunction, "lambda")

        paramTypes := make([]*types.Type, len(e.Params))
        for i, p := range e.Params {
                pt := types.BuiltinAny
                if p.TypeExpr != nil {
                        pt = c.resolveTypeExpr(p.TypeExpr)
                }
                paramTypes[i] = pt
                c.symTable.Define(&symbols.Symbol{
                        Name: p.Name,
                        Kind: symbols.SymParam,
                        Span: p.Span,
                })
        }

        var retType *types.Type
        if e.Body != nil {
                retType = c.inferExpr(e.Body)
        } else if len(e.Block) > 0 {
                for _, stmt := range e.Block {
                        c.checkStatement(stmt)
                }
                retType = types.BuiltinNone
        }

        c.symTable.PopScope()
        return types.NewFunctionType(paramTypes, retType, nil)
}

func (c *Checker) inferOptionPropagate(e *ast.OptionPropagate) *types.Type {
        innerType := c.inferExpr(e.Expr)
        if innerType == nil {
                return types.BuiltinAny
        }

        underlying := types.Underlying(innerType)
        if underlying != nil && underlying.Kind == types.KindOption {
                return underlying.ElementT
        }
        if underlying != nil && underlying.Kind == types.KindResult {
                return underlying.ElementT // Ok type
        }

        return types.BuiltinAny
}

// --- Type Resolution ---

func (c *Checker) resolveTypeExpr(te ast.TypeExpr) *types.Type {
        if te == nil {
                return types.BuiltinNone
        }

        switch t := te.(type) {
        case *ast.NamedType:
                return c.resolveNamedType(t)
        case *ast.QualifiedType:
                // For qualified types like time.Instant, return Any for now
                return types.BuiltinAny
        case *ast.UnionType:
                left := c.resolveTypeExpr(t.Left)
                right := c.resolveTypeExpr(t.Right)
                return types.NewUnionType([]*types.Type{left, right})
        case *ast.IntersectionType:
                left := c.resolveTypeExpr(t.Left)
                right := c.resolveTypeExpr(t.Right)
                return &types.Type{Kind: types.KindIntersection, Members: []*types.Type{left, right}}
        case *ast.FunctionType:
                params := make([]*types.Type, len(t.Params))
                for i, p := range t.Params {
                        params[i] = c.resolveTypeExpr(p)
                }
                ret := c.resolveTypeExpr(t.ReturnType)
                return types.NewFunctionType(params, ret, nil)
        case *ast.TupleType:
                elems := make([]*types.Type, len(t.Elements))
                for i, e := range t.Elements {
                        elems[i] = c.resolveTypeExpr(e)
                }
                return types.NewTupleType(elems)
        case *ast.ListType:
                return types.NewListType(c.resolveTypeExpr(t.Element))
        case *ast.MapType:
                return types.NewMapType(c.resolveTypeExpr(t.Key), c.resolveTypeExpr(t.Value))
        case *ast.SetType:
                return types.NewSetType(c.resolveTypeExpr(t.Element))
        case *ast.OptionType:
                return types.NewOptionType(c.resolveTypeExpr(t.Inner))
        case *ast.RefinementType:
                base := c.resolveTypeExpr(t.Base)
                predStr := c.exprToString(t.Predicate)
                return types.NewRefinementType(base, predStr)
        case *ast.StringLitType:
                return types.NewStringLitType(t.Value)
        default:
                return types.BuiltinAny
        }
}

func (c *Checker) resolveNamedType(t *ast.NamedType) *types.Type {
        // Check builtins and registered types
        if resolved, ok := c.typeReg.Lookup(t.Name); ok {
                if len(t.Args) > 0 && len(resolved.TypeParams) > 0 {
                        // Instantiate generic type
                        args := make([]*types.Type, len(t.Args))
                        for i, a := range t.Args {
                                args[i] = c.resolveTypeExpr(a)
                        }
                        // Create instantiated copy
                        inst := *resolved
                        inst.TypeArgs = args
                        return &inst
                }
                return resolved
        }

        // Special handling for Option and Result
        switch t.Name {
        case "Option":
                if len(t.Args) == 1 {
                        return types.NewOptionType(c.resolveTypeExpr(t.Args[0]))
                }
        case "Result":
                if len(t.Args) == 2 {
                        return types.NewResultType(
                                c.resolveTypeExpr(t.Args[0]),
                                c.resolveTypeExpr(t.Args[1]))
                }
        }

        c.addError(newError(ErrUndefinedType, t.Span,
                fmt.Sprintf("undefined type %q", t.Name)))
        return types.BuiltinAny
}

// --- Helpers ---

func (c *Checker) exprToString(expr ast.Expr) string {
        if expr == nil {
                return ""
        }
        switch e := expr.(type) {
        case *ast.Identifier:
                return e.Name
        case *ast.IntLiteral:
                return e.Value
        case *ast.StringLiteral:
                return e.Value
        case *ast.BoolLiteral:
                if e.Value {
                        return "true"
                }
                return "false"
        case *ast.BinaryOp:
                return c.exprToString(e.Left) + " " + e.Op + " " + c.exprToString(e.Right)
        case *ast.UnaryOp:
                return e.Op + " " + c.exprToString(e.Operand)
        case *ast.FieldAccess:
                return c.exprToString(e.Object) + "." + e.Field
        case *ast.OptionalFieldAccess:
                return c.exprToString(e.Object) + "?." + e.Field
        case *ast.PipelineExpr:
                return c.exprToString(e.Left) + " |> " + c.exprToString(e.Right)
        default:
                return "<expr>"
        }
}

func (c *Checker) typeExprString(te ast.TypeExpr) string {
        if te == nil {
                return "None"
        }
        t := c.resolveTypeExpr(te)
        return t.String()
}

// FormatErrors returns all errors as a formatted string suitable for display.
func FormatErrors(errs []*CheckError) string {
        if len(errs) == 0 {
                return ""
        }
        var sb strings.Builder
        for _, e := range errs {
                sb.WriteString(e.Error())
                if e.Expected != "" || e.Got != "" {
                        sb.WriteString(fmt.Sprintf("\n  expected: %s\n  got:      %s", e.Expected, e.Got))
                }
                if e.Fix != "" {
                        sb.WriteString(fmt.Sprintf("\n  fix: %s", e.Fix))
                }
                sb.WriteString("\n")
        }
        return sb.String()
}

// FormatErrorsJSON returns all errors as a JSON array string for AI consumption.
func FormatErrorsJSON(errs []*CheckError) string {
        if len(errs) == 0 {
                return "[]"
        }
        var parts []string
        for _, e := range errs {
                parts = append(parts, e.JSON())
        }
        return "[" + strings.Join(parts, ",") + "]"
}
