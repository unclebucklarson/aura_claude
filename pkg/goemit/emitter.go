// Package goemit compiles Aura ASTs to Go source code.
//
// The generated Go file is self-contained: it includes a small runtime
// preamble defining Option[T] and Result[T,E] and maps Aura built-ins
// (print, len, append, etc.) to Go equivalents.
//
// Type mapping:
//
//	Int             → int64
//	Float           → float64
//	String          → string
//	Bool            → bool
//	[T]             → []T
//	{K: V}          → map[K]V
//	Option[T] / T?  → auraOption[T]
//	Result[T, E]    → auraResult[T, E]
//	void / ()       → (no return)
//
// Not yet supported (emit a comment): effects, refinement types,
// trait/impl dispatch, pipeline operator, list comprehensions.
package goemit

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/unclebucklarson/aura/pkg/ast"
)

// runtimePreamble is prepended to every emitted file.
const runtimePreamble = `
// --- Aura runtime types ---

type auraOption[T any] struct {
	v  T
	ok bool
}

func auraSome[T any](v T) auraOption[T] { return auraOption[T]{v: v, ok: true} }
func auraNone[T any]() auraOption[T]    { return auraOption[T]{} }

type auraResult[T, E any] struct {
	v  T
	e  E
	ok bool
}

func auraOk[T, E any](v T) auraResult[T, E]  { return auraResult[T, E]{v: v, ok: true} }
func auraErr[T, E any](e E) auraResult[T, E] { return auraResult[T, E]{e: e} }

`

// Emitter walks an Aura AST and produces Go source.
type Emitter struct {
	sb     strings.Builder
	indent int
	errors []string

	// enumVariants tracks enum name → variant names, so constructor
	// calls like Color.Red can be resolved.
	enumVariants map[string][]string
	// variantEnum maps variant name → enum name for reverse lookup.
	variantEnum map[string]string
}

// New returns a fresh Emitter.
func New() *Emitter {
	return &Emitter{
		enumVariants: make(map[string][]string),
		variantEnum:  make(map[string]string),
	}
}

// Emit converts mod to a complete Go source file.
// Returns the source string and any non-fatal warnings.
func (e *Emitter) Emit(mod *ast.Module) (string, []string) {
	e.sb.Reset()
	e.errors = nil
	e.indent = 0

	// Pre-scan enums so constructor calls can be resolved.
	e.scanEnums(mod)

	// Package declaration.
	pkgName := "main"
	if mod.Name != nil {
		pkgName = goIdent(mod.Name.Parts[len(mod.Name.Parts)-1])
		// Programs with a main() always emit package main.
		for _, item := range mod.Items {
			if fn, ok := item.(*ast.FnDef); ok && fn.Name == "main" {
				pkgName = "main"
				break
			}
		}
	}
	e.line("package " + pkgName)
	e.line("")

	// Collect required imports.
	imports := e.collectImports(mod)
	if len(imports) > 0 {
		e.line("import (")
		e.indent++
		for _, imp := range imports {
			e.line(`"` + imp + `"`)
		}
		e.indent--
		e.line(")")
		e.line("")
	}

	// Runtime preamble.
	e.sb.WriteString(runtimePreamble)

	// Top-level items.
	for _, item := range mod.Items {
		e.emitTopLevel(item)
	}

	return e.sb.String(), e.errors
}

// --- pre-scan ---

func (e *Emitter) scanEnums(mod *ast.Module) {
	for _, item := range mod.Items {
		if en, ok := item.(*ast.EnumDef); ok {
			var variants []string
			for _, v := range en.Variants {
				variants = append(variants, v.Name)
				e.variantEnum[v.Name] = en.Name
			}
			e.enumVariants[en.Name] = variants
		}
	}
}

// collectImports determines which standard library packages are needed.
func (e *Emitter) collectImports(mod *ast.Module) []string {
	src := e.sb.String() // not yet written, so walk AST heuristically
	_ = src
	var imps []string
	// Always include fmt (for print/println).
	imps = append(imps, "fmt")
	// Scan for string formatting (interpolation) or os usage.
	for _, item := range mod.Items {
		if fn, ok := item.(*ast.FnDef); ok {
			if usesOS(fn.Body) {
				imps = append(imps, "os")
				break
			}
		}
	}
	return imps
}

