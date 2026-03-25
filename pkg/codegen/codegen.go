// Package codegen implements the AI-assisted spec-to-implementation pipeline.
//
// Given an Aura source file containing spec blocks, codegen builds a
// structured prompt, calls the Anthropic Messages API, and validates the
// generated code with the type checker.
//
// No external dependencies — uses net/http and encoding/json from stdlib.
package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/unclebucklarson/aura/pkg/ast"
	"github.com/unclebucklarson/aura/pkg/checker"
	"github.com/unclebucklarson/aura/pkg/lexer"
	"github.com/unclebucklarson/aura/pkg/parser"
)

// DefaultModel is the Anthropic model used for code generation.
const DefaultModel = "claude-opus-4-6"

const anthropicMessagesURL = "https://api.anthropic.com/v1/messages"

// --- Context extraction ---

// ModuleContext holds type and function context extracted from a module.
// It is passed to BuildPrompt so the AI has the full picture.
type ModuleContext struct {
	ModuleName string
	Structs    []*ast.StructDef
	Enums      []*ast.EnumDef
	Types      []*ast.TypeDef
	Functions  []*ast.FnDef // existing function signatures (for context)
}

// ExtractContext builds a ModuleContext from a parsed module.
func ExtractContext(mod *ast.Module) *ModuleContext {
	ctx := &ModuleContext{ModuleName: mod.Name.String()}
	for _, item := range mod.Items {
		switch n := item.(type) {
		case *ast.StructDef:
			ctx.Structs = append(ctx.Structs, n)
		case *ast.EnumDef:
			ctx.Enums = append(ctx.Enums, n)
		case *ast.TypeDef:
			ctx.Types = append(ctx.Types, n)
		case *ast.FnDef:
			ctx.Functions = append(ctx.Functions, n)
		}
	}
	return ctx
}

// FindUnimplementedSpecs returns specs in mod that have no satisfying function.
func FindUnimplementedSpecs(mod *ast.Module) []*ast.SpecBlock {
	satisfied := map[string]bool{}
	for _, item := range mod.Items {
		if fn, ok := item.(*ast.FnDef); ok && fn.Satisfies != "" {
			satisfied[fn.Satisfies] = true
		}
	}
	var out []*ast.SpecBlock
	for _, item := range mod.Items {
		if spec, ok := item.(*ast.SpecBlock); ok && !satisfied[spec.Name] {
			out = append(out, spec)
		}
	}
	return out
}

// --- Prompt builder ---

