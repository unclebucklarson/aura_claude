package compiler

import (
	"strings"
	"testing"

	"github.com/unclebucklarson/aura/pkg/lexer"
	"github.com/unclebucklarson/aura/pkg/parser"
)

// parseAndCompile parses the given Aura source and compiles it to a Chunk.
// It fails the test if parsing produces errors or if compilation produces errors.
func parseAndCompile(t *testing.T, src string) *Chunk {
	t.Helper()
	l := lexer.New(src, "<test>")
	tokens, lexErrs := l.Tokenize()
	if len(lexErrs) > 0 {
		t.Fatalf("lex errors: %v", lexErrs)
	}
	p := parser.New(tokens, "<test>")
	mod, parseErrs := p.Parse()
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	chunk, errs := CompileModule(mod)
	if len(errs) > 0 {
		msgs := make([]string, len(errs))
		for i, e := range errs {
			msgs[i] = e.Error()
		}
		t.Fatalf("compile errors: %s", strings.Join(msgs, "; "))
	}
	return chunk
}

// hasOpCode returns true if the chunk's bytecode contains at least one
// instruction with the given opcode.
func hasOpCode(ch *Chunk, op OpCode) bool {
	for i := 0; i+InstructionSize <= len(ch.Code); i += InstructionSize {
		if OpCode(ch.Code[i]) == op {
			return true
		}
	}
	return false
}

// countOpCode counts how many instructions with the given opcode appear in ch.
func countOpCode(ch *Chunk, op OpCode) int {
	n := 0
	for i := 0; i+InstructionSize <= len(ch.Code); i += InstructionSize {
		if OpCode(ch.Code[i]) == op {
			n++
		}
	}
	return n
}

// firstArg returns the 16-bit argument of the first instruction with op.
func firstArg(ch *Chunk, op OpCode) uint16 {
	for i := 0; i+InstructionSize <= len(ch.Code); i += InstructionSize {
		if OpCode(ch.Code[i]) == op {
			return ReadArg(ch.Code, i)
		}
	}
	return 0
}

// --- Opcode / chunk structure tests ---

func TestInstructionSize(t *testing.T) {
	if InstructionSize != 3 {
		t.Fatalf("expected InstructionSize=3, got %d", InstructionSize)
	}
}

func TestReadWriteArg(t *testing.T) {
	code := make([]byte, 3)
	WriteArg(code, 0, 0x1234)
	if got := ReadArg(code, 0); got != 0x1234 {
		t.Fatalf("expected 0x1234, got 0x%04X", got)
	}
	WriteArg(code, 0, 0)
	if got := ReadArg(code, 0); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
	WriteArg(code, 0, 0xFFFF)
	if got := ReadArg(code, 0); got != 0xFFFF {
		t.Fatalf("expected 0xFFFF, got 0x%04X", got)
	}
}

func TestOpCodeNames(t *testing.T) {
	cases := []struct {
		op   OpCode
		want string
	}{
		{OpConst, "CONST"},
		{OpAdd, "ADD"},
		{OpReturn, "RETURN"},
		{OpHalt, "HALT"},
		{OpMakeList, "MAKE_LIST"},
		{OpJumpIfFalse, "JUMP_IF_FALSE"},
	}
	for _, tc := range cases {
		if got := tc.op.Name(); got != tc.want {
			t.Errorf("opcode 0x%02X: want %q got %q", byte(tc.op), tc.want, got)
		}
	}
	// Unknown opcode.
	if name := OpCode(0xFF).Name(); !strings.HasPrefix(name, "UNKNOWN") {
		t.Errorf("expected UNKNOWN prefix, got %q", name)
	}
}

func TestChunkEmit(t *testing.T) {
	ch := NewChunk("test")
	off := ch.emit(OpConst, 42, 1)
	if off != 0 {
		t.Fatalf("expected offset 0, got %d", off)
	}
	if len(ch.Code) != InstructionSize {
		t.Fatalf("expected %d bytes, got %d", InstructionSize, len(ch.Code))
	}
	if OpCode(ch.Code[0]) != OpConst {
		t.Fatalf("expected OpConst, got 0x%02X", ch.Code[0])
	}
	if ReadArg(ch.Code, 0) != 42 {
		t.Fatalf("expected arg=42, got %d", ReadArg(ch.Code, 0))
	}
}

func TestChunkAddConstant(t *testing.T) {
	ch := NewChunk("test")
	idx, err := ch.addConstant(Constant{Kind: ConstInt, Val: int64(99)})
	if err != nil {
		t.Fatal(err)
	}
	if idx != 0 {
		t.Fatalf("expected idx=0, got %d", idx)
	}
	if len(ch.Constants) != 1 {
		t.Fatalf("expected 1 constant, got %d", len(ch.Constants))
	}
}

