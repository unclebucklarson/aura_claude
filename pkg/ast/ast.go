// Package ast defines all AST node types for the Aura language.
package ast

import "github.com/unclebucklarson/aura/pkg/token"

// Node is the interface implemented by all AST nodes.
type Node interface {
        nodeType() string
        GetSpan() token.Span
}

// Visibility represents pub/private visibility.
type Visibility int

const (
        Private Visibility = iota
        Public
)

func (v Visibility) String() string {
        if v == Public {
                return "pub"
        }
        return "private"
}

// --- Top-Level Nodes ---

// Module is the root AST node for an Aura source file.
type Module struct {
        Span    token.Span
        Name    *QualifiedName
        Imports []*ImportNode
        Items   []TopLevelItem
        // Comments that appear before the module declaration or at top of file
        LeadingComments []Comment
}

func (n *Module) nodeType() string     { return "Module" }
func (n *Module) GetSpan() token.Span  { return n.Span }

// Comment represents a comment in source code.
type Comment struct {
        Span  token.Span
        Text  string
        IsDoc bool
}

// QualifiedName represents a dotted name like "std.time".
type QualifiedName struct {
        Span  token.Span
        Parts []string
}

func (n *QualifiedName) String() string {
        result := ""
        for i, p := range n.Parts {
                if i > 0 {
                        result += "."
                }
                result += p
        }
        return result
}
func (n *QualifiedName) nodeType() string    { return "QualifiedName" }
func (n *QualifiedName) GetSpan() token.Span { return n.Span }

// ImportNode represents an import statement.
type ImportNode struct {
        Span  token.Span
        Path  *QualifiedName
        Alias string   // for "import X as Y"
        Names []string // for "from X import a, b"; nil = whole module, ["*"] = wildcard
        Comments []Comment
}

func (n *ImportNode) nodeType() string    { return "ImportNode" }
func (n *ImportNode) GetSpan() token.Span { return n.Span }

// TopLevelItem is an interface for items that appear at module level.
type TopLevelItem interface {
        Node
        isTopLevel()
}

// --- Type Definitions ---

// TypeDef represents "type Name = TypeExpr".
type TypeDef struct {
        Span       token.Span
        Name       string
        TypeParams []string
        Visibility Visibility
        Body       TypeExpr
        Comments   []Comment
}

func (n *TypeDef) nodeType() string    { return "TypeDef" }
func (n *TypeDef) GetSpan() token.Span { return n.Span }
func (n *TypeDef) isTopLevel()         {}

// StructDef represents a struct definition.
type StructDef struct {
        Span       token.Span
        Name       string
        TypeParams []string
        Visibility Visibility
        Fields     []*FieldDef
        Comments   []Comment
}

func (n *StructDef) nodeType() string    { return "StructDef" }
func (n *StructDef) GetSpan() token.Span { return n.Span }
func (n *StructDef) isTopLevel()         {}

// FieldDef represents a field in a struct.
type FieldDef struct {
        Span       token.Span
        Name       string
        TypeExpr   TypeExpr
        Default    Expr
        Visibility Visibility
        Comments   []Comment
}

func (n *FieldDef) nodeType() string    { return "FieldDef" }
func (n *FieldDef) GetSpan() token.Span { return n.Span }

// EnumDef represents an enum definition.
type EnumDef struct {
        Span       token.Span
        Name       string
        TypeParams []string
        Visibility Visibility
        Variants   []*VariantDef
        Comments   []Comment
}

func (n *EnumDef) nodeType() string    { return "EnumDef" }
func (n *EnumDef) GetSpan() token.Span { return n.Span }
func (n *EnumDef) isTopLevel()         {}

// VariantDef represents an enum variant.
type VariantDef struct {
        Span   token.Span
        Name   string
        Fields []TypeExpr
}

func (n *VariantDef) nodeType() string    { return "VariantDef" }
func (n *VariantDef) GetSpan() token.Span { return n.Span }

// TraitDef represents a trait definition.
type TraitDef struct {
        Span       token.Span
        Name       string
        TypeParams []string
        Visibility Visibility
        Members    []TraitMember
        Comments   []Comment
}

