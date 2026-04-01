// Package compiler translates typed Aura ASTs to a stack-based bytecode IR.
//
// Instruction encoding: fixed 3 bytes — [OpCode][ArgHi][ArgLo]
// Argument is a 16-bit unsigned integer: arg = uint16(hi)<<8 | uint16(lo)
// Instructions that take no argument still occupy 3 bytes (arg ignored).
package compiler

import "fmt"

// InstructionSize is the fixed byte width of every instruction.
const InstructionSize = 3

// OpCode identifies a bytecode instruction.
type OpCode byte

// --- Constants (0x00–0x0F) ---
const (
	// OpConst pushes Constants[arg] onto the stack.
	OpConst OpCode = 0x00
	// OpNil pushes the nil/none value.
	OpNil OpCode = 0x01
	// OpTrue pushes bool true.
	OpTrue OpCode = 0x02
	// OpFalse pushes bool false.
	OpFalse OpCode = 0x03
)

// --- Locals / Globals / Upvalues (0x10–0x1F) ---
const (
	// OpGetLocal pushes the local at stack slot arg.
	OpGetLocal OpCode = 0x10
	// OpSetLocal pops TOS and stores it in stack slot arg.
	OpSetLocal OpCode = 0x11
	// OpGetGlobal pushes the value of global named Constants[arg].
	OpGetGlobal OpCode = 0x12
	// OpSetGlobal pops TOS and stores it as global named Constants[arg].
	OpSetGlobal OpCode = 0x13
	// OpDefGlobal pops TOS and defines (not re-assigns) global Constants[arg].
	OpDefGlobal OpCode = 0x14
	// OpGetUpvalue pushes upvalue[arg] from the current closure.
	OpGetUpvalue OpCode = 0x15
	// OpSetUpvalue pops TOS and stores into upvalue[arg].
	OpSetUpvalue OpCode = 0x16
)

// --- Stack manipulation (0x20–0x2F) ---
const (
	// OpPop discards TOS.
	OpPop OpCode = 0x20
	// OpDup duplicates TOS.
	OpDup OpCode = 0x21
	// OpSwap swaps TOS and TOS-1.
	OpSwap OpCode = 0x22
	// OpPopN pops arg items from the stack.
	OpPopN OpCode = 0x23
)

// --- Arithmetic (0x30–0x3F) ---
const (
	// OpAdd pops two values, pushes their sum.
	OpAdd OpCode = 0x30
	// OpSub pops two values, pushes their difference (TOS-1 - TOS).
	OpSub OpCode = 0x31
	// OpMul pops two values, pushes their product.
	OpMul OpCode = 0x32
	// OpDiv pops two values, pushes their quotient.
	OpDiv OpCode = 0x33
	// OpMod pops two values, pushes their remainder.
	OpMod OpCode = 0x34
	// OpPow pops two values, pushes TOS-1 ** TOS.
	OpPow OpCode = 0x35
	// OpNeg pops one value, pushes its arithmetic negation.
	OpNeg OpCode = 0x36
	// OpConcat pops two strings, pushes their concatenation.
	OpConcat OpCode = 0x37
)

// --- Comparison (0x40–0x4F) ---
const (
	// OpEq pops two values, pushes true if equal.
	OpEq OpCode = 0x40
	// OpNeq pops two values, pushes true if not equal.
	OpNeq OpCode = 0x41
	// OpLt pops two values, pushes true if TOS-1 < TOS.
	OpLt OpCode = 0x42
	// OpLte pops two values, pushes true if TOS-1 <= TOS.
	OpLte OpCode = 0x43
	// OpGt pops two values, pushes true if TOS-1 > TOS.
	OpGt OpCode = 0x44
	// OpGte pops two values, pushes true if TOS-1 >= TOS.
	OpGte OpCode = 0x45
	// OpIn pops two values, pushes true if TOS-1 is in TOS.
	OpIn OpCode = 0x46
	// OpNotIn pops two values, pushes true if TOS-1 is not in TOS.
	OpNotIn OpCode = 0x47
)