func usesOS(stmts []ast.Statement) bool {
	for _, s := range stmts {
		switch st := s.(type) {
		case *ast.ExprStmt:
			if usesOSExpr(st.Expr) {
				return true
			}
		case *ast.LetStmt:
			if st.Value != nil && usesOSExpr(st.Value) {
				return true
			}
		}
	}
	return false
}

func usesOSExpr(e ast.Expr) bool {
	if call, ok := e.(*ast.CallExpr); ok {
		if fa, ok := call.Callee.(*ast.FieldAccess); ok {
			if id, ok := fa.Object.(*ast.Identifier); ok && id.Name == "os" {
				return true
			}
		}
	}
	return false
}

// --- top-level ---

func (e *Emitter) emitTopLevel(item ast.TopLevelItem) {
	switch it := item.(type) {
	case *ast.FnDef:
		e.emitFnDef(it)
	case *ast.StructDef:
		e.emitStructDef(it)
	case *ast.EnumDef:
		e.emitEnumDef(it)
	case *ast.TypeDef:
		e.emitTypeDef(it)
	case *ast.ConstDef:
		e.emitConstDef(it)
	case *ast.SpecBlock, *ast.TraitDef, *ast.ImplBlock:
		// Specs and traits are compile-time constructs; skip.
		e.line("// [skipped: " + fmt.Sprintf("%T", item) + "]")
	}
}

func (e *Emitter) emitFnDef(fn *ast.FnDef) {
	e.line("")
	// Signature.
	var sig strings.Builder
	sig.WriteString("func ")
	sig.WriteString(goExportedIdent(fn.Name, fn.Visibility == ast.Public))
	sig.WriteString("(")
	for i, p := range fn.Params {
		if i > 0 {
			sig.WriteString(", ")
		}
		sig.WriteString(goIdent(p.Name) + " " + e.goType(p.TypeExpr))
	}
	sig.WriteString(")")
	if fn.ReturnType != nil {
		sig.WriteString(" " + e.goType(fn.ReturnType))
	}
	sig.WriteString(" {")
	e.line(sig.String())
	e.indent++
	for _, stmt := range fn.Body {
		e.emitStmt(stmt)
	}
	e.indent--
	e.line("}")
}

func (e *Emitter) emitStructDef(s *ast.StructDef) {
	e.line("")
	e.line("type " + goExportedIdent(s.Name, s.Visibility == ast.Public) + " struct {")
	e.indent++
	for _, f := range s.Fields {
		exported := f.Visibility == ast.Public || s.Visibility == ast.Public
		e.line(goExportedIdent(f.Name, exported) + " " + e.goType(f.TypeExpr))
	}
	e.indent--
	e.line("}")
}

func (e *Emitter) emitEnumDef(en *ast.EnumDef) {
	e.line("")
	exported := en.Visibility == ast.Public
	goName := goExportedIdent(en.Name, exported)

	// Check if all variants are unit (no fields).
	allUnit := true
	for _, v := range en.Variants {
		if len(v.Fields) > 0 {
			allUnit = false
			break
		}
	}

	if allUnit {
		// Simple enum: emit as int type + const block.
		e.line("type " + goName + " int")
		e.line("")
		e.line("const (")
		e.indent++
		for i, v := range en.Variants {
			if i == 0 {
				e.line(goExportedIdent(v.Name, exported) + " " + goName + " = iota")
			} else {
				e.line(goExportedIdent(v.Name, exported))
			}
		}
		e.indent--
		e.line(")")
	} else {
		// Tagged enum: emit interface + one struct per variant.
		e.line("type " + goName + " interface { is" + goName + "() }")
		for _, v := range en.Variants {
			varName := goExportedIdent(en.Name+"_"+v.Name, exported)
			e.line("")
			e.line("type " + varName + " struct {")
			e.indent++
			for i, f := range v.Fields {
				e.line(fmt.Sprintf("F%d %s", i, e.goType(f)))
			}
			e.indent--
			e.line("}")
			e.line("func (" + varName + ") is" + goName + "() {}")
		}
	}
}