func (n *TraitDef) nodeType() string    { return "TraitDef" }
func (n *TraitDef) GetSpan() token.Span { return n.Span }
func (n *TraitDef) isTopLevel()         {}

// TraitMember can be a function signature or a default implementation.
type TraitMember interface {
        Node
        isTraitMember()
}

// ImplBlock represents an impl block.
type ImplBlock struct {
        Span       token.Span
        TraitName  string   // empty for inherent impl
        TargetType TypeExpr // the type being implemented
        Methods    []*FnDef
        Comments   []Comment
}

func (n *ImplBlock) nodeType() string    { return "ImplBlock" }
func (n *ImplBlock) GetSpan() token.Span { return n.Span }
func (n *ImplBlock) isTopLevel()         {}

// --- Spec Blocks ---

// SpecBlock represents a spec block.
type SpecBlock struct {
        Span       token.Span
        Name       string
        Doc        string
        Inputs     []*SpecInput
        Guarantees []*SpecGuarantee
        Effects    []string
        Errors     []*SpecError
        Comments   []Comment
}

func (n *SpecBlock) nodeType() string    { return "SpecBlock" }
func (n *SpecBlock) GetSpan() token.Span { return n.Span }
func (n *SpecBlock) isTopLevel()         {}

// SpecInput represents an input declaration in a spec.
type SpecInput struct {
        Span        token.Span
        Name        string
        TypeExpr    TypeExpr
        Default     Expr
        Description string
}

func (n *SpecInput) nodeType() string    { return "SpecInput" }
func (n *SpecInput) GetSpan() token.Span { return n.Span }

// SpecGuarantee represents a guarantee in a spec.
type SpecGuarantee struct {
        Span      token.Span
        Condition string
}

func (n *SpecGuarantee) nodeType() string    { return "SpecGuarantee" }
func (n *SpecGuarantee) GetSpan() token.Span { return n.Span }

// SpecError represents an error declaration in a spec.
type SpecError struct {
        Span        token.Span
        TypeName    string
        Fields      []TypeExpr
        Description string
}

func (n *SpecError) nodeType() string    { return "SpecError" }
func (n *SpecError) GetSpan() token.Span { return n.Span }

// --- Function Definitions ---

// FnDef represents a function definition.
type FnDef struct {
        Span       token.Span
        Name       string
        TypeParams []string
        Params     []*Param
        ReturnType TypeExpr
        Effects    []string
        Satisfies  string
        Visibility Visibility
        Body       []Statement
        Comments   []Comment
}

func (n *FnDef) nodeType() string    { return "FnDef" }
func (n *FnDef) GetSpan() token.Span { return n.Span }
func (n *FnDef) isTopLevel()         {}
func (n *FnDef) isTraitMember()      {}

// FnSignature represents just a function signature (in traits).
type FnSignature struct {
        Span       token.Span
        Name       string
        TypeParams []string
        Params     []*Param
        ReturnType TypeExpr
        Effects    []string
        Visibility Visibility
}

func (n *FnSignature) nodeType() string    { return "FnSignature" }
func (n *FnSignature) GetSpan() token.Span { return n.Span }
func (n *FnSignature) isTraitMember()      {}

// Param represents a function parameter.
type Param struct {
        Span     token.Span
        Name     string
        TypeExpr TypeExpr
        Default  Expr
}

func (n *Param) nodeType() string    { return "Param" }
func (n *Param) GetSpan() token.Span { return n.Span }

// --- Const Definitions ---

// ConstDef represents a top-level constant.
type ConstDef struct {
        Span       token.Span
        Name       string
        TypeExpr   TypeExpr
        Value      Expr
        Visibility Visibility
        Comments   []Comment
}

func (n *ConstDef) nodeType() string    { return "ConstDef" }
func (n *ConstDef) GetSpan() token.Span { return n.Span }
func (n *ConstDef) isTopLevel()         {}

// --- Test Block ---

// TestBlock represents a test block.
type TestBlock struct {
        Span     token.Span
        Name     string
        Body     []Statement
        Comments []Comment
}