// --- Logical (0x50–0x5F) ---
const (
	// OpNot pops one value, pushes its boolean negation.
	OpNot OpCode = 0x50
	// OpAnd short-circuit AND: if TOS is falsy, jump arg bytes forward; else pop.
	OpAnd OpCode = 0x51
	// OpOr short-circuit OR: if TOS is truthy, jump arg bytes forward; else pop.
	OpOr OpCode = 0x52
)

// --- Jumps (0x60–0x6F) ---
const (
	// OpJump unconditionally jumps arg bytes forward from the start of this instruction.
	OpJump OpCode = 0x60
	// OpJumpIfFalse pops TOS; if falsy, jumps arg bytes forward.
	OpJumpIfFalse OpCode = 0x61
	// OpJumpIfTrue pops TOS; if truthy, jumps arg bytes forward.
	OpJumpIfTrue OpCode = 0x62
	// OpLoop jumps arg bytes backward (for while/for loops).
	OpLoop OpCode = 0x63
	// OpJumpBack jumps arg bytes backward (unconditional).
	OpJumpBack OpCode = 0x64
)

// --- Calls and closures (0x70–0x7F) ---
const (
	// OpCall calls the function at TOS-(arg+1) with arg arguments.
	// After call, leaves return value on stack.
	OpCall OpCode = 0x70
	// OpCallMethod calls a method on a receiver.
	// arg packs: high byte = method-name constant index, low byte = argument count.
	// Receiver is on the stack below the arguments.
	OpCallMethod OpCode = 0x71
	// OpReturn returns TOS to the caller.
	OpReturn OpCode = 0x72
	// OpClosure creates a closure from the chunk at Constants[arg], then reads
	// 2*upvalueCount additional bytes (isLocal byte + index byte per upvalue).
	OpClosure OpCode = 0x73
	// OpPipeline is equivalent to OpCall with 1 argument (pipeline expression marker).
	OpPipeline OpCode = 0x74
)

// --- Collections (0x80–0x8F) ---
const (
	// OpMakeList pops arg items and builds a List (bottom-most is first element).
	OpMakeList OpCode = 0x80
	// OpMakeMap pops arg*2 items (alternating key, value) and builds a Map.
	OpMakeMap OpCode = 0x81
	// OpMakeSet pops arg items and builds a Set.
	OpMakeSet OpCode = 0x82
	// OpMakeTuple pops arg items and builds a Tuple.
	OpMakeTuple OpCode = 0x83
)

// --- Field / Index access (0x90–0x9F) ---
const (
	// OpGetField pops TOS (struct/map), pushes field Constants[arg].
	OpGetField OpCode = 0x90
	// OpSetField pops TOS (value) and TOS-1 (object), sets field Constants[arg].
	OpSetField OpCode = 0x91
	// OpGetIndex pops TOS (index) and TOS-1 (collection), pushes the element.
	OpGetIndex OpCode = 0x92
	// OpSetIndex pops TOS (value), TOS-1 (index), TOS-2 (collection), sets element.
	OpSetIndex OpCode = 0x93
	// OpOptChain pops TOS; if None pushes None; else pushes TOS.Constants[arg].
	OpOptChain OpCode = 0x94
	// OpUnwrap pops TOS; if Some(x) pushes x; if None panics with runtime error.
	OpUnwrap OpCode = 0x95
)

// --- Struct / Enum construction (0xA0–0xAF) ---
const (
	// OpMakeStruct pops arg*2 key/value pairs and builds a struct instance.
	// Constants[arg] holds the struct name.
	OpMakeStruct OpCode = 0xA0
	// OpMakeVariant constructs an enum variant named Constants[high byte of arg].
	// Low byte of arg is the number of payload fields to pop.
	OpMakeVariant OpCode = 0xA1
)

