// Package docgen extracts documentation from parsed Aura ASTs and renders
// it as Markdown or structured JSON.
package docgen

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/unclebucklarson/aura/pkg/ast"
)

// --- Public data model ---

// DocPage holds all extracted documentation for one Aura module.
type DocPage struct {
	Module    string      `json:"module"`
	Functions []*FnDoc    `json:"functions,omitempty"`
	Types     []*TypeDoc  `json:"types,omitempty"`
	Structs   []*StructDoc `json:"structs,omitempty"`
	Enums     []*EnumDoc  `json:"enums,omitempty"`
	Traits    []*TraitDoc `json:"traits,omitempty"`
	Specs     []*SpecDoc  `json:"specs,omitempty"`
}

type FnDoc struct {
	Signature   string   `json:"signature"`
	Doc         string   `json:"doc,omitempty"`
	Effects     []string `json:"effects,omitempty"`
	Constraints []string `json:"constraints,omitempty"`
	Satisfies   string   `json:"satisfies,omitempty"`
}

type TypeDoc struct {
	Name string `json:"name"`
	Body string `json:"body"`
	Doc  string `json:"doc,omitempty"`
}

type StructDoc struct {
	Name       string       `json:"name"`
	TypeParams []string     `json:"type_params,omitempty"`
	Fields     []*FieldDoc  `json:"fields"`
	Methods    []*FnDoc     `json:"methods,omitempty"`
	Doc        string       `json:"doc,omitempty"`
}

type FieldDoc struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type EnumDoc struct {
	Name       string        `json:"name"`
	TypeParams []string      `json:"type_params,omitempty"`
	Variants   []*VariantDoc `json:"variants"`
	Doc        string        `json:"doc,omitempty"`
}

type VariantDoc struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields,omitempty"`
}

type TraitDoc struct {
	Name    string   `json:"name"`
	Methods []string `json:"methods"`
	Doc     string   `json:"doc,omitempty"`
}

type SpecDoc struct {
	Name       string          `json:"name"`
	Doc        string          `json:"doc,omitempty"`
	Inputs     []*SpecInputDoc `json:"inputs,omitempty"`
	Guarantees []string        `json:"guarantees,omitempty"`
	Effects    []string        `json:"effects,omitempty"`
	Errors     []*SpecErrorDoc `json:"errors,omitempty"`
}

type SpecInputDoc struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type SpecErrorDoc struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// --- Generator ---

