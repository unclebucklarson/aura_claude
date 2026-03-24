// Package token defines the token types for the Aura language.
package token

import "fmt"

// Type represents a token type in the Aura language.
type Type int

const (
        // Special tokens
        ILLEGAL Type = iota
        EOF
        NEWLINE
        INDENT
        DEDENT
        COMMENT
        DOC_COMMENT

        // Literals
        IDENT     // identifier (starts with lowercase or _)
        TYPE_IDENT // type identifier (starts with uppercase)
        INT_LIT
        FLOAT_LIT
        STRING_LIT
        BOOL_LIT
        NONE_LIT

        // Operators and delimiters
        PLUS     // +
        MINUS    // -
        STAR     // *
        SLASH    // /
        PERCENT  // %
        POWER    // **
        EQ       // ==
        NEQ      // !=
        LT       // <
        GT       // >
        LTE      // <=
        GTE      // >=
        ASSIGN   // =
        LPAREN   // (
        RPAREN   // )
        LBRACKET // [
        RBRACKET // ]
        LBRACE   // {
        RBRACE   // }
        COLON    // :
        COMMA    // ,
        DOT      // .
        QUESTION // ?
        PIPE         // |
        PIPE_GT      // |>
        AMP          // &
        BANG         // !
        ARROW        // ->
        QUESTION_DOT // ?.
        HASH         // #
        DOTDOTDOT    // ...

        // Keywords
        MODULE
        IMPORT
        FROM
        AS
        PUB
        FN
        RETURN
        LET
        MUT
        IF
        ELIF
        ELSE
        MATCH
        CASE
        FOR
        IN
        WHILE
        BREAK
        CONTINUE
        STRUCT
        ENUM
        TYPE
        TRAIT
        IMPL
        SPEC
        SATISFIES
        REQUIRES
        GUARANTEES
        ERRORS
        WITH
        DO
        AND
        OR
        NOT
        IS
        OPTION_KW // Option
        RESULT_KW // Result
        OK
        ERR
        SOME
        NONE_KW // None keyword (distinct from none literal)
        TRUE
        FALSE
        NONE_VAL // none (value)
        ASSERT
        TEST
        DOC_KW       // doc
        INPUTS_KW    // inputs
        EFFECTS_KW   // effects
        WHERE
        MATCHES
        THEN
)

var tokenNames = map[Type]string{
        ILLEGAL:     "ILLEGAL",
        EOF:         "EOF",
        NEWLINE:     "NEWLINE",
        INDENT:      "INDENT",
        DEDENT:      "DEDENT",
        COMMENT:     "COMMENT",
        DOC_COMMENT: "DOC_COMMENT",
        IDENT:       "IDENT",
        TYPE_IDENT:  "TYPE_IDENT",
        INT_LIT:     "INT_LIT",
        FLOAT_LIT:   "FLOAT_LIT",
        STRING_LIT:  "STRING_LIT",
        BOOL_LIT:    "BOOL_LIT",
        NONE_LIT:    "NONE_LIT",
        PLUS:        "+",
        MINUS:       "-",
        STAR:        "*",
        SLASH:       "/",
        PERCENT:     "%",
        POWER:       "**",
        EQ:          "==",
        NEQ:         "!=",
        LT:          "<",
        GT:          ">",
        LTE:         "<=",
        GTE:         ">=",
        ASSIGN:      "=",
        LPAREN:      "(",
        RPAREN:      ")",
        LBRACKET:    "[",
        RBRACKET:    "]",
        LBRACE:      "{",
        RBRACE:      "}",
        COLON:       ":",
        COMMA:       ",",
        DOT:         ".",
        QUESTION:    "?",
        PIPE:         "|",
        PIPE_GT:      "|>",
        AMP:          "&",
        BANG:         "!",
        ARROW:        "->",
        QUESTION_DOT: "?.",
        HASH:         "#",
        DOTDOTDOT:   "...",
        MODULE:      "module",
        IMPORT:      "import",
        FROM:        "from",
        AS:          "as",
        PUB:         "pub",
        FN:          "fn",
        RETURN:      "return",
        LET:         "let",
        MUT:         "mut",
        IF:          "if",
        ELIF:        "elif",
        ELSE:        "else",
        MATCH:       "match",
        CASE:        "case",
        FOR:         "for",
        IN:          "in",
        WHILE:       "while",
        BREAK:       "break",
        CONTINUE:    "continue",
        STRUCT:      "struct",
        ENUM:        "enum",
        TYPE:        "type",
        TRAIT:       "trait",
        IMPL:        "impl",
        SPEC:        "spec",
        SATISFIES:   "satisfies",
        REQUIRES:    "requires",
        GUARANTEES:  "guarantees",
        ERRORS:      "errors",
        WITH:        "with",
        DO:          "do",
        AND:         "and",
        OR:          "or",
        NOT:         "not",
        IS:          "is",
        OPTION_KW:   "Option",
        RESULT_KW:   "Result",
        OK:          "Ok",
        ERR:         "Err",
        SOME:        "Some",
        NONE_KW:     "None",
        TRUE:        "true",
        FALSE:       "false",
        NONE_VAL:    "none",
        ASSERT:      "assert",
        TEST:        "test",
        DOC_KW:      "doc",
        INPUTS_KW:   "inputs",
        EFFECTS_KW:  "effects",
        WHERE:       "where",
        MATCHES:     "matches",
        THEN:        "then",
}

