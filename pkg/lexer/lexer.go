// Package lexer implements the tokenizer for the Aura language.
// It handles indentation-based blocks by emitting INDENT and DEDENT tokens.
package lexer

import (
        "fmt"
        "strings"
        "unicode"
        "unicode/utf8"

        "github.com/unclebucklarson/aura/pkg/token"
)

// Lexer tokenizes Aura source code.
type Lexer struct {
        input   string
        file    string
        pos     int // current position in input
        line    int // current line (1-based)
        col     int // current column (1-based)
        tokens  []token.Token
        errors  []Error

        // Indentation tracking
        indentStack []int // stack of indentation levels
        atLineStart bool  // whether we're at the start of a line
        parenDepth  int   // depth of () [] {} nesting (suppress NEWLINE inside)
}

// Error represents a lexer error with position information.
type Error struct {
        Pos     token.Position
        Message string
}

func (e Error) Error() string {
        return fmt.Sprintf("%s: %s", e.Pos, e.Message)
}

// New creates a new Lexer for the given source code.
func New(input string, file string) *Lexer {
        l := &Lexer{
                input:       input,
                file:        file,
                pos:         0,
                line:        1,
                col:         1,
                indentStack: []int{0},
                atLineStart: true,
                parenDepth:  0,
        }
        return l
}

// Tokenize processes the entire input and returns all tokens.
func (l *Lexer) Tokenize() ([]token.Token, []Error) {
        for l.pos < len(l.input) {
                if l.atLineStart {
                        if l.parenDepth > 0 {
                                // Inside brackets/parens, skip leading whitespace but don't emit INDENT/DEDENT
                                for l.pos < len(l.input) && (l.peek() == ' ' || l.peek() == '\t') {
                                        l.advance()
                                }
                                l.atLineStart = false
                        } else {
                                l.handleIndentation()
                                // handleIndentation may set atLineStart=true for blank lines;
                                // if so, loop again to re-process indentation.
                                if l.atLineStart {
                                        continue
                                }
                        }
                        if l.pos >= len(l.input) {
                                break
                        }
                }
                l.scanToken()
        }

        // Emit DEDENTs for remaining indent levels
        for len(l.indentStack) > 1 {
                l.indentStack = l.indentStack[:len(l.indentStack)-1]
                l.emit(token.DEDENT, "")
        }

        l.emit(token.EOF, "")
        return l.tokens, l.errors
}

// Errors returns any errors encountered during tokenization.
func (l *Lexer) Errors() []Error {
        return l.errors
}

func (l *Lexer) currentPos() token.Position {
        return token.Position{Line: l.line, Column: l.col, Offset: l.pos}
}

func (l *Lexer) emit(typ token.Type, literal string) {
        l.tokens = append(l.tokens, token.Token{
                Type:    typ,
                Literal: literal,
                Pos:     l.currentPos(),
        })
}

func (l *Lexer) emitAt(typ token.Type, literal string, pos token.Position) {
        l.tokens = append(l.tokens, token.Token{
                Type:    typ,
                Literal: literal,
                Pos:     pos,
        })
}

func (l *Lexer) addError(msg string) {
        l.errors = append(l.errors, Error{
                Pos:     l.currentPos(),
                Message: msg,
        })
}

func (l *Lexer) peek() byte {
        if l.pos >= len(l.input) {
                return 0
        }
        return l.input[l.pos]
}

func (l *Lexer) peekAt(offset int) byte {
        p := l.pos + offset
        if p >= len(l.input) {
                return 0
        }
        return l.input[p]
}

func (l *Lexer) advance() byte {
        if l.pos >= len(l.input) {
                return 0
        }
        ch := l.input[l.pos]
        l.pos++
        if ch == '\n' {
                l.line++
                l.col = 1
        } else {
                l.col++
        }
        return ch
}

func (l *Lexer) handleIndentation() {
        pos := l.currentPos()
        indent := 0

        for l.pos < len(l.input) {
                ch := l.peek()
                if ch == ' ' {
                        indent++
                        l.advance()
                } else if ch == '\t' {
                        l.addError("tabs are not allowed for indentation; use 4 spaces")
                        l.advance()
                        indent += 4
                } else {
                        break
                }
        }

        // Skip blank lines and comment-only lines
        if l.pos >= len(l.input) || l.peek() == '\n' {
                if l.pos < len(l.input) {
                        l.advance() // consume the newline
                        l.atLineStart = true
                }
                return
        }

        // Handle comments at any indentation (don't change indent level for comment-only lines)
        if l.peek() == '#' {
                l.scanComment()
                // After comment, check for newline
                if l.pos < len(l.input) && l.peek() == '\n' {
                        l.advance()
                        l.atLineStart = true
                }
                return
        }

        currentIndent := l.indentStack[len(l.indentStack)-1]

        if indent > currentIndent {
                l.indentStack = append(l.indentStack, indent)
                l.emitAt(token.INDENT, "", pos)
        } else if indent < currentIndent {
                for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > indent {
                        l.indentStack = l.indentStack[:len(l.indentStack)-1]
                        l.emitAt(token.DEDENT, "", pos)
                }
                if l.indentStack[len(l.indentStack)-1] != indent {
                        l.addError(fmt.Sprintf("inconsistent indentation: expected %d spaces, got %d", l.indentStack[len(l.indentStack)-1], indent))
                }
        }
        // Mark that we've processed indentation for a real line
        l.atLineStart = false
}