func (e *Emitter) emitTypeDef(td *ast.TypeDef) {
	exported := td.Visibility == ast.Public
	e.line("type " + goExportedIdent(td.Name, exported) + " = " + e.goType(td.Body))
}

func (e *Emitter) emitConstDef(cd *ast.ConstDef) {
	exported := cd.Visibility == ast.Public
	val := ""
	if cd.Value != nil {
		val = " = " + e.emitExpr(cd.Value)
	}
	e.line("const " + goExportedIdent(cd.Name, exported) + val)
}

// --- statements ---

func (e *Emitter) emitStmt(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		e.emitLetStmt(s)
	case *ast.AssignStmt:
		e.emitAssignStmt(s)
	case *ast.ReturnStmt:
		e.emitReturnStmt(s)
	case *ast.IfStmt:
		e.emitIfStmt(s)
	case *ast.MatchStmt:
		e.emitMatchStmt(s)
	case *ast.ForStmt:
		e.emitForStmt(s)
	case *ast.WhileStmt:
		e.emitWhileStmt(s)
	case *ast.ExprStmt:
		e.line(e.emitExpr(s.Expr))
	case *ast.AssertStmt:
		e.line(`if !(` + e.emitExpr(s.Condition) + `) { panic("assertion failed") }`)
	case *ast.BreakStmt:
		e.line("break")
	case *ast.ContinueStmt:
		e.line("continue")
	case *ast.LetTupleDestructure:
		e.emitTupleDestructure(s)
	case *ast.WithStmt:
		e.line("// [with stmt: effects not compiled]")
	default:
		e.line(fmt.Sprintf("// [unsupported stmt: %T]", stmt))
	}
}

func (e *Emitter) emitLetStmt(s *ast.LetStmt) {
	name := goIdent(s.Name)
	if s.Value == nil {
		// Declaration with type only.
		var zero string
		if s.TypeHint != nil {
			zero = "var " + name + " " + e.goType(s.TypeHint)
		} else {
			zero = "var " + name + " any"
		}
		e.line(zero)
		return
	}
	val := e.emitExpr(s.Value)
	if s.Mutable {
		e.line(name + " := " + val)
	} else {
		e.line(name + " := " + val)
	}
}

func (e *Emitter) emitAssignStmt(s *ast.AssignStmt) {
	target := e.emitExpr(s.Target)
	val := e.emitExpr(s.Value)
	e.line(target + " = " + val)
}

func (e *Emitter) emitReturnStmt(s *ast.ReturnStmt) {
	if s.Value == nil {
		e.line("return")
	} else {
		e.line("return " + e.emitExpr(s.Value))
	}
}

func (e *Emitter) emitIfStmt(s *ast.IfStmt) {
	e.line("if " + e.emitExpr(s.Condition) + " {")
	e.indent++
	for _, stmt := range s.ThenBody {
		e.emitStmt(stmt)
	}
	e.indent--
	for _, elif := range s.ElifClauses {
		e.line("} else if " + e.emitExpr(elif.Condition) + " {")
		e.indent++
		for _, stmt := range elif.Body {
			e.emitStmt(stmt)
		}
		e.indent--
	}
	if len(s.ElseBody) > 0 {
		e.line("} else {")
		e.indent++
		for _, stmt := range s.ElseBody {
			e.emitStmt(stmt)
		}
		e.indent--
	}
	e.line("}")
}

func (e *Emitter) emitMatchStmt(s *ast.MatchStmt) {
	subject := e.emitExpr(s.Subject)
	e.line("switch _match := " + subject + "; {")
	for _, c := range s.Cases {
		cond := e.emitPatternCond("_match", c.Pattern)
		if c.Guard != nil {
			cond += " && " + e.emitExpr(c.Guard)
		}
		e.line("case " + cond + ":")
		e.indent++
		e.emitPatternBindings("_match", c.Pattern)
		for _, stmt := range c.Body {
			e.emitStmt(stmt)
		}
		e.indent--
	}
	e.line("}")
}