func (n *TestBlock) nodeType() string    { return "TestBlock" }
func (n *TestBlock) GetSpan() token.Span { return n.Span }
func (n *TestBlock) isTopLevel()         {}

// --- Type Expressions ---

// TypeExpr is the interface for all type expression nodes.
type TypeExpr interface {
        Node
        isTypeExpr()
}

// NamedType represents a named type like "Int" or "List[String]".
type NamedType struct {
        Span token.Span
        Name string
        Args []TypeExpr
}

func (n *NamedType) nodeType() string    { return "NamedType" }
func (n *NamedType) GetSpan() token.Span { return n.Span }
func (n *NamedType) isTypeExpr()         {}

// QualifiedType represents a qualified type like "time.Instant".
type QualifiedType struct {
        Span      token.Span
        Qualifier string
        Name      string
        Args      []TypeExpr
}

func (n *QualifiedType) nodeType() string    { return "QualifiedType" }
func (n *QualifiedType) GetSpan() token.Span { return n.Span }
func (n *QualifiedType) isTypeExpr()         {}

// UnionType represents T | U.
type UnionType struct {
        Span  token.Span
        Left  TypeExpr
        Right TypeExpr
}

func (n *UnionType) nodeType() string    { return "UnionType" }
func (n *UnionType) GetSpan() token.Span { return n.Span }
func (n *UnionType) isTypeExpr()         {}

// IntersectionType represents T & U.
type IntersectionType struct {
        Span  token.Span
        Left  TypeExpr
        Right TypeExpr
}

func (n *IntersectionType) nodeType() string    { return "IntersectionType" }
func (n *IntersectionType) GetSpan() token.Span { return n.Span }
func (n *IntersectionType) isTypeExpr()         {}

// FunctionType represents fn(T, U) -> V.
type FunctionType struct {
        Span       token.Span
        Params     []TypeExpr
        ReturnType TypeExpr
}

func (n *FunctionType) nodeType() string    { return "FunctionType" }
func (n *FunctionType) GetSpan() token.Span { return n.Span }
func (n *FunctionType) isTypeExpr()         {}

// TupleType represents (T, U, V).
type TupleType struct {
        Span     token.Span
        Elements []TypeExpr
}

func (n *TupleType) nodeType() string    { return "TupleType" }
func (n *TupleType) GetSpan() token.Span { return n.Span }
func (n *TupleType) isTypeExpr()         {}

// ListType represents [T].
type ListType struct {
        Span    token.Span
        Element TypeExpr
}

func (n *ListType) nodeType() string    { return "ListType" }
func (n *ListType) GetSpan() token.Span { return n.Span }
func (n *ListType) isTypeExpr()         {}

// MapType represents {K: V}.
type MapType struct {
        Span  token.Span
        Key   TypeExpr
        Value TypeExpr
}

func (n *MapType) nodeType() string    { return "MapType" }
func (n *MapType) GetSpan() token.Span { return n.Span }
func (n *MapType) isTypeExpr()         {}

// SetType represents {T}.
type SetType struct {
        Span    token.Span
        Element TypeExpr
}

func (n *SetType) nodeType() string    { return "SetType" }
func (n *SetType) GetSpan() token.Span { return n.Span }
func (n *SetType) isTypeExpr()         {}

// OptionType represents T?.
type OptionType struct {
        Span  token.Span
        Inner TypeExpr
}

func (n *OptionType) nodeType() string    { return "OptionType" }
func (n *OptionType) GetSpan() token.Span { return n.Span }
func (n *OptionType) isTypeExpr()         {}

// RefinementType represents T where predicate.
type RefinementType struct {
        Span      token.Span
        Base      TypeExpr
        Predicate Expr // The predicate expression
}

func (n *RefinementType) nodeType() string    { return "RefinementType" }
func (n *RefinementType) GetSpan() token.Span { return n.Span }
func (n *RefinementType) isTypeExpr()         {}

// StringLitType represents a string literal used as a type (in union types).
type StringLitType struct {
        Span  token.Span
        Value string
}

