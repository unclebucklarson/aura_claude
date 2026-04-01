package compiler

import (
	"fmt"
	"strings"
)

// ConstantKind identifies the Go type stored in a Constant.
type ConstantKind byte

const (
	ConstInt    ConstantKind = iota // Val is int64
	ConstFloat                      // Val is float64
	ConstString                     // Val is string
	ConstBool                       // Val is bool
	ConstChunk                      // Val is *Chunk (nested function/lambda)
)

// Constant is a single entry in a Chunk's constant pool.
type Constant struct {
	Kind ConstantKind
	Val  any // int64 | float64 | string | bool | *Chunk
}

func (c Constant) String() string {
	switch c.Kind {
	case ConstInt:
		return fmt.Sprintf("%d", c.Val.(int64))
	case ConstFloat:
		return fmt.Sprintf("%g", c.Val.(float64))
	case ConstString:
		return fmt.Sprintf("%q", c.Val.(string))
	case ConstBool:
		if c.Val.(bool) {
			return "true"
		}
		return "false"
	case ConstChunk:
		return fmt.Sprintf("<chunk %s>", c.Val.(*Chunk).Name)
	default:
		return "<?>"
	}
}

// UpvalueDesc describes a captured variable in a closure.
// If IsLocal is true, Index is the stack slot in the enclosing function.
// If IsLocal is false, Index is the upvalue index in the enclosing function.
type UpvalueDesc struct {
	IsLocal bool
	Index   uint16
}

// LocalInfo records a local variable's name and the stack depth at which it
// was declared. Used by the disassembler and for debuggability.
type LocalInfo struct {
	Name  string
	Depth int  // scope depth (0 = module level)
}

// SourceMapEntry maps a bytecode offset to a source line number.
type SourceMapEntry struct {
	Offset int
	Line   int
}

// Chunk is the compiled representation of a single function or module body.
// It contains flat bytecode (3 bytes per instruction), a constant pool,
// upvalue descriptors for closures, a source map for error reporting,
// and metadata.
type Chunk struct {
	Name      string           // function name, "<module>", or "<lambda>"
	Code      []byte           // flat bytecode; len always a multiple of InstructionSize
	Constants []Constant       // constant pool indexed by uint16
	Upvalues  []UpvalueDesc    // upvalue descriptors for closures
	Locals    []LocalInfo      // local variable metadata (parallel to stack slots)
	SourceMap []SourceMapEntry // sorted by Offset; used for error messages
	Arity     int              // number of required parameters
	MaxStack  int              // high-water mark of stack depth (set by compiler)
	HasVararg bool             // true if the last param is variadic
}

// NewChunk allocates a Chunk with the given name.
func NewChunk(name string) *Chunk {
	return &Chunk{Name: name}
}

// emit appends a 3-byte instruction to ch.Code.
// op is the opcode; arg is the 16-bit argument.
func (ch *Chunk) emit(op OpCode, arg uint16, line int) int {
	offset := len(ch.Code)
	ch.Code = append(ch.Code, byte(op), byte(arg>>8), byte(arg))
	// Record source mapping; only add new entry when line changes.
	if len(ch.SourceMap) == 0 || ch.SourceMap[len(ch.SourceMap)-1].Line != line {
		ch.SourceMap = append(ch.SourceMap, SourceMapEntry{Offset: offset, Line: line})
	}
	return offset
}

// emit0 emits an instruction with a zero argument.
func (ch *Chunk) emit0(op OpCode, line int) int {
	return ch.emit(op, 0, line)
}

// addConstant adds val to the constant pool and returns its index.
// Returns an error if the pool is full (> 65535 entries).
func (ch *Chunk) addConstant(c Constant) (uint16, error) {
	if len(ch.Constants) >= 0xFFFF {
		return 0, fmt.Errorf("constant pool overflow in chunk %q", ch.Name)
	}
	ch.Constants = append(ch.Constants, c)
	return uint16(len(ch.Constants) - 1), nil
}

// addStringConst interns a string into the constant pool.
// If the same string already exists it re-uses the existing index.
func (ch *Chunk) addStringConst(s string) (uint16, error) {
	for i, c := range ch.Constants {
		if c.Kind == ConstString && c.Val.(string) == s {
			return uint16(i), nil
		}
	}
	return ch.addConstant(Constant{Kind: ConstString, Val: s})
}

// patchArg overwrites the argument bytes at the given instruction offset.
func (ch *Chunk) patchArg(offset int, arg uint16) {
	WriteArg(ch.Code, offset, arg)
}

// LineForOffset returns the source line number for the given bytecode offset
// using the SourceMap. Returns -1 if the map is empty.
func (ch *Chunk) LineForOffset(offset int) int {
	line := -1
	for _, e := range ch.SourceMap {
		if e.Offset > offset {
			break
		}
		line = e.Line
	}
	return line
}

// Summary returns a compact human-readable description of the chunk's size.
func (ch *Chunk) Summary() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "chunk %q: %d bytes, %d constants, %d upvalues",
		ch.Name, len(ch.Code), len(ch.Constants), len(ch.Upvalues))
	return sb.String()
}