// emitPatternCond emits the boolean condition for a case pattern.
func (e *Emitter) emitPatternCond(subj string, pat ast.Pattern) string {
	switch p := pat.(type) {
	case *ast.WildcardPattern:
		return "true"
	case *ast.BindingPattern:
		return "true"
	case *ast.LiteralPattern:
		return subj + " == " + p.Value
	case *ast.ConstructorPattern:
		// Variant constructor: e.g., Some(x) or Color.Red
		name := p.TypeName
		if strings.Contains(name, ".") {
			parts := strings.SplitN(name, ".", 2)
			// enumName.VariantName
			enumName := parts[0]
			variantName := parts[1]
			if variants, ok := e.enumVariants[enumName]; ok {
				// unit enum: compare as int const
				_ = variants
				return subj + " == " + goExportedIdent(variantName, true)
			}
			// tagged enum: type assertion
			return "func() bool { _, ok := " + subj + ".(" + goExportedIdent(enumName+"_"+variantName, true) + "); return ok }()"
		}
		// Could be a bare variant name or Option/Result
		switch name {
		case "Some":
			return subj + ".ok"
		case "None":
			return "!" + subj + ".ok"
		case "Ok":
			return subj + ".ok"
		case "Err":
			return "!" + subj + ".ok"
		}
		// Tagged enum variant without enum prefix
		if enumName, ok := e.variantEnum[name]; ok {
			return "func() bool { _, ok := " + subj + ".(" + goExportedIdent(enumName+"_"+name, true) + "); return ok }()"
		}
		return "true"
	default:
		return "true"
	}
}

// emitPatternBindings emits variable bindings extracted from a pattern match.
func (e *Emitter) emitPatternBindings(subj string, pat ast.Pattern) {
	switch p := pat.(type) {
	case *ast.BindingPattern:
		e.line(goIdent(p.Name) + " := _match")
	case *ast.ConstructorPattern:
		switch p.TypeName {
		case "Some":
			if len(p.Fields) == 1 {
				if bp, ok := p.Fields[0].(*ast.BindingPattern); ok {
					e.line(goIdent(bp.Name) + " := _match.v")
				}
			}
		case "Ok":
			if len(p.Fields) == 1 {
				if bp, ok := p.Fields[0].(*ast.BindingPattern); ok {
					e.line(goIdent(bp.Name) + " := _match.v")
				}
			}
		case "Err":
			if len(p.Fields) == 1 {
				if bp, ok := p.Fields[0].(*ast.BindingPattern); ok {
					e.line(goIdent(bp.Name) + " := _match.e")
				}
			}
		default:
			if strings.Contains(p.TypeName, ".") {
				parts := strings.SplitN(p.TypeName, ".", 2)
				enumName := parts[0]
				variantName := parts[1]
				variantType := goExportedIdent(enumName+"_"+variantName, true)
				if len(p.Fields) > 0 {
					e.line("_v, _ := _match.(" + variantType + ")")
					for i, f := range p.Fields {
						if bp, ok := f.(*ast.BindingPattern); ok {
							e.line(fmt.Sprintf("%s := _v.F%d", goIdent(bp.Name), i))
						}
					}
				}
			}
		}
	}
}

func (e *Emitter) emitForStmt(s *ast.ForStmt) {
	e.line("for _, " + goIdent(s.Variable) + " := range " + e.emitExpr(s.Iterable) + " {")
	e.indent++
	for _, stmt := range s.Body {
		e.emitStmt(stmt)
	}
	e.indent--
	e.line("}")
}

func (e *Emitter) emitWhileStmt(s *ast.WhileStmt) {
	e.line("for " + e.emitExpr(s.Condition) + " {")
	e.indent++
	for _, stmt := range s.Body {
		e.emitStmt(stmt)
	}
	e.indent--
	e.line("}")
}

func (e *Emitter) emitTupleDestructure(s *ast.LetTupleDestructure) {
	names := make([]string, len(s.Names))
	for i, n := range s.Names {
		names[i] = goIdent(n)
	}
	e.line(strings.Join(names, ", ") + " := " + e.emitExpr(s.Value) + ".unpack()")
	// Note: Aura tuples don't have an unpack method; this emits a placeholder.
	// For now we emit a direct tuple access pattern.
	e.warn("tuple destructure: partial support")
}

// --- expressions ---