func (n *StringLitType) nodeType() string    { return "StringLitType" }
func (n *StringLitType) GetSpan() token.Span { return n.Span }
func (n *StringLitType) isTypeExpr()         {}

// --- Statements ---

// Statement is the interface for all statement nodes.
type Statement interface {
        Node
        isStatement()
}

// LetStmt represents "let [mut] name [: type] = expr".
type LetStmt struct {
        Span     token.Span
        Name     string
        Mutable  bool
        TypeHint TypeExpr
        Value    Expr
        Comments []Comment
}

func (n *LetStmt) nodeType() string    { return "LetStmt" }
func (n *LetStmt) GetSpan() token.Span { return n.Span }
func (n *LetStmt) isStatement()        {}

// AssignStmt represents "target = expr".
type AssignStmt struct {
        Span   token.Span
        Target Expr
        Value  Expr
}

func (n *AssignStmt) nodeType() string    { return "AssignStmt" }
func (n *AssignStmt) GetSpan() token.Span { return n.Span }
func (n *AssignStmt) isStatement()        {}

// ReturnStmt represents "return [expr]".
type ReturnStmt struct {
        Span  token.Span
        Value Expr
}

func (n *ReturnStmt) nodeType() string    { return "ReturnStmt" }
func (n *ReturnStmt) GetSpan() token.Span { return n.Span }
func (n *ReturnStmt) isStatement()        {}

// IfStmt represents an if/elif/else statement.
type IfStmt struct {
        Span        token.Span
        Condition   Expr
        ThenBody    []Statement
        ElifClauses []*ElifClause
        ElseBody    []Statement
}

func (n *IfStmt) nodeType() string    { return "IfStmt" }
func (n *IfStmt) GetSpan() token.Span { return n.Span }
func (n *IfStmt) isStatement()        {}

// ElifClause represents an elif branch.
type ElifClause struct {
        Span      token.Span
        Condition Expr
        Body      []Statement
}

// MatchStmt represents a match statement.
type MatchStmt struct {
        Span    token.Span
        Subject Expr
        Cases   []*CaseClause
}

func (n *MatchStmt) nodeType() string    { return "MatchStmt" }
func (n *MatchStmt) GetSpan() token.Span { return n.Span }
func (n *MatchStmt) isStatement()        {}

// CaseClause represents a case in a match statement.
type CaseClause struct {
        Span    token.Span
        Pattern Pattern
        Guard   Expr
        Body    []Statement
}

// ForStmt represents "for var in iterable: body".
type ForStmt struct {
        Span     token.Span
        Variable string
        Iterable Expr
        Body     []Statement
}

func (n *ForStmt) nodeType() string    { return "ForStmt" }
func (n *ForStmt) GetSpan() token.Span { return n.Span }
func (n *ForStmt) isStatement()        {}

// WhileStmt represents "while cond: body".
type WhileStmt struct {
        Span      token.Span
        Condition Expr
        Body      []Statement
}

func (n *WhileStmt) nodeType() string    { return "WhileStmt" }
func (n *WhileStmt) GetSpan() token.Span { return n.Span }
func (n *WhileStmt) isStatement()        {}

// ExprStmt wraps an expression used as a statement.
type ExprStmt struct {
        Span token.Span
        Expr Expr
}

func (n *ExprStmt) nodeType() string    { return "ExprStmt" }
func (n *ExprStmt) GetSpan() token.Span { return n.Span }
func (n *ExprStmt) isStatement()        {}

// AssertStmt represents "assert expr [, message]".
type AssertStmt struct {
        Span      token.Span
        Condition Expr
        Message   string
}

func (n *AssertStmt) nodeType() string    { return "AssertStmt" }
func (n *AssertStmt) GetSpan() token.Span { return n.Span }
func (n *AssertStmt) isStatement()        {}

// BreakStmt represents "break".
type BreakStmt struct {
        Span token.Span
}

func (n *BreakStmt) nodeType() string    { return "BreakStmt" }
func (n *BreakStmt) GetSpan() token.Span { return n.Span }
func (n *BreakStmt) isStatement()        {}

