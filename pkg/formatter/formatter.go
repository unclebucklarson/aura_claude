// Package formatter converts an Aura AST back to canonical source code.
// It enforces consistent 4-space indentation and deterministic output.
package formatter

import (
        "fmt"
        "strings"

        "github.com/unclebucklarson/aura/pkg/ast"
        "github.com/unclebucklarson/aura/pkg/token"
)

// Formatter converts AST nodes to canonical Aura source.
type Formatter struct {
        indent     int
        indentStr  string // "    " (4 spaces)
        buf        strings.Builder
}

// New creates a new Formatter with 4-space indentation.
func New() *Formatter {
        return &Formatter{
                indentStr: "    ",
        }
}

// Format formats an entire Module to a string.
func (f *Formatter) Format(module *ast.Module) string {
        f.buf.Reset()
        f.indent = 0
        f.formatModule(module)
        return f.buf.String()
}

// FormatNode formats a single AST node.
func (f *Formatter) FormatNode(node ast.Node) string {
        f.buf.Reset()
        f.indent = 0
        switch n := node.(type) {
        case *ast.Module:
                f.formatModule(n)
        case ast.TopLevelItem:
                f.formatTopLevelItem(n)
        case ast.Statement:
                f.formatStatement(n)
        case ast.Expr:
                f.formatExpr(n)
        }
        return f.buf.String()
}

func (f *Formatter) write(s string) {
        f.buf.WriteString(s)
}

func (f *Formatter) writeln(s string) {
        f.buf.WriteString(s)
        f.buf.WriteByte('\n')
}

func (f *Formatter) writeIndent() {
        for i := 0; i < f.indent; i++ {
                f.buf.WriteString(f.indentStr)
        }
}

func (f *Formatter) writeLine(s string) {
        f.writeIndent()
        f.writeln(s)
}

// --- Module ---

func (f *Formatter) formatModule(m *ast.Module) {
        if m.Name != nil {
                f.writeLine("module " + m.Name.String())
                f.write("\n")
        }

        if len(m.Imports) > 0 {
                for _, imp := range m.Imports {
                        f.formatImport(imp)
                }
                f.write("\n")
        }

        for i, item := range m.Items {
                f.formatTopLevelItem(item)
                if i < len(m.Items)-1 {
                        f.write("\n")
                }
        }
}

func (f *Formatter) formatImport(imp *ast.ImportNode) {
        if imp.Names != nil {
                // from X import a, b
                f.writeIndent()
                f.write("from " + imp.Path.String() + " import ")
                f.writeln(strings.Join(imp.Names, ", "))
        } else if imp.Alias != "" {
                f.writeLine("import " + imp.Path.String() + " as " + imp.Alias)
        } else {
                f.writeLine("import " + imp.Path.String())
        }
}

// --- Top-level items ---

func (f *Formatter) formatTopLevelItem(item ast.TopLevelItem) {
        switch n := item.(type) {
        case *ast.TypeDef:
                f.formatTypeDef(n)
        case *ast.StructDef:
                f.formatStructDef(n)
        case *ast.EnumDef:
                f.formatEnumDef(n)
        case *ast.TraitDef:
                f.formatTraitDef(n)
        case *ast.ImplBlock:
                f.formatImplBlock(n)
        case *ast.SpecBlock:
                f.formatSpecBlock(n)
        case *ast.FnDef:
                f.formatFnDef(n)
        case *ast.ConstDef:
                f.formatConstDef(n)
        case *ast.TestBlock:
                f.formatTestBlock(n)
        }
}

func (f *Formatter) formatComments(comments []ast.Comment) {
        for _, c := range comments {
                f.writeIndent()
                if c.IsDoc {
                        f.writeln("## " + c.Text)
                } else {
                        f.writeln("# " + c.Text)
                }
        }
}

func (f *Formatter) visPrefix(vis ast.Visibility) string {
        if vis == ast.Public {
                return "pub "
        }
        return ""
}

