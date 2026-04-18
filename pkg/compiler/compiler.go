package compiler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/unclebucklarson/aura/pkg/ast"
	"github.com/unclebucklarson/aura/pkg/token"
)

// CompileError records a compilation error with a source position.
type CompileError struct {
	Message string
	Line    int
	Col     int
}

func (e CompileError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("compile error at line %d: %s", e.Line, e.Message)
	}
	return "compile error: " + e.Message
}

// scope tracks local variables at a single function/lambda scope level.
type scope struct {
	locals []LocalInfo // parallel to VM stack slots (slot 0 = first local)
	depth  int         // current block nesting depth within this function
}

// loopContext captures the start offset and pending break-jump patches for
// a loop body so that break/continue can be wired up.
type loopContext struct {
	startOffset int     // loop top — target for continue
	breakJumps  []int   // offsets of JUMP instructions to patch for break
	label       *string // loop label (used for labeled break/continue)
}

// Compiler walks a typed AST and emits bytecode into a Chunk.
type Compiler struct {
	chunk   *Chunk
	parent  *Compiler    // enclosing compiler (nil at module level)
	sc      scope        // local variable tracking
	loops   []loopContext
	globals map[string]bool // set of names defined at module level
	errors  []CompileError
}

// newCompiler creates a compiler for a new function/lambda chunk.
// parent is nil for module-level compilation.
func newCompiler(name string, parent *Compiler) *Compiler {
	globals := map[string]bool{}
	if parent != nil {
		globals = parent.globals
	}
	return &Compiler{
		chunk:   NewChunk(name),
		parent:  parent,
		globals: globals,
	}
}

// CompileModule compiles a parsed Aura module into a top-level Chunk.
// Returns the chunk and any compile errors encountered.
func CompileModule(mod *ast.Module) (*Chunk, []CompileError) {
	c := newCompiler("<module:"+mod.Name.String()+">", nil)

	// First pass: register all top-level function/const/type names as globals.
	for _, item := range mod.Items {
		switch d := item.(type) {
		case *ast.FnDef:
			c.globals[d.Name] = true
		case *ast.ConstDef:
			c.globals[d.Name] = true
		}
	}

	// Second pass: emit definitions.
	for _, item := range mod.Items {
		c.compileTopLevel(item)
	}

	// Emit HALT at end of module.
	line := 0
	if len(mod.Items) > 0 {
		line = mod.Items[len(mod.Items)-1].GetSpan().Start.Line
	}
	c.chunk.emit0(OpHalt, line)

	return c.chunk, c.errors
}

// --- Top-level declarations ---

func (c *Compiler) compileTopLevel(item ast.TopLevelItem) {
	switch d := item.(type) {
	case *ast.FnDef:
		c.compileFnDef(d)
	case *ast.ConstDef:
		c.compileConstDef(d)
	case *ast.TypeDef, *ast.StructDef, *ast.EnumDef,
		*ast.TraitDef, *ast.ImplBlock, *ast.SpecBlock:
		// These are type-system constructs; the runtime doesn't need bytecode for them.
		// The VM will use the type registry populated by the checker.
	case *ast.TestBlock:
		// Test blocks are compiled separately when running tests.
	default:
		c.errorf(item.GetSpan(), "unsupported top-level item %T", item)
	}
}

func (c *Compiler) compileFnDef(d *ast.FnDef) {
	line := d.Span.Start.Line
	// Compile the function body into a new chunk.
	inner := newCompiler(d.Name, c)
	inner.chunk.Arity = len(d.Params)
	inner.globals = c.globals

	// Define params as the first locals (slot 0..n-1).
	for _, p := range d.Params {
		inner.sc.locals = append(inner.sc.locals, LocalInfo{Name: p.Name, Depth: 0})
		inner.chunk.Locals = append(inner.chunk.Locals, LocalInfo{Name: p.Name, Depth: 0})
	}

	inner.compileBody(d.Body)

	// Ensure there is always a return at the end.
	if len(inner.chunk.Code) == 0 ||
		OpCode(inner.chunk.Code[len(inner.chunk.Code)-InstructionSize]) != OpReturn {
		inner.chunk.emit0(OpNil, line)
		inner.chunk.emit0(OpReturn, line)
	}

	c.errors = append(c.errors, inner.errors...)

	// Add the inner chunk as a constant and emit OpClosure.
	idx, err := c.chunk.addConstant(Constant{Kind: ConstChunk, Val: inner.chunk})
	if err != nil {
		c.errorf(d.Span, "%v", err)
		return
	}

	c.chunk.emit(OpClosure, idx, line)

	// Emit upvalue descriptors as raw bytes following the OpClosure instruction.
	// (2 bytes per upvalue: isLocal byte + index byte)
	for _, uv := range inner.chunk.Upvalues {
		isLocalByte := byte(0)
		if uv.IsLocal {
			isLocalByte = 1
		}
		c.chunk.Code = append(c.chunk.Code, isLocalByte, byte(uv.Index))
	}

	// Define the function as a global.
	nameIdx, err2 := c.chunk.addStringConst(d.Name)
	if err2 != nil {
		c.errorf(d.Span, "%v", err2)
		return
	}
	c.chunk.emit(OpDefGlobal, nameIdx, line)
}

func (c *Compiler) compileConstDef(d *ast.ConstDef) {
	line := d.Span.Start.Line
	c.compileExpr(d.Value)
	nameIdx, err := c.chunk.addStringConst(d.Name)
	if err != nil {
		c.errorf(d.Span, "%v", err)
		return
	}
	c.chunk.emit(OpDefGlobal, nameIdx, line)
}

// --- Statements ---

func (c *Compiler) compileBody(stmts []ast.Statement) {
	for _, s := range stmts {
		c.compileStmt(s)
	}
}

