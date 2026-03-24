// Package parser implements a recursive descent parser for the Aura language.
package parser

import (
        "fmt"
        "strings"

        "github.com/unclebucklarson/aura/pkg/ast"
        "github.com/unclebucklarson/aura/pkg/lexer"
        "github.com/unclebucklarson/aura/pkg/token"
)

// Parser transforms a token stream into an AST.
type Parser struct {
        tokens             []token.Token
        pos                int
        file               string
        errors             []Error
        blockExprJustEnded bool // set when a block expression (e.g. match expr) just consumed its DEDENT
}

// Error represents a parser error.
type Error struct {
        Pos     token.Position
        Message string
}

func (e Error) Error() string {
        return fmt.Sprintf("%s: %s", e.Pos, e.Message)
}

// New creates a new Parser from a slice of tokens.
func New(tokens []token.Token, file string) *Parser {
        return &Parser{
                tokens: tokens,
                pos:    0,
                file:   file,
        }
}

// Parse parses the token stream and returns a Module AST.
func (p *Parser) Parse() (*ast.Module, []Error) {
        module := p.parseModule()
        return module, p.errors
}

// Errors returns parser errors.
func (p *Parser) Errors() []Error {
        return p.errors
}

// --- Token access helpers ---

func (p *Parser) current() token.Token {
        if p.pos >= len(p.tokens) {
                return token.Token{Type: token.EOF}
        }
        return p.tokens[p.pos]
}

func (p *Parser) peek() token.Type {
        return p.current().Type
}

func (p *Parser) peekAt(offset int) token.Type {
        idx := p.pos + offset
        if idx >= len(p.tokens) {
                return token.EOF
        }
        return p.tokens[idx].Type
}

func (p *Parser) advance() token.Token {
        tok := p.current()
        if p.pos < len(p.tokens) {
                p.pos++
        }
        return tok
}

func (p *Parser) expect(typ token.Type) token.Token {
        tok := p.current()
        if tok.Type != typ {
                p.addError(fmt.Sprintf("expected %s, got %s (%q)", typ, tok.Type, tok.Literal))
                return tok
        }
        return p.advance()
}

func (p *Parser) check(typ token.Type) bool {
        return p.peek() == typ
}

func (p *Parser) match(types ...token.Type) bool {
        for _, t := range types {
                if p.check(t) {
                        return true
                }
        }
        return false
}

func (p *Parser) addError(msg string) {
        tok := p.current()
        p.errors = append(p.errors, Error{
                Pos:     tok.Pos,
                Message: msg,
        })
}

func (p *Parser) skipNewlines() {
        for p.check(token.NEWLINE) || p.check(token.COMMENT) || p.check(token.DOC_COMMENT) {
                p.advance()
        }
}

func (p *Parser) expectNewline() {
        if p.blockExprJustEnded {
                p.blockExprJustEnded = false
                return // block expression already consumed its terminator
        }
        if p.check(token.NEWLINE) {
                p.advance()
        } else if p.check(token.EOF) || p.check(token.DEDENT) {
                // OK, implicit newline at end
        } else {
                p.addError(fmt.Sprintf("expected newline, got %s", p.peek()))
        }
}

func (p *Parser) makeSpan(start token.Position) token.Span {
        end := p.current().Pos
        return token.Span{File: p.file, Start: start, End: end}
}

// --- Module parsing ---

func (p *Parser) parseModule() *ast.Module {
        start := p.current().Pos
        p.skipNewlines()

        module := &ast.Module{}

        // Parse optional module declaration
        if p.check(token.MODULE) {
                p.advance()
                module.Name = p.parseQualifiedName()
                p.expectNewline()
                p.skipNewlines()
        }

        // Parse imports
        for p.check(token.IMPORT) || p.check(token.FROM) {
                imp := p.parseImport()
                if imp != nil {
                        module.Imports = append(module.Imports, imp)
                }
                p.skipNewlines()
        }

        // Parse top-level items
        for !p.check(token.EOF) {
                p.skipNewlines()
                if p.check(token.EOF) {
                        break
                }
                item := p.parseTopLevelItem()
                if item != nil {
                        module.Items = append(module.Items, item)
                }
                p.skipNewlines()
        }

        module.Span = p.makeSpan(start)
        return module
}

func (p *Parser) parseQualifiedName() *ast.QualifiedName {
        start := p.current().Pos
        qn := &ast.QualifiedName{}

        tok := p.current()
        if tok.Type == token.IDENT || tok.Type == token.TYPE_IDENT || tok.Type.IsKeyword() {
                qn.Parts = append(qn.Parts, tok.Literal)
                p.advance()
        } else {
                p.addError(fmt.Sprintf("expected identifier in qualified name, got %s", tok.Type))
                return qn
        }

        for p.check(token.DOT) {
                p.advance() // consume .
                tok = p.current()
                if tok.Type == token.IDENT || tok.Type == token.TYPE_IDENT || tok.Type.IsKeyword() {
                        qn.Parts = append(qn.Parts, tok.Literal)
                        p.advance()
                } else {
                        p.addError("expected identifier after '.'")
                        break
                }
        }

        qn.Span = p.makeSpan(start)
        return qn
}

func (p *Parser) parseImport() *ast.ImportNode {
        start := p.current().Pos
        imp := &ast.ImportNode{}

        if p.check(token.IMPORT) {
                p.advance()
                imp.Path = p.parseQualifiedName()
                if p.check(token.AS) {
                        p.advance()
                        imp.Alias = p.expect(token.IDENT).Literal
                }
        } else if p.check(token.FROM) {
                p.advance()
                imp.Path = p.parseQualifiedName()
                p.expect(token.IMPORT)
                if p.check(token.STAR) {
                        p.advance()
                        imp.Names = []string{"*"}
                } else {
                        imp.Names = p.parseIdentList()
                }
        }

        p.expectNewline()
        imp.Span = p.makeSpan(start)
        return imp
}

func (p *Parser) parseIdentList() []string {
        var names []string
        tok := p.current()
        if tok.Type == token.IDENT || tok.Type == token.TYPE_IDENT || tok.Type.IsKeyword() {
                names = append(names, tok.Literal)
                p.advance()
        }
        for p.check(token.COMMA) {
                p.advance()
                tok = p.current()
                if tok.Type == token.IDENT || tok.Type == token.TYPE_IDENT || tok.Type.IsKeyword() {
                        names = append(names, tok.Literal)
                        p.advance()
                }
        }
        return names
}

// --- Top-level item parsing ---

func (p *Parser) parseTopLevelItem() ast.TopLevelItem {
        vis := ast.Private
        if p.check(token.PUB) {
                vis = ast.Public
                p.advance()
        }

        switch p.peek() {
        case token.TYPE:
                return p.parseTypeDef(vis)
        case token.STRUCT:
                return p.parseStructDef(vis)
        case token.ENUM:
                return p.parseEnumDef(vis)
        case token.TRAIT:
                return p.parseTraitDef(vis)
        case token.IMPL:
                return p.parseImplBlock()
        case token.SPEC:
                return p.parseSpecBlock()
        case token.FN:
                return p.parseFnDef(vis)
        case token.LET:
                return p.parseConstDef(vis)
        case token.TEST:
                return p.parseTestBlock()
        default:
                p.addError(fmt.Sprintf("unexpected token at top level: %s (%q)", p.peek(), p.current().Literal))
                p.advance()
                return nil
        }
}

// --- Type definitions ---

func (p *Parser) parseTypeDef(vis ast.Visibility) *ast.TypeDef {
        start := p.current().Pos
        p.expect(token.TYPE)

        td := &ast.TypeDef{Visibility: vis}
        td.Name = p.expectTypeName()
        td.TypeParams = p.parseOptionalTypeParams()
        p.expect(token.ASSIGN)
        td.Body = p.parseTypeExpr()
        p.expectNewline()
        td.Span = p.makeSpan(start)
        return td
}