// --- Option / Result constructors (0xB0–0xBF) ---
const (
	// OpSome pops TOS and wraps it in Some(x).
	OpSome OpCode = 0xB0
	// OpNone pushes the None value.
	OpNone OpCode = 0xB1
	// OpOk pops TOS and wraps it in Ok(x).
	OpOk OpCode = 0xB2
	// OpErr pops TOS and wraps it in Err(x).
	OpErr OpCode = 0xB3
	// OpIsSome pops TOS, pushes true if it is Some(_).
	OpIsSome OpCode = 0xB4
	// OpIsNone pops TOS, pushes true if it is None.
	OpIsNone OpCode = 0xB5
	// OpIsOk pops TOS, pushes true if it is Ok(_).
	OpIsOk OpCode = 0xB6
	// OpIsErr pops TOS, pushes true if it is Err(_).
	OpIsErr OpCode = 0xB7
	// OpUnwrapSome pops TOS; if Some(x) pushes x, else runtime error.
	OpUnwrapSome OpCode = 0xB8
	// OpUnwrapOk pops TOS; if Ok(x) pushes x, else runtime error.
	OpUnwrapOk OpCode = 0xB9
	// OpUnwrapErr pops TOS; if Err(e) pushes e, else runtime error.
	OpUnwrapErr OpCode = 0xBA
)

// --- String interpolation (0xC0–0xCF) ---
const (
	// OpInterpolate pops arg items from the stack and joins them as a string.
	OpInterpolate OpCode = 0xC0
)

// --- Pattern matching (0xD0–0xDF) ---
const (
	// OpMatchLiteral pops TOS, compares with Constants[arg]; pushes bool.
	OpMatchLiteral OpCode = 0xD0
	// OpMatchType checks that TOS is an instance of the type named Constants[arg]; pushes bool.
	OpMatchType OpCode = 0xD1
	// OpMatchVariant checks that TOS is the enum variant named Constants[arg]; pushes bool.
	OpMatchVariant OpCode = 0xD2
	// OpDestructureVariant pops TOS (enum variant), pushes its payload fields.
	// arg is the expected number of payload fields.
	OpDestructureVariant OpCode = 0xD3
	// OpDestructureStruct pops TOS (struct), pushes the value of field Constants[arg].
	OpDestructureStruct OpCode = 0xD4
	// OpMatchSome checks that TOS is Some(_); pushes bool.
	OpMatchSome OpCode = 0xD5
	// OpMatchNone checks that TOS is None; pushes bool.
	OpMatchNone OpCode = 0xD6
	// OpMatchOk checks that TOS is Ok(_); pushes bool.
	OpMatchOk OpCode = 0xD7
	// OpMatchErr checks that TOS is Err(_); pushes bool.
	OpMatchErr OpCode = 0xD8
)

// --- Misc (0xE0–0xFF) ---
const (
	// OpAssert pops TOS; if falsy raises runtime error with message Constants[arg].
	OpAssert OpCode = 0xE0
	// OpPrint pops TOS and prints it (used in REPL mode).
	OpPrint OpCode = 0xE1
	// OpBuildRange pops two ints (lo, hi), pushes a Range value.
	OpBuildRange OpCode = 0xE2
	// OpForIter advances the iterator at TOS; if exhausted jumps arg bytes forward,
	// else pushes next item onto stack.
	OpForIter OpCode = 0xE3
	// OpMakeIter pops TOS (list/map/range/string), pushes an iterator object.
	OpMakeIter OpCode = 0xE4
	// OpHalt stops execution (end of top-level module).
	OpHalt OpCode = 0xE5
)