func (c *Compiler) compileStmt(s ast.Statement) {
	switch st := s.(type) {
	case *ast.LetStmt:
		c.compileLetStmt(st)
	case *ast.LetTupleDestructure:
		c.compileLetTuple(st)
	case *ast.AssignStmt:
		c.compileAssignStmt(st)
	case *ast.ReturnStmt:
		c.compileReturnStmt(st)
	case *ast.ExprStmt:
		c.compileExpr(st.Expr)
		c.chunk.emit0(OpPop, st.Span.Start.Line)
	case *ast.IfStmt:
		c.compileIfStmt(st)
	case *ast.MatchStmt:
		c.compileMatchStmt(st)
	case *ast.ForStmt:
		c.compileForStmt(st)
	case *ast.WhileStmt:
		c.compileWhileStmt(st)
	case *ast.BreakStmt:
		c.compileBreakStmt(st)
	case *ast.ContinueStmt:
		c.compileContinueStmt(st)
	case *ast.AssertStmt:
		c.compileAssertStmt(st)
	case *ast.WithStmt:
		c.compileWithStmt(st)
	default:
		c.errorf(s.GetSpan(), "unsupported statement %T", s)
	}
}

func (c *Compiler) compileLetStmt(s *ast.LetStmt) {
	line := s.Span.Start.Line
	if s.Value != nil {
		c.compileExpr(s.Value)
	} else {
		c.chunk.emit0(OpNil, line)
	}
	slot := c.declareLocal(s.Name, line)
	_ = slot // value is already on top of stack at this slot
}

func (c *Compiler) compileLetTuple(s *ast.LetTupleDestructure) {
	line := s.Span.Start.Line
	c.compileExpr(s.Value)
	// The value on stack is a tuple; emit GET_INDEX for each position.
	for i, name := range s.Names {
		c.chunk.emit0(OpDup, line)
		idx, _ := c.chunk.addConstant(Constant{Kind: ConstInt, Val: int64(i)})
		c.chunk.emit(OpConst, idx, line)
		c.chunk.emit0(OpGetIndex, line)
		c.declareLocal(name, line)
	}
	// Pop the original tuple.
	c.chunk.emit0(OpPop, line)
}

func (c *Compiler) compileAssignStmt(s *ast.AssignStmt) {
	line := s.Span.Start.Line
	c.compileExpr(s.Value)

	switch t := s.Target.(type) {
	case *ast.Identifier:
		if slot, ok := c.resolveLocal(t.Name); ok {
			c.chunk.emit(OpSetLocal, uint16(slot), line)
		} else if uvIdx, ok := c.resolveUpvalue(t.Name); ok {
			c.chunk.emit(OpSetUpvalue, uint16(uvIdx), line)
		} else {
			nameIdx, _ := c.chunk.addStringConst(t.Name)
			c.chunk.emit(OpSetGlobal, nameIdx, line)
		}
		c.chunk.emit0(OpPop, line)

	case *ast.FieldAccess:
		// target.field = value
		// stack: value → need: object, then SET_FIELD
		// We need to push object, swap value to TOS, SET_FIELD.
		c.compileExpr(t.Object)
		c.chunk.emit0(OpSwap, line)
		fieldIdx, _ := c.chunk.addStringConst(t.Field)
		c.chunk.emit(OpSetField, fieldIdx, line)
		c.chunk.emit0(OpPop, line)

	case *ast.IndexExpr:
		// target[index] = value
		c.compileExpr(t.Object)
		c.compileExpr(t.Index)
		c.chunk.emit0(OpSetIndex, line)
		c.chunk.emit0(OpPop, line)

	default:
		c.errorf(s.Span, "unsupported assignment target %T", s.Target)
	}
}

func (c *Compiler) compileReturnStmt(s *ast.ReturnStmt) {
	line := s.Span.Start.Line
	if s.Value != nil {
		c.compileExpr(s.Value)
	} else {
		c.chunk.emit0(OpNil, line)
	}
	c.chunk.emit0(OpReturn, line)
}

func (c *Compiler) compileIfStmt(s *ast.IfStmt) {
	line := s.Span.Start.Line
	var jumpsToEnd []int

	c.compileExpr(s.Condition)
	jumpFalse := c.chunk.emit(OpJumpIfFalse, 0, line)

	c.enterScope()
	c.compileBody(s.ThenBody)
	c.exitScope(line)

	if len(s.ElifClauses) > 0 || len(s.ElseBody) > 0 {
		jumpsToEnd = append(jumpsToEnd, c.chunk.emit(OpJump, 0, line))
	}
	c.patchJump(jumpFalse)

	for _, elif := range s.ElifClauses {
		elifLine := elif.Span.Start.Line
		c.compileExpr(elif.Condition)
		nextFalse := c.chunk.emit(OpJumpIfFalse, 0, elifLine)
		c.enterScope()
		c.compileBody(elif.Body)
		c.exitScope(elifLine)
		jumpsToEnd = append(jumpsToEnd, c.chunk.emit(OpJump, 0, elifLine))
		c.patchJump(nextFalse)
	}

	if len(s.ElseBody) > 0 {
		c.enterScope()
		c.compileBody(s.ElseBody)
		c.exitScope(line)
	}

	for _, j := range jumpsToEnd {
		c.patchJump(j)
	}
}