// ContinueStmt represents "continue".
type ContinueStmt struct {
        Span token.Span
}

func (n *ContinueStmt) nodeType() string    { return "ContinueStmt" }
func (n *ContinueStmt) GetSpan() token.Span { return n.Span }
func (n *ContinueStmt) isStatement()        {}

// WithStmt represents "with expr [as name][, expr [as name]]: body".
type WithStmt struct {
        Span     token.Span
        Bindings []*WithBinding
        Body     []Statement
}

func (n *WithStmt) nodeType() string    { return "WithStmt" }
func (n *WithStmt) GetSpan() token.Span { return n.Span }
func (n *WithStmt) isStatement()        {}

// WithBinding represents a single binding in a with statement.
type WithBinding struct {
        Expr  Expr
        Alias string
}

// --- Expressions ---

// Expr is the interface for all expression nodes.
type Expr interface {
        Node
        isExpr()
}

// Identifier represents a variable reference.
type Identifier struct {
        Span token.Span
        Name string
}

func (n *Identifier) nodeType() string    { return "Identifier" }
func (n *Identifier) GetSpan() token.Span { return n.Span }
func (n *Identifier) isExpr()             {}

// IntLiteral represents an integer literal.
type IntLiteral struct {
        Span  token.Span
        Value string
}

func (n *IntLiteral) nodeType() string    { return "IntLiteral" }
func (n *IntLiteral) GetSpan() token.Span { return n.Span }
func (n *IntLiteral) isExpr()             {}

// FloatLiteral represents a float literal.
type FloatLiteral struct {
        Span  token.Span
        Value string
}

func (n *FloatLiteral) nodeType() string    { return "FloatLiteral" }
func (n *FloatLiteral) GetSpan() token.Span { return n.Span }
func (n *FloatLiteral) isExpr()             {}

// StringLiteral represents a string literal.
type StringLiteral struct {
        Span  token.Span
        Value string
        // Parts for interpolated strings: alternating string and Expr
        Parts []StringPart
}

func (n *StringLiteral) nodeType() string    { return "StringLiteral" }
func (n *StringLiteral) GetSpan() token.Span { return n.Span }
func (n *StringLiteral) isExpr()             {}

// StringPart represents a part of an interpolated string.
type StringPart struct {
        IsExpr bool
        Text   string
        Expr   Expr
}

// BoolLiteral represents a boolean literal.
type BoolLiteral struct {
        Span  token.Span
        Value bool
}

func (n *BoolLiteral) nodeType() string    { return "BoolLiteral" }
func (n *BoolLiteral) GetSpan() token.Span { return n.Span }
func (n *BoolLiteral) isExpr()             {}

// NoneLiteral represents the none literal.
type NoneLiteral struct {
        Span token.Span
}

func (n *NoneLiteral) nodeType() string    { return "NoneLiteral" }
func (n *NoneLiteral) GetSpan() token.Span { return n.Span }
func (n *NoneLiteral) isExpr()             {}

// BinaryOp represents a binary operation.
type BinaryOp struct {
        Span  token.Span
        Op    string
        Left  Expr
        Right Expr
}

func (n *BinaryOp) nodeType() string    { return "BinaryOp" }
func (n *BinaryOp) GetSpan() token.Span { return n.Span }
func (n *BinaryOp) isExpr()             {}

// UnaryOp represents a unary operation.
type UnaryOp struct {
        Span    token.Span
        Op      string
        Operand Expr
}

func (n *UnaryOp) nodeType() string    { return "UnaryOp" }
func (n *UnaryOp) GetSpan() token.Span { return n.Span }
func (n *UnaryOp) isExpr()             {}

// CallExpr represents a function call.
type CallExpr struct {
        Span   token.Span
        Callee Expr
        Args   []*Arg
}

func (n *CallExpr) nodeType() string    { return "CallExpr" }
func (n *CallExpr) GetSpan() token.Span { return n.Span }
func (n *CallExpr) isExpr()             {}

// Arg represents a function argument (possibly named).
type Arg struct {
        Span  token.Span
        Name  string // empty for positional args
        Value Expr
}

