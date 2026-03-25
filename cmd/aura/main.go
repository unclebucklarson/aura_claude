// Package main provides the CLI entry point for the Aura toolchain.
//
// Usage:
//
//	aura format   <file.aura>                    - Parse and reformat an Aura source file
//	aura parse    <file.aura>                    - Parse and display the AST (for debugging)
//	aura check    [--json] <file.aura>           - Type-check an Aura source file
//	aura run      <file.aura>                    - Execute an Aura program
//	aura test     <file.aura>                    - Run test blocks in an Aura file
//	aura doc      [--json] <file.aura>           - Generate documentation from an Aura source file
//	aura generate [--dry-run] [--json] <file>    - Generate implementations for unimplemented specs
//	aura init     [name]                         - Create a new aura.pkg manifest in the current directory
//	aura add      <alias> <path>                 - Add a local package dependency
//	aura build                                   - Verify all dependencies resolve
//	aura repl                                    - Interactive REPL
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/unclebucklarson/aura/pkg/ast"
	"github.com/unclebucklarson/aura/pkg/checker"
	"github.com/unclebucklarson/aura/pkg/codegen"
	"github.com/unclebucklarson/aura/pkg/docgen"
	"github.com/unclebucklarson/aura/pkg/formatter"
	"github.com/unclebucklarson/aura/pkg/interpreter"
	"github.com/unclebucklarson/aura/pkg/lexer"
	"github.com/unclebucklarson/aura/pkg/module"
	"github.com/unclebucklarson/aura/pkg/parser"
	"github.com/unclebucklarson/aura/pkg/pkgmgr"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		name := ""
		if len(os.Args) >= 3 {
			name = os.Args[2]
		}
		os.Exit(runInit(name))

	case "add":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "error: usage: aura add <alias> <path>")
			printUsage()
			os.Exit(1)
		}
		os.Exit(runAdd(os.Args[2], os.Args[3]))

	case "build":
		os.Exit(runBuild())

	case "repl":
		os.Exit(runRepl())

	case "doc":
		jsonOutput := false
		filePath := ""
		for _, arg := range os.Args[2:] {
			if arg == "--json" {
				jsonOutput = true
			} else {
				filePath = arg
			}
		}
		if filePath == "" {
			fmt.Fprintln(os.Stderr, "error: no file specified")
			printUsage()
			os.Exit(1)
		}
		src, err := readFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(runDoc(src, filePath, jsonOutput))

	case "generate":
		dryRun := false
		jsonOutput := false
		filePath := ""
		for _, arg := range os.Args[2:] {
			switch arg {
			case "--dry-run":
				dryRun = true
			case "--json":
				jsonOutput = true
			default:
				filePath = arg
			}
		}
		if filePath == "" {
			fmt.Fprintln(os.Stderr, "error: no file specified")
			printUsage()
			os.Exit(1)
		}
		src, err := readFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(runGenerate(src, filePath, dryRun, jsonOutput))

	case "format", "parse", "check", "run", "test":
		if len(os.Args) < 3 && command != "check" {
			fmt.Fprintf(os.Stderr, "error: no file specified for '%s'\n", command)
			printUsage()
			os.Exit(1)
		}

		switch command {
		case "format":
			filePath := os.Args[2]
			src, err := readFile(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(runFormat(src, filePath))

		case "parse":
			filePath := os.Args[2]
			src, err := readFile(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(runParse(src, filePath))

		case "check":
			jsonOutput := false
			filePath := ""
			for _, arg := range os.Args[2:] {
				if arg == "--json" {
					jsonOutput = true
				} else {
					filePath = arg
				}
			}
			if filePath == "" {
				fmt.Fprintln(os.Stderr, "error: no file specified")
				printUsage()
				os.Exit(1)
			}
			src, err := readFile(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(runCheck(src, filePath, jsonOutput))

		case "run":
			filePath := os.Args[2]
			src, err := readFile(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(runRun(src, filePath))

		case "test":
			filePath := os.Args[2]
			src, err := readFile(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(runTest(src, filePath))
		}

	default:
		fmt.Fprintf(os.Stderr, "error: unknown command %q\n", command)
		printUsage()
		os.Exit(1)
	}
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read file %q: %v", path, err)
	}
	return string(data), nil
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: aura <command> [options] <file.aura>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  format    Parse and reformat an Aura source file")
	fmt.Fprintln(os.Stderr, "  parse     Parse and display the AST (for debugging)")
	fmt.Fprintln(os.Stderr, "  check     Type-check an Aura source file")
	fmt.Fprintln(os.Stderr, "  run       Execute an Aura program")
	fmt.Fprintln(os.Stderr, "  test      Run test blocks in an Aura file")
	fmt.Fprintln(os.Stderr, "  doc       Generate documentation from an Aura source file")
	fmt.Fprintln(os.Stderr, "  generate  Generate implementations for unimplemented specs (requires ANTHROPIC_API_KEY)")
	fmt.Fprintln(os.Stderr, "  init      Create a new aura.pkg manifest in the current directory")
	fmt.Fprintln(os.Stderr, "  add       Add a local package dependency: aura add <alias> <path>")
	fmt.Fprintln(os.Stderr, "  build     Verify all dependencies in aura.pkg resolve correctly")
	fmt.Fprintln(os.Stderr, "  repl      Interactive REPL")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  --json      Output in JSON format (AI-parseable)")
	fmt.Fprintln(os.Stderr, "  --dry-run   Print prompt without calling the API (generate only)")
}

func runDoc(src, file string, jsonOutput bool) int {
	mod, code := parseSource(src, file)
	if code != 0 {
		return code
	}
	page := docgen.Generate(mod)
	if jsonOutput {
		fmt.Println(page.JSON())
	} else {
		fmt.Print(page.Markdown())
	}
	return 0
}

func runGenerate(src, file string, dryRun, jsonOutput bool) int {
	mod, code := parseSource(src, file)
	if code != 0 {
		return code
	}

	specs := codegen.FindUnimplementedSpecs(mod)
	if len(specs) == 0 {
		if jsonOutput {
			fmt.Println("[]")
		} else {
			fmt.Println("No unimplemented specs found.")
		}
		return 0
	}

	ctx := codegen.ExtractContext(mod)
	apiKey := os.Getenv("ANTHROPIC_API_KEY")

	var results []*codegen.Result

	for _, spec := range specs {
		r := &codegen.Result{SpecName: spec.Name}

		if dryRun {
			r.Prompt = codegen.BuildPrompt(spec, ctx)
			r.Valid = true
		} else {
			generated, err := codegen.Generate(spec, ctx, apiKey, "")
			if err != nil {
				r.Errors = []string{err.Error()}
			} else {
				r.Generated = generated
				errs := codegen.Validate(src, generated, file)
				r.Valid = len(errs) == 0
				r.Errors = errs
			}
		}

		results = append(results, r)
	}

	if jsonOutput {
		printGenerateJSON(results)
	} else {
		printGenerateText(results, dryRun)
	}

	for _, r := range results {
		if !r.Valid || len(r.Errors) > 0 {
			return 1
		}
	}
	return 0
}

func printGenerateJSON(results []*codegen.Result) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	_ = enc.Encode(results)
}

func printGenerateText(results []*codegen.Result, dryRun bool) {
	for _, r := range results {
		fmt.Printf("=== Spec: %s ===\n", r.SpecName)
		if dryRun {
			fmt.Println("--- Prompt ---")
			fmt.Println(r.Prompt)
		} else if r.Generated != "" {
			fmt.Println("--- Generated ---")
			fmt.Println(r.Generated)
			if r.Valid {
				fmt.Println("✓ Validated successfully")
			} else {
				fmt.Printf("✗ Validation failed (%d error(s)):\n", len(r.Errors))
				for _, e := range r.Errors {
					fmt.Printf("  - %s\n", e)
				}
			}
		} else if len(r.Errors) > 0 {
			fmt.Printf("✗ Generation failed: %s\n", r.Errors[0])
		}
		fmt.Println()
	}
}

func runInit(name string) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if name == "" {
		name = filepath.Base(cwd)
	}
	if err := pkgmgr.Init(cwd, name); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Printf("Created aura.pkg for package %q\n", name)
	return 0
}