// --- Type definitions ---

func (f *Formatter) formatTypeDef(td *ast.TypeDef) {
        f.formatComments(td.Comments)
        f.writeIndent()
        f.write(f.visPrefix(td.Visibility))
        f.write("type " + td.Name)
        if len(td.TypeParams) > 0 {
                f.write("[" + strings.Join(td.TypeParams, ", ") + "]")
        }
        f.write(" = ")
        f.formatTypeExpr(td.Body)
        f.write("\n")
}

func (f *Formatter) formatTypeExpr(te ast.TypeExpr) {
        switch t := te.(type) {
        case *ast.NamedType:
                f.write(t.Name)
                if len(t.Args) > 0 {
                        f.write("[")
                        for i, arg := range t.Args {
                                if i > 0 {
                                        f.write(", ")
                                }
                                f.formatTypeExpr(arg)
                        }
                        f.write("]")
                }
        case *ast.QualifiedType:
                f.write(t.Qualifier + "." + t.Name)
                if len(t.Args) > 0 {
                        f.write("[")
                        for i, arg := range t.Args {
                                if i > 0 {
                                        f.write(", ")
                                }
                                f.formatTypeExpr(arg)
                        }
                        f.write("]")
                }
        case *ast.UnionType:
                f.formatTypeExpr(t.Left)
                f.write(" | ")
                f.formatTypeExpr(t.Right)
        case *ast.IntersectionType:
                f.formatTypeExpr(t.Left)
                f.write(" & ")
                f.formatTypeExpr(t.Right)
        case *ast.FunctionType:
                f.write("fn(")
                for i, p := range t.Params {
                        if i > 0 {
                                f.write(", ")
                        }
                        f.formatTypeExpr(p)
                }
                f.write(") -> ")
                f.formatTypeExpr(t.ReturnType)
        case *ast.TupleType:
                f.write("(")
                for i, e := range t.Elements {
                        if i > 0 {
                                f.write(", ")
                        }
                        f.formatTypeExpr(e)
                }
                f.write(")")
        case *ast.ListType:
                f.write("[")
                f.formatTypeExpr(t.Element)
                f.write("]")
        case *ast.MapType:
                f.write("{")
                f.formatTypeExpr(t.Key)
                f.write(": ")
                f.formatTypeExpr(t.Value)
                f.write("}")
        case *ast.SetType:
                f.write("{")
                f.formatTypeExpr(t.Element)
                f.write("}")
        case *ast.OptionType:
                f.formatTypeExpr(t.Inner)
                f.write("?")
        case *ast.RefinementType:
                f.formatTypeExpr(t.Base)
                f.write(" where ")
                f.formatExpr(t.Predicate)
        case *ast.StringLitType:
                f.write(fmt.Sprintf("%q", t.Value))
        default:
                f.write("_unknown_type_")
        }
}

// --- Struct ---

func (f *Formatter) formatStructDef(sd *ast.StructDef) {
        f.formatComments(sd.Comments)
        f.writeIndent()
        f.write(f.visPrefix(sd.Visibility))
        f.write("struct " + sd.Name)
        if len(sd.TypeParams) > 0 {
                f.write("[" + strings.Join(sd.TypeParams, ", ") + "]")
        }
        f.writeln(":")
        f.indent++
        for _, field := range sd.Fields {
                f.formatFieldDef(field)
        }
        f.indent--
}

func (f *Formatter) formatFieldDef(fd *ast.FieldDef) {
        f.formatComments(fd.Comments)
        f.writeIndent()
        f.write(f.visPrefix(fd.Visibility))
        f.write(fd.Name + ": ")
        f.formatTypeExpr(fd.TypeExpr)
        if fd.Default != nil {
                f.write(" = ")
                f.formatExpr(fd.Default)
        }
        f.write("\n")
}

// --- Enum ---