func (p *Parser) expectTypeName() string {
        tok := p.current()
        if tok.Type == token.TYPE_IDENT || tok.Type == token.IDENT || tok.Type.IsKeyword() {
                p.advance()
                return tok.Literal
        }
        p.addError(fmt.Sprintf("expected type name, got %s", tok.Type))
        return "_error_"
}

func (p *Parser) parseOptionalTypeParams() []string {
        if !p.check(token.LBRACKET) {
                return nil
        }
        p.advance()
        var params []string
        params = append(params, p.expectTypeName())
        for p.check(token.COMMA) {
                p.advance()
                params = append(params, p.expectTypeName())
        }
        p.expect(token.RBRACKET)
        return params
}

// --- Type expressions ---

func (p *Parser) parseTypeExpr() ast.TypeExpr {
        return p.parseUnionType()
}

func (p *Parser) parseUnionType() ast.TypeExpr {
        left := p.parseIntersectionType()
        for p.check(token.PIPE) {
                p.advance()
                right := p.parseIntersectionType()
                left = &ast.UnionType{
                        Span:  token.Span{File: p.file, Start: left.GetSpan().Start, End: right.GetSpan().End},
                        Left:  left,
                        Right: right,
                }
        }
        return left
}

func (p *Parser) parseIntersectionType() ast.TypeExpr {
        left := p.parseOptionOrRefinementType()
        for p.check(token.AMP) {
                p.advance()
                right := p.parseOptionOrRefinementType()
                left = &ast.IntersectionType{
                        Span:  token.Span{File: p.file, Start: left.GetSpan().Start, End: right.GetSpan().End},
                        Left:  left,
                        Right: right,
                }
        }
        return left
}

func (p *Parser) parseOptionOrRefinementType() ast.TypeExpr {
        base := p.parsePrimaryType()

        // Option shorthand: T?
        if p.check(token.QUESTION) {
                p.advance()
                base = &ast.OptionType{
                        Span:  token.Span{File: p.file, Start: base.GetSpan().Start, End: p.current().Pos},
                        Inner: base,
                }
        }

        // Refinement type: T where predicate
        if p.check(token.WHERE) {
                p.advance()
                pred := p.parseExpr()
                base = &ast.RefinementType{
                        Span:      token.Span{File: p.file, Start: base.GetSpan().Start, End: pred.GetSpan().End},
                        Base:      base,
                        Predicate: pred,
                }
        }

        return base
}

func (p *Parser) parsePrimaryType() ast.TypeExpr {
        start := p.current().Pos

        // String literal as type (for union string types)
        if p.check(token.STRING_LIT) {
                tok := p.advance()
                return &ast.StringLitType{
                        Span:  p.makeSpan(start),
                        Value: tok.Literal,
                }
        }

        // Function type: fn(params) -> ret
        if p.check(token.FN) {
                p.advance()
                p.expect(token.LPAREN)
                var params []ast.TypeExpr
                if !p.check(token.RPAREN) {
                        params = append(params, p.parseTypeExpr())
                        for p.check(token.COMMA) {
                                p.advance()
                                params = append(params, p.parseTypeExpr())
                        }
                }
                p.expect(token.RPAREN)
                p.expect(token.ARROW)
                ret := p.parseTypeExpr()
                return &ast.FunctionType{
                        Span:       p.makeSpan(start),
                        Params:     params,
                        ReturnType: ret,
                }
        }

        // List type: [T]
        if p.check(token.LBRACKET) {
                p.advance()
                elem := p.parseTypeExpr()
                p.expect(token.RBRACKET)
                return &ast.ListType{
                        Span:    p.makeSpan(start),
                        Element: elem,
                }
        }

        // Map type {K: V} or Set type {T} or Tuple (T, U)
        if p.check(token.LBRACE) {
                p.advance()
                first := p.parseTypeExpr()
                if p.check(token.COLON) {
                        p.advance()
                        val := p.parseTypeExpr()
                        p.expect(token.RBRACE)
                        return &ast.MapType{
                                Span:  p.makeSpan(start),
                                Key:   first,
                                Value: val,
                        }
                }
                p.expect(token.RBRACE)
                return &ast.SetType{
                        Span:    p.makeSpan(start),
                        Element: first,
                }
        }

        // Tuple type (T, U)
        if p.check(token.LPAREN) {
                p.advance()
                var elems []ast.TypeExpr
                elems = append(elems, p.parseTypeExpr())
                if p.check(token.COMMA) {
                        for p.check(token.COMMA) {
                                p.advance()
                                elems = append(elems, p.parseTypeExpr())
                        }
                        p.expect(token.RPAREN)
                        return &ast.TupleType{
                                Span:     p.makeSpan(start),
                                Elements: elems,
                        }
                }
                p.expect(token.RPAREN)
                // Single element in parens - just return the inner type
                return elems[0]
        }

        // Named type (possibly qualified: mod.Type)
        tok := p.current()
        if tok.Type == token.TYPE_IDENT || tok.Type == token.IDENT || tok.Type == token.OPTION_KW || tok.Type == token.RESULT_KW || tok.Type == token.OK || tok.Type == token.ERR || tok.Type == token.SOME || tok.Type == token.NONE_KW {
                name := tok.Literal
                p.advance()

                // Check for qualified type: name.Type
                if p.check(token.DOT) {
                        p.advance()
                        typeName := p.current().Literal
                        p.advance()
                        var args []ast.TypeExpr
                        if p.check(token.LBRACKET) {
                                args = p.parseTypeArgs()
                        }
                        return &ast.QualifiedType{
                                Span:      p.makeSpan(start),
                                Qualifier: name,
                                Name:      typeName,
                                Args:      args,
                        }
                }

                var args []ast.TypeExpr
                if p.check(token.LBRACKET) {
                        args = p.parseTypeArgs()
                }
                return &ast.NamedType{
                        Span: p.makeSpan(start),
                        Name: name,
                        Args: args,
                }
        }

        p.addError(fmt.Sprintf("expected type expression, got %s (%q)", tok.Type, tok.Literal))
        p.advance()
        return &ast.NamedType{Span: p.makeSpan(start), Name: "_error_"}
}

func (p *Parser) parseTypeArgs() []ast.TypeExpr {
        p.expect(token.LBRACKET)
        var args []ast.TypeExpr
        args = append(args, p.parseTypeExpr())
        for p.check(token.COMMA) {
                p.advance()
                args = append(args, p.parseTypeExpr())
        }
        p.expect(token.RBRACKET)
        return args
}

// --- Struct definitions ---

func (p *Parser) parseStructDef(vis ast.Visibility) *ast.StructDef {
        start := p.current().Pos
        p.expect(token.STRUCT)

        sd := &ast.StructDef{Visibility: vis}
        sd.Name = p.expectTypeName()
        sd.TypeParams = p.parseOptionalTypeParams()
        p.expect(token.COLON)
        p.expectNewline()
        p.expect(token.INDENT)

        for !p.check(token.DEDENT) && !p.check(token.EOF) {
                p.skipNewlines()
                if p.check(token.DEDENT) || p.check(token.EOF) {
                        break
                }
                field := p.parseFieldDef()
                if field != nil {
                        sd.Fields = append(sd.Fields, field)
                }
        }

        if p.check(token.DEDENT) {
                p.advance()
        }

        sd.Span = p.makeSpan(start)
        return sd
}