// opNames maps each OpCode to a human-readable mnemonic for the disassembler.
var opNames = map[OpCode]string{
	OpConst:  "CONST",
	OpNil:    "NIL",
	OpTrue:   "TRUE",
	OpFalse:  "FALSE",

	OpGetLocal:   "GET_LOCAL",
	OpSetLocal:   "SET_LOCAL",
	OpGetGlobal:  "GET_GLOBAL",
	OpSetGlobal:  "SET_GLOBAL",
	OpDefGlobal:  "DEF_GLOBAL",
	OpGetUpvalue: "GET_UPVALUE",
	OpSetUpvalue: "SET_UPVALUE",

	OpPop:  "POP",
	OpDup:  "DUP",
	OpSwap: "SWAP",
	OpPopN: "POP_N",

	OpAdd:    "ADD",
	OpSub:    "SUB",
	OpMul:    "MUL",
	OpDiv:    "DIV",
	OpMod:    "MOD",
	OpPow:    "POW",
	OpNeg:    "NEG",
	OpConcat: "CONCAT",

	OpEq:    "EQ",
	OpNeq:   "NEQ",
	OpLt:    "LT",
	OpLte:   "LTE",
	OpGt:    "GT",
	OpGte:   "GTE",
	OpIn:    "IN",
	OpNotIn: "NOT_IN",

	OpNot: "NOT",
	OpAnd: "AND",
	OpOr:  "OR",

	OpJump:        "JUMP",
	OpJumpIfFalse: "JUMP_IF_FALSE",
	OpJumpIfTrue:  "JUMP_IF_TRUE",
	OpLoop:        "LOOP",
	OpJumpBack:    "JUMP_BACK",

	OpCall:       "CALL",
	OpCallMethod: "CALL_METHOD",
	OpReturn:     "RETURN",
	OpClosure:    "CLOSURE",
	OpPipeline:   "PIPELINE",

	OpMakeList:  "MAKE_LIST",
	OpMakeMap:   "MAKE_MAP",
	OpMakeSet:   "MAKE_SET",
	OpMakeTuple: "MAKE_TUPLE",

	OpGetField: "GET_FIELD",
	OpSetField: "SET_FIELD",
	OpGetIndex: "GET_INDEX",
	OpSetIndex: "SET_INDEX",
	OpOptChain: "OPT_CHAIN",
	OpUnwrap:   "UNWRAP",

	OpMakeStruct:  "MAKE_STRUCT",
	OpMakeVariant: "MAKE_VARIANT",

	OpSome:       "SOME",
	OpNone:       "NONE",
	OpOk:         "OK",
	OpErr:        "ERR",
	OpIsSome:     "IS_SOME",
	OpIsNone:     "IS_NONE",
	OpIsOk:       "IS_OK",
	OpIsErr:      "IS_ERR",
	OpUnwrapSome: "UNWRAP_SOME",
	OpUnwrapOk:   "UNWRAP_OK",
	OpUnwrapErr:  "UNWRAP_ERR",

	OpInterpolate: "INTERPOLATE",

	OpMatchLiteral:       "MATCH_LITERAL",
	OpMatchType:          "MATCH_TYPE",
	OpMatchVariant:       "MATCH_VARIANT",
	OpDestructureVariant: "DESTRUCTURE_VARIANT",
	OpDestructureStruct:  "DESTRUCTURE_STRUCT",
	OpMatchSome:          "MATCH_SOME",
	OpMatchNone:          "MATCH_NONE",
	OpMatchOk:            "MATCH_OK",
	OpMatchErr:           "MATCH_ERR",

	OpAssert:     "ASSERT",
	OpPrint:      "PRINT",
	OpBuildRange: "BUILD_RANGE",
	OpForIter:    "FOR_ITER",
	OpMakeIter:   "MAKE_ITER",
	OpHalt:       "HALT",
}

// Name returns the mnemonic for op, or "UNKNOWN(0xXX)" if unrecognised.
func (op OpCode) Name() string {
	if name, ok := opNames[op]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(0x%02X)", byte(op))
}

// ReadArg reads the 16-bit argument from code[offset+1..offset+2].
func ReadArg(code []byte, offset int) uint16 {
	return uint16(code[offset+1])<<8 | uint16(code[offset+2])
}

// WriteArg encodes a 16-bit argument into code[offset+1..offset+2].
func WriteArg(code []byte, offset int, arg uint16) {
	code[offset+1] = byte(arg >> 8)
	code[offset+2] = byte(arg)
}
