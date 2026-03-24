package interpreter

import (
	"fmt"

	"github.com/unclebucklarson/aura/pkg/ast"
)

// TestResult holds the result of a single test block.
type TestResult struct {
	Name    string
	Passed  bool
	Error   string
}

// RunTests executes all test blocks in the module and returns results.
func RunTests(module *ast.Module) []TestResult {
	// Create interpreter and register everything
	interp := New(module)
	if _, err := interp.Run(); err != nil {
		return []TestResult{{
			Name:   "<module>",
			Passed: false,
			Error:  err.Error(),
		}}
	}

	var results []TestResult

	for _, item := range module.Items {
		tb, ok := item.(*ast.TestBlock)
		if !ok {
			continue
		}

		tr := runSingleTest(tb, interp.env)
		results = append(results, tr)
	}

	return results
}

func runSingleTest(tb *ast.TestBlock, parentEnv *Environment) TestResult {
	testEnv := NewEnclosedEnvironment(parentEnv)
	result := TestResult{Name: tb.Name, Passed: true}

	defer func() {
		if r := recover(); r != nil {
			result.Passed = false
			switch e := r.(type) {
			case *RuntimeError:
				result.Error = e.Message
			default:
				result.Error = fmt.Sprintf("%v", r)
			}
		}
	}()

	for _, stmt := range tb.Body {
		ExecStmt(stmt, testEnv)
	}

	return result
}

// FormatTestResults formats test results for display.
func FormatTestResults(results []TestResult) string {
	passed := 0
	failed := 0
	output := ""

	for _, r := range results {
		if r.Passed {
			passed++
			output += fmt.Sprintf("  ✓ %s\n", r.Name)
		} else {
			failed++
			output += fmt.Sprintf("  ✗ %s: %s\n", r.Name, r.Error)
		}
	}

	output += fmt.Sprintf("\n%d passed, %d failed, %d total\n", passed, failed, passed+failed)
	return output
}