func (c *Compiler) compileMatchStmt(s *ast.MatchStmt) {
	line := s.Span.Start.Line
	// Compile subject onto stack; it stays there across all case tests.
	c.compileExpr(s.Subject)

	var jumpsToEnd []int

	for i, cas := range s.Cases {
		caseLine := cas.Span.Start.Line
		// Duplicate subject for this case's pattern test (all but last case).
		isLast := i == len(s.Cases)-1
		if !isLast {
			c.chunk.emit0(OpDup, caseLine)
		}
		// Emit pattern matching code; leaves bool on stack (or skipped for wildcard).
		jumpToNext := c.compilePattern(cas.Pattern, caseLine, !isLast)

		// If there's a guard, test it too.
		var guardJump int = -1
		if cas.Guard != nil {
			c.compileExpr(cas.Guard)
			guardJump = c.chunk.emit(OpJumpIfFalse, 0, caseLine)
		}

		// Body.
		c.enterScope()
		c.compileBody(cas.Body)
		c.exitScope(caseLine)

		// Jump to end after body (except last).
		if !isLast {
			jumpsToEnd = append(jumpsToEnd, c.chunk.emit(OpJump, 0, caseLine))
		}

		// Patch the jump-to-next-case.
		if jumpToNext >= 0 {
			c.patchJump(jumpToNext)
		}
		if guardJump >= 0 {
			c.patchJump(guardJump)
		}
	}

	// Pop subject (it's still on stack if no case consumed it, which shouldn't happen).
	c.chunk.emit0(OpPop, line)

	for _, j := range jumpsToEnd {
		c.patchJump(j)
	}
}

// compilePattern emits code that tests whether the value currently on TOS
// matches the given pattern.
//
// If needsDup is true, the subject is a Dup'd copy; it will be consumed (popped)
// by this function. The pattern code emits:
//  1. A boolean test result on TOS (for non-wildcard patterns).
//  2. Returns the offset of a JUMP_IF_FALSE instruction that skips to the next case.
//     Returns -1 if the pattern is a wildcard (unconditionally matches).
func (c *Compiler) compilePattern(pat ast.Pattern, line int, needsDup bool) int {
	switch p := pat.(type) {
	case *ast.WildcardPattern:
		// Wildcard: always matches. Pop the dup'd subject if present.
		if needsDup {
			c.chunk.emit0(OpPop, line)
		}
		return -1 // no jump needed

	case *ast.BindingPattern:
		// Binds the subject to a name. If needsDup the dup is the subject.
		if needsDup {
			// subject is already on stack; declare it as a local
		}
		c.declareLocal(p.Name, line)
		return -1

	case *ast.LiteralPattern:
		// Emit the literal value and compare.
		litIdx := c.compileLiteralPattern(p)
		c.chunk.emit(OpMatchLiteral, litIdx, line)
		return c.chunk.emit(OpJumpIfFalse, 0, line)

	case *ast.ConstructorPattern:
		return c.compileConstructorPattern(p, line, needsDup)

	case *ast.OrPattern:
		return c.compileOrPattern(p, line)

	case *ast.ListPattern:
		return c.compileListPattern(p, line)

	default:
		c.errorf(pat.GetSpan(), "unsupported pattern %T", pat)
		return -1
	}
}

func (c *Compiler) compileLiteralPattern(p *ast.LiteralPattern) uint16 {
	line := p.Span.Start.Line
	var con Constant
	switch p.Kind {
	case token.INT_LIT:
		v, _ := strconv.ParseInt(p.Value, 10, 64)
		con = Constant{Kind: ConstInt, Val: v}
	case token.FLOAT_LIT:
		v, _ := strconv.ParseFloat(p.Value, 64)
		con = Constant{Kind: ConstFloat, Val: v}
	case token.STRING_LIT:
		con = Constant{Kind: ConstString, Val: unquoteString(p.Value)}
	case token.TRUE:
		con = Constant{Kind: ConstBool, Val: true}
	case token.FALSE:
		con = Constant{Kind: ConstBool, Val: false}
	default:
		con = Constant{Kind: ConstString, Val: p.Value}
	}
	idx, err := c.chunk.addConstant(con)
	if err != nil {
		c.errorf(p.Span, "%v", err)
		_ = line
	}
	return idx
}

func (c *Compiler) compileConstructorPattern(p *ast.ConstructorPattern, line int, needsDup bool) int {
	// Handles: Some(x), None, Ok(x), Err(e), EnumName.Variant(fields...)
	name := p.TypeName
	switch name {
	case "Some":
		c.chunk.emit0(OpMatchSome, line)
		jumpFalse := c.chunk.emit(OpJumpIfFalse, 0, line)
		if len(p.Fields) == 1 {
			c.chunk.emit0(OpUnwrapSome, line)
			c.compilePatternBindOrPop(p.Fields[0], line)
		}
		return jumpFalse
	case "None":
		c.chunk.emit0(OpMatchNone, line)
		return c.chunk.emit(OpJumpIfFalse, 0, line)
	case "Ok":
		c.chunk.emit0(OpMatchOk, line)
		jumpFalse := c.chunk.emit(OpJumpIfFalse, 0, line)
		if len(p.Fields) == 1 {
			c.chunk.emit0(OpUnwrapOk, line)
			c.compilePatternBindOrPop(p.Fields[0], line)
		}
		return jumpFalse
	case "Err":
		c.chunk.emit0(OpMatchErr, line)
		jumpFalse := c.chunk.emit(OpJumpIfFalse, 0, line)
		if len(p.Fields) == 1 {
			c.chunk.emit0(OpUnwrapErr, line)
			c.compilePatternBindOrPop(p.Fields[0], line)
		}
		return jumpFalse
	default:
		// General enum variant or struct.
		// Strip qualifier (e.g. "TaskError.NotFound" → "NotFound" for match).
		variantName := name
		if idx := strings.LastIndex(name, "."); idx >= 0 {
			variantName = name[idx+1:]
		}
		nameIdx, _ := c.chunk.addStringConst(variantName)
		c.chunk.emit(OpMatchVariant, nameIdx, line)
		jumpFalse := c.chunk.emit(OpJumpIfFalse, 0, line)
		if len(p.Fields) > 0 {
			c.chunk.emit(OpDestructureVariant, uint16(len(p.Fields)), line)
			for _, field := range p.Fields {
				c.compilePatternBindOrPop(field, line)
			}
		}
		return jumpFalse
	}
}