func (f *Formatter) formatEnumDef(ed *ast.EnumDef) {
        f.formatComments(ed.Comments)
        f.writeIndent()
        f.write(f.visPrefix(ed.Visibility))
        f.write("enum " + ed.Name)
        if len(ed.TypeParams) > 0 {
                f.write("[" + strings.Join(ed.TypeParams, ", ") + "]")
        }
        f.writeln(":")
        f.indent++
        for _, v := range ed.Variants {
                f.writeIndent()
                f.write(v.Name)
                if len(v.Fields) > 0 {
                        f.write("(")
                        for i, ft := range v.Fields {
                                if i > 0 {
                                        f.write(", ")
                                }
                                f.formatTypeExpr(ft)
                        }
                        f.write(")")
                }
                f.write("\n")
        }
        f.indent--
}

// --- Trait ---

func (f *Formatter) formatTraitDef(td *ast.TraitDef) {
        f.formatComments(td.Comments)
        f.writeIndent()
        f.write(f.visPrefix(td.Visibility))
        f.write("trait " + td.Name)
        if len(td.TypeParams) > 0 {
                f.write("[" + strings.Join(td.TypeParams, ", ") + "]")
        }
        f.writeln(":")
        f.indent++
        for _, m := range td.Members {
                if fn, ok := m.(*ast.FnDef); ok {
                        f.formatFnDef(fn)
                }
        }
        f.indent--
}

// --- Impl ---

func (f *Formatter) formatImplBlock(ib *ast.ImplBlock) {
        f.formatComments(ib.Comments)
        f.writeIndent()
        if ib.TraitName != "" {
                f.write("impl " + ib.TraitName + " for ")
                f.formatTypeExpr(ib.TargetType)
        } else {
                f.write("impl ")
                f.formatTypeExpr(ib.TargetType)
        }
        f.writeln(":")
        f.indent++
        for _, m := range ib.Methods {
                f.formatFnDef(m)
                f.write("\n")
        }
        f.indent--
}

// --- Spec ---

func (f *Formatter) formatSpecBlock(sb *ast.SpecBlock) {
        f.formatComments(sb.Comments)
        f.writeLine("spec " + sb.Name + ":")
        f.indent++

        if sb.Doc != "" {
                f.writeLine(fmt.Sprintf("doc: %q", sb.Doc))
                f.write("\n")
        }

        if len(sb.Inputs) > 0 {
                f.writeLine("inputs:")
                f.indent++
                for _, inp := range sb.Inputs {
                        f.writeIndent()
                        f.write(inp.Name + ": ")
                        f.formatTypeExpr(inp.TypeExpr)
                        if inp.Default != nil {
                                f.write(" = ")
                                f.formatExpr(inp.Default)
                        }
                        if inp.Description != "" {
                                f.write(fmt.Sprintf(" - %q", inp.Description))
                        }
                        f.write("\n")
                }
                f.indent--
                f.write("\n")
        }

        if len(sb.Guarantees) > 0 {
                f.writeLine("guarantees:")
                f.indent++
                for _, g := range sb.Guarantees {
                        f.writeLine(fmt.Sprintf("- %q", g.Condition))
                }
                f.indent--
                f.write("\n")
        }

        if len(sb.Effects) > 0 {
                f.writeLine("effects: " + strings.Join(sb.Effects, ", "))
                f.write("\n")
        }

        if len(sb.Errors) > 0 {
                f.writeLine("errors:")
                f.indent++
                for _, e := range sb.Errors {
                        f.writeIndent()
                        f.write(e.TypeName)
                        if len(e.Fields) > 0 {
                                f.write("(")
                                for i, ft := range e.Fields {
                                        if i > 0 {
                                                f.write(", ")
                                        }
                                        f.formatTypeExpr(ft)
                                }
                                f.write(")")
                        }
                        if e.Description != "" {
                                f.write(fmt.Sprintf(" - %q", e.Description))
                        }
                        f.write("\n")
                }
                f.indent--
        } else if sb.Errors != nil {
                // explicitly empty errors
        }

        f.indent--
}