func (e *Emitter) emitExpr(expr ast.Expr) string {
	if expr == nil {
		return "nil"
	}
	switch ex := expr.(type) {
	case *ast.Identifier:
		return goIdent(ex.Name)
	case *ast.IntLiteral:
		return fmt.Sprintf("int64(%s)", ex.Value)
	case *ast.FloatLiteral:
		return ex.Value
	case *ast.StringLiteral:
		return e.emitStringLiteral(ex)
	case *ast.BoolLiteral:
		if ex.Value {
			return "true"
		}
		return "false"
	case *ast.NoneLiteral:
		return "auraNone[any]()"
	case *ast.BinaryOp:
		return e.emitBinaryOp(ex)
	case *ast.UnaryOp:
		return e.emitUnaryOp(ex)
	case *ast.CallExpr:
		return e.emitCallExpr(ex)
	case *ast.FieldAccess:
		return e.emitExpr(ex.Object) + "." + goExportedIdent(ex.Field, true)
	case *ast.OptionalFieldAccess:
		return e.emitExpr(ex.Object) + ".v." + goExportedIdent(ex.Field, true)
	case *ast.IndexExpr:
		return e.emitExpr(ex.Object) + "[" + e.emitExpr(ex.Index) + "]"
	case *ast.ListExpr:
		return e.emitListExpr(ex)
	case *ast.MapExpr:
		return e.emitMapExpr(ex)
	case *ast.StructExpr:
		return e.emitStructExpr(ex)
	case *ast.IfExpr:
		return e.emitIfExpr(ex)
	case *ast.TupleLiteral:
		return e.emitTupleLiteral(ex)
	case *ast.Lambda:
		return e.emitLambda(ex)
	case *ast.OptionPropagate:
		// x? in Aura — propagate None upward. Emit as a simple unwrap for now.
		return e.emitExpr(ex.Expr) + ".v"
	case *ast.PipelineExpr:
		// x |> f — emit as f(x)
		return e.emitExpr(ex.Right) + "(" + e.emitExpr(ex.Left) + ")"
	default:
		e.warn(fmt.Sprintf("unsupported expr: %T", expr))
		return `"?"`
	}
}

func (e *Emitter) emitStringLiteral(s *ast.StringLiteral) string {
	if len(s.Parts) == 0 {
		// Plain string.
		return `"` + strings.ReplaceAll(s.Value, `"`, `\"`) + `"`
	}
	// Interpolated string: build fmt.Sprintf call.
	var fmtStr strings.Builder
	var args []string
	fmtStr.WriteString(`"`)
	for _, part := range s.Parts {
		if part.IsExpr {
			fmtStr.WriteString("%v")
			args = append(args, e.emitExpr(part.Expr))
		} else {
			fmtStr.WriteString(strings.ReplaceAll(part.Text, `"`, `\"`))
		}
	}
	fmtStr.WriteString(`"`)
	if len(args) == 0 {
		return fmtStr.String()
	}
	return "fmt.Sprintf(" + fmtStr.String() + ", " + strings.Join(args, ", ") + ")"
}

func (e *Emitter) emitBinaryOp(ex *ast.BinaryOp) string {
	l := e.emitExpr(ex.Left)
	r := e.emitExpr(ex.Right)
	op := ex.Op
	// Aura uses "and"/"or"/"not" — map to Go equivalents.
	switch op {
	case "and":
		op = "&&"
	case "or":
		op = "||"
	case "==", "!=", "<", ">", "<=", ">=":
		// direct
	case "+", "-", "*", "/", "%":
		// direct
	}
	return "(" + l + " " + op + " " + r + ")"
}

func (e *Emitter) emitUnaryOp(ex *ast.UnaryOp) string {
	op := ex.Op
	if op == "not" {
		op = "!"
	}
	return op + e.emitExpr(ex.Operand)
}

