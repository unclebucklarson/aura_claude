package lexer

import (
        "testing"

        "github.com/unclebucklarson/aura/pkg/token"
)

func TestSimpleTokens(t *testing.T) {
        input := `let x = 42`
        l := New(input, "test.aura")
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("unexpected errors: %v", errs)
        }

        expected := []token.Type{token.LET, token.IDENT, token.ASSIGN, token.INT_LIT, token.EOF}
        // Filter out non-essential tokens for matching
        var got []token.Type
        for _, tok := range tokens {
                got = append(got, tok.Type)
        }

        if len(got) != len(expected) {
                t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(got), got)
        }
        for i, exp := range expected {
                if got[i] != exp {
                        t.Errorf("token[%d]: expected %s, got %s", i, exp, got[i])
                }
        }
}

func TestKeywords(t *testing.T) {
        input := `module import from as pub fn return let mut if elif else match case for in while break continue struct enum type trait impl spec satisfies with and or not is`
        l := New(input, "test.aura")
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("unexpected errors: %v", errs)
        }

        expectedTypes := []token.Type{
                token.MODULE, token.IMPORT, token.FROM, token.AS, token.PUB,
                token.FN, token.RETURN, token.LET, token.MUT,
                token.IF, token.ELIF, token.ELSE, token.MATCH, token.CASE,
                token.FOR, token.IN, token.WHILE, token.BREAK, token.CONTINUE,
                token.STRUCT, token.ENUM, token.TYPE, token.TRAIT, token.IMPL,
                token.SPEC, token.SATISFIES, token.WITH, token.AND, token.OR, token.NOT, token.IS,
        }

        idx := 0
        for _, tok := range tokens {
                if tok.Type == token.NEWLINE || tok.Type == token.EOF {
                        continue
                }
                if idx >= len(expectedTypes) {
                        break
                }
                if tok.Type != expectedTypes[idx] {
                        t.Errorf("token[%d]: expected %s, got %s (%q)", idx, expectedTypes[idx], tok.Type, tok.Literal)
                }
                idx++
        }
}

func TestOperators(t *testing.T) {
        input := `+ - * / % ** == != < > <= >= = -> : , . ? | &`
        l := New(input, "test.aura")
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("unexpected errors: %v", errs)
        }

        expectedTypes := []token.Type{
                token.PLUS, token.MINUS, token.STAR, token.SLASH, token.PERCENT, token.POWER,
                token.EQ, token.NEQ, token.LT, token.GT, token.LTE, token.GTE,
                token.ASSIGN, token.ARROW, token.COLON, token.COMMA, token.DOT, token.QUESTION,
                token.PIPE, token.AMP,
        }

        idx := 0
        for _, tok := range tokens {
                if tok.Type == token.NEWLINE || tok.Type == token.EOF {
                        continue
                }
                if idx >= len(expectedTypes) {
                        break
                }
                if tok.Type != expectedTypes[idx] {
                        t.Errorf("token[%d]: expected %s, got %s (%q)", idx, expectedTypes[idx], tok.Type, tok.Literal)
                }
                idx++
        }
}

func TestStringLiteral(t *testing.T) {
        input := `"hello world"`
        l := New(input, "test.aura")
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("unexpected errors: %v", errs)
        }

        found := false
        for _, tok := range tokens {
                if tok.Type == token.STRING_LIT {
                        if tok.Literal != "hello world" {
                                t.Errorf("expected 'hello world', got %q", tok.Literal)
                        }
                        found = true
                }
        }
        if !found {
                t.Error("no STRING_LIT token found")
        }
}

func TestIndentation(t *testing.T) {
        input := "if x:\n    let y = 1\n    let z = 2\nlet w = 3\n"
        l := New(input, "test.aura")
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("unexpected errors: %v", errs)
        }

        // Check for INDENT and DEDENT
        hasIndent := false
        hasDedent := false
        for _, tok := range tokens {
                if tok.Type == token.INDENT {
                        hasIndent = true
                }
                if tok.Type == token.DEDENT {
                        hasDedent = true
                }
        }
        if !hasIndent {
                t.Error("expected INDENT token")
        }
        if !hasDedent {
                t.Error("expected DEDENT token")
        }
}