func runAdd(alias, depPath string) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	manifestPath, err := pkgmgr.Find(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if manifestPath == "" {
		fmt.Fprintln(os.Stderr, "error: no aura.pkg found; run 'aura init' first")
		return 1
	}
	m, err := pkgmgr.Load(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if err := pkgmgr.AddDep(m, alias, depPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if err := pkgmgr.Write(m); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	abs, _ := filepath.Abs(depPath)
	fmt.Printf("Added dependency %q -> %s\n", alias, abs)
	return 0
}

func runBuild() int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	m, err := pkgmgr.FindAndLoad(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if m == nil {
		fmt.Fprintln(os.Stderr, "error: no aura.pkg found; run 'aura init' first")
		return 1
	}

	fmt.Printf("Package: %s  v%s\n", m.Name, m.Version)
	if len(m.Deps) == 0 {
		fmt.Println("Dependencies: none")
		fmt.Println("Build OK")
		return 0
	}

	fmt.Printf("Dependencies (%d):\n", len(m.Deps))
	ok := true
	for _, d := range m.Deps {
		fi, err := os.Stat(d.Path)
		if err != nil || !fi.IsDir() {
			fmt.Printf("  ✗ %s -> %s (not found)\n", d.Alias, d.Path)
			ok = false
		} else {
			fmt.Printf("  ✓ %s -> %s\n", d.Alias, d.Path)
		}
	}
	if ok {
		fmt.Println("Build OK")
		return 0
	}
	fmt.Fprintln(os.Stderr, "Build failed: one or more dependencies could not be resolved")
	return 1
}

func parseSource(src, file string) (*ast.Module, int) {
	l := lexer.New(src, file)
	tokens, lexErrors := l.Tokenize()
	if len(lexErrors) > 0 {
		for _, e := range lexErrors {
			fmt.Fprintf(os.Stderr, "%s:%s\n", file, e)
		}
		return nil, 1
	}

	p := parser.New(tokens, file)
	module, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		for _, e := range parseErrors {
			fmt.Fprintf(os.Stderr, "%s:%s\n", file, e)
		}
		return nil, 1
	}

	return module, 0
}

func runFormat(src, file string) int {
	module, code := parseSource(src, file)
	if code != 0 {
		return code
	}
	f := formatter.New()
	output := f.Format(module)
	fmt.Print(output)
	return 0
}

func runParse(src, file string) int {
	// Lex
	l := lexer.New(src, file)
	tokens, lexErrors := l.Tokenize()
	if len(lexErrors) > 0 {
		for _, e := range lexErrors {
			fmt.Fprintf(os.Stderr, "%s:%s\n", file, e)
		}
		return 1
	}

	// Print tokens
	fmt.Println("=== Tokens ===")
	for _, tok := range tokens {
		fmt.Printf("  %s\n", tok)
	}

	// Parse
	p := parser.New(tokens, file)
	module, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		fmt.Println("\n=== Parse Errors ===")
		for _, e := range parseErrors {
			fmt.Fprintf(os.Stderr, "%s:%s\n", file, e)
		}
		return 1
	}

	// Print AST
	fmt.Println("\n=== AST ===")
	printAST(module, 0)
	return 0
}