func (p *Parser) parseFieldDef() *ast.FieldDef {
        start := p.current().Pos
        fd := &ast.FieldDef{}

        if p.check(token.PUB) {
                fd.Visibility = ast.Public
                p.advance()
        }

        fd.Name = p.expect(token.IDENT).Literal
        p.expect(token.COLON)
        fd.TypeExpr = p.parseTypeExpr()

        if p.check(token.ASSIGN) {
                p.advance()
                fd.Default = p.parseExpr()
        }

        p.expectNewline()
        fd.Span = p.makeSpan(start)
        return fd
}

// --- Enum definitions ---

func (p *Parser) parseEnumDef(vis ast.Visibility) *ast.EnumDef {
        start := p.current().Pos
        p.expect(token.ENUM)

        ed := &ast.EnumDef{Visibility: vis}
        ed.Name = p.expectTypeName()
        ed.TypeParams = p.parseOptionalTypeParams()
        p.expect(token.COLON)
        p.expectNewline()
        p.expect(token.INDENT)

        for !p.check(token.DEDENT) && !p.check(token.EOF) {
                p.skipNewlines()
                if p.check(token.DEDENT) || p.check(token.EOF) {
                        break
                }
                variant := p.parseVariantDef()
                if variant != nil {
                        ed.Variants = append(ed.Variants, variant)
                }
        }

        if p.check(token.DEDENT) {
                p.advance()
        }

        ed.Span = p.makeSpan(start)
        return ed
}

func (p *Parser) parseVariantDef() *ast.VariantDef {
        start := p.current().Pos
        vd := &ast.VariantDef{}
        vd.Name = p.expectTypeName()

        if p.check(token.LPAREN) {
                p.advance()
                if !p.check(token.RPAREN) {
                        vd.Fields = append(vd.Fields, p.parseTypeExpr())
                        for p.check(token.COMMA) {
                                p.advance()
                                vd.Fields = append(vd.Fields, p.parseTypeExpr())
                        }
                }
                p.expect(token.RPAREN)
        }

        p.expectNewline()
        vd.Span = p.makeSpan(start)
        return vd
}

// --- Trait definitions ---

func (p *Parser) parseTraitDef(vis ast.Visibility) *ast.TraitDef {
        start := p.current().Pos
        p.expect(token.TRAIT)

        td := &ast.TraitDef{Visibility: vis}
        td.Name = p.expectTypeName()
        td.TypeParams = p.parseOptionalTypeParams()
        p.expect(token.COLON)
        p.expectNewline()
        p.expect(token.INDENT)

        for !p.check(token.DEDENT) && !p.check(token.EOF) {
                p.skipNewlines()
                if p.check(token.DEDENT) || p.check(token.EOF) {
                        break
                }
                // Parse function signature or definition
                member := p.parseFnDef(ast.Private)
                td.Members = append(td.Members, member)
        }

        if p.check(token.DEDENT) {
                p.advance()
        }

        td.Span = p.makeSpan(start)
        return td
}

// --- Impl blocks ---

func (p *Parser) parseImplBlock() *ast.ImplBlock {
        start := p.current().Pos
        p.expect(token.IMPL)

        ib := &ast.ImplBlock{}

        // "impl TraitName for Type:" or "impl Type:"
        firstName := p.expectTypeName()
        if p.check(token.FOR) {
                p.advance()
                ib.TraitName = firstName
                ib.TargetType = p.parseTypeExpr()
        } else {
                ib.TargetType = &ast.NamedType{
                        Span: p.makeSpan(start),
                        Name: firstName,
                }
        }

        p.expect(token.COLON)
        p.expectNewline()
        p.expect(token.INDENT)

        for !p.check(token.DEDENT) && !p.check(token.EOF) {
                p.skipNewlines()
                if p.check(token.DEDENT) || p.check(token.EOF) {
                        break
                }
                vis := ast.Private
                if p.check(token.PUB) {
                        vis = ast.Public
                        p.advance()
                }
                fn := p.parseFnDef(vis)
                ib.Methods = append(ib.Methods, fn)
        }

        if p.check(token.DEDENT) {
                p.advance()
        }

        ib.Span = p.makeSpan(start)
        return ib
}

// --- Spec blocks ---

func (p *Parser) parseSpecBlock() *ast.SpecBlock {
        start := p.current().Pos
        p.expect(token.SPEC)

        sb := &ast.SpecBlock{}
        sb.Name = p.expectTypeName()
        p.expect(token.COLON)
        p.expectNewline()
        p.expect(token.INDENT)

        for !p.check(token.DEDENT) && !p.check(token.EOF) {
                p.skipNewlines()
                if p.check(token.DEDENT) || p.check(token.EOF) {
                        break
                }

                switch p.peek() {
                case token.DOC_KW:
                        p.advance()
                        p.expect(token.COLON)
                        sb.Doc = p.expect(token.STRING_LIT).Literal
                        p.expectNewline()

                case token.INPUTS_KW:
                        p.advance()
                        p.expect(token.COLON)
                        p.expectNewline()
                        p.expect(token.INDENT)
                        for !p.check(token.DEDENT) && !p.check(token.EOF) {
                                p.skipNewlines()
                                if p.check(token.DEDENT) || p.check(token.EOF) {
                                        break
                                }
                                input := p.parseSpecInput()
                                if input != nil {
                                        sb.Inputs = append(sb.Inputs, input)
                                }
                        }
                        if p.check(token.DEDENT) {
                                p.advance()
                        }

                case token.GUARANTEES:
                        p.advance()
                        p.expect(token.COLON)
                        p.expectNewline()
                        p.expect(token.INDENT)
                        for !p.check(token.DEDENT) && !p.check(token.EOF) {
                                p.skipNewlines()
                                if p.check(token.DEDENT) || p.check(token.EOF) {
                                        break
                                }
                                g := p.parseSpecGuarantee()
                                if g != nil {
                                        sb.Guarantees = append(sb.Guarantees, g)
                                }
                        }
                        if p.check(token.DEDENT) {
                                p.advance()
                        }

                case token.EFFECTS_KW:
                        p.advance()
                        p.expect(token.COLON)
                        sb.Effects = p.parseEffectList()
                        p.expectNewline()

                case token.ERRORS:
                        p.advance()
                        p.expect(token.COLON)
                        // errors: {} (empty) or errors: NEWLINE INDENT ... DEDENT
                        if p.check(token.LBRACE) {
                                p.advance()
                                p.expect(token.RBRACE)
                                p.expectNewline()
                        } else {
                                p.expectNewline()
                                p.expect(token.INDENT)
                                for !p.check(token.DEDENT) && !p.check(token.EOF) {
                                        p.skipNewlines()
                                        if p.check(token.DEDENT) || p.check(token.EOF) {
                                                break
                                        }
                                        se := p.parseSpecError()
                                        if se != nil {
                                                sb.Errors = append(sb.Errors, se)
                                        }
                                }
                                if p.check(token.DEDENT) {
                                        p.advance()
                                }
                        }

                default:
                        p.addError(fmt.Sprintf("unexpected token in spec block: %s (%q)", p.peek(), p.current().Literal))
                        p.advance()
                }
        }

        if p.check(token.DEDENT) {
                p.advance()
        }

        sb.Span = p.makeSpan(start)
        return sb
}

func (p *Parser) parseSpecInput() *ast.SpecInput {
        start := p.current().Pos
        si := &ast.SpecInput{}

        si.Name = p.expect(token.IDENT).Literal
        p.expect(token.COLON)
        si.TypeExpr = p.parseTypeExpr()

        if p.check(token.ASSIGN) {
                p.advance()
                si.Default = p.parseExpr()
        }

        if p.check(token.MINUS) {
                p.advance()
                si.Description = p.expect(token.STRING_LIT).Literal
        }

        p.expectNewline()
        si.Span = p.makeSpan(start)
        return si
}

func (p *Parser) parseSpecGuarantee() *ast.SpecGuarantee {
        start := p.current().Pos
        p.expect(token.MINUS)
        cond := p.expect(token.STRING_LIT).Literal
        p.expectNewline()
        return &ast.SpecGuarantee{
                Span:      p.makeSpan(start),
                Condition: cond,
        }
}