// --- Function ---

func (f *Formatter) formatFnDef(fd *ast.FnDef) {
        f.formatComments(fd.Comments)
        f.writeIndent()
        f.write(f.visPrefix(fd.Visibility))
        f.write("fn " + fd.Name)
        if len(fd.TypeParams) > 0 {
                f.write("[" + strings.Join(fd.TypeParams, ", ") + "]")
        }
        f.write("(")
        for i, p := range fd.Params {
                if i > 0 {
                        f.write(", ")
                }
                f.formatParam(p)
        }
        f.write(")")
        if fd.ReturnType != nil {
                f.write(" -> ")
                f.formatTypeExpr(fd.ReturnType)
        }
        if len(fd.Effects) > 0 {
                f.write(" with " + strings.Join(fd.Effects, ", "))
        }
        if fd.Satisfies != "" {
                f.write(" satisfies " + fd.Satisfies)
        }
        f.writeln(":")

        f.indent++
        for _, stmt := range fd.Body {
                f.formatStatement(stmt)
        }
        f.indent--
}

func (f *Formatter) formatParam(p *ast.Param) {
        f.write(p.Name)
        if p.TypeExpr != nil {
                f.write(": ")
                f.formatTypeExpr(p.TypeExpr)
        }
        if p.Default != nil {
                f.write(" = ")
                f.formatExpr(p.Default)
        }
}

// --- Const ---

func (f *Formatter) formatConstDef(cd *ast.ConstDef) {
        f.formatComments(cd.Comments)
        f.writeIndent()
        f.write(f.visPrefix(cd.Visibility))
        f.write("let " + cd.Name)
        if cd.TypeExpr != nil {
                f.write(": ")
                f.formatTypeExpr(cd.TypeExpr)
        }
        f.write(" = ")
        f.formatExpr(cd.Value)
        f.write("\n")
}

// --- Test ---

func (f *Formatter) formatTestBlock(tb *ast.TestBlock) {
        f.formatComments(tb.Comments)
        f.writeLine(fmt.Sprintf("test %q:", tb.Name))
        f.indent++
        for _, stmt := range tb.Body {
                f.formatStatement(stmt)
        }
        f.indent--
}

// --- Statements ---