func runCheck(src, file string, jsonOutput bool) int {
	module, code := parseSource(src, file)
	if code != 0 {
		return code
	}

	c := checker.New(module)
	errs := c.Check()

	if len(errs) == 0 {
		if jsonOutput {
			fmt.Println(checker.FormatErrorsJSON(errs))
		} else {
			fmt.Println("✓ No type errors found.")
		}
		return 0
	}

	if jsonOutput {
		fmt.Println(checker.FormatErrorsJSON(errs))
	} else {
		fmt.Fprint(os.Stderr, checker.FormatErrors(errs))
	}

	for _, e := range errs {
		if e.Severity == checker.SeverityError {
			return 1
		}
	}
	return 0
}

func runRun(src, file string) int {
	mod, code := parseSource(src, file)
	if code != 0 {
		return code
	}

	// Resolve the absolute path of the source file for the module resolver
	absPath, pathErr := filepath.Abs(file)
	if pathErr != nil {
		fmt.Fprintf(os.Stderr, "error resolving path: %v\n", pathErr)
		return 1
	}
	baseDir := filepath.Dir(absPath)

	// Create module resolver and effect context for full stdlib/import support
	resolver := module.NewResolver(baseDir)
	// Auto-detect aura.pkg and inject dep search paths into resolver.
	if m, err := pkgmgr.FindAndLoad(baseDir); err == nil && m != nil {
		pkgmgr.ApplyToResolver(m, resolver)
	}
	effects := interpreter.NewEffectContext()

	interp := interpreter.NewWithResolverAndEffects(mod, absPath, resolver, effects)
	_, err := interp.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Run main function if it exists
	if mainFn, ok := interp.Env().Get("main"); ok {
		_ = mainFn
		_, err = interp.RunFunction("main", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
	}

	return 0
}

func runTest(src, file string) int {
	module, code := parseSource(src, file)
	if code != 0 {
		return code
	}

	results := interpreter.RunTests(module)
	fmt.Print(interpreter.FormatTestResults(results))

	for _, r := range results {
		if !r.Passed {
			return 1
		}
	}
	return 0
}


func printAST(node ast.Node, depth int) {
	indent := strings.Repeat("  ", depth)

	switch n := node.(type) {
	case *ast.Module:
		fmt.Printf("%sModule", indent)
		if n.Name != nil {
			fmt.Printf(" %s", n.Name.String())
		}
		fmt.Println()
		for _, imp := range n.Imports {
			printAST(imp, depth+1)
		}
		for _, item := range n.Items {
			printAST(item, depth+1)
		}

	case *ast.ImportNode:
		fmt.Printf("%sImport %s", indent, n.Path.String())
		if n.Alias != "" {
			fmt.Printf(" as %s", n.Alias)
		}
		if n.Names != nil {
			fmt.Printf(" names=%v", n.Names)
		}
		fmt.Println()

	case *ast.TypeDef:
		fmt.Printf("%sTypeDef %s%s = ", indent, visStr(n.Visibility), n.Name)
		printTypeExpr(n.Body)
		fmt.Println()

	case *ast.StructDef:
		fmt.Printf("%sStructDef %s%s\n", indent, visStr(n.Visibility), n.Name)
		for _, f := range n.Fields {
			fmt.Printf("%s  Field %s%s: ", indent, visStr(f.Visibility), f.Name)
			printTypeExpr(f.TypeExpr)
			if f.Default != nil {
				fmt.Printf(" = ...")
			}
			fmt.Println()
		}

	case *ast.EnumDef:
		fmt.Printf("%sEnumDef %s%s\n", indent, visStr(n.Visibility), n.Name)
		for _, v := range n.Variants {
			fmt.Printf("%s  Variant %s", indent, v.Name)
			if len(v.Fields) > 0 {
				fmt.Printf("(")
				for i, f := range v.Fields {
					if i > 0 {
						fmt.Printf(", ")
					}
					printTypeExpr(f)
				}
				fmt.Printf(")")
			}
			fmt.Println()
		}

	case *ast.SpecBlock:
		fmt.Printf("%sSpecBlock %s\n", indent, n.Name)
		if n.Doc != "" {
			fmt.Printf("%s  doc: %q\n", indent, n.Doc)
		}
		for _, inp := range n.Inputs {
			fmt.Printf("%s  input: %s: ", indent, inp.Name)
			printTypeExpr(inp.TypeExpr)
			fmt.Println()
		}
		for _, g := range n.Guarantees {
			fmt.Printf("%s  guarantee: %q\n", indent, g.Condition)
		}
		if len(n.Effects) > 0 {
			fmt.Printf("%s  effects: %v\n", indent, n.Effects)
		}
		for _, e := range n.Errors {
			fmt.Printf("%s  error: %s\n", indent, e.TypeName)
		}

	case *ast.FnDef:
		fmt.Printf("%sFnDef %s%s(", indent, visStr(n.Visibility), n.Name)
		for i, p := range n.Params {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", p.Name)
			if p.TypeExpr != nil {
				fmt.Printf(": ")
				printTypeExpr(p.TypeExpr)
			}
		}
		fmt.Printf(")")
		if n.ReturnType != nil {
			fmt.Printf(" -> ")
			printTypeExpr(n.ReturnType)
		}
		if len(n.Effects) > 0 {
			fmt.Printf(" with %s", strings.Join(n.Effects, ", "))
		}
		if n.Satisfies != "" {
			fmt.Printf(" satisfies %s", n.Satisfies)
		}
		fmt.Println()
		for _, stmt := range n.Body {
			printStmt(stmt, depth+1)
		}

	case *ast.ConstDef:
		fmt.Printf("%sConstDef %s%s\n", indent, visStr(n.Visibility), n.Name)

	case *ast.TestBlock:
		fmt.Printf("%sTestBlock %q\n", indent, n.Name)
		for _, stmt := range n.Body {
			printStmt(stmt, depth+1)
		}

	case *ast.TraitDef:
		fmt.Printf("%sTraitDef %s%s\n", indent, visStr(n.Visibility), n.Name)

	case *ast.ImplBlock:
		fmt.Printf("%sImplBlock", indent)
		if n.TraitName != "" {
			fmt.Printf(" %s for", n.TraitName)
		}
		fmt.Println()
	}
}

func printStmt(stmt ast.Statement, depth int) {
	indent := strings.Repeat("  ", depth)
	switch s := stmt.(type) {
	case *ast.LetStmt:
		fmt.Printf("%sLet %s\n", indent, s.Name)
	case *ast.ReturnStmt:
		fmt.Printf("%sReturn\n", indent)
	case *ast.IfStmt:
		fmt.Printf("%sIf\n", indent)
	case *ast.MatchStmt:
		fmt.Printf("%sMatch\n", indent)
	case *ast.ForStmt:
		fmt.Printf("%sFor %s\n", indent, s.Variable)
	case *ast.WhileStmt:
		fmt.Printf("%sWhile\n", indent)
	case *ast.ExprStmt:
		fmt.Printf("%sExprStmt\n", indent)
	case *ast.AssertStmt:
		fmt.Printf("%sAssert\n", indent)
	case *ast.BreakStmt:
		fmt.Printf("%sBreak\n", indent)
	case *ast.ContinueStmt:
		fmt.Printf("%sContinue\n", indent)
	case *ast.AssignStmt:
		fmt.Printf("%sAssign\n", indent)
	case *ast.WithStmt:
		fmt.Printf("%sWith\n", indent)
	default:
		fmt.Printf("%s<stmt>\n", indent)
	}
}

func printTypeExpr(te ast.TypeExpr) {
	switch t := te.(type) {
	case *ast.NamedType:
		fmt.Printf("%s", t.Name)
		if len(t.Args) > 0 {
			fmt.Printf("[")
			for i, a := range t.Args {
				if i > 0 {
					fmt.Printf(", ")
				}
				printTypeExpr(a)
			}
			fmt.Printf("]")
		}
	case *ast.QualifiedType:
		fmt.Printf("%s.%s", t.Qualifier, t.Name)
	case *ast.UnionType:
		printTypeExpr(t.Left)
		fmt.Printf(" | ")
		printTypeExpr(t.Right)
	case *ast.ListType:
		fmt.Printf("[")
		printTypeExpr(t.Element)
		fmt.Printf("]")
	case *ast.OptionType:
		printTypeExpr(t.Inner)
		fmt.Printf("?")
	case *ast.RefinementType:
		printTypeExpr(t.Base)
		fmt.Printf(" where ...")
	case *ast.StringLitType:
		fmt.Printf("%q", t.Value)
	default:
		fmt.Printf("<type>")
	}
}

func visStr(v ast.Visibility) string {
	if v == ast.Public {
		return "pub "
	}
	return ""
}