// FieldAccess represents "obj.field".
type FieldAccess struct {
        Span   token.Span
        Object Expr
        Field  string
}

func (n *FieldAccess) nodeType() string    { return "FieldAccess" }
func (n *FieldAccess) GetSpan() token.Span { return n.Span }
func (n *FieldAccess) isExpr()             {}

// OptionalFieldAccess represents "obj?.field" (option chaining).
type OptionalFieldAccess struct {
        Span   token.Span
        Object Expr
        Field  string
}

func (n *OptionalFieldAccess) nodeType() string    { return "OptionalFieldAccess" }
func (n *OptionalFieldAccess) GetSpan() token.Span { return n.Span }
func (n *OptionalFieldAccess) isExpr()             {}

// PipelineExpr represents "expr |> func" (pipeline operator).
type PipelineExpr struct {
        Span  token.Span
        Left  Expr
        Right Expr
}

func (n *PipelineExpr) nodeType() string    { return "PipelineExpr" }
func (n *PipelineExpr) GetSpan() token.Span { return n.Span }
func (n *PipelineExpr) isExpr()             {}

// IndexExpr represents "obj[index]".
type IndexExpr struct {
        Span   token.Span
        Object Expr
        Index  Expr
}

func (n *IndexExpr) nodeType() string    { return "IndexExpr" }
func (n *IndexExpr) GetSpan() token.Span { return n.Span }
func (n *IndexExpr) isExpr()             {}

// OptionPropagate represents "expr?".
type OptionPropagate struct {
        Span token.Span
        Expr Expr
}

func (n *OptionPropagate) nodeType() string    { return "OptionPropagate" }
func (n *OptionPropagate) GetSpan() token.Span { return n.Span }
func (n *OptionPropagate) isExpr()             {}

// ListExpr represents "[a, b, c]".
type ListExpr struct {
        Span     token.Span
        Elements []Expr
}

func (n *ListExpr) nodeType() string    { return "ListExpr" }
func (n *ListExpr) GetSpan() token.Span { return n.Span }
func (n *ListExpr) isExpr()             {}

// ListComp represents "[expr for var in iterable if filter]".
type ListComp struct {
        Span     token.Span
        Element  Expr
        Variable string
        Iterable Expr
        Filter   Expr
}

func (n *ListComp) nodeType() string    { return "ListComp" }
func (n *ListComp) GetSpan() token.Span { return n.Span }
func (n *ListComp) isExpr()             {}

// MapExpr represents "{k1: v1, k2: v2}".
type MapExpr struct {
        Span    token.Span
        Entries []*MapEntry
}

func (n *MapExpr) nodeType() string    { return "MapExpr" }
func (n *MapExpr) GetSpan() token.Span { return n.Span }
func (n *MapExpr) isExpr()             {}

// MapEntry represents a key-value pair in a map expression.
type MapEntry struct {
        Key   Expr
        Value Expr
}

// StructExpr represents "TypeName(field: value, ...)".
type StructExpr struct {
        Span     token.Span
        TypeName string
        Fields   []*FieldInit
}

func (n *StructExpr) nodeType() string    { return "StructExpr" }
func (n *StructExpr) GetSpan() token.Span { return n.Span }
func (n *StructExpr) isExpr()             {}

// FieldInit represents a field initializer in struct construction.
type FieldInit struct {
        Name  string
        Value Expr
}

// IfExpr represents "if cond then a else b".
type IfExpr struct {
        Span      token.Span
        Condition Expr
        ThenExpr  Expr
        ElseExpr  Expr
}

func (n *IfExpr) nodeType() string    { return "IfExpr" }
func (n *IfExpr) GetSpan() token.Span { return n.Span }
func (n *IfExpr) isExpr()             {}

// Lambda represents "|params| -> expr" or "|params|: body".
type Lambda struct {
        Span   token.Span
        Params []*Param
        Body   Expr       // for single-expression lambdas
        Block  []Statement // for block lambdas
}

func (n *Lambda) nodeType() string    { return "Lambda" }
func (n *Lambda) GetSpan() token.Span { return n.Span }
func (n *Lambda) isExpr()             {}