func TestChunkStringIntern(t *testing.T) {
	ch := NewChunk("test")
	idx1, _ := ch.addStringConst("hello")
	idx2, _ := ch.addStringConst("hello")
	if idx1 != idx2 {
		t.Fatalf("expected same index for duplicate string, got %d vs %d", idx1, idx2)
	}
	if len(ch.Constants) != 1 {
		t.Fatalf("expected 1 constant after dedup, got %d", len(ch.Constants))
	}
}

func TestChunkLineForOffset(t *testing.T) {
	ch := NewChunk("test")
	ch.emit(OpConst, 0, 5)
	ch.emit(OpConst, 0, 5)
	ch.emit(OpAdd, 0, 10)
	if got := ch.LineForOffset(0); got != 5 {
		t.Errorf("offset 0: expected line 5, got %d", got)
	}
	if got := ch.LineForOffset(6); got != 10 {
		t.Errorf("offset 6: expected line 10, got %d", got)
	}
}

// --- Compiler tests ---

func TestCompileIntLiteral(t *testing.T) {
	src := `module test
fn main():
    return 42`
	ch := parseAndCompile(t, src)
	// There should be at least one OpConst for 42.
	if !hasOpCode(ch, OpClosure) {
		t.Fatal("expected OpClosure for fn main")
	}
	// Find the inner chunk for main.
	var mainChunk *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "main" {
			mainChunk = c.Val.(*Chunk)
		}
	}
	if mainChunk == nil {
		t.Fatal("could not find chunk for main")
	}
	if !hasOpCode(mainChunk, OpConst) {
		t.Fatal("expected OpConst in main")
	}
	if !hasOpCode(mainChunk, OpReturn) {
		t.Fatal("expected OpReturn in main")
	}
	// The constant should be 42.
	found := false
	for _, c := range mainChunk.Constants {
		if c.Kind == ConstInt && c.Val.(int64) == 42 {
			found = true
		}
	}
	if !found {
		t.Fatal("expected constant 42 in main chunk")
	}
}

func TestCompileBoolLiterals(t *testing.T) {
	src := `module test
fn f():
    let x = true
    let y = false
    return x`
	ch := parseAndCompile(t, src)
	var fc *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "f" {
			fc = c.Val.(*Chunk)
		}
	}
	if fc == nil {
		t.Fatal("no chunk for f")
	}
	if !hasOpCode(fc, OpTrue) {
		t.Error("expected OpTrue")
	}
	if !hasOpCode(fc, OpFalse) {
		t.Error("expected OpFalse")
	}
}

func TestCompileArithmetic(t *testing.T) {
	src := `module test
fn add(a, b):
    return a + b`
	ch := parseAndCompile(t, src)
	var fc *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "add" {
			fc = c.Val.(*Chunk)
		}
	}
	if fc == nil {
		t.Fatal("no chunk for add")
	}
	if !hasOpCode(fc, OpAdd) {
		t.Error("expected OpAdd")
	}
	if !hasOpCode(fc, OpGetLocal) {
		t.Error("expected OpGetLocal")
	}
}

func TestCompileLetStmt(t *testing.T) {
	src := `module test
fn f():
    let x = 10
    let y = 20
    return x`
	ch := parseAndCompile(t, src)
	var fc *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "f" {
			fc = c.Val.(*Chunk)
		}
	}
	if fc == nil {
		t.Fatal("no chunk for f")
	}
	// Two OpConst for 10 and 20.
	if countOpCode(fc, OpConst) < 2 {
		t.Errorf("expected at least 2 OpConst, got %d", countOpCode(fc, OpConst))
	}
}

func TestCompileIfStmt(t *testing.T) {
	src := `module test
fn check(x):
    if x > 0:
        return 1
    else:
        return 0`
	ch := parseAndCompile(t, src)
	var fc *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "check" {
			fc = c.Val.(*Chunk)
		}
	}
	if fc == nil {
		t.Fatal("no chunk for check")
	}
	if !hasOpCode(fc, OpJumpIfFalse) {
		t.Error("expected OpJumpIfFalse")
	}
	if !hasOpCode(fc, OpJump) {
		t.Error("expected OpJump for else branch")
	}
}

func TestCompileWhileStmt(t *testing.T) {
	src := `module test
fn loop():
    let mut i = 0
    while i < 10:
        i = i + 1
    return i`
	ch := parseAndCompile(t, src)
	var fc *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "loop" {
			fc = c.Val.(*Chunk)
		}
	}
	if fc == nil {
		t.Fatal("no chunk for loop")
	}
	if !hasOpCode(fc, OpLoop) {
		t.Error("expected OpLoop for while back-edge")
	}
	if !hasOpCode(fc, OpJumpIfFalse) {
		t.Error("expected OpJumpIfFalse for while condition")
	}
}