func (e *Emitter) emitCallExpr(ex *ast.CallExpr) string {
	// Handle built-ins and special constructors.
	if id, ok := ex.Callee.(*ast.Identifier); ok {
		switch id.Name {
		case "print":
			args := e.emitArgs(ex.Args)
			return "fmt.Print(" + strings.Join(args, ", ") + ")"
		case "println":
			args := e.emitArgs(ex.Args)
			return "fmt.Println(" + strings.Join(args, ", ") + ")"
		case "len":
			args := e.emitArgs(ex.Args)
			if len(args) == 1 {
				return "int64(len(" + args[0] + "))"
			}
		case "append":
			args := e.emitArgs(ex.Args)
			if len(args) == 2 {
				return "append(" + args[0] + ", " + args[1] + ")"
			}
		case "Some":
			args := e.emitArgs(ex.Args)
			if len(args) == 1 {
				return "auraSome(" + args[0] + ")"
			}
		case "None":
			return "auraNone[any]()"
		case "Ok":
			args := e.emitArgs(ex.Args)
			if len(args) == 1 {
				return "auraOk[any, any](" + args[0] + ")"
			}
		case "Err":
			args := e.emitArgs(ex.Args)
			if len(args) == 1 {
				return "auraErr[any, any](" + args[0] + ")"
			}
		case "int", "Int":
			args := e.emitArgs(ex.Args)
			if len(args) == 1 {
				return "int64(" + args[0] + ")"
			}
		case "float", "Float":
			args := e.emitArgs(ex.Args)
			if len(args) == 1 {
				return "float64(" + args[0] + ")"
			}
		case "str", "String":
			args := e.emitArgs(ex.Args)
			if len(args) == 1 {
				return "fmt.Sprintf(\"%v\", " + args[0] + ")"
			}
		}
	}

	// Method calls on known objects.
	if fa, ok := ex.Callee.(*ast.FieldAccess); ok {
		obj := e.emitExpr(fa.Object)
		args := e.emitArgs(ex.Args)
		switch fa.Field {
		case "append":
			if len(args) == 1 {
				return "append(" + obj + ", " + args[0] + ")"
			}
		case "len":
			return "int64(len(" + obj + "))"
		}
		allArgs := append([]string{}, args...)
		return obj + "." + goExportedIdent(fa.Field, true) + "(" + strings.Join(allArgs, ", ") + ")"
	}

	// General call.
	callee := e.emitExpr(ex.Callee)
	args := e.emitArgs(ex.Args)
	return callee + "(" + strings.Join(args, ", ") + ")"
}

func (e *Emitter) emitArgs(args []*ast.Arg) []string {
	out := make([]string, len(args))
	for i, a := range args {
		out[i] = e.emitExpr(a.Value)
	}
	return out
}

func (e *Emitter) emitListExpr(ex *ast.ListExpr) string {
	if len(ex.Elements) == 0 {
		return "[]any{}"
	}
	elems := make([]string, len(ex.Elements))
	for i, el := range ex.Elements {
		elems[i] = e.emitExpr(el)
	}
	return "[]any{" + strings.Join(elems, ", ") + "}"
}

func (e *Emitter) emitMapExpr(ex *ast.MapExpr) string {
	if len(ex.Entries) == 0 {
		return "map[any]any{}"
	}
	entries := make([]string, len(ex.Entries))
	for i, entry := range ex.Entries {
		entries[i] = e.emitExpr(entry.Key) + ": " + e.emitExpr(entry.Value)
	}
	return "map[any]any{" + strings.Join(entries, ", ") + "}"
}

func (e *Emitter) emitStructExpr(ex *ast.StructExpr) string {
	name := ex.TypeName
	fields := make([]string, len(ex.Fields))
	for i, f := range ex.Fields {
		fields[i] = goExportedIdent(f.Name, true) + ": " + e.emitExpr(f.Value)
	}
	return goExportedIdent(name, true) + "{" + strings.Join(fields, ", ") + "}"
}

func (e *Emitter) emitIfExpr(ex *ast.IfExpr) string {
	// Go has no ternary; emit an immediately-invoked func literal.
	cond := e.emitExpr(ex.Condition)
	then := e.emitExpr(ex.ThenExpr)
	els := e.emitExpr(ex.ElseExpr)
	return "func() any { if " + cond + " { return " + then + " }; return " + els + " }()"
}