func (f *Formatter) formatStatement(stmt ast.Statement) {
        switch s := stmt.(type) {
        case *ast.LetStmt:
                f.writeIndent()
                f.write("let ")
                if s.Mutable {
                        f.write("mut ")
                }
                f.write(s.Name)
                if s.TypeHint != nil {
                        f.write(": ")
                        f.formatTypeExpr(s.TypeHint)
                }
                f.write(" = ")
                f.formatExpr(s.Value)
                f.write("\n")

        case *ast.AssignStmt:
                f.writeIndent()
                f.formatExpr(s.Target)
                f.write(" = ")
                f.formatExpr(s.Value)
                f.write("\n")

        case *ast.ReturnStmt:
                f.writeIndent()
                if s.Value != nil {
                        f.write("return ")
                        f.formatExpr(s.Value)
                } else {
                        f.write("return")
                }
                f.write("\n")

        case *ast.IfStmt:
                f.writeIndent()
                f.write("if ")
                f.formatExpr(s.Condition)
                f.writeln(":")
                f.indent++
                for _, stmt := range s.ThenBody {
                        f.formatStatement(stmt)
                }
                f.indent--
                for _, elif := range s.ElifClauses {
                        f.writeIndent()
                        f.write("elif ")
                        f.formatExpr(elif.Condition)
                        f.writeln(":")
                        f.indent++
                        for _, stmt := range elif.Body {
                                f.formatStatement(stmt)
                        }
                        f.indent--
                }
                if len(s.ElseBody) > 0 {
                        f.writeLine("else:")
                        f.indent++
                        for _, stmt := range s.ElseBody {
                                f.formatStatement(stmt)
                        }
                        f.indent--
                }

        case *ast.MatchStmt:
                f.writeIndent()
                f.write("match ")
                f.formatExpr(s.Subject)
                f.writeln(":")
                f.indent++
                for _, c := range s.Cases {
                        f.writeIndent()
                        f.write("case ")
                        f.formatPattern(c.Pattern)
                        if c.Guard != nil {
                                f.write(" if ")
                                f.formatExpr(c.Guard)
                        }
                        f.writeln(":")
                        f.indent++
                        for _, stmt := range c.Body {
                                f.formatStatement(stmt)
                        }
                        f.indent--
                }
                f.indent--

        case *ast.ForStmt:
                f.writeIndent()
                f.write("for " + s.Variable + " in ")
                f.formatExpr(s.Iterable)
                f.writeln(":")
                f.indent++
                for _, stmt := range s.Body {
                        f.formatStatement(stmt)
                }
                f.indent--

        case *ast.WhileStmt:
                f.writeIndent()
                f.write("while ")
                f.formatExpr(s.Condition)
                f.writeln(":")
                f.indent++
                for _, stmt := range s.Body {
                        f.formatStatement(stmt)
                }
                f.indent--

        case *ast.ExprStmt:
                f.writeIndent()
                f.formatExpr(s.Expr)
                f.write("\n")

        case *ast.AssertStmt:
                f.writeIndent()
                f.write("assert ")
                f.formatExpr(s.Condition)
                if s.Message != "" {
                        f.write(fmt.Sprintf(", %q", s.Message))
                }
                f.write("\n")

        case *ast.BreakStmt:
                f.writeLine("break")

        case *ast.ContinueStmt:
                f.writeLine("continue")

        case *ast.WithStmt:
                f.writeIndent()
                f.write("with ")
                for i, b := range s.Bindings {
                        if i > 0 {
                                f.write(", ")
                        }
                        f.formatExpr(b.Expr)
                        if b.Alias != "" {
                                f.write(" as " + b.Alias)
                        }
                }
                f.writeln(":")
                f.indent++
                for _, stmt := range s.Body {
                        f.formatStatement(stmt)
                }
                f.indent--
        }
}

// --- Patterns ---

func (f *Formatter) formatPattern(p ast.Pattern) {
        switch pat := p.(type) {
        case *ast.WildcardPattern:
                f.write("_")
        case *ast.BindingPattern:
                f.write(pat.Name)
        case *ast.LiteralPattern:
                if pat.Kind == token.STRING_LIT {
                        f.write(fmt.Sprintf("%q", pat.Value))
                } else {
                        f.write(pat.Value)
                }
        case *ast.ConstructorPattern:
                f.write(pat.TypeName)
                if len(pat.Fields) > 0 {
                        f.write("(")
                        for i, fp := range pat.Fields {
                                if i > 0 {
                                        f.write(", ")
                                }
                                f.formatPattern(fp)
                        }
                        f.write(")")
                }
        case *ast.ListPattern:
                f.write("[")
                for i, e := range pat.Elements {
                        if i > 0 {
                                f.write(", ")
                        }
                        f.formatPattern(e)
                }
                f.write("]")
        case *ast.TuplePattern:
                f.write("(")
                for i, e := range pat.Elements {
                        if i > 0 {
                                f.write(", ")
                        }
                        f.formatPattern(e)
                }
                f.write(")")
        }
}

// --- Expressions ---