func TestCompileListExpr(t *testing.T) {
	src := `module test
fn f():
    return [1, 2, 3]`
	ch := parseAndCompile(t, src)
	var fc *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "f" {
			fc = c.Val.(*Chunk)
		}
	}
	if fc == nil {
		t.Fatal("no chunk for f")
	}
	if !hasOpCode(fc, OpMakeList) {
		t.Error("expected OpMakeList")
	}
	arg := firstArg(fc, OpMakeList)
	if arg != 3 {
		t.Errorf("expected list count=3, got %d", arg)
	}
}

func TestCompileMapExpr(t *testing.T) {
	src := `module test
fn f():
    return {"a": 1, "b": 2}`
	ch := parseAndCompile(t, src)
	var fc *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "f" {
			fc = c.Val.(*Chunk)
		}
	}
	if fc == nil {
		t.Fatal("no chunk for f")
	}
	if !hasOpCode(fc, OpMakeMap) {
		t.Error("expected OpMakeMap")
	}
	arg := firstArg(fc, OpMakeMap)
	if arg != 2 {
		t.Errorf("expected map count=2, got %d", arg)
	}
}

func TestCompileLambdaEmitsClosure(t *testing.T) {
	src := `module test
fn f():
    let double = |x| -> x * 2
    return double`
	ch := parseAndCompile(t, src)
	var fc *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "f" {
			fc = c.Val.(*Chunk)
		}
	}
	if fc == nil {
		t.Fatal("no chunk for f")
	}
	if !hasOpCode(fc, OpClosure) {
		t.Error("expected OpClosure for lambda")
	}
}

func TestCompileOptionConstructors(t *testing.T) {
	src := `module test
fn wrap(x):
    return Some(x)
fn empty():
    return none`
	ch := parseAndCompile(t, src)
	var wrapChunk, emptyChunk *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk {
			switch c.Val.(*Chunk).Name {
			case "wrap":
				wrapChunk = c.Val.(*Chunk)
			case "empty":
				emptyChunk = c.Val.(*Chunk)
			}
		}
	}
	if wrapChunk == nil || emptyChunk == nil {
		t.Fatal("could not find chunks")
	}
	if !hasOpCode(wrapChunk, OpSome) {
		t.Error("expected OpSome in wrap")
	}
	if !hasOpCode(emptyChunk, OpNone) {
		t.Error("expected OpNone in empty")
	}
}

func TestCompileResultConstructors(t *testing.T) {
	src := `module test
fn good(x):
    return Ok(x)
fn bad(e):
    return Err(e)`
	ch := parseAndCompile(t, src)
	var goodChunk, badChunk *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk {
			switch c.Val.(*Chunk).Name {
			case "good":
				goodChunk = c.Val.(*Chunk)
			case "bad":
				badChunk = c.Val.(*Chunk)
			}
		}
	}
	if goodChunk == nil || badChunk == nil {
		t.Fatal("could not find chunks")
	}
	if !hasOpCode(goodChunk, OpOk) {
		t.Error("expected OpOk in good")
	}
	if !hasOpCode(badChunk, OpErr) {
		t.Error("expected OpErr in bad")
	}
}

func TestCompileStringInterpolation(t *testing.T) {
	src := `module test
fn greet(name):
    return "Hello, {name}!"`
	ch := parseAndCompile(t, src)
	var fc *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "greet" {
			fc = c.Val.(*Chunk)
		}
	}
	if fc == nil {
		t.Fatal("no chunk for greet")
	}
	if !hasOpCode(fc, OpInterpolate) {
		t.Error("expected OpInterpolate for string interpolation")
	}
}

func TestCompileFieldAccess(t *testing.T) {
	src := `module test
fn get_name(obj):
    return obj.name`
	ch := parseAndCompile(t, src)
	var fc *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "get_name" {
			fc = c.Val.(*Chunk)
		}
	}
	if fc == nil {
		t.Fatal("no chunk for get_name")
	}
	if !hasOpCode(fc, OpGetField) {
		t.Error("expected OpGetField")
	}
	// The field name "name" should be in constants.
	found := false
	for _, c := range fc.Constants {
		if c.Kind == ConstString && c.Val.(string) == "name" {
			found = true
		}
	}
	if !found {
		t.Error("expected string constant 'name' for field access")
	}
}

func TestCompileMethodCall(t *testing.T) {
	src := `module test
fn f(s):
    return s.upper()`
	ch := parseAndCompile(t, src)
	var fc *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "f" {
			fc = c.Val.(*Chunk)
		}
	}
	if fc == nil {
		t.Fatal("no chunk for f")
	}
	if !hasOpCode(fc, OpCallMethod) {
		t.Error("expected OpCallMethod")
	}
}