func (c *Compiler) compilePatternBindOrPop(pat ast.Pattern, line int) {
	switch p := pat.(type) {
	case *ast.BindingPattern:
		c.declareLocal(p.Name, line)
	case *ast.WildcardPattern:
		c.chunk.emit0(OpPop, line)
	default:
		// For nested patterns, recursively compile.
		c.compilePattern(p, line, false)
	}
}

func (c *Compiler) compileOrPattern(p *ast.OrPattern, line int) int {
	// Emit left pattern check; if it succeeds, jump over right check.
	// Both branches need to consume the subject identically.
	c.chunk.emit0(OpDup, line)
	leftJump := c.compilePattern(p.Left, line, true)
	// If left matched (no jump taken), jump to body.
	jumpToBody := c.chunk.emit(OpJump, 0, line)
	// Patch left jump to here (try right side).
	if leftJump >= 0 {
		c.patchJump(leftJump)
	}
	rightJump := c.compilePattern(p.Right, line, true)
	c.patchJump(jumpToBody)
	return rightJump
}

func (c *Compiler) compileListPattern(p *ast.ListPattern, line int) int {
	// Simple list pattern: check length and bind elements.
	// For now, emit length check using a runtime helper.
	// Build a constant with the expected length.
	lenIdx, _ := c.chunk.addConstant(Constant{Kind: ConstInt, Val: int64(len(p.Elements))})
	c.chunk.emit(OpConst, lenIdx, line)
	// OpMatchLiteral repurposed: we emit a method call to check len instead.
	// Actually, emit: DUP, CALL_METHOD "len", CONST n, EQ, JUMP_IF_FALSE
	c.chunk.emit0(OpDup, line)
	// Call .len() on it using OpCallMethod
	methodNameIdx, _ := c.chunk.addStringConst("len")
	// pack: high byte = method name idx (if fits in 8 bits), low = 0 args
	if methodNameIdx <= 0xFF {
		c.chunk.emit(OpCallMethod, methodNameIdx<<8|0, line)
	} else {
		c.errorf(p.Span, "method name constant index too large for OpCallMethod encoding")
	}
	c.chunk.emit(OpConst, lenIdx, line)
	c.chunk.emit0(OpEq, line)
	jumpFalse := c.chunk.emit(OpJumpIfFalse, 0, line)

	// Bind each element.
	for i, elem := range p.Elements {
		c.chunk.emit0(OpDup, line)
		idxConst, _ := c.chunk.addConstant(Constant{Kind: ConstInt, Val: int64(i)})
		c.chunk.emit(OpConst, idxConst, line)
		c.chunk.emit0(OpGetIndex, line)
		c.compilePatternBindOrPop(elem, line)
	}
	// Pop the original list.
	c.chunk.emit0(OpPop, line)
	return jumpFalse
}

func (c *Compiler) compileForStmt(s *ast.ForStmt) {
	line := s.Span.Start.Line
	// Compile iterable and wrap in iterator.
	c.compileExpr(s.Iterable)
	c.chunk.emit0(OpMakeIter, line)

	loopStart := len(c.chunk.Code)
	c.pushLoop(loopStart)

	// FOR_ITER: if exhausted, jump past body (patched below).
	forIterOffset := c.chunk.emit(OpForIter, 0, line)

	// The iterator pushes the next item onto the stack; declare loop variable.
	c.enterScope()
	slot := c.declareLocal(s.Variable, line)
	_ = slot

	c.compileBody(s.Body)
	c.exitScope(line)

	// Jump back to loop start.
	backDist := uint16(len(c.chunk.Code) - loopStart + InstructionSize)
	c.chunk.emit(OpLoop, backDist, line)

	// Patch ForIter to jump here (past loop).
	c.patchJump(forIterOffset)

	// Pop the iterator.
	c.chunk.emit0(OpPop, line)

	c.popLoop()
}

func (c *Compiler) compileWhileStmt(s *ast.WhileStmt) {
	line := s.Span.Start.Line
	loopStart := len(c.chunk.Code)
	c.pushLoop(loopStart)

	c.compileExpr(s.Condition)
	jumpFalse := c.chunk.emit(OpJumpIfFalse, 0, line)

	c.enterScope()
	c.compileBody(s.Body)
	c.exitScope(line)

	// Jump back to condition.
	backDist := uint16(len(c.chunk.Code) - loopStart + InstructionSize)
	c.chunk.emit(OpLoop, backDist, line)

	c.patchJump(jumpFalse)
	c.popLoop()
}

func (c *Compiler) compileBreakStmt(s *ast.BreakStmt) {
	line := s.Span.Start.Line
	if len(c.loops) == 0 {
		c.errorf(s.Span, "break outside loop")
		return
	}
	
	// If there's a label, find the matching loop
	if s.Label != nil {
		label := *s.Label
		// Find the matching loop from innermost to outermost
		for i := len(c.loops) - 1; i >= 0; i-- {
			if c.loops[i].label != nil && *c.loops[i].label == label {
				// Jump to end of that loop
				jumpOffset := c.chunk.emit(OpJump, 0, line)
				// We patch all break jumps to the matching loop 
				// But in current code, the patching logic isn't implemented
				// This would need a more complex system
				// For now, we'll treat as regular break (to innermost)
				c.loops[len(c.loops)-1].breakJumps = append(c.loops[len(c.loops)-1].breakJumps, jumpOffset)
				return
			}
		}
		// No matching label found
		c.errorf(s.Span, "break label %q not found", label)
		return
	}
	
	// Regular break to innermost loop
	jumpOffset := c.chunk.emit(OpJump, 0, line)
	c.loops[len(c.loops)-1].breakJumps = append(c.loops[len(c.loops)-1].breakJumps, jumpOffset)
}