func (p *Parser) parseSpecError() *ast.SpecError {
        start := p.current().Pos
        se := &ast.SpecError{}
        se.TypeName = p.expectTypeName()

        if p.check(token.LPAREN) {
                p.advance()
                if !p.check(token.RPAREN) {
                        se.Fields = append(se.Fields, p.parseTypeExpr())
                        for p.check(token.COMMA) {
                                p.advance()
                                se.Fields = append(se.Fields, p.parseTypeExpr())
                        }
                }
                p.expect(token.RPAREN)
        }

        if p.check(token.MINUS) {
                p.advance()
                se.Description = p.expect(token.STRING_LIT).Literal
        }

        p.expectNewline()
        se.Span = p.makeSpan(start)
        return se
}

func (p *Parser) parseEffectList() []string {
        var effects []string
        tok := p.current()
        if tok.Type == token.IDENT || tok.Type.IsKeyword() {
                effects = append(effects, tok.Literal)
                p.advance()
        }
        for p.check(token.COMMA) {
                p.advance()
                tok = p.current()
                if tok.Type == token.IDENT || tok.Type.IsKeyword() {
                        effects = append(effects, tok.Literal)
                        p.advance()
                }
        }
        return effects
}

// --- Function definitions ---

func (p *Parser) parseFnDef(vis ast.Visibility) *ast.FnDef {
        start := p.current().Pos
        p.expect(token.FN)

        fd := &ast.FnDef{Visibility: vis}

        // Function name
        tok := p.current()
        if tok.Type == token.IDENT || tok.Type == token.TYPE_IDENT || tok.Type.IsKeyword() {
                fd.Name = tok.Literal
                p.advance()
        } else {
                p.addError(fmt.Sprintf("expected function name, got %s", tok.Type))
        }

        fd.TypeParams = p.parseOptionalTypeParams()

        // Parameters
        p.expect(token.LPAREN)
        if !p.check(token.RPAREN) {
                fd.Params = p.parseParamList()
        }
        p.expect(token.RPAREN)

        // Return type
        if p.check(token.ARROW) {
                p.advance()
                fd.ReturnType = p.parseTypeExpr()
        }

        // Effects
        if p.check(token.WITH) {
                p.advance()
                fd.Effects = p.parseEffectList()
        }

        // Satisfies
        if p.check(token.SATISFIES) {
                p.advance()
                fd.Satisfies = p.expectTypeName()
        }

        // Body
        p.expect(token.COLON)
        p.expectNewline()
        p.expect(token.INDENT)
        fd.Body = p.parseStatementBlock()
        if p.check(token.DEDENT) {
                p.advance()
        }

        fd.Span = p.makeSpan(start)
        return fd
}

func (p *Parser) parseParamList() []*ast.Param {
        var params []*ast.Param
        params = append(params, p.parseParam())
        for p.check(token.COMMA) {
                p.advance()
                if p.check(token.RPAREN) {
                        break // trailing comma
                }
                params = append(params, p.parseParam())
        }
        return params
}

func (p *Parser) parseParam() *ast.Param {
        start := p.current().Pos
        param := &ast.Param{}

        tok := p.current()
        if tok.Type == token.IDENT || tok.Type.IsKeyword() {
                param.Name = tok.Literal
                p.advance()
        } else {
                p.addError("expected parameter name")
        }

        if p.check(token.COLON) {
                p.advance()
                param.TypeExpr = p.parseTypeExpr()
        }

        if p.check(token.ASSIGN) {
                p.advance()
                param.Default = p.parseExpr()
        }

        param.Span = p.makeSpan(start)
        return param
}

// --- Const definitions ---

func (p *Parser) parseConstDef(vis ast.Visibility) *ast.ConstDef {
        start := p.current().Pos
        p.expect(token.LET)

        cd := &ast.ConstDef{Visibility: vis}
        cd.Name = p.expect(token.IDENT).Literal

        if p.check(token.COLON) {
                p.advance()
                cd.TypeExpr = p.parseTypeExpr()
        }

        p.expect(token.ASSIGN)
        cd.Value = p.parseExpr()
        p.expectNewline()

        cd.Span = p.makeSpan(start)
        return cd
}

// --- Test blocks ---

func (p *Parser) parseTestBlock() *ast.TestBlock {
        start := p.current().Pos
        p.expect(token.TEST)

        tb := &ast.TestBlock{}
        tb.Name = p.expect(token.STRING_LIT).Literal
        p.expect(token.COLON)
        p.expectNewline()
        p.expect(token.INDENT)
        tb.Body = p.parseStatementBlock()
        if p.check(token.DEDENT) {
                p.advance()
        }

        tb.Span = p.makeSpan(start)
        return tb
}

// --- Statements ---

func (p *Parser) parseStatementBlock() []Statement {
        var stmts []Statement
        for !p.check(token.DEDENT) && !p.check(token.EOF) {
                p.skipNewlines()
                if p.check(token.DEDENT) || p.check(token.EOF) {
                        break
                }
                stmt := p.parseStatement()
                if stmt != nil {
                        stmts = append(stmts, stmt)
                }
        }
        return stmts
}

// Statement type alias for ast.Statement
type Statement = ast.Statement

func (p *Parser) parseStatement() ast.Statement {
        switch p.peek() {
        case token.LET:
                return p.parseLetStmt()
        case token.RETURN:
                return p.parseReturnStmt()
        case token.IF:
                return p.parseIfStmt()
        case token.MATCH:
                return p.parseMatchStmtOrExpr()
        case token.FOR:
                return p.parseForStmt()
        case token.WHILE:
                return p.parseWhileStmt()
        case token.BREAK:
                start := p.current().Pos
                p.advance()
                p.expectNewline()
                return &ast.BreakStmt{Span: p.makeSpan(start)}
        case token.CONTINUE:
                start := p.current().Pos
                p.advance()
                p.expectNewline()
                return &ast.ContinueStmt{Span: p.makeSpan(start)}
        case token.ASSERT:
                return p.parseAssertStmt()
        case token.WITH:
                return p.parseWithStmt()
        default:
                return p.parseExprOrAssignStmt()
        }
}

func (p *Parser) parseLetStmt() ast.Statement {
        start := p.current().Pos
        p.expect(token.LET)

        mutable := false
        if p.check(token.MUT) {
                mutable = true
                p.advance()
        }

        // Check for tuple destructuring: let (x, y) = expr
        if p.check(token.LPAREN) {
                p.advance()
                var names []string
                for !p.check(token.RPAREN) && !p.check(token.EOF) {
                        tok := p.current()
                        if tok.Type == token.IDENT || tok.Type.IsKeyword() {
                                names = append(names, tok.Literal)
                                p.advance()
                        } else {
                                p.addError("expected variable name in tuple destructuring")
                                p.advance()
                        }
                        if p.check(token.COMMA) {
                                p.advance()
                        }
                }
                p.expect(token.RPAREN)
                p.expect(token.ASSIGN)
                value := p.parseExpr()
                p.expectNewline()
                return &ast.LetTupleDestructure{
                        Span:    p.makeSpan(start),
                        Names:   names,
                        Mutable: mutable,
                        Value:   value,
                }
        }

        ls := &ast.LetStmt{}
        ls.Mutable = mutable

        tok := p.current()
        if tok.Type == token.IDENT || tok.Type.IsKeyword() {
                ls.Name = tok.Literal
                p.advance()
        } else {
                // Could be a wildcard pattern: let _ = expr
                if tok.Literal == "_" {
                        ls.Name = "_"
                        p.advance()
                } else {
                        p.addError("expected variable name in let statement")
                }
        }

        if p.check(token.COLON) {
                p.advance()
                ls.TypeHint = p.parseTypeExpr()
        }

        p.expect(token.ASSIGN)
        ls.Value = p.parseExpr()
        p.expectNewline()

        ls.Span = p.makeSpan(start)
        return ls
}