func TestCompilePipelineExpr(t *testing.T) {
	src := `module test
fn f(x):
    return x |> double`
	ch := parseAndCompile(t, src)
	var fc *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "f" {
			fc = c.Val.(*Chunk)
		}
	}
	if fc == nil {
		t.Fatal("no chunk for f")
	}
	if !hasOpCode(fc, OpPipeline) {
		t.Error("expected OpPipeline")
	}
}

func TestCompileIfExpr(t *testing.T) {
	src := `module test
fn abs(x):
    return if x >= 0 then x else -x`
	ch := parseAndCompile(t, src)
	var fc *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "abs" {
			fc = c.Val.(*Chunk)
		}
	}
	if fc == nil {
		t.Fatal("no chunk for abs")
	}
	if !hasOpCode(fc, OpJumpIfFalse) {
		t.Error("expected OpJumpIfFalse for ternary if")
	}
	if !hasOpCode(fc, OpNeg) {
		t.Error("expected OpNeg for -x")
	}
}

func TestCompileGlobalConstDef(t *testing.T) {
	src := `module test
let max_val = 100`
	ch := parseAndCompile(t, src)
	if !hasOpCode(ch, OpDefGlobal) {
		t.Error("expected OpDefGlobal for let max_val")
	}
	// The global name "max_val" should be interned.
	found := false
	for _, c := range ch.Constants {
		if c.Kind == ConstString && c.Val.(string) == "max_val" {
			found = true
		}
	}
	if !found {
		t.Error("expected string constant 'max_val' in module chunk")
	}
}

func TestDisassembler(t *testing.T) {
	src := `module test
fn f(x):
    return x + 1`
	ch := parseAndCompile(t, src)
	dis := Disassemble(ch)
	if !strings.Contains(dis, "CLOSURE") {
		t.Errorf("disassembly missing CLOSURE: %s", dis)
	}
	if !strings.Contains(dis, "== f ==") {
		t.Errorf("disassembly missing nested chunk header: %s", dis)
	}
}

func TestCompileForStmt(t *testing.T) {
	src := `module test
fn sum(xs):
    let mut total = 0
    for x in xs:
        total = total + x
    return total`
	ch := parseAndCompile(t, src)
	var fc *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "sum" {
			fc = c.Val.(*Chunk)
		}
	}
	if fc == nil {
		t.Fatal("no chunk for sum")
	}
	if !hasOpCode(fc, OpMakeIter) {
		t.Error("expected OpMakeIter")
	}
	if !hasOpCode(fc, OpForIter) {
		t.Error("expected OpForIter")
	}
	if !hasOpCode(fc, OpLoop) {
		t.Error("expected OpLoop back-edge")
	}
}

func TestCompileUpvalue(t *testing.T) {
	src := `module test
fn outer(x):
    let inner = |y| -> x + y
    return inner`
	ch := parseAndCompile(t, src)
	var outerChunk *Chunk
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "outer" {
			outerChunk = c.Val.(*Chunk)
		}
	}
	if outerChunk == nil {
		t.Fatal("no chunk for outer")
	}
	// The lambda inside outer should have an upvalue for x.
	var lambdaChunk *Chunk
	for _, c := range outerChunk.Constants {
		if c.Kind == ConstChunk && c.Val.(*Chunk).Name == "<lambda>" {
			lambdaChunk = c.Val.(*Chunk)
		}
	}
	if lambdaChunk == nil {
		t.Fatal("no lambda chunk inside outer")
	}
	if len(lambdaChunk.Upvalues) == 0 {
		t.Error("expected upvalue descriptor for captured x")
	}
	if !lambdaChunk.Upvalues[0].IsLocal {
		t.Error("expected upvalue to be local (captured from direct parent)")
	}
}

func TestChunkCodeAlwaysMultipleOfInstructionSize(t *testing.T) {
	src := `module test
fn f(a, b, c):
    if a > b:
        return a + c
    elif b > c:
        return b
    else:
        return c`
	ch := parseAndCompile(t, src)
	// Verify the module chunk.
	if len(ch.Code)%InstructionSize != 0 {
		t.Errorf("module chunk code length %d is not a multiple of %d", len(ch.Code), InstructionSize)
	}
	// And all nested chunks.
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk {
			inner := c.Val.(*Chunk)
			if len(inner.Code)%InstructionSize != 0 {
				t.Errorf("inner chunk %q code length %d is not a multiple of %d",
					inner.Name, len(inner.Code), InstructionSize)
			}
		}
	}
}