func TestNumbers(t *testing.T) {
        input := `42 3.14 1e10`
        l := New(input, "test.aura")
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("unexpected errors: %v", errs)
        }

        var nums []token.Token
        for _, tok := range tokens {
                if tok.Type == token.INT_LIT || tok.Type == token.FLOAT_LIT {
                        nums = append(nums, tok)
                }
        }
        if len(nums) != 3 {
                t.Fatalf("expected 3 number tokens, got %d", len(nums))
        }
        if nums[0].Type != token.INT_LIT || nums[0].Literal != "42" {
                t.Errorf("expected INT 42, got %s %q", nums[0].Type, nums[0].Literal)
        }
        if nums[1].Type != token.FLOAT_LIT || nums[1].Literal != "3.14" {
                t.Errorf("expected FLOAT 3.14, got %s %q", nums[1].Type, nums[1].Literal)
        }
        if nums[2].Type != token.FLOAT_LIT || nums[2].Literal != "1e10" {
                t.Errorf("expected FLOAT 1e10, got %s %q", nums[2].Type, nums[2].Literal)
        }
}

func TestBoolAndNone(t *testing.T) {
        input := `true false none`
        l := New(input, "test.aura")
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("unexpected errors: %v", errs)
        }

        expected := []token.Type{token.BOOL_LIT, token.BOOL_LIT, token.NONE_LIT}
        idx := 0
        for _, tok := range tokens {
                if tok.Type == token.NEWLINE || tok.Type == token.EOF {
                        continue
                }
                if idx < len(expected) {
                        if tok.Type != expected[idx] {
                                t.Errorf("token[%d]: expected %s, got %s", idx, expected[idx], tok.Type)
                        }
                        idx++
                }
        }
}

func TestComments(t *testing.T) {
        input := "# comment\n## doc comment\nlet x = 1\n"
        l := New(input, "test.aura")
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("unexpected errors: %v", errs)
        }

        hasComment := false
        hasDocComment := false
        for _, tok := range tokens {
                if tok.Type == token.COMMENT {
                        hasComment = true
                }
                if tok.Type == token.DOC_COMMENT {
                        hasDocComment = true
                }
        }
        if !hasComment {
                t.Error("expected COMMENT token")
        }
        if !hasDocComment {
                t.Error("expected DOC_COMMENT token")
        }
}

func TestParenSuppressesNewline(t *testing.T) {
        input := "f(\n    1,\n    2\n)\n"
        l := New(input, "test.aura")
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("unexpected errors: %v", errs)
        }

        // Inside parens, NEWLINEs should be suppressed
        for _, tok := range tokens {
                if tok.Type == token.INDENT || tok.Type == token.DEDENT {
                        t.Errorf("unexpected %s token inside parentheses", tok.Type)
                }
        }
}

func TestEmptyInput(t *testing.T) {
        input := ""
        l := New(input, "test.aura")
        tokens, errs := l.Tokenize()
        if len(errs) > 0 {
                t.Fatalf("unexpected errors: %v", errs)
        }
        if len(tokens) != 1 || tokens[0].Type != token.EOF {
                t.Errorf("expected single EOF token, got %v", tokens)
        }
}

func TestPositionTracking(t *testing.T) {
        input := "let x = 1\nlet y = 2\n"
        l := New(input, "test.aura")
        tokens, _ := l.Tokenize()

        // First token should be at line 1, col 1
        if tokens[0].Pos.Line != 1 || tokens[0].Pos.Column != 1 {
                t.Errorf("first token position: expected 1:1, got %d:%d", tokens[0].Pos.Line, tokens[0].Pos.Column)
        }

        // Find second 'let' - should be at line 2
        for _, tok := range tokens {
                if tok.Type == token.LET && tok.Pos.Line == 2 {
                        if tok.Pos.Column != 1 {
                                t.Errorf("second let position: expected col 1, got %d", tok.Pos.Column)
                        }
                        return
                }
        }
        t.Error("second 'let' token at line 2 not found")
}