// BuildPrompt constructs the full prompt for generating a function from a spec.
// The prompt includes: spec details, available types, existing functions,
// an Aura syntax reference, and a clear generation request.
func BuildPrompt(spec *ast.SpecBlock, ctx *ModuleContext) string {
	var sb strings.Builder

	sb.WriteString("You are implementing a function in the Aura programming language.\n")
	sb.WriteString("Output ONLY the Aura function implementation — no markdown fences, no explanations.\n\n")

	// Spec section
	fmt.Fprintf(&sb, "## Spec: %s\n\n", spec.Name)
	if spec.Doc != "" {
		fmt.Fprintf(&sb, "Description: %s\n\n", spec.Doc)
	}
	if len(spec.Inputs) > 0 {
		sb.WriteString("Inputs:\n")
		for _, inp := range spec.Inputs {
			line := fmt.Sprintf("  %s: %s", inp.Name, typeExprStr(inp.TypeExpr))
			if inp.Description != "" {
				line += " — " + inp.Description
			}
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}
	if len(spec.Guarantees) > 0 {
		sb.WriteString("Guarantees (postconditions the implementation must satisfy):\n")
		for _, g := range spec.Guarantees {
			fmt.Fprintf(&sb, "  - %s\n", g.Condition)
		}
		sb.WriteString("\n")
	}
	if len(spec.Effects) > 0 {
		fmt.Fprintf(&sb, "Permitted effects: %s\n\n", strings.Join(spec.Effects, ", "))
	}
	if len(spec.Errors) > 0 {
		sb.WriteString("Error conditions:\n")
		for _, e := range spec.Errors {
			line := fmt.Sprintf("  %s", e.TypeName)
			if e.Description != "" {
				line += " — " + e.Description
			}
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}

	// Type context
	hasTypes := len(ctx.Types) > 0 || len(ctx.Structs) > 0 || len(ctx.Enums) > 0
	if hasTypes {
		sb.WriteString("## Types available in this module:\n\n")
		for _, t := range ctx.Types {
			fmt.Fprintf(&sb, "type %s = %s\n", typeDefName(t), typeExprStr(t.Body))
		}
		for _, s := range ctx.Structs {
			name := s.Name
			if len(s.TypeParams) > 0 {
				name += "[" + strings.Join(s.TypeParams, ", ") + "]"
			}
			fmt.Fprintf(&sb, "\nstruct %s:\n", name)
			for _, f := range s.Fields {
				fmt.Fprintf(&sb, "    %s: %s\n", f.Name, typeExprStr(f.TypeExpr))
			}
		}
		for _, e := range ctx.Enums {
			name := e.Name
			if len(e.TypeParams) > 0 {
				name += "[" + strings.Join(e.TypeParams, ", ") + "]"
			}
			fmt.Fprintf(&sb, "\nenum %s:\n", name)
			for _, v := range e.Variants {
				if len(v.Fields) == 0 {
					fmt.Fprintf(&sb, "    %s\n", v.Name)
				} else {
					fs := make([]string, len(v.Fields))
					for i, f := range v.Fields {
						fs[i] = typeExprStr(f)
					}
					fmt.Fprintf(&sb, "    %s(%s)\n", v.Name, strings.Join(fs, ", "))
				}
			}
		}
		sb.WriteString("\n")
	}

	// Existing functions (context only)
	if len(ctx.Functions) > 0 {
		sb.WriteString("## Existing functions (context only — do not re-implement):\n\n")
		for _, fn := range ctx.Functions {
			sb.WriteString(fnSigStr(fn) + "\n")
		}
		sb.WriteString("\n")
	}

	// Aura syntax cheat sheet
	sb.WriteString("## Aura language rules:\n\n")
	sb.WriteString("- Function declaration: `[pub] fn name[TypeParams](param: Type) -> ReturnType [with effect1, effect2] [satisfies SpecName]:`\n")
	sb.WriteString("- Indented body with `return value`\n")
	sb.WriteString("- Result type: `Ok(value)` or `Err(ErrorVariant(\"message\"))`\n")
	sb.WriteString("- Option type: `Some(value)` or `None`\n")
	sb.WriteString("- String interpolation: `\"text {variable} more\"`\n")
	sb.WriteString("- Effect capabilities: `db.insert(table, value)`, `time.now()`, `log.info(msg)`, `net.get(url)`, `env.get(key)`\n")
	sb.WriteString("- Let: `let [mut] name [: Type] = expr`\n")
	sb.WriteString("- If/elif/else, for variable in iterable:, while condition: — all use indented blocks\n")
	sb.WriteString("- Match: `match expr:` with `case Pattern:` arms\n")
	sb.WriteString("- Struct construction: `StructName(field1: value1, field2: value2)`\n\n")

	// Task
	fmt.Fprintf(&sb, "## Task:\n\nImplement the `%s` spec. ", spec.Name)
	sb.WriteString("Start with the function declaration line (`pub fn` or `fn`), ")
	sb.WriteString("use `satisfies " + spec.Name + "` in the declaration, ")
	sb.WriteString("then provide the indented body.\n")
	sb.WriteString("Output ONLY valid Aura code — no prose, no markdown.\n")

	return sb.String()
}

// --- Generation result ---

// Result holds the output of a single spec generation.
type Result struct {
	SpecName  string   `json:"spec"`
	Prompt    string   `json:"prompt,omitempty"` // only in dry-run / verbose mode
	Generated string   `json:"generated,omitempty"`
	Valid     bool     `json:"valid"`
	Errors    []string `json:"errors,omitempty"`
}

// --- API client ---

type apiRequest struct {
	Model     string       `json:"model"`
	MaxTokens int          `json:"max_tokens"`
	Messages  []apiMessage `json:"messages"`
}

type apiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type apiResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Generate calls the Anthropic API to generate a function implementation
// for the given spec. Returns the raw generated code string.
func Generate(spec *ast.SpecBlock, ctx *ModuleContext, apiKey, model string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}
	if model == "" {
		model = DefaultModel
	}

	prompt := BuildPrompt(spec, ctx)

	reqBody, err := json.Marshal(apiRequest{
		Model:     model,
		MaxTokens: 2048,
		Messages:  []apiMessage{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, anthropicMessagesURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API call: %w", err)
	}
	defer resp.Body.Close()

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if apiResp.Error != nil {
		return "", fmt.Errorf("API error (%s): %s", apiResp.Error.Type, apiResp.Error.Message)
	}
	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("API returned empty content")
	}

	code := strings.TrimSpace(apiResp.Content[0].Text)
	// Strip markdown fences if the model added them despite instructions.
	code = stripFences(code)
	return code, nil
}

// --- Validation ---

// Validate appends the generated code to the original source and runs the
// type checker. Returns the list of checker error messages (nil = valid).
func Validate(originalSrc, generatedCode, file string) []string {
	combined := originalSrc + "\n\n" + generatedCode + "\n"

	l := lexer.New(combined, file)
	tokens, lexErrs := l.Tokenize()
	if len(lexErrs) > 0 {
		msgs := make([]string, len(lexErrs))
		for i, e := range lexErrs {
			msgs[i] = e.Error()
		}
		return msgs
	}

	p := parser.New(tokens, file)
	mod, parseErrs := p.Parse()
	if len(parseErrs) > 0 {
		msgs := make([]string, len(parseErrs))
		for i, e := range parseErrs {
			msgs[i] = e.Error()
		}
		return msgs
	}

	c := checker.New(mod)
	errs := c.Check()
	if len(errs) == 0 {
		return nil
	}
	msgs := make([]string, len(errs))
	for i, e := range errs {
		msgs[i] = e.Error()
	}
	return msgs
}

// --- Helpers ---

// stripFences removes ```aura / ``` markdown fences if present.
func stripFences(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "```") {
			continue
		}
		out = append(out, l)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
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
		base := t.Qualifier + "." + t.Name
		if len(t.Args) == 0 {
			return base
		}
		args := make([]string, len(t.Args))
		for i, a := range t.Args {
			args[i] = typeExprStr(a)
		}
		return base + "[" + strings.Join(args, ", ") + "]"
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
	case *ast.IntersectionType:
		return typeExprStr(t.Left) + " & " + typeExprStr(t.Right)
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

// typeDefName renders a TypeDef name with type params.
func typeDefName(n *ast.TypeDef) string {
	if len(n.TypeParams) == 0 {
		return n.Name
	}
	return n.Name + "[" + strings.Join(n.TypeParams, ", ") + "]"
}

// fnSigStr renders a FnDef's signature (no body).
func fnSigStr(fn *ast.FnDef) string {
	var sb strings.Builder
	if fn.Visibility == ast.Public {
		sb.WriteString("pub ")
	}
	sb.WriteString("fn ")
	sb.WriteString(fn.Name)
	if len(fn.TypeParams) > 0 {
		sb.WriteString("[" + strings.Join(fn.TypeParams, ", ") + "]")
	}
	sb.WriteString("(")
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
	if fn.Satisfies != "" {
		sb.WriteString(" satisfies " + fn.Satisfies)
	}
	return sb.String()
}