// TupleLiteral represents "(a, b, c)" or "(a,)" for single-element tuples.
type TupleLiteral struct {
        Span     token.Span
        Elements []Expr
}

func (n *TupleLiteral) nodeType() string    { return "TupleLiteral" }
func (n *TupleLiteral) GetSpan() token.Span { return n.Span }
func (n *TupleLiteral) isExpr()             {}

// LetTupleDestructure represents "let (x, y, z) = expr".
type LetTupleDestructure struct {
        Span    token.Span
        Names   []string
        Mutable bool
        Value   Expr
}

func (n *LetTupleDestructure) nodeType() string    { return "LetTupleDestructure" }
func (n *LetTupleDestructure) GetSpan() token.Span { return n.Span }
func (n *LetTupleDestructure) isStatement()        {}

// --- Patterns ---

// Pattern is the interface for match patterns.
type Pattern interface {
        Node
        isPattern()
}

// WildcardPattern represents "_".
type WildcardPattern struct {
        Span token.Span
}

func (n *WildcardPattern) nodeType() string    { return "WildcardPattern" }
func (n *WildcardPattern) GetSpan() token.Span { return n.Span }
func (n *WildcardPattern) isPattern()          {}

// BindingPattern represents a variable binding in a pattern.
type BindingPattern struct {
        Span token.Span
        Name string
}

func (n *BindingPattern) nodeType() string    { return "BindingPattern" }
func (n *BindingPattern) GetSpan() token.Span { return n.Span }
func (n *BindingPattern) isPattern()          {}

// LiteralPattern represents a literal match in a pattern.
type LiteralPattern struct {
        Span  token.Span
        Value string
        Kind  token.Type // INT_LIT, STRING_LIT, BOOL_LIT, etc.
}

func (n *LiteralPattern) nodeType() string    { return "LiteralPattern" }
func (n *LiteralPattern) GetSpan() token.Span { return n.Span }
func (n *LiteralPattern) isPattern()          {}

// ConstructorPattern represents "TypeName(patterns...)".
type ConstructorPattern struct {
        Span     token.Span
        TypeName string // can be dotted: "TaskError.NotFound"
        Fields   []Pattern
}

func (n *ConstructorPattern) nodeType() string    { return "ConstructorPattern" }
func (n *ConstructorPattern) GetSpan() token.Span { return n.Span }
func (n *ConstructorPattern) isPattern()          {}

// ListPattern represents "[p1, p2, ...]".
type ListPattern struct {
        Span     token.Span
        Elements []Pattern
}

func (n *ListPattern) nodeType() string    { return "ListPattern" }
func (n *ListPattern) GetSpan() token.Span { return n.Span }
func (n *ListPattern) isPattern()          {}

// TuplePattern represents "(p1, p2, ...)".
type TuplePattern struct {
        Span     token.Span
        Elements []Pattern
}

func (n *TuplePattern) nodeType() string    { return "TuplePattern" }
func (n *TuplePattern) GetSpan() token.Span { return n.Span }
func (n *TuplePattern) isPattern()          {}

// SpreadPattern represents "...rest" in list patterns.
type SpreadPattern struct {
        Span token.Span
        Name string // the variable name to bind (e.g., "rest" in ...rest)
}

func (n *SpreadPattern) nodeType() string    { return "SpreadPattern" }
func (n *SpreadPattern) GetSpan() token.Span { return n.Span }
func (n *SpreadPattern) isPattern()          {}

// --- Match Expression ---

// MatchExpr represents a match expression: match value: pattern -> expr, ...
// Unlike MatchStmt, this is an expression that returns a value.
type MatchExpr struct {
        Span    token.Span
        Subject Expr
        Arms    []*MatchArm
}

func (n *MatchExpr) nodeType() string    { return "MatchExpr" }
func (n *MatchExpr) GetSpan() token.Span { return n.Span }
func (n *MatchExpr) isExpr()             {}

// MatchArm represents a single arm in a match expression: pattern -> expr
type MatchArm struct {
        Span    token.Span
        Pattern Pattern
        Body    Expr
}