func (c *Compiler) compileContinueStmt(s *ast.ContinueStmt) {
	line := s.Span.Start.Line
	if len(c.loops) == 0 {
		c.errorf(s.Span, "continue outside loop")
		return
	}
	
	// If there's a label, find the matching loop and continue there
	if s.Label != nil {
		label := *s.Label
		// Find the matching loop from innermost to outermost
		for i := len(c.loops) - 1; i >= 0; i-- {
			if c.loops[i].label != nil && *c.loops[i].label == label {
				// Continue to that loop (emit loop back to start)
				backDist := uint16(len(c.chunk.Code) - c.loops[i].startOffset + InstructionSize)
				c.chunk.emit(OpLoop, backDist, line)
				return
			}
		}
		// No matching label found
		c.errorf(s.Span, "continue label %q not found", label)
		return
	}
	
	// Regular continue to innermost loop
	loopCtx := c.loops[len(c.loops)-1]
	backDist := uint16(len(c.chunk.Code) - loopCtx.startOffset + InstructionSize)
	c.chunk.emit(OpLoop, backDist, line)
}

func (c *Compiler) compileAssertStmt(s *ast.AssertStmt) {
	line := s.Span.Start.Line
	c.compileExpr(s.Condition)
	msgIdx, _ := c.chunk.addStringConst(s.Message)
	c.chunk.emit(OpAssert, msgIdx, line)
}

func (c *Compiler) compileWithStmt(s *ast.WithStmt) {
	// WithStmt is an effect tracking annotation; at runtime it's a no-op block.
	// Just compile the body statements.
	c.compileBody(s.Body)
}

// --- Expressions ---

func (c *Compiler) compileExpr(e ast.Expr) {
	switch ex := e.(type) {
	case *ast.IntLiteral:
		v, err := strconv.ParseInt(ex.Value, 10, 64)
		if err != nil {
			c.errorf(ex.Span, "invalid integer literal %q", ex.Value)
			v = 0
		}
		idx, _ := c.chunk.addConstant(Constant{Kind: ConstInt, Val: v})
		c.chunk.emit(OpConst, idx, ex.Span.Start.Line)
	case *ast.FloatLiteral:
		v, err := strconv.ParseFloat(ex.Value, 64)
		if err != nil {
			c.errorf(ex.Span, "invalid float literal %q", ex.Value)
			v = 0
		}
		idx, _ := c.chunk.addConstant(Constant{Kind: ConstFloat, Val: v})
		c.chunk.emit(OpConst, idx, ex.Span.Start.Line)
	case *ast.StringLiteral:
		c.compileStringLiteral(ex)
	case *ast.BoolLiteral:
		if ex.Value {
			c.chunk.emit0(OpTrue, ex.Span.Start.Line)
		} else {
			c.chunk.emit0(OpFalse, ex.Span.Start.Line)
		}
	case *ast.NoneLiteral:
		c.chunk.emit0(OpNone, ex.Span.Start.Line)
	case *ast.Identifier:
		c.compileIdentifier(ex)
	case *ast.BinaryOp:
		c.compileBinaryOp(ex)
	case *ast.UnaryOp:
		c.compileUnaryOp(ex)
	case *ast.CallExpr:
		c.compileCallExpr(ex)
	case *ast.FieldAccess:
		c.compileFieldAccess(ex)
	case *ast.OptionalFieldAccess:
		c.compileOptFieldAccess(ex)
	case *ast.IndexExpr:
		c.compileIndexExpr(ex)
	case *ast.OptionPropagate:
		c.compileOptionPropagate(ex)
	case *ast.PipelineExpr:
		c.compilePipelineExpr(ex)
	case *ast.ListExpr:
		c.compileListExpr(ex)
	case *ast.ListComp:
		c.compileListComp(ex)
	case *ast.MapExpr:
		c.compileMapExpr(ex)
	case *ast.StructExpr:
		c.compileStructExpr(ex)
	case *ast.TupleLiteral:
		c.compileTupleLiteral(ex)
	case *ast.Lambda:
		c.compileLambda(ex)
	case *ast.IfExpr:
		c.compileIfExpr(ex)
	case *ast.MatchExpr:
		c.compileMatchExpr(ex)
	default:
		c.errorf(e.GetSpan(), "unsupported expression %T", e)
	}
}

func (c *Compiler) compileStringLiteral(ex *ast.StringLiteral) {
	line := ex.Span.Start.Line
	// Check if any part is a dynamic expression.
	hasDynamic := false
	for _, p := range ex.Parts {
		if p.IsExpr {
			hasDynamic = true
			break
		}
	}
	if !hasDynamic {
		// Simple string — collect all text parts.
		var sb strings.Builder
		for _, p := range ex.Parts {
			sb.WriteString(p.Text)
		}
		idx, _ := c.chunk.addStringConst(sb.String())
		c.chunk.emit(OpConst, idx, line)
		return
	}
	// String interpolation: push each part, emit OpInterpolate.
	count := uint16(0)
	for _, p := range ex.Parts {
		if p.IsExpr {
			c.compileExpr(p.Expr)
		} else if p.Text != "" {
			idx, _ := c.chunk.addStringConst(p.Text)
			c.chunk.emit(OpConst, idx, line)
		} else {
			continue
		}
		count++
	}
	c.chunk.emit(OpInterpolate, count, line)
}

func (c *Compiler) compileIdentifier(ex *ast.Identifier) {
	line := ex.Span.Start.Line
	name := ex.Name
	// Built-in constructors.
	switch name {
	case "none":
		c.chunk.emit0(OpNone, line)
		return
	case "true":
		c.chunk.emit0(OpTrue, line)
		return
	case "false":
		c.chunk.emit0(OpFalse, line)
		return
	}
	// Local first, then upvalue, then global.
	if slot, ok := c.resolveLocal(name); ok {
		c.chunk.emit(OpGetLocal, uint16(slot), line)
	} else if uvIdx, ok := c.resolveUpvalue(name); ok {
		c.chunk.emit(OpGetUpvalue, uint16(uvIdx), line)
	} else {
		nameIdx, _ := c.chunk.addStringConst(name)
		c.chunk.emit(OpGetGlobal, nameIdx, line)
	}
}