func (p *Parser) parseReturnStmt() *ast.ReturnStmt {
        start := p.current().Pos
        p.expect(token.RETURN)

        rs := &ast.ReturnStmt{}
        if !p.check(token.NEWLINE) && !p.check(token.DEDENT) && !p.check(token.EOF) {
                rs.Value = p.parseExpr()
        }
        p.expectNewline()

        rs.Span = p.makeSpan(start)
        return rs
}

func (p *Parser) parseIfStmt() *ast.IfStmt {
        start := p.current().Pos
        p.expect(token.IF)

        is := &ast.IfStmt{}
        is.Condition = p.parseExpr()
        p.expect(token.COLON)
        p.expectNewline()
        p.expect(token.INDENT)
        is.ThenBody = p.parseStatementBlock()
        if p.check(token.DEDENT) {
                p.advance()
        }

        for p.check(token.ELIF) {
                elifStart := p.current().Pos
                p.advance()
                cond := p.parseExpr()
                p.expect(token.COLON)
                p.expectNewline()
                p.expect(token.INDENT)
                body := p.parseStatementBlock()
                if p.check(token.DEDENT) {
                        p.advance()
                }
                is.ElifClauses = append(is.ElifClauses, &ast.ElifClause{
                        Span:      p.makeSpan(elifStart),
                        Condition: cond,
                        Body:      body,
                })
        }

        if p.check(token.ELSE) {
                p.advance()
                p.expect(token.COLON)
                p.expectNewline()
                p.expect(token.INDENT)
                is.ElseBody = p.parseStatementBlock()
                if p.check(token.DEDENT) {
                        p.advance()
                }
        }

        is.Span = p.makeSpan(start)
        return is
}

// parseMatchStmtOrExpr determines whether to parse a match statement (case-based)
// or a match expression (arrow-based) by peeking ahead after the subject.
func (p *Parser) parseMatchStmtOrExpr() ast.Statement {
        // Save position to peek ahead
        saved := p.pos
        savedErrors := len(p.errors)
        savedBlockExpr := p.blockExprJustEnded
        p.advance() // consume 'match'

        // Skip the subject expression tokens until we hit ':'
        depth := 0
        for !p.check(token.EOF) {
                if p.check(token.COLON) && depth == 0 {
                        break
                }
                if p.check(token.LPAREN) || p.check(token.LBRACKET) {
                        depth++
                }
                if p.check(token.RPAREN) || p.check(token.RBRACKET) {
                        depth--
                }
                p.advance()
        }

        if p.check(token.COLON) {
                p.advance() // skip ':'
        }
        // Skip newlines
        for p.check(token.NEWLINE) {
                p.advance()
        }
        // Skip indent
        if p.check(token.INDENT) {
                p.advance()
        }
        // Skip newlines after indent
        for p.check(token.NEWLINE) {
                p.advance()
        }

        // Check if the first token is 'case' -> MatchStmt, otherwise -> MatchExpr
        isStmt := p.check(token.CASE)

        // Restore position
        p.pos = saved
        p.errors = p.errors[:savedErrors]
        p.blockExprJustEnded = savedBlockExpr

        if isStmt {
                return p.parseMatchStmt()
        }
        // Parse as expression statement wrapping a match expression
        start := p.current().Pos
        expr := p.parseMatchExpr()
        return &ast.ExprStmt{Span: p.makeSpan(start), Expr: expr}
}

func (p *Parser) parseMatchStmt() *ast.MatchStmt {
        start := p.current().Pos
        p.expect(token.MATCH)

        ms := &ast.MatchStmt{}
        ms.Subject = p.parseExpr()
        p.expect(token.COLON)
        p.expectNewline()
        p.expect(token.INDENT)

        for !p.check(token.DEDENT) && !p.check(token.EOF) {
                p.skipNewlines()
                if p.check(token.DEDENT) || p.check(token.EOF) {
                        break
                }
                cc := p.parseCaseClause()
                if cc != nil {
                        ms.Cases = append(ms.Cases, cc)
                }
        }

        if p.check(token.DEDENT) {
                p.advance()
        }

        ms.Span = p.makeSpan(start)
        return ms
}

func (p *Parser) parseCaseClause() *ast.CaseClause {
        start := p.current().Pos
        p.expect(token.CASE)

        cc := &ast.CaseClause{}
        cc.Pattern = p.parsePattern()

        if p.check(token.IF) {
                p.advance()
                cc.Guard = p.parseExpr()
        }

        p.expect(token.COLON)
        p.expectNewline()
        p.expect(token.INDENT)
        cc.Body = p.parseStatementBlock()
        if p.check(token.DEDENT) {
                p.advance()
        }

        cc.Span = p.makeSpan(start)
        return cc
}

func (p *Parser) parseForStmt() *ast.ForStmt {
        start := p.current().Pos
        p.expect(token.FOR)

        fs := &ast.ForStmt{}

        tok := p.current()
        if tok.Type == token.IDENT || tok.Type.IsKeyword() {
                fs.Variable = tok.Literal
                p.advance()
        } else {
                p.addError("expected variable name in for statement")
        }

        p.expect(token.IN)
        fs.Iterable = p.parseExpr()
        p.expect(token.COLON)
        p.expectNewline()
        p.expect(token.INDENT)
        fs.Body = p.parseStatementBlock()
        if p.check(token.DEDENT) {
                p.advance()
        }

        fs.Span = p.makeSpan(start)
        return fs
}

func (p *Parser) parseWhileStmt() *ast.WhileStmt {
        start := p.current().Pos
        p.expect(token.WHILE)

        ws := &ast.WhileStmt{}
        ws.Condition = p.parseExpr()
        p.expect(token.COLON)
        p.expectNewline()
        p.expect(token.INDENT)
        ws.Body = p.parseStatementBlock()
        if p.check(token.DEDENT) {
                p.advance()
        }

        ws.Span = p.makeSpan(start)
        return ws
}

func (p *Parser) parseAssertStmt() *ast.AssertStmt {
        start := p.current().Pos
        p.expect(token.ASSERT)

        as := &ast.AssertStmt{}
        as.Condition = p.parseExpr()

        if p.check(token.COMMA) {
                p.advance()
                as.Message = p.expect(token.STRING_LIT).Literal
        }

        p.expectNewline()
        as.Span = p.makeSpan(start)
        return as
}

func (p *Parser) parseWithStmt() *ast.WithStmt {
        start := p.current().Pos
        p.expect(token.WITH)

        ws := &ast.WithStmt{}

        // Parse bindings
        binding := p.parseWithBinding()
        ws.Bindings = append(ws.Bindings, binding)
        for p.check(token.COMMA) {
                p.advance()
                binding = p.parseWithBinding()
                ws.Bindings = append(ws.Bindings, binding)
        }

        p.expect(token.COLON)
        p.expectNewline()
        p.expect(token.INDENT)
        ws.Body = p.parseStatementBlock()
        if p.check(token.DEDENT) {
                p.advance()
        }

        ws.Span = p.makeSpan(start)
        return ws
}

func (p *Parser) parseWithBinding() *ast.WithBinding {
        wb := &ast.WithBinding{}
        wb.Expr = p.parseExpr()
        if p.check(token.AS) {
                p.advance()
                wb.Alias = p.expect(token.IDENT).Literal
        }
        return wb
}