func (f *Formatter) formatExpr(expr ast.Expr) {
        switch e := expr.(type) {
        case *ast.Identifier:
                f.write(e.Name)

        case *ast.IntLiteral:
                f.write(e.Value)

        case *ast.FloatLiteral:
                f.write(e.Value)

        case *ast.StringLiteral:
                f.write(fmt.Sprintf("%q", e.Value))

        case *ast.BoolLiteral:
                if e.Value {
                        f.write("true")
                } else {
                        f.write("false")
                }

        case *ast.NoneLiteral:
                f.write("none")

        case *ast.BinaryOp:
                f.formatExpr(e.Left)
                f.write(" " + e.Op + " ")
                f.formatExpr(e.Right)

        case *ast.UnaryOp:
                if e.Op == "!" {
                        f.formatExpr(e.Operand)
                        f.write("!")
                } else {
                        f.write(e.Op)
                        if e.Op == "not" {
                                f.write(" ")
                        }
                        f.formatExpr(e.Operand)
                }

        case *ast.CallExpr:
                f.formatExpr(e.Callee)
                f.write("(")
                for i, arg := range e.Args {
                        if i > 0 {
                                f.write(", ")
                        }
                        if arg.Name != "" {
                                f.write(arg.Name + ": ")
                        }
                        f.formatExpr(arg.Value)
                }
                f.write(")")

        case *ast.FieldAccess:
                f.formatExpr(e.Object)
                f.write("." + e.Field)

        case *ast.OptionalFieldAccess:
                f.formatExpr(e.Object)
                f.write("?." + e.Field)

        case *ast.PipelineExpr:
                f.formatExpr(e.Left)
                f.write(" |> ")
                f.formatExpr(e.Right)

        case *ast.IndexExpr:
                f.formatExpr(e.Object)
                f.write("[")
                f.formatExpr(e.Index)
                f.write("]")

        case *ast.OptionPropagate:
                f.formatExpr(e.Expr)
                f.write("?")

        case *ast.ListExpr:
                f.write("[")
                for i, el := range e.Elements {
                        if i > 0 {
                                f.write(", ")
                        }
                        f.formatExpr(el)
                }
                f.write("]")

        case *ast.ListComp:
                f.write("[")
                f.formatExpr(e.Element)
                f.write(" for " + e.Variable + " in ")
                f.formatExpr(e.Iterable)
                if e.Filter != nil {
                        f.write(" if ")
                        f.formatExpr(e.Filter)
                }
                f.write("]")

        case *ast.MapExpr:
                f.write("{")
                for i, entry := range e.Entries {
                        if i > 0 {
                                f.write(", ")
                        }
                        f.formatExpr(entry.Key)
                        f.write(": ")
                        f.formatExpr(entry.Value)
                }
                f.write("}")

        case *ast.StructExpr:
                f.write(e.TypeName + "(")
                // Check if we should format multi-line
                if len(e.Fields) > 3 {
                        f.write("\n")
                        f.indent++
                        for i, fi := range e.Fields {
                                f.writeIndent()
                                f.write(fi.Name + ": ")
                                f.formatExpr(fi.Value)
                                if i < len(e.Fields)-1 {
                                        f.write(",")
                                } else {
                                        f.write(",")
                                }
                                f.write("\n")
                        }
                        f.indent--
                        f.writeIndent()
                } else {
                        for i, fi := range e.Fields {
                                if i > 0 {
                                        f.write(", ")
                                }
                                f.write(fi.Name + ": ")
                                f.formatExpr(fi.Value)
                        }
                }
                f.write(")")

        case *ast.IfExpr:
                f.write("if ")
                f.formatExpr(e.Condition)
                f.write(" then ")
                f.formatExpr(e.ThenExpr)
                f.write(" else ")
                f.formatExpr(e.ElseExpr)

        case *ast.Lambda:
                f.write("|")
                for i, p := range e.Params {
                        if i > 0 {
                                f.write(", ")
                        }
                        f.formatParam(p)
                }
                f.write("|")
                if e.Body != nil {
                        f.write(" -> ")
                        f.formatExpr(e.Body)
                }

        default:
                f.write("_unknown_expr_")
        }
}