func (c *Compiler) compileBinaryOp(ex *ast.BinaryOp) {
	line := ex.Span.Start.Line

	// Short-circuit operators.
	if ex.Op == "and" {
		c.compileExpr(ex.Left)
		jumpFalse := c.chunk.emit(OpAnd, 0, line)
		c.compileExpr(ex.Right)
		c.patchJump(jumpFalse)
		return
	}
	if ex.Op == "or" {
		c.compileExpr(ex.Left)
		jumpTrue := c.chunk.emit(OpOr, 0, line)
		c.compileExpr(ex.Right)
		c.patchJump(jumpTrue)
		return
	}

	c.compileExpr(ex.Left)
	c.compileExpr(ex.Right)

	switch ex.Op {
	case "+":
		c.chunk.emit0(OpAdd, line)
	case "-":
		c.chunk.emit0(OpSub, line)
	case "*":
		c.chunk.emit0(OpMul, line)
	case "/":
		c.chunk.emit0(OpDiv, line)
	case "%":
		c.chunk.emit0(OpMod, line)
	case "**":
		c.chunk.emit0(OpPow, line)
	case "++":
		c.chunk.emit0(OpConcat, line)
	case "==":
		c.chunk.emit0(OpEq, line)
	case "!=":
		c.chunk.emit0(OpNeq, line)
	case "<":
		c.chunk.emit0(OpLt, line)
	case "<=":
		c.chunk.emit0(OpLte, line)
	case ">":
		c.chunk.emit0(OpGt, line)
	case ">=":
		c.chunk.emit0(OpGte, line)
	case "in":
		c.chunk.emit0(OpIn, line)
	case "not in":
		c.chunk.emit0(OpNotIn, line)
	default:
		c.errorf(ex.Span, "unsupported binary operator %q", ex.Op)
	}
}

func (c *Compiler) compileUnaryOp(ex *ast.UnaryOp) {
	line := ex.Span.Start.Line
	c.compileExpr(ex.Operand)
	switch ex.Op {
	case "-":
		c.chunk.emit0(OpNeg, line)
	case "not":
		c.chunk.emit0(OpNot, line)
	default:
		c.errorf(ex.Span, "unsupported unary operator %q", ex.Op)
	}
}

func (c *Compiler) compileCallExpr(ex *ast.CallExpr) {
	line := ex.Span.Start.Line

	// Check for special constructors.
	if ident, ok := ex.Callee.(*ast.Identifier); ok {
		switch ident.Name {
		case "Some":
			if len(ex.Args) == 1 {
				c.compileExpr(ex.Args[0].Value)
				c.chunk.emit0(OpSome, line)
				return
			}
		case "Ok":
			if len(ex.Args) == 1 {
				c.compileExpr(ex.Args[0].Value)
				c.chunk.emit0(OpOk, line)
				return
			}
		case "Err":
			if len(ex.Args) == 1 {
				c.compileExpr(ex.Args[0].Value)
				c.chunk.emit0(OpErr, line)
				return
			}
		}
	}

	// Method call: receiver.method(args...)
	if fa, ok := ex.Callee.(*ast.FieldAccess); ok {
		c.compileExpr(fa.Object)
		for _, arg := range ex.Args {
			c.compileExpr(arg.Value)
		}
		argCount := len(ex.Args)
		if argCount > 0xFF {
			c.errorf(ex.Span, "too many method arguments")
			return
		}
		methodIdx, err := c.chunk.addStringConst(fa.Field)
		if err != nil {
			c.errorf(ex.Span, "%v", err)
			return
		}
		if methodIdx > 0xFF {
			c.errorf(ex.Span, "method name constant index too large for packed encoding")
			return
		}
		packed := (methodIdx << 8) | uint16(argCount)
		c.chunk.emit(OpCallMethod, packed, line)
		return
	}

	// Regular function call.
	c.compileExpr(ex.Callee)
	for _, arg := range ex.Args {
		c.compileExpr(arg.Value)
	}
	c.chunk.emit(OpCall, uint16(len(ex.Args)), line)
}

func (c *Compiler) compileFieldAccess(ex *ast.FieldAccess) {
	line := ex.Span.Start.Line
	c.compileExpr(ex.Object)
	fieldIdx, _ := c.chunk.addStringConst(ex.Field)
	c.chunk.emit(OpGetField, fieldIdx, line)
}

func (c *Compiler) compileOptFieldAccess(ex *ast.OptionalFieldAccess) {
	line := ex.Span.Start.Line
	c.compileExpr(ex.Object)
	fieldIdx, _ := c.chunk.addStringConst(ex.Field)
	c.chunk.emit(OpOptChain, fieldIdx, line)
}

func (c *Compiler) compileIndexExpr(ex *ast.IndexExpr) {
	line := ex.Span.Start.Line
	c.compileExpr(ex.Object)
	c.compileExpr(ex.Index)
	c.chunk.emit0(OpGetIndex, line)
}

func (c *Compiler) compileOptionPropagate(ex *ast.OptionPropagate) {
	// expr? — if the result is Some(x), push x; if None, return None from function.
	line := ex.Span.Start.Line
	c.compileExpr(ex.Expr)
	c.chunk.emit0(OpIsSome, line)
	jumpIfSome := c.chunk.emit(OpJumpIfTrue, 0, line)
	// It's None: return None from current function.
	c.chunk.emit0(OpNone, line)
	c.chunk.emit0(OpReturn, line)
	// It's Some: unwrap.
	c.patchJump(jumpIfSome)
	c.chunk.emit0(OpUnwrapSome, line)
}