func (l *Lexer) scanToken() {
        ch := l.peek()

        // Skip spaces (not at line start)
        if ch == ' ' || ch == '\t' {
                l.advance()
                return
        }

        // Newline
        if ch == '\n' {
                if l.parenDepth == 0 {
                        l.emit(token.NEWLINE, "\n")
                }
                l.advance()
                l.atLineStart = true
                return
        }

        // Comment
        if ch == '#' {
                l.scanComment()
                return
        }

        // String literal
        if ch == '"' {
                l.scanString()
                return
        }

        // Number literal
        if ch >= '0' && ch <= '9' {
                l.scanNumber()
                return
        }

        // Negative number (only if followed by digit and preceded by operator context)
        // We'll let the parser handle unary minus instead

        // Identifier / keyword
        if isIdentStart(ch) {
                l.scanIdentifier()
                return
        }

        // Operators and delimiters
        pos := l.currentPos()
        l.advance()

        switch ch {
        case '+':
                l.emitAt(token.PLUS, "+", pos)
        case '-':
                if l.peek() == '>' {
                        l.advance()
                        l.emitAt(token.ARROW, "->", pos)
                } else {
                        l.emitAt(token.MINUS, "-", pos)
                }
        case '*':
                if l.peek() == '*' {
                        l.advance()
                        l.emitAt(token.POWER, "**", pos)
                } else {
                        l.emitAt(token.STAR, "*", pos)
                }
        case '/':
                l.emitAt(token.SLASH, "/", pos)
        case '%':
                l.emitAt(token.PERCENT, "%", pos)
        case '=':
                if l.peek() == '=' {
                        l.advance()
                        l.emitAt(token.EQ, "==", pos)
                } else {
                        l.emitAt(token.ASSIGN, "=", pos)
                }
        case '!':
                if l.peek() == '=' {
                        l.advance()
                        l.emitAt(token.NEQ, "!=", pos)
                } else {
                        l.emitAt(token.BANG, "!", pos)
                }
        case '<':
                if l.peek() == '=' {
                        l.advance()
                        l.emitAt(token.LTE, "<=", pos)
                } else {
                        l.emitAt(token.LT, "<", pos)
                }
        case '>':
                if l.peek() == '=' {
                        l.advance()
                        l.emitAt(token.GTE, ">=", pos)
                } else {
                        l.emitAt(token.GT, ">", pos)
                }
        case '(':
                l.parenDepth++
                l.emitAt(token.LPAREN, "(", pos)
        case ')':
                l.parenDepth--
                if l.parenDepth < 0 {
                        l.parenDepth = 0
                }
                l.emitAt(token.RPAREN, ")", pos)
        case '[':
                l.parenDepth++
                l.emitAt(token.LBRACKET, "[", pos)
        case ']':
                l.parenDepth--
                if l.parenDepth < 0 {
                        l.parenDepth = 0
                }
                l.emitAt(token.RBRACKET, "]", pos)
        case '{':
                l.parenDepth++
                l.emitAt(token.LBRACE, "{", pos)
        case '}':
                l.parenDepth--
                if l.parenDepth < 0 {
                        l.parenDepth = 0
                }
                l.emitAt(token.RBRACE, "}", pos)
        case ':':
                l.emitAt(token.COLON, ":", pos)
        case ',':
                l.emitAt(token.COMMA, ",", pos)
        case '.':
                if l.peek() == '.' && l.peekAt(1) == '.' {
                        l.advance()
                        l.advance()
                        l.emitAt(token.DOTDOTDOT, "...", pos)
                } else {
                        l.emitAt(token.DOT, ".", pos)
                }
        case '?':
                if l.peek() == '.' {
                        l.advance()
                        l.emitAt(token.QUESTION_DOT, "?.", pos)
                } else {
                        l.emitAt(token.QUESTION, "?", pos)
                }
        case '|':
                if l.peek() == '>' {
                        l.advance()
                        l.emitAt(token.PIPE_GT, "|>", pos)
                } else {
                        l.emitAt(token.PIPE, "|", pos)
                }
        case '&':
                l.emitAt(token.AMP, "&", pos)
        default:
                l.addError(fmt.Sprintf("unexpected character: %q", ch))
                l.emitAt(token.ILLEGAL, string(ch), pos)
        }
}