func (p *Parser) parseExprOrAssignStmt() ast.Statement {
        start := p.current().Pos
        expr := p.parseExpr()

        // Check if this is an assignment
        if p.check(token.ASSIGN) {
                p.advance()
                value := p.parseExpr()
                p.expectNewline()
                return &ast.AssignStmt{
                        Span:   p.makeSpan(start),
                        Target: expr,
                        Value:  value,
                }
        }

        p.expectNewline()
        return &ast.ExprStmt{
                Span: p.makeSpan(start),
                Expr: expr,
        }
}

// --- Patterns ---

func (p *Parser) parsePattern() ast.Pattern {
        start := p.current().Pos

        // Wildcard
        if p.check(token.IDENT) && p.current().Literal == "_" {
                p.advance()
                return &ast.WildcardPattern{Span: p.makeSpan(start)}
        }

        // Literal patterns
        if p.check(token.INT_LIT) {
                tok := p.advance()
                return &ast.LiteralPattern{Span: p.makeSpan(start), Value: tok.Literal, Kind: token.INT_LIT}
        }
        if p.check(token.FLOAT_LIT) {
                tok := p.advance()
                return &ast.LiteralPattern{Span: p.makeSpan(start), Value: tok.Literal, Kind: token.FLOAT_LIT}
        }
        if p.check(token.STRING_LIT) {
                tok := p.advance()
                return &ast.LiteralPattern{Span: p.makeSpan(start), Value: tok.Literal, Kind: token.STRING_LIT}
        }
        if p.check(token.BOOL_LIT) {
                tok := p.advance()
                return &ast.LiteralPattern{Span: p.makeSpan(start), Value: tok.Literal, Kind: token.BOOL_LIT}
        }
        if p.check(token.NONE_LIT) {
                tok := p.advance()
                return &ast.LiteralPattern{Span: p.makeSpan(start), Value: tok.Literal, Kind: token.NONE_LIT}
        }

        // Constructor or binding pattern
        if p.check(token.TYPE_IDENT) || p.check(token.OK) || p.check(token.ERR) || p.check(token.SOME) || p.check(token.NONE_KW) {
                name := p.current().Literal
                p.advance()

                // Check for dotted constructor: Type.Variant
                if p.check(token.DOT) {
                        p.advance()
                        variantName := p.current().Literal
                        p.advance()
                        name = name + "." + variantName
                }

                if p.check(token.LPAREN) {
                        p.advance()
                        var fields []ast.Pattern
                        if !p.check(token.RPAREN) {
                                fields = append(fields, p.parsePattern())
                                for p.check(token.COMMA) {
                                        p.advance()
                                        fields = append(fields, p.parsePattern())
                                }
                        }
                        p.expect(token.RPAREN)
                        return &ast.ConstructorPattern{
                                Span:     p.makeSpan(start),
                                TypeName: name,
                                Fields:   fields,
                        }
                }

                // Just a constructor without args
                return &ast.ConstructorPattern{
                        Span:     p.makeSpan(start),
                        TypeName: name,
                }
        }

        // Spread pattern: ...rest
        if p.check(token.DOTDOTDOT) {
                p.advance()
                if !p.check(token.IDENT) {
                        p.addError("expected identifier after '...'")
                        return &ast.WildcardPattern{Span: p.makeSpan(start)}
                }
                name := p.current().Literal
                p.advance()
                return &ast.SpreadPattern{Span: p.makeSpan(start), Name: name}
        }

        // Binding pattern (identifier)
        if p.check(token.IDENT) {
                name := p.current().Literal
                p.advance()
                return &ast.BindingPattern{Span: p.makeSpan(start), Name: name}
        }

        // List pattern
        if p.check(token.LBRACKET) {
                p.advance()
                var elems []ast.Pattern
                if !p.check(token.RBRACKET) {
                        elems = append(elems, p.parsePattern())
                        for p.check(token.COMMA) {
                                p.advance()
                                elems = append(elems, p.parsePattern())
                        }
                }
                p.expect(token.RBRACKET)
                return &ast.ListPattern{Span: p.makeSpan(start), Elements: elems}
        }

        // Tuple pattern
        if p.check(token.LPAREN) {
                p.advance()
                var elems []ast.Pattern
                elems = append(elems, p.parsePattern())
                for p.check(token.COMMA) {
                        p.advance()
                        elems = append(elems, p.parsePattern())
                }
                p.expect(token.RPAREN)
                return &ast.TuplePattern{Span: p.makeSpan(start), Elements: elems}
        }

        p.addError(fmt.Sprintf("expected pattern, got %s (%q)", p.peek(), p.current().Literal))
        p.advance()
        return &ast.WildcardPattern{Span: p.makeSpan(start)}
}

// --- Expressions (precedence climbing) ---

func (p *Parser) parseExpr() ast.Expr {
        return p.parsePipelineExpr()
}

func (p *Parser) parsePipelineExpr() ast.Expr {
        left := p.parseOrExpr()
        for p.check(token.PIPE_GT) {
                p.advance()
                right := p.parseOrExpr()
                left = &ast.PipelineExpr{
                        Span:  token.Span{File: p.file, Start: left.GetSpan().Start, End: right.GetSpan().End},
                        Left:  left,
                        Right: right,
                }
        }
        return left
}

func (p *Parser) parseOrExpr() ast.Expr {
        left := p.parseAndExpr()
        for p.check(token.OR) {
                op := p.advance()
                right := p.parseAndExpr()
                left = &ast.BinaryOp{
                        Span:  token.Span{File: p.file, Start: left.GetSpan().Start, End: right.GetSpan().End},
                        Op:    op.Literal,
                        Left:  left,
                        Right: right,
                }
        }
        return left
}

func (p *Parser) parseAndExpr() ast.Expr {
        left := p.parseNotExpr()
        for p.check(token.AND) {
                op := p.advance()
                right := p.parseNotExpr()
                left = &ast.BinaryOp{
                        Span:  token.Span{File: p.file, Start: left.GetSpan().Start, End: right.GetSpan().End},
                        Op:    op.Literal,
                        Left:  left,
                        Right: right,
                }
        }
        return left
}

func (p *Parser) parseNotExpr() ast.Expr {
        if p.check(token.NOT) {
                start := p.current().Pos
                p.advance()
                operand := p.parseNotExpr()
                return &ast.UnaryOp{
                        Span:    p.makeSpan(start),
                        Op:      "not",
                        Operand: operand,
                }
        }
        return p.parseComparison()
}

func (p *Parser) parseComparison() ast.Expr {
        left := p.parseAddition()
        for p.match(token.EQ, token.NEQ, token.LT, token.GT, token.LTE, token.GTE, token.IS, token.IN) {
                op := p.advance()
                opStr := op.Literal
                // Handle "is not" as a compound operator
                if op.Type == token.IS && p.check(token.NOT) {
                        p.advance()
                        opStr = "is not"
                }
                // Handle "not in" as a compound operator
                if op.Type == token.NOT && p.check(token.IN) {
                        p.advance()
                        opStr = "not in"
                }
                right := p.parseAddition()
                left = &ast.BinaryOp{
                        Span:  token.Span{File: p.file, Start: left.GetSpan().Start, End: right.GetSpan().End},
                        Op:    opStr,
                        Left:  left,
                        Right: right,
                }
        }
        return left
}

func (p *Parser) parseAddition() ast.Expr {
        left := p.parseMultiplication()
        for p.match(token.PLUS, token.MINUS) {
                op := p.advance()
                right := p.parseMultiplication()
                left = &ast.BinaryOp{
                        Span:  token.Span{File: p.file, Start: left.GetSpan().Start, End: right.GetSpan().End},
                        Op:    op.Literal,
                        Left:  left,
                        Right: right,
                }
        }
        return left
}