// Generate extracts documentation from a parsed module.
// Only pub declarations are included; private items are silently skipped.
func Generate(mod *ast.Module) *DocPage {
	page := &DocPage{Module: mod.Name.String()}

	// Collect impl methods per type name so they can be attached to StructDoc.
	implMethods := map[string][]*FnDoc{}
	for _, item := range mod.Items {
		if impl, ok := item.(*ast.ImplBlock); ok && impl.TraitName == "" {
			typeName := typeExprString(impl.TargetType)
			for _, m := range impl.Methods {
				if m.Visibility == ast.Public {
					implMethods[typeName] = append(implMethods[typeName], extractFnDoc(m))
				}
			}
		}
	}

	for _, item := range mod.Items {
		switch n := item.(type) {
		case *ast.FnDef:
			if n.Visibility == ast.Public {
				page.Functions = append(page.Functions, extractFnDoc(n))
			}
		case *ast.TypeDef:
			if n.Visibility == ast.Public {
				page.Types = append(page.Types, &TypeDoc{
					Name: typeDefName(n),
					Body: typeExprString(n.Body),
					Doc:  docComment(n.Comments),
				})
			}
		case *ast.StructDef:
			if n.Visibility == ast.Public {
				sd := &StructDoc{
					Name:       n.Name,
					TypeParams: n.TypeParams,
					Doc:        docComment(n.Comments),
					Methods:    implMethods[n.Name],
				}
				for _, f := range n.Fields {
					sd.Fields = append(sd.Fields, &FieldDoc{
						Name: f.Name,
						Type: typeExprString(f.TypeExpr),
					})
				}
				page.Structs = append(page.Structs, sd)
			}
		case *ast.EnumDef:
			if n.Visibility == ast.Public {
				ed := &EnumDoc{
					Name:       n.Name,
					TypeParams: n.TypeParams,
					Doc:        docComment(n.Comments),
				}
				for _, v := range n.Variants {
					vd := &VariantDoc{Name: v.Name}
					for _, f := range v.Fields {
						vd.Fields = append(vd.Fields, typeExprString(f))
					}
					ed.Variants = append(ed.Variants, vd)
				}
				page.Enums = append(page.Enums, ed)
			}
		case *ast.TraitDef:
			if n.Visibility == ast.Public {
				td := &TraitDoc{
					Name: n.Name,
					Doc:  docComment(n.Comments),
				}
				for _, m := range n.Members {
					if sig, ok := m.(*ast.FnSignature); ok {
						td.Methods = append(td.Methods, fnSignatureString(sig))
					} else if fn, ok := m.(*ast.FnDef); ok {
						td.Methods = append(td.Methods, fnDefSignature(fn)+" (default)")
					}
				}
				page.Traits = append(page.Traits, td)
			}
		case *ast.SpecBlock:
			sd := &SpecDoc{
				Name:    n.Name,
				Doc:     n.Doc,
				Effects: n.Effects,
			}
			for _, inp := range n.Inputs {
				sd.Inputs = append(sd.Inputs, &SpecInputDoc{
					Name:        inp.Name,
					Type:        typeExprString(inp.TypeExpr),
					Description: inp.Description,
				})
			}
			for _, g := range n.Guarantees {
				sd.Guarantees = append(sd.Guarantees, g.Condition)
			}
			for _, e := range n.Errors {
				sd.Errors = append(sd.Errors, &SpecErrorDoc{
					Name:        e.TypeName,
					Description: e.Description,
				})
			}
			page.Specs = append(page.Specs, sd)
		}
	}

	return page
}

// --- Renderers ---