func (c *Compiler) compilePipelineExpr(ex *ast.PipelineExpr) {
	line := ex.Span.Start.Line
	// Push the LHS value.
	c.compileExpr(ex.Left)
	// If RHS is a function reference, call it with 1 argument.
	c.compileExpr(ex.Right)
	c.chunk.emit(OpPipeline, 1, line)
}

func (c *Compiler) compileListExpr(ex *ast.ListExpr) {
	line := ex.Span.Start.Line
	for _, elem := range ex.Elements {
		c.compileExpr(elem)
	}
	c.chunk.emit(OpMakeList, uint16(len(ex.Elements)), line)
}

func (c *Compiler) compileListComp(ex *ast.ListComp) {
	// Compile as: let _result = []; for var in iterable: if filter: _result.append(element)
	// Use a temporary local for the accumulator.
	line := ex.Span.Start.Line

	// Create empty list.
	c.chunk.emit(OpMakeList, 0, line)
	resultSlot := c.declareLocal("_listcomp", line)

	// Compile iterable and make iterator.
	c.compileExpr(ex.Iterable)
	c.chunk.emit0(OpMakeIter, line)

	loopStart := len(c.chunk.Code)
	c.pushLoop(loopStart)
	forIterOffset := c.chunk.emit(OpForIter, 0, line)

	// Declare loop variable.
	c.enterScope()
	c.declareLocal(ex.Variable, line)

	// Optional filter.
	var filterJump int = -1
	if ex.Filter != nil {
		c.compileExpr(ex.Filter)
		filterJump = c.chunk.emit(OpJumpIfFalse, 0, line)
	}

	// Append element to result list.
	c.chunk.emit(OpGetLocal, uint16(resultSlot), line)
	c.compileExpr(ex.Element)
	methodIdx, _ := c.chunk.addStringConst("push")
	if methodIdx <= 0xFF {
		c.chunk.emit(OpCallMethod, methodIdx<<8|1, line)
	}
	c.chunk.emit0(OpPop, line) // discard push return value

	if filterJump >= 0 {
		c.patchJump(filterJump)
	}

	c.exitScope(line)

	backDist := uint16(len(c.chunk.Code) - loopStart + InstructionSize)
	c.chunk.emit(OpLoop, backDist, line)
	c.patchJump(forIterOffset)
	c.chunk.emit0(OpPop, line) // pop iterator

	c.popLoop()

	// Push result list (it's still in resultSlot).
	c.chunk.emit(OpGetLocal, uint16(resultSlot), line)
}

func (c *Compiler) compileMapExpr(ex *ast.MapExpr) {
	line := ex.Span.Start.Line
	for _, entry := range ex.Entries {
		c.compileExpr(entry.Key)
		c.compileExpr(entry.Value)
	}
	c.chunk.emit(OpMakeMap, uint16(len(ex.Entries)), line)
}

func (c *Compiler) compileStructExpr(ex *ast.StructExpr) {
	line := ex.Span.Start.Line
	// Push field name-value pairs, then MAKE_STRUCT.
	for _, f := range ex.Fields {
		nameIdx, _ := c.chunk.addStringConst(f.Name)
		c.chunk.emit(OpConst, nameIdx, line)
		c.compileExpr(f.Value)
	}
	typeNameIdx, _ := c.chunk.addStringConst(ex.TypeName)
	// Pack: use the type name index as arg for MAKE_STRUCT;
	// field count is embedded in the stack (each field pushed as key+value pair).
	// The VM reads 2*N items from the stack for MAKE_STRUCT where N = field count.
	// Encode field count in arg's high byte and type name idx in low byte?
	// Simpler: use arg = typeNameIdx; VM knows field count from name+type lookup,
	// or we encode it differently. Use arg = field count and a separate constant for name.
	// Let's use: CONST typeNameIdx, MAKE_STRUCT fieldCount.
	c.chunk.emit(OpConst, typeNameIdx, line)
	c.chunk.emit(OpMakeStruct, uint16(len(ex.Fields)), line)
}

func (c *Compiler) compileTupleLiteral(ex *ast.TupleLiteral) {
	line := ex.Span.Start.Line
	for _, elem := range ex.Elements {
		c.compileExpr(elem)
	}
	c.chunk.emit(OpMakeTuple, uint16(len(ex.Elements)), line)
}

func (c *Compiler) compileLambda(ex *ast.Lambda) {
	line := ex.Span.Start.Line
	name := "<lambda>"
	inner := newCompiler(name, c)
	inner.chunk.Arity = len(ex.Params)
	inner.globals = c.globals

	for _, p := range ex.Params {
		inner.sc.locals = append(inner.sc.locals, LocalInfo{Name: p.Name, Depth: 0})
		inner.chunk.Locals = append(inner.chunk.Locals, LocalInfo{Name: p.Name, Depth: 0})
	}

	if ex.Body != nil {
		// Single-expression lambda.
		inner.compileExpr(ex.Body)
		inner.chunk.emit0(OpReturn, line)
	} else {
		// Block lambda.
		inner.compileBody(ex.Block)
		if len(inner.chunk.Code) == 0 ||
			OpCode(inner.chunk.Code[len(inner.chunk.Code)-InstructionSize]) != OpReturn {
			inner.chunk.emit0(OpNil, line)
			inner.chunk.emit0(OpReturn, line)
		}
	}

	c.errors = append(c.errors, inner.errors...)

	idx, err := c.chunk.addConstant(Constant{Kind: ConstChunk, Val: inner.chunk})
	if err != nil {
		c.errorf(ex.Span, "%v", err)
		return
	}
	c.chunk.emit(OpClosure, idx, line)

	for _, uv := range inner.chunk.Upvalues {
		isLocalByte := byte(0)
		if uv.IsLocal {
			isLocalByte = 1
		}
		c.chunk.Code = append(c.chunk.Code, isLocalByte, byte(uv.Index))
	}
}