func (p *Parser) parseMultiplication() ast.Expr {
        left := p.parseUnary()
        for p.match(token.STAR, token.SLASH, token.PERCENT) {
                op := p.advance()
                right := p.parseUnary()
                left = &ast.BinaryOp{
                        Span:  token.Span{File: p.file, Start: left.GetSpan().Start, End: right.GetSpan().End},
                        Op:    op.Literal,
                        Left:  left,
                        Right: right,
                }
        }
        return left
}

func (p *Parser) parseUnary() ast.Expr {
        if p.check(token.MINUS) {
                start := p.current().Pos
                p.advance()
                operand := p.parseUnary()
                return &ast.UnaryOp{
                        Span:    p.makeSpan(start),
                        Op:      "-",
                        Operand: operand,
                }
        }
        return p.parsePower()
}

func (p *Parser) parsePower() ast.Expr {
        base := p.parsePostfix()
        if p.check(token.POWER) {
                p.advance()
                exp := p.parseUnary() // right-associative
                return &ast.BinaryOp{
                        Span:  token.Span{File: p.file, Start: base.GetSpan().Start, End: exp.GetSpan().End},
                        Op:    "**",
                        Left:  base,
                        Right: exp,
                }
        }
        return base
}

func (p *Parser) parsePostfix() ast.Expr {
        expr := p.parsePrimary()

        for {
                if p.check(token.QUESTION_DOT) {
                        p.advance()
                        field := p.current().Literal
                        p.advance()
                        expr = &ast.OptionalFieldAccess{
                                Span:   token.Span{File: p.file, Start: expr.GetSpan().Start, End: p.current().Pos},
                                Object: expr,
                                Field:  field,
                        }
                } else if p.check(token.DOT) {
                        p.advance()
                        field := p.current().Literal
                        p.advance()
                        expr = &ast.FieldAccess{
                                Span:   token.Span{File: p.file, Start: expr.GetSpan().Start, End: p.current().Pos},
                                Object: expr,
                                Field:  field,
                        }
                } else if p.check(token.LPAREN) {
                        p.advance()
                        args := p.parseArgList()
                        end := p.expect(token.RPAREN)
                        expr = &ast.CallExpr{
                                Span:   token.Span{File: p.file, Start: expr.GetSpan().Start, End: end.Pos},
                                Callee: expr,
                                Args:   args,
                        }
                } else if p.check(token.LBRACKET) {
                        p.advance()
                        index := p.parseExpr()
                        p.expect(token.RBRACKET)
                        expr = &ast.IndexExpr{
                                Span:   token.Span{File: p.file, Start: expr.GetSpan().Start, End: p.current().Pos},
                                Object: expr,
                                Index:  index,
                        }
                } else if p.check(token.QUESTION) {
                        p.advance()
                        expr = &ast.OptionPropagate{
                                Span: token.Span{File: p.file, Start: expr.GetSpan().Start, End: p.current().Pos},
                                Expr: expr,
                        }
                } else if p.check(token.BANG) {
                        p.advance()
                        // Unwrap operator - treat as postfix unary
                        expr = &ast.UnaryOp{
                                Span:    token.Span{File: p.file, Start: expr.GetSpan().Start, End: p.current().Pos},
                                Op:      "!",
                                Operand: expr,
                        }
                } else {
                        break
                }
        }

        return expr
}

func (p *Parser) parseArgList() []*ast.Arg {
        var args []*ast.Arg
        if p.check(token.RPAREN) {
                return args
        }

        args = append(args, p.parseArg())
        for p.check(token.COMMA) {
                p.advance()
                if p.check(token.RPAREN) {
                        break // trailing comma
                }
                args = append(args, p.parseArg())
        }
        return args
}

func (p *Parser) parseArg() *ast.Arg {
        start := p.current().Pos
        arg := &ast.Arg{}

        // Check for named argument: name: value
        if (p.check(token.IDENT) || p.peek().IsKeyword()) && p.peekAt(1) == token.COLON {
                arg.Name = p.current().Literal
                p.advance()
                p.advance() // consume ':'
        }

        arg.Value = p.parseExpr()
        arg.Span = p.makeSpan(start)
        return arg
}

func (p *Parser) parsePrimary() ast.Expr {
        start := p.current().Pos

        switch p.peek() {
        case token.INT_LIT:
                tok := p.advance()
                return &ast.IntLiteral{Span: p.makeSpan(start), Value: tok.Literal}

        case token.FLOAT_LIT:
                tok := p.advance()
                return &ast.FloatLiteral{Span: p.makeSpan(start), Value: tok.Literal}

        case token.STRING_LIT:
                tok := p.advance()
                sl := &ast.StringLiteral{Span: p.makeSpan(start), Value: tok.Literal}
                // Check for string interpolation: parse {expr} parts
                if strings.Contains(tok.Literal, "{") {
                        sl.Parts = parseStringInterpolation(tok.Literal, p.file)
                }
                return sl

        case token.BOOL_LIT:
                tok := p.advance()
                return &ast.BoolLiteral{Span: p.makeSpan(start), Value: tok.Literal == "true"}

        case token.NONE_LIT:
                p.advance()
                return &ast.NoneLiteral{Span: p.makeSpan(start)}

        case token.IF:
                // Could be if expression: if cond then a else b
                // We need to check if the next token after condition is "then"
                return p.parseIfExprOrPrimary()

        case token.MATCH:
                return p.parseMatchExpr()

        case token.PIPE:
                return p.parseLambda()

        case token.LPAREN:
                p.advance()
                // Empty tuple: ()
                if p.check(token.RPAREN) {
                        p.advance()
                        return &ast.TupleLiteral{Span: p.makeSpan(start), Elements: nil}
                }
                expr := p.parseExpr()
                // If followed by comma, it's a tuple
                if p.check(token.COMMA) {
                        elements := []ast.Expr{expr}
                        p.advance() // consume first comma
                        // Parse remaining elements (if any)
                        for !p.check(token.RPAREN) && !p.check(token.EOF) {
                                elements = append(elements, p.parseExpr())
                                if !p.check(token.COMMA) {
                                        break
                                }
                                p.advance() // consume comma
                        }
                        p.expect(token.RPAREN)
                        return &ast.TupleLiteral{Span: p.makeSpan(start), Elements: elements}
                }
                // Otherwise it's a grouped expression
                p.expect(token.RPAREN)
                return expr

        case token.LBRACKET:
                return p.parseListExpr()

        case token.LBRACE:
                return p.parseMapExpr()

        case token.TYPE_IDENT, token.OPTION_KW, token.RESULT_KW, token.OK, token.ERR, token.SOME, token.NONE_KW:
                name := p.current().Literal
                p.advance()

                // Check for struct construction: TypeName(field: value, ...)
                if p.check(token.LPAREN) {
                        return p.parseStructOrCallExpr(name, start)
                }

                // Check for qualified access: Type.member
                if p.check(token.DOT) {
                        p.advance()
                        member := p.current().Literal
                        p.advance()
                        fa := &ast.FieldAccess{
                                Span:   p.makeSpan(start),
                                Object: &ast.Identifier{Span: p.makeSpan(start), Name: name},
                                Field:  member,
                        }
                        return fa
                }

                return &ast.Identifier{Span: p.makeSpan(start), Name: name}

        case token.IDENT:
                name := p.current().Literal
                p.advance()
                return &ast.Identifier{Span: p.makeSpan(start), Name: name}

        default:
                // Try to handle keywords used as identifiers in some contexts
                if p.peek().IsKeyword() {
                        name := p.current().Literal
                        p.advance()
                        return &ast.Identifier{Span: p.makeSpan(start), Name: name}
                }

                p.addError(fmt.Sprintf("expected expression, got %s (%q)", p.peek(), p.current().Literal))
                p.advance()
                return &ast.Identifier{Span: p.makeSpan(start), Name: "_error_"}
        }
}

