package lsp

import (
	"strings"

	"github.com/unclebucklarson/aura/pkg/ast"
	"github.com/unclebucklarson/aura/pkg/lexer"
	"github.com/unclebucklarson/aura/pkg/parser"
	"github.com/unclebucklarson/aura/pkg/token"
)

// computeHover returns hover information for the identifier at pos in src.
// Returns nil if no meaningful hover info is available.
func computeHover(src, filePath string, pos Position) *Hover {
	mod := parseOnly(src, filePath)
	if mod == nil {
		return nil
	}
	word, wordRange := wordAt(src, pos)
	if word == "" {
		return nil
	}
	// Search top-level definitions for a matching name.
	for _, item := range mod.Items {
		switch it := item.(type) {
		case *ast.FnDef:
			if it.Name == word {
				content := fnHoverContent(it)
				return &Hover{Contents: MarkupContent{Kind: "markdown", Value: content}, Range: &wordRange}
			}
		case *ast.StructDef:
			if it.Name == word {
				return &Hover{
					Contents: MarkupContent{Kind: "markdown", Value: "```aura\nstruct " + it.Name + "\n```"},
					Range:    &wordRange,
				}
			}
		case *ast.EnumDef:
			if it.Name == word {
				return &Hover{
					Contents: MarkupContent{Kind: "markdown", Value: "```aura\nenum " + it.Name + "\n```"},
					Range:    &wordRange,
				}
			}
		case *ast.TypeDef:
			if it.Name == word {
				body := ""
				if it.Body != nil {
					body = " = " + typeExprStr(it.Body)
				}
				return &Hover{
					Contents: MarkupContent{Kind: "markdown", Value: "```aura\ntype " + it.Name + body + "\n```"},
					Range:    &wordRange,
				}
			}
		}
	}
	return nil
}

// computeDefinition returns the Location of the definition for the identifier
// at pos in src, or nil if not found.
func computeDefinition(src, filePath string, pos Position) *Location {
	mod := parseOnly(src, filePath)
	if mod == nil {
		return nil
	}
	word, _ := wordAt(src, pos)
	if word == "" {
		return nil
	}
	for _, item := range mod.Items {
		switch it := item.(type) {
		case *ast.FnDef:
			if it.Name == word {
				return spanToLocation(it.Span, filePath)
			}
		case *ast.StructDef:
			if it.Name == word {
				return spanToLocation(it.Span, filePath)
			}
		case *ast.EnumDef:
			if it.Name == word {
				return spanToLocation(it.Span, filePath)
			}
		case *ast.TypeDef:
			if it.Name == word {
				return spanToLocation(it.Span, filePath)
			}
		case *ast.ConstDef:
			if it.Name == word {
				return spanToLocation(it.GetSpan(), filePath)
			}
		}
	}
	return nil
}

// parseOnly lexes and parses src, returning the AST or nil on error.
func parseOnly(src, filePath string) *ast.Module {
	l := lexer.New(src, filePath)
	tokens, lexErrs := l.Tokenize()
	if len(lexErrs) > 0 {
		return nil
	}
	p := parser.New(tokens, filePath)
	mod, parseErrs := p.Parse()
	if len(parseErrs) > 0 {
		return nil
	}
	return mod
}

// wordAt extracts the identifier word at the given cursor position from src
// and returns the word and its Range (0-based).
func wordAt(src string, pos Position) (string, Range) {
	lines := strings.Split(src, "\n")
	if pos.Line >= len(lines) {
		return "", Range{}
	}
	line := lines[pos.Line]
	col := pos.Character
	if col > len(line) {
		col = len(line)
	}
	start := col
	for start > 0 && isIdentChar(line[start-1]) {
		start--
	}
	end := col
	for end < len(line) && isIdentChar(line[end]) {
		end++
	}
	if start == end {
		return "", Range{}
	}
	word := line[start:end]
	r := Range{
		Start: Position{Line: pos.Line, Character: start},
		End:   Position{Line: pos.Line, Character: end},
	}
	return word, r
}

// isIdentChar returns true for characters valid inside an Aura identifier.
func isIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') || b == '_'
}

// spanToLocation converts a token.Span to an LSP Location.
func spanToLocation(span token.Span, filePath string) *Location {
	return &Location{
		URI: pathToURI(filePath),
		Range: Range{
			Start: Position{Line: span.Start.Line - 1, Character: span.Start.Column - 1},
			End:   Position{Line: span.End.Line - 1, Character: span.End.Column - 1},
		},
	}
}

// fnHoverContent builds a markdown hover string for a function definition.
func fnHoverContent(fn *ast.FnDef) string {
	var sb strings.Builder
	sb.WriteString("```aura\n")
	if fn.Visibility == ast.Public {
		sb.WriteString("pub ")
	}
	sb.WriteString("fn " + fn.Name + "(")
	for i, p := range fn.Params {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(p.Name)
		if p.TypeExpr != nil {
			sb.WriteString(": " + typeExprStr(p.TypeExpr))
		}
	}
	sb.WriteString(")")
	if fn.ReturnType != nil {
		sb.WriteString(" -> " + typeExprStr(fn.ReturnType))
	}
	if len(fn.Effects) > 0 {
		sb.WriteString(" with " + strings.Join(fn.Effects, ", "))
	}
	sb.WriteString("\n```")
	for _, c := range fn.Comments {
		if c.IsDoc {
			sb.WriteString("\n\n" + strings.TrimPrefix(strings.TrimSpace(c.Text), "##"))
			break
		}
	}
	return sb.String()
}

// typeExprStr renders an ast.TypeExpr as a source string.
func typeExprStr(te ast.TypeExpr) string {
	if te == nil {
		return ""
	}
	switch t := te.(type) {
	case *ast.NamedType:
		if len(t.Args) == 0 {
			return t.Name
		}
		args := make([]string, len(t.Args))
		for i, a := range t.Args {
			args[i] = typeExprStr(a)
		}
		return t.Name + "[" + strings.Join(args, ", ") + "]"
	case *ast.QualifiedType:
		return t.Qualifier + "." + t.Name
	case *ast.ListType:
		return "[" + typeExprStr(t.Element) + "]"
	case *ast.MapType:
		return "{" + typeExprStr(t.Key) + ": " + typeExprStr(t.Value) + "}"
	case *ast.SetType:
		return "{" + typeExprStr(t.Element) + "}"
	case *ast.OptionType:
		return typeExprStr(t.Inner) + "?"
	case *ast.UnionType:
		return typeExprStr(t.Left) + " | " + typeExprStr(t.Right)
	case *ast.TupleType:
		parts := make([]string, len(t.Elements))
		for i, e := range t.Elements {
			parts[i] = typeExprStr(e)
		}
		return "(" + strings.Join(parts, ", ") + ")"
	case *ast.FunctionType:
		params := make([]string, len(t.Params))
		for i, p := range t.Params {
			params[i] = typeExprStr(p)
		}
		return "fn(" + strings.Join(params, ", ") + ") -> " + typeExprStr(t.ReturnType)
	case *ast.RefinementType:
		return typeExprStr(t.Base) + " where ..."
	case *ast.StringLitType:
		return `"` + t.Value + `"`
	default:
		return "?"
	}
}