func (e *Emitter) emitTupleLiteral(ex *ast.TupleLiteral) string {
	elems := make([]string, len(ex.Elements))
	for i, el := range ex.Elements {
		elems[i] = e.emitExpr(el)
	}
	// Emit as a struct literal — real tuple support would need codegen types.
	return "struct{ V []any }{V: []any{" + strings.Join(elems, ", ") + "}}"
}

func (e *Emitter) emitLambda(ex *ast.Lambda) string {
	params := make([]string, len(ex.Params))
	for i, p := range ex.Params {
		t := "any"
		if p.TypeExpr != nil {
			t = e.goType(p.TypeExpr)
		}
		params[i] = goIdent(p.Name) + " " + t
	}
	body := e.emitExpr(ex.Body)
	return "func(" + strings.Join(params, ", ") + ") any { return " + body + " }"
}

// --- type mapping ---

func (e *Emitter) goType(te ast.TypeExpr) string {
	if te == nil {
		return "any"
	}
	switch t := te.(type) {
	case *ast.NamedType:
		switch t.Name {
		case "Int":
			return "int64"
		case "Float":
			return "float64"
		case "String":
			return "string"
		case "Bool":
			return "bool"
		case "Any":
			return "any"
		case "Option":
			if len(t.Args) == 1 {
				return "auraOption[" + e.goType(t.Args[0]) + "]"
			}
			return "auraOption[any]"
		case "Result":
			if len(t.Args) == 2 {
				return "auraResult[" + e.goType(t.Args[0]) + ", " + e.goType(t.Args[1]) + "]"
			}
			return "auraResult[any, any]"
		case "List":
			if len(t.Args) == 1 {
				return "[]" + e.goType(t.Args[0])
			}
			return "[]any"
		case "Map":
			if len(t.Args) == 2 {
				return "map[" + e.goType(t.Args[0]) + "]" + e.goType(t.Args[1])
			}
			return "map[any]any"
		case "Set":
			if len(t.Args) == 1 {
				return "map[" + e.goType(t.Args[0]) + "]struct{}"
			}
			return "map[any]struct{}"
		default:
			return goExportedIdent(t.Name, true)
		}
	case *ast.ListType:
		return "[]" + e.goType(t.Element)
	case *ast.MapType:
		return "map[" + e.goType(t.Key) + "]" + e.goType(t.Value)
	case *ast.SetType:
		return "map[" + e.goType(t.Element) + "]struct{}"
	case *ast.OptionType:
		return "auraOption[" + e.goType(t.Inner) + "]"
	case *ast.TupleType:
		// Represent as []any for now.
		return "[]any"
	case *ast.FunctionType:
		params := make([]string, len(t.Params))
		for i, p := range t.Params {
			params[i] = e.goType(p)
		}
		return "func(" + strings.Join(params, ", ") + ") " + e.goType(t.ReturnType)
	case *ast.UnionType, *ast.IntersectionType, *ast.RefinementType:
		return "any"
	case *ast.QualifiedType:
		return t.Qualifier + "." + t.Name
	default:
		return "any"
	}
}

// --- helpers ---

func (e *Emitter) line(s string) {
	if s == "" {
		e.sb.WriteString("\n")
		return
	}
	e.sb.WriteString(strings.Repeat("\t", e.indent))
	e.sb.WriteString(s)
	e.sb.WriteString("\n")
}

func (e *Emitter) warn(msg string) {
	e.errors = append(e.errors, "warning: "+msg)
}

// goIdent converts an Aura identifier to a Go identifier.
// Replaces any character not valid in Go identifiers with _.
func goIdent(name string) string {
	if name == "_" {
		return "_"
	}
	// Reserved Go keywords → add underscore suffix.
	switch name {
	case "type", "range", "func", "var", "const", "map", "chan",
		"go", "select", "defer", "fallthrough", "interface",
		"package", "import", "struct", "switch", "case", "default",
		"for", "if", "else", "return", "break", "continue",
		"goto", "nil", "true", "false":
		return name + "_"
	}
	return name
}

// goExportedIdent produces an exported (uppercase) or unexported name.
func goExportedIdent(name string, exported bool) string {
	if name == "" {
		return "_"
	}
	name = goIdent(name)
	if exported {
		r := []rune(name)
		r[0] = unicode.ToUpper(r[0])
		return string(r)
	}
	r := []rune(name)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