func (p *Parser) parseIfExprOrPrimary() ast.Expr {
        // This is used when we see 'if' in expression position.
        // We don't use this for if statements—those are parsed in parseStatement.
        // For inline if expressions: if cond then a else b
        start := p.current().Pos
        p.advance() // consume 'if'
        cond := p.parseExpr()

        if p.check(token.THEN) {
                p.advance()
                thenExpr := p.parseExpr()
                p.expect(token.ELSE)
                elseExpr := p.parseExpr()
                return &ast.IfExpr{
                        Span:      p.makeSpan(start),
                        Condition: cond,
                        ThenExpr:  thenExpr,
                        ElseExpr:  elseExpr,
                }
        }

        // If no 'then', just return the condition (shouldn't normally happen)
        return cond
}

func (p *Parser) parseMatchExpr() ast.Expr {
        start := p.current().Pos
        p.advance() // consume 'match'

        subject := p.parseExpr()
        p.expect(token.COLON)
        p.expectNewline()
        p.expect(token.INDENT)

        var arms []*ast.MatchArm
        for !p.check(token.DEDENT) && !p.check(token.EOF) {
                p.skipNewlines()
                if p.check(token.DEDENT) || p.check(token.EOF) {
                        break
                }
                armStart := p.current().Pos
                pattern := p.parsePattern()
                p.expect(token.ARROW)
                body := p.parseExpr()
                arms = append(arms, &ast.MatchArm{
                        Span:    p.makeSpan(armStart),
                        Pattern: pattern,
                        Body:    body,
                })
                // Consume newline after arm if present
                p.skipNewlines()
        }

        // Consume the DEDENT that closes the match body.
        // Mark that we just exited a block expression, so the enclosing
        // statement doesn't need an explicit newline terminator.
        if p.check(token.DEDENT) {
                p.advance()
                p.blockExprJustEnded = true
        }

        return &ast.MatchExpr{
                Span:    p.makeSpan(start),
                Subject: subject,
                Arms:    arms,
        }
}

func (p *Parser) parseLambda() ast.Expr {
        start := p.current().Pos
        p.expect(token.PIPE)

        var params []*ast.Param
        if !p.check(token.PIPE) {
                params = p.parseParamList()
        }
        p.expect(token.PIPE)

        lambda := &ast.Lambda{Params: params}

        if p.check(token.ARROW) {
                p.advance()
                lambda.Body = p.parseExpr()
        } else if p.check(token.COLON) {
                p.advance()
                p.expectNewline()
                p.expect(token.INDENT)
                lambda.Block = p.parseStatementBlock()
                if p.check(token.DEDENT) {
                        p.advance()
                }
        }

        lambda.Span = p.makeSpan(start)
        return lambda
}

func (p *Parser) parseListExpr() ast.Expr {
        start := p.current().Pos
        p.expect(token.LBRACKET)

        if p.check(token.RBRACKET) {
                p.advance()
                return &ast.ListExpr{Span: p.makeSpan(start)}
        }

        first := p.parseExpr()

        // Check for list comprehension: [expr for var in iterable]
        if p.check(token.FOR) {
                p.advance()
                variable := p.expect(token.IDENT).Literal
                p.expect(token.IN)
                iterable := p.parseExpr()
                var filter ast.Expr
                if p.check(token.IF) {
                        p.advance()
                        filter = p.parseExpr()
                }
                p.expect(token.RBRACKET)
                return &ast.ListComp{
                        Span:     p.makeSpan(start),
                        Element:  first,
                        Variable: variable,
                        Iterable: iterable,
                        Filter:   filter,
                }
        }

        // Regular list
        elems := []ast.Expr{first}
        for p.check(token.COMMA) {
                p.advance()
                if p.check(token.RBRACKET) {
                        break
                }
                elems = append(elems, p.parseExpr())
        }
        p.expect(token.RBRACKET)
        return &ast.ListExpr{Span: p.makeSpan(start), Elements: elems}
}

func (p *Parser) parseMapExpr() ast.Expr {
        start := p.current().Pos
        p.expect(token.LBRACE)

        if p.check(token.RBRACE) {
                p.advance()
                return &ast.MapExpr{Span: p.makeSpan(start)}
        }

        var entries []*ast.MapEntry
        key := p.parseExpr()
        p.expect(token.COLON)
        value := p.parseExpr()
        entries = append(entries, &ast.MapEntry{Key: key, Value: value})

        for p.check(token.COMMA) {
                p.advance()
                if p.check(token.RBRACE) {
                        break
                }
                key = p.parseExpr()
                p.expect(token.COLON)
                value = p.parseExpr()
                entries = append(entries, &ast.MapEntry{Key: key, Value: value})
        }

        p.expect(token.RBRACE)
        return &ast.MapExpr{Span: p.makeSpan(start), Entries: entries}
}

func (p *Parser) parseStructOrCallExpr(name string, start token.Position) ast.Expr {
        p.advance() // consume (

        // Check if this is a struct construction (has named args) or function call
        if p.check(token.RPAREN) {
                end := p.advance()
                return &ast.CallExpr{
                        Span:   token.Span{File: p.file, Start: start, End: end.Pos},
                        Callee: &ast.Identifier{Span: token.Span{File: p.file, Start: start}, Name: name},
                }
        }

        // Try to determine if first arg is named
        args := p.parseArgList()
        p.expect(token.RPAREN)

        // If all args are named, treat as struct expr
        allNamed := true
        for _, arg := range args {
                if arg.Name == "" {
                        allNamed = false
                        break
                }
        }

        if allNamed && len(args) > 0 {
                var fields []*ast.FieldInit
                for _, arg := range args {
                        fields = append(fields, &ast.FieldInit{
                                Name:  arg.Name,
                                Value: arg.Value,
                        })
                }
                return &ast.StructExpr{
                        Span:     p.makeSpan(start),
                        TypeName: name,
                        Fields:   fields,
                }
        }

        // Otherwise, function call
        return &ast.CallExpr{
                Span:   p.makeSpan(start),
                Callee: &ast.Identifier{Span: token.Span{File: p.file, Start: start}, Name: name},
                Args:   args,
        }
}

// parseStringInterpolation extracts parts from a string with {expr} interpolation.
// It properly lexes and parses embedded expressions inside {braces}.
func parseStringInterpolation(s string, file string) []ast.StringPart {
        var parts []ast.StringPart
        var current strings.Builder
        i := 0
        for i < len(s) {
                if s[i] == '{' {
                        if current.Len() > 0 {
                                parts = append(parts, ast.StringPart{Text: current.String()})
                                current.Reset()
                        }
                        // Find matching }
                        j := i + 1
                        depth := 1
                        for j < len(s) && depth > 0 {
                                if s[j] == '{' {
                                        depth++
                                } else if s[j] == '}' {
                                        depth--
                                }
                                j++
                        }
                        exprStr := strings.TrimSpace(s[i+1 : j-1])

                        // Parse the expression using the real lexer and parser
                        expr := parseEmbeddedExpr(exprStr, file)
                        parts = append(parts, ast.StringPart{
                                IsExpr: true,
                                Text:   exprStr,
                                Expr:   expr,
                        })
                        i = j
                } else {
                        current.WriteByte(s[i])
                        i++
                }
        }
        if current.Len() > 0 {
                parts = append(parts, ast.StringPart{Text: current.String()})
        }
        return parts
}

// parseEmbeddedExpr lexes and parses a single expression string.
func parseEmbeddedExpr(exprStr string, file string) ast.Expr {
        l := lexer.New(exprStr, file)
        tokens, _ := l.Tokenize()
        if len(tokens) == 0 {
                return &ast.Identifier{Name: exprStr}
        }
        p := New(tokens, file)
        expr := p.parseExpr()
        if expr == nil {
                return &ast.Identifier{Name: exprStr}
        }
        return expr
}