func (c *Compiler) compileIfExpr(ex *ast.IfExpr) {
	line := ex.Span.Start.Line
	c.compileExpr(ex.Condition)
	jumpFalse := c.chunk.emit(OpJumpIfFalse, 0, line)
	c.compileExpr(ex.ThenExpr)
	jumpEnd := c.chunk.emit(OpJump, 0, line)
	c.patchJump(jumpFalse)
	c.compileExpr(ex.ElseExpr)
	c.patchJump(jumpEnd)
}

func (c *Compiler) compileMatchExpr(ex *ast.MatchExpr) {
	line := ex.Span.Start.Line
	c.compileExpr(ex.Subject)

	var jumpsToEnd []int

	for i, arm := range ex.Arms {
		armLine := arm.Span.Start.Line
		isLast := i == len(ex.Arms)-1
		if !isLast {
			c.chunk.emit0(OpDup, armLine)
		}
		jumpToNext := c.compilePattern(arm.Pattern, armLine, !isLast)

		var guardJump int = -1
		if arm.Guard != nil {
			c.compileExpr(arm.Guard)
			guardJump = c.chunk.emit(OpJumpIfFalse, 0, armLine)
		}

		// Pop subject before evaluating body expression.
		if !isLast {
			c.chunk.emit0(OpSwap, armLine)
			c.chunk.emit0(OpPop, armLine)
		}

		c.compileExpr(arm.Body)

		if !isLast {
			jumpsToEnd = append(jumpsToEnd, c.chunk.emit(OpJump, 0, armLine))
		}

		if jumpToNext >= 0 {
			c.patchJump(jumpToNext)
		}
		if guardJump >= 0 {
			c.patchJump(guardJump)
		}
	}

	// Pop subject (TOS after last case body is the result value).
	// The last case doesn't dup, so subject was consumed by the last compilePattern.
	// Actually: for the last case, subject is still on stack below the result.
	// Pop it now.
	c.chunk.emit0(OpSwap, line)
	c.chunk.emit0(OpPop, line)

	for _, j := range jumpsToEnd {
		c.patchJump(j)
	}
}

// --- Scope / local variable management ---

func (c *Compiler) enterScope() {
	c.sc.depth++
}

func (c *Compiler) exitScope(line int) {
	// Pop all locals declared at this depth.
	n := 0
	for len(c.sc.locals) > 0 && c.sc.locals[len(c.sc.locals)-1].Depth == c.sc.depth {
		c.sc.locals = c.sc.locals[:len(c.sc.locals)-1]
		n++
	}
	if n == 1 {
		c.chunk.emit0(OpPop, line)
	} else if n > 1 {
		c.chunk.emit(OpPopN, uint16(n), line)
	}
	c.sc.depth--
}

func (c *Compiler) declareLocal(name string, line int) int {
	slot := len(c.sc.locals)
	info := LocalInfo{Name: name, Depth: c.sc.depth}
	c.sc.locals = append(c.sc.locals, info)
	c.chunk.Locals = append(c.chunk.Locals, info)
	return slot
}

func (c *Compiler) resolveLocal(name string) (int, bool) {
	for i := len(c.sc.locals) - 1; i >= 0; i-- {
		if c.sc.locals[i].Name == name {
			return i, true
		}
	}
	return -1, false
}

// resolveUpvalue resolves a name that is not a local in this scope but may
// exist in an enclosing scope. It records an UpvalueDesc in this chunk.
func (c *Compiler) resolveUpvalue(name string) (int, bool) {
	if c.parent == nil {
		return -1, false
	}
	// Is it a local in the parent?
	if slot, ok := c.parent.resolveLocal(name); ok {
		return c.addUpvalue(uint16(slot), true), true
	}
	// Is it an upvalue in the parent?
	if uvIdx, ok := c.parent.resolveUpvalue(name); ok {
		return c.addUpvalue(uint16(uvIdx), false), true
	}
	return -1, false
}

func (c *Compiler) addUpvalue(index uint16, isLocal bool) int {
	// Deduplicate.
	for i, uv := range c.chunk.Upvalues {
		if uv.Index == index && uv.IsLocal == isLocal {
			return i
		}
	}
	c.chunk.Upvalues = append(c.chunk.Upvalues, UpvalueDesc{IsLocal: isLocal, Index: index})
	return len(c.chunk.Upvalues) - 1
}

// --- Jump patching ---

// patchJump sets the argument of the jump instruction at `offset` to jump to
// the current end of the code (i.e., forward-jump to "here").
func (c *Compiler) patchJump(offset int) {
	dist := len(c.chunk.Code) - offset
	if dist > 0xFFFF {
		c.errorf(token.Span{}, "jump distance too large")
		return
	}
	c.chunk.patchArg(offset, uint16(dist))
}

// --- Loop stack ---

func (c *Compiler) pushLoop(startOffset int) {
	c.loops = append(c.loops, loopContext{startOffset: startOffset})
}

func (c *Compiler) popLoop() {
	if len(c.loops) == 0 {
		return
	}
	ctx := c.loops[len(c.loops)-1]
	c.loops = c.loops[:len(c.loops)-1]
	// Patch all break jumps to current position.
	for _, j := range ctx.breakJumps {
		c.patchJump(j)
	}
}

// --- Helpers ---

func (c *Compiler) errorf(span token.Span, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	line := 0
	if span.Start.Line > 0 {
		line = span.Start.Line
	}
	c.errors = append(c.errors, CompileError{Message: msg, Line: line, Col: span.Start.Column})
}

// unquoteString removes surrounding quotes from a string literal value.
func unquoteString(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