func (t Type) String() string {
        if name, ok := tokenNames[t]; ok {
                return name
        }
        return fmt.Sprintf("Token(%d)", int(t))
}

// keywords maps keyword strings to their token types.
var keywords = map[string]Type{
        "module":     MODULE,
        "import":     IMPORT,
        "from":       FROM,
        "as":         AS,
        "pub":        PUB,
        "fn":         FN,
        "return":     RETURN,
        "let":        LET,
        "mut":        MUT,
        "if":         IF,
        "elif":       ELIF,
        "else":       ELSE,
        "match":      MATCH,
        "case":       CASE,
        "for":        FOR,
        "in":         IN,
        "while":      WHILE,
        "break":      BREAK,
        "continue":   CONTINUE,
        "struct":     STRUCT,
        "enum":       ENUM,
        "type":       TYPE,
        "trait":      TRAIT,
        "impl":       IMPL,
        "spec":       SPEC,
        "satisfies":  SATISFIES,
        "requires":   REQUIRES,
        "guarantees": GUARANTEES,
        "errors":     ERRORS,
        "with":       WITH,
        "do":         DO,
        "and":        AND,
        "or":         OR,
        "not":        NOT,
        "is":         IS,
        "Option":     OPTION_KW,
        "Result":     RESULT_KW,
        "Ok":         OK,
        "Err":        ERR,
        "Some":       SOME,
        "None":       NONE_KW,
        "true":       TRUE,
        "false":      FALSE,
        "none":       NONE_VAL,
        "assert":     ASSERT,
        "test":       TEST,
        "doc":        DOC_KW,
        "inputs":     INPUTS_KW,
        "effects":    EFFECTS_KW,
        "where":      WHERE,
        "matches":    MATCHES,
        "then":       THEN,
}

// LookupIdent returns the token type for the given identifier string.
// If the identifier is a keyword, the keyword token type is returned.
// Otherwise, IDENT or TYPE_IDENT is returned based on the first character.
func LookupIdent(ident string) Type {
        if tok, ok := keywords[ident]; ok {
                return tok
        }
        if len(ident) > 0 && ident[0] >= 'A' && ident[0] <= 'Z' {
                return TYPE_IDENT
        }
        return IDENT
}

// Position represents a position in source code.
type Position struct {
        Line   int // 1-based line number
        Column int // 1-based column number
        Offset int // 0-based byte offset
}

func (p Position) String() string {
        return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// Span represents a range in source code.
type Span struct {
        File  string
        Start Position
        End   Position
}

// Token represents a lexical token.
type Token struct {
        Type    Type
        Literal string
        Pos     Position
}

func (t Token) String() string {
        return fmt.Sprintf("Token(%s, %q, %s)", t.Type, t.Literal, t.Pos)
}

// IsKeyword returns true if the token type is a keyword.
func (t Type) IsKeyword() bool {
        return t >= MODULE && t <= THEN
}