// Markdown renders the DocPage as GitHub-flavoured Markdown.
func (p *DocPage) Markdown() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s\n", p.Module)

	if len(p.Functions) > 0 {
		sb.WriteString("\n## Functions\n")
		for _, f := range p.Functions {
			fmt.Fprintf(&sb, "\n### `%s`\n", f.Signature)
			if f.Doc != "" {
				fmt.Fprintf(&sb, "\n%s\n", f.Doc)
			}
			if len(f.Effects) > 0 {
				fmt.Fprintf(&sb, "\n**Effects:** %s\n", strings.Join(f.Effects, ", "))
			}
			if len(f.Constraints) > 0 {
				fmt.Fprintf(&sb, "\n**Constraints:** %s\n", strings.Join(f.Constraints, ", "))
			}
			if f.Satisfies != "" {
				fmt.Fprintf(&sb, "\n**Satisfies:** `%s`\n", f.Satisfies)
			}
		}
	}

	if len(p.Types) > 0 {
		sb.WriteString("\n## Types\n")
		for _, t := range p.Types {
			fmt.Fprintf(&sb, "\n### `type %s = %s`\n", t.Name, t.Body)
			if t.Doc != "" {
				fmt.Fprintf(&sb, "\n%s\n", t.Doc)
			}
		}
	}

	if len(p.Structs) > 0 {
		sb.WriteString("\n## Structs\n")
		for _, s := range p.Structs {
			name := s.Name
			if len(s.TypeParams) > 0 {
				name += "[" + strings.Join(s.TypeParams, ", ") + "]"
			}
			fmt.Fprintf(&sb, "\n### `struct %s`\n", name)
			if s.Doc != "" {
				fmt.Fprintf(&sb, "\n%s\n", s.Doc)
			}
			if len(s.Fields) > 0 {
				sb.WriteString("\n**Fields:**\n")
				for _, f := range s.Fields {
					fmt.Fprintf(&sb, "- `%s: %s`\n", f.Name, f.Type)
				}
			}
			if len(s.Methods) > 0 {
				sb.WriteString("\n**Methods:**\n")
				for _, m := range s.Methods {
					fmt.Fprintf(&sb, "- `%s`\n", m.Signature)
					if m.Doc != "" {
						fmt.Fprintf(&sb, "  %s\n", m.Doc)
					}
				}
			}
		}
	}

	if len(p.Enums) > 0 {
		sb.WriteString("\n## Enums\n")
		for _, e := range p.Enums {
			name := e.Name
			if len(e.TypeParams) > 0 {
				name += "[" + strings.Join(e.TypeParams, ", ") + "]"
			}
			fmt.Fprintf(&sb, "\n### `enum %s`\n", name)
			if e.Doc != "" {
				fmt.Fprintf(&sb, "\n%s\n", e.Doc)
			}
			if len(e.Variants) > 0 {
				sb.WriteString("\n**Variants:**\n")
				for _, v := range e.Variants {
					if len(v.Fields) == 0 {
						fmt.Fprintf(&sb, "- `%s`\n", v.Name)
					} else {
						fmt.Fprintf(&sb, "- `%s(%s)`\n", v.Name, strings.Join(v.Fields, ", "))
					}
				}
			}
		}
	}

	if len(p.Traits) > 0 {
		sb.WriteString("\n## Traits\n")
		for _, t := range p.Traits {
			fmt.Fprintf(&sb, "\n### `trait %s`\n", t.Name)
			if t.Doc != "" {
				fmt.Fprintf(&sb, "\n%s\n", t.Doc)
			}
			if len(t.Methods) > 0 {
				sb.WriteString("\n**Methods:**\n")
				for _, m := range t.Methods {
					fmt.Fprintf(&sb, "- `%s`\n", m)
				}
			}
		}
	}

	if len(p.Specs) > 0 {
		sb.WriteString("\n## Specs\n")
		for _, s := range p.Specs {
			fmt.Fprintf(&sb, "\n### `spec %s`\n", s.Name)
			if s.Doc != "" {
				fmt.Fprintf(&sb, "\n%s\n", s.Doc)
			}
			if len(s.Inputs) > 0 {
				sb.WriteString("\n**Inputs:**\n")
				for _, inp := range s.Inputs {
					if inp.Description != "" {
						fmt.Fprintf(&sb, "- `%s: %s` — %s\n", inp.Name, inp.Type, inp.Description)
					} else {
						fmt.Fprintf(&sb, "- `%s: %s`\n", inp.Name, inp.Type)
					}
				}
			}
			if len(s.Guarantees) > 0 {
				sb.WriteString("\n**Guarantees:**\n")
				for _, g := range s.Guarantees {
					fmt.Fprintf(&sb, "- %s\n", g)
				}
			}
			if len(s.Effects) > 0 {
				fmt.Fprintf(&sb, "\n**Effects:** %s\n", strings.Join(s.Effects, ", "))
			}
			if len(s.Errors) > 0 {
				sb.WriteString("\n**Errors:**\n")
				for _, e := range s.Errors {
					if e.Description != "" {
						fmt.Fprintf(&sb, "- `%s` — %s\n", e.Name, e.Description)
					} else {
						fmt.Fprintf(&sb, "- `%s`\n", e.Name)
					}
				}
			}
		}
	}

	return sb.String()
}

// JSON renders the DocPage as indented JSON.
// HTML escaping is disabled so -> is preserved as-is (not \u003e).
func (p *DocPage) JSON() string {
	var buf strings.Builder
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	_ = enc.Encode(p)
	return strings.TrimRight(buf.String(), "\n")
}

// --- Helpers ---

func extractFnDoc(fn *ast.FnDef) *FnDoc {
	d := &FnDoc{
		Signature: fnDefSignature(fn),
		Doc:       docComment(fn.Comments),
		Effects:   fn.Effects,
		Satisfies: fn.Satisfies,
	}
	for _, c := range fn.Constraints {
		d.Constraints = append(d.Constraints, c.TypeParam+": "+c.TraitName)
	}
	return d
}