func (l *Lexer) scanComment() {
        pos := l.currentPos()
        l.advance() // consume first #

        isDoc := l.peek() == '#'
        if isDoc {
                l.advance() // consume second #
        }

        start := l.pos
        for l.pos < len(l.input) && l.peek() != '\n' {
                l.advance()
        }

        text := l.input[start:l.pos]
        if isDoc {
                l.emitAt(token.DOC_COMMENT, strings.TrimSpace(text), pos)
        } else {
                l.emitAt(token.COMMENT, strings.TrimSpace(text), pos)
        }
}

func (l *Lexer) scanString() {
        pos := l.currentPos()
        l.advance() // consume opening "

        var buf strings.Builder
        for l.pos < len(l.input) {
                ch := l.peek()
                if ch == '"' {
                        l.advance() // consume closing "
                        l.emitAt(token.STRING_LIT, buf.String(), pos)
                        return
                }
                if ch == '\\' {
                        l.advance()
                        esc := l.peek()
                        switch esc {
                        case 'n':
                                buf.WriteByte('\n')
                        case 't':
                                buf.WriteByte('\t')
                        case '\\':
                                buf.WriteByte('\\')
                        case '"':
                                buf.WriteByte('"')
                        case '{':
                                buf.WriteByte('{')
                        case '}':
                                buf.WriteByte('}')
                        default:
                                l.addError(fmt.Sprintf("unknown escape sequence: \\%c", esc))
                                buf.WriteByte(esc)
                        }
                        l.advance()
                        continue
                }
                if ch == '\n' {
                        l.addError("unterminated string literal")
                        l.emitAt(token.STRING_LIT, buf.String(), pos)
                        return
                }
                buf.WriteByte(ch)
                l.advance()
        }
        l.addError("unterminated string literal at end of file")
        l.emitAt(token.STRING_LIT, buf.String(), pos)
}

func (l *Lexer) scanNumber() {
        pos := l.currentPos()
        start := l.pos
        isFloat := false

        for l.pos < len(l.input) && l.peek() >= '0' && l.peek() <= '9' {
                l.advance()
        }

        // Check for decimal point
        if l.pos < len(l.input) && l.peek() == '.' && l.peekAt(1) >= '0' && l.peekAt(1) <= '9' {
                isFloat = true
                l.advance() // consume .
                for l.pos < len(l.input) && l.peek() >= '0' && l.peek() <= '9' {
                        l.advance()
                }
        }

        // Check for exponent
        if l.pos < len(l.input) && (l.peek() == 'e' || l.peek() == 'E') {
                isFloat = true
                l.advance()
                if l.pos < len(l.input) && (l.peek() == '+' || l.peek() == '-') {
                        l.advance()
                }
                for l.pos < len(l.input) && l.peek() >= '0' && l.peek() <= '9' {
                        l.advance()
                }
        }

        literal := l.input[start:l.pos]
        if isFloat {
                l.emitAt(token.FLOAT_LIT, literal, pos)
        } else {
                l.emitAt(token.INT_LIT, literal, pos)
        }
}

func (l *Lexer) scanIdentifier() {
        pos := l.currentPos()
        start := l.pos

        for l.pos < len(l.input) {
                ch := l.peek()
                if isIdentPart(ch) {
                        l.advance()
                } else {
                        break
                }
        }

        literal := l.input[start:l.pos]
        typ := token.LookupIdent(literal)

        // Handle bool literals
        if typ == token.TRUE || typ == token.FALSE {
                l.emitAt(token.BOOL_LIT, literal, pos)
                return
        }

        // Handle none literal
        if typ == token.NONE_VAL {
                l.emitAt(token.NONE_LIT, literal, pos)
                return
        }

        // Check for string prefix 'r' for regex strings
        if literal == "r" && l.pos < len(l.input) && l.peek() == '"' {
                l.scanString()
                // Modify the last token to be a string with 'r' prefix indicator
                if len(l.tokens) > 0 {
                        last := &l.tokens[len(l.tokens)-1]
                        last.Literal = "r" + last.Literal
                        last.Pos = pos
                }
                return
        }

        l.emitAt(typ, literal, pos)
}

func isIdentStart(ch byte) bool {
        return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentPart(ch byte) bool {
        return isIdentStart(ch) || (ch >= '0' && ch <= '9')
}

// Helper for rune-based validation (unused but kept for future UTF-8 support)
func isLetter(r rune) bool {
        return unicode.IsLetter(r)
}

func isDigit(r rune) bool {
        return unicode.IsDigit(r)
}

// Ensure utf8 import is used
var _ = utf8.RuneCountInString