// fnDefSignature renders a FnDef as a signature string.
func fnDefSignature(fn *ast.FnDef) string {
	var sb strings.Builder
	sb.WriteString("fn ")
	sb.WriteString(fn.Name)
	if len(fn.TypeParams) > 0 {
		sb.WriteString("[")
		sb.WriteString(strings.Join(fn.TypeParams, ", "))
		sb.WriteString("]")
	}
	sb.WriteString("(")
	for i, p := range fn.Params {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(p.Name)
		if p.TypeExpr != nil {
			sb.WriteString(": ")
			sb.WriteString(typeExprString(p.TypeExpr))
		}
	}
	sb.WriteString(")")
	if fn.ReturnType != nil {
		sb.WriteString(" -> ")
		sb.WriteString(typeExprString(fn.ReturnType))
	}
	return sb.String()
}

// fnSignatureString renders a FnSignature (trait method) as a string.
func fnSignatureString(sig *ast.FnSignature) string {
	var sb strings.Builder
	sb.WriteString("fn ")
	sb.WriteString(sig.Name)
	if len(sig.TypeParams) > 0 {
		sb.WriteString("[")
		sb.WriteString(strings.Join(sig.TypeParams, ", "))
		sb.WriteString("]")
	}
	sb.WriteString("(")
	for i, p := range sig.Params {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(p.Name)
		if p.TypeExpr != nil {
			sb.WriteString(": ")
			sb.WriteString(typeExprString(p.TypeExpr))
		}
	}
	sb.WriteString(")")
	if sig.ReturnType != nil {
		sb.WriteString(" -> ")
		sb.WriteString(typeExprString(sig.ReturnType))
	}
	return sb.String()
}

// typeDefName renders a TypeDef name including type params.
func typeDefName(n *ast.TypeDef) string {
	if len(n.TypeParams) == 0 {
		return n.Name
	}
	return n.Name + "[" + strings.Join(n.TypeParams, ", ") + "]"
}

// docComment extracts the text of the first doc comment (##) in a comment list.
func docComment(comments []ast.Comment) string {
	var parts []string
	for _, c := range comments {
		if c.IsDoc {
			parts = append(parts, strings.TrimSpace(c.Text))
		}
	}
	return strings.Join(parts, "\n")
}

// typeExprString renders an ast.TypeExpr back to its source representation.
func typeExprString(te ast.TypeExpr) string {
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
			args[i] = typeExprString(a)
		}
		return t.Name + "[" + strings.Join(args, ", ") + "]"
	case *ast.QualifiedType:
		base := t.Qualifier + "." + t.Name
		if len(t.Args) == 0 {
			return base
		}
		args := make([]string, len(t.Args))
		for i, a := range t.Args {
			args[i] = typeExprString(a)
		}
		return base + "[" + strings.Join(args, ", ") + "]"
	case *ast.ListType:
		return "[" + typeExprString(t.Element) + "]"
	case *ast.MapType:
		return "{" + typeExprString(t.Key) + ": " + typeExprString(t.Value) + "}"
	case *ast.SetType:
		return "{" + typeExprString(t.Element) + "}"
	case *ast.OptionType:
		return typeExprString(t.Inner) + "?"
	case *ast.UnionType:
		return typeExprString(t.Left) + " | " + typeExprString(t.Right)
	case *ast.IntersectionType:
		return typeExprString(t.Left) + " & " + typeExprString(t.Right)
	case *ast.TupleType:
		parts := make([]string, len(t.Elements))
		for i, e := range t.Elements {
			parts[i] = typeExprString(e)
		}
		return "(" + strings.Join(parts, ", ") + ")"
	case *ast.FunctionType:
		params := make([]string, len(t.Params))
		for i, p := range t.Params {
			params[i] = typeExprString(p)
		}
		return "fn(" + strings.Join(params, ", ") + ") -> " + typeExprString(t.ReturnType)
	case *ast.RefinementType:
		return typeExprString(t.Base) + " where ..."
	case *ast.StringLitType:
		return `"` + t.Value + `"`
	default:
		return "?"
	}
}
