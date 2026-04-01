package compiler

import (
	"fmt"
	"strings"
)

// Disassemble returns a human-readable listing of all instructions in ch.
// Nested chunks (closures) are expanded inline, indented.
func Disassemble(ch *Chunk) string {
	var sb strings.Builder
	disassembleChunk(&sb, ch, 0)
	return sb.String()
}

func disassembleChunk(sb *strings.Builder, ch *Chunk, indent int) {
	prefix := strings.Repeat("  ", indent)
	fmt.Fprintf(sb, "%s== %s ==\n", prefix, ch.Name)
	fmt.Fprintf(sb, "%s  constants: %d  upvalues: %d  arity: %d\n",
		prefix, len(ch.Constants), len(ch.Upvalues), ch.Arity)

	i := 0
	for i < len(ch.Code) {
		offset := i
		op := OpCode(ch.Code[i])
		arg := ReadArg(ch.Code, i)
		i += InstructionSize

		line := ch.LineForOffset(offset)
		lineStr := fmt.Sprintf("%4d", line)
		if line < 0 {
			lineStr = "   ?"
		}

		fmt.Fprintf(sb, "%s  %04d %s  %-20s", prefix, offset, lineStr, op.Name())

		switch op {
		case OpConst:
			if int(arg) < len(ch.Constants) {
				fmt.Fprintf(sb, " [%d] %s", arg, ch.Constants[arg])
			} else {
				fmt.Fprintf(sb, " [%d] <out of range>", arg)
			}

		case OpGetGlobal, OpSetGlobal, OpDefGlobal,
			OpGetField, OpSetField,
			OpMatchLiteral, OpMatchType, OpMatchVariant,
			OpDestructureStruct, OpMakeStruct, OpMakeVariant,
			OpAssert, OpOptChain, OpInterpolate:
			if int(arg) < len(ch.Constants) {
				fmt.Fprintf(sb, " [%d] %s", arg, ch.Constants[arg])
			} else {
				fmt.Fprintf(sb, " [%d]", arg)
			}

		case OpGetLocal, OpSetLocal:
			name := ""
			if int(arg) < len(ch.Locals) {
				name = " (" + ch.Locals[arg].Name + ")"
			}
			fmt.Fprintf(sb, " slot=%d%s", arg, name)

		case OpGetUpvalue, OpSetUpvalue:
			if int(arg) < len(ch.Upvalues) {
				uv := ch.Upvalues[arg]
				if uv.IsLocal {
					fmt.Fprintf(sb, " upvalue=%d (local %d)", arg, uv.Index)
				} else {
					fmt.Fprintf(sb, " upvalue=%d (outer upvalue %d)", arg, uv.Index)
				}
			} else {
				fmt.Fprintf(sb, " upvalue=%d", arg)
			}

		case OpCall, OpMakeList, OpMakeMap, OpMakeSet, OpMakeTuple,
			OpPopN, OpDestructureVariant:
			fmt.Fprintf(sb, " count=%d", arg)

		case OpJump, OpJumpIfFalse, OpJumpIfTrue:
			// arg is bytes forward from start of this instruction
			target := offset + int(arg)
			fmt.Fprintf(sb, " -> %04d", target)

		case OpLoop, OpJumpBack:
			// arg is bytes backward from start of this instruction
			target := offset - int(arg)
			fmt.Fprintf(sb, " -> %04d", target)

		case OpForIter:
			// if iterator exhausted, jump arg bytes forward
			target := offset + int(arg)
			fmt.Fprintf(sb, " done -> %04d", target)

		case OpCallMethod:
			// arg encodes: high byte = method name constant index, low byte = arg count
			// But since arg is 16 bits and we want separate fields, we use the full arg
			// for method name index, and the arg count is in the next instruction slot.
			// Actually: see compiler.go — we pack methodIdx<<8 | argCount when both fit in 8 bits.
			// Disassemble generically here.
			methodIdx := arg >> 8
			argCount := arg & 0xFF
			methodName := ""
			if int(methodIdx) < len(ch.Constants) {
				methodName = ch.Constants[methodIdx].Val.(string)
			}
			fmt.Fprintf(sb, " method=%q args=%d", methodName, argCount)

		case OpClosure:
			if int(arg) < len(ch.Constants) && ch.Constants[arg].Kind == ConstChunk {
				inner := ch.Constants[arg].Val.(*Chunk)
				fmt.Fprintf(sb, " <closure %q> upvalues=%d", inner.Name, len(inner.Upvalues))
			} else {
				fmt.Fprintf(sb, " [%d]", arg)
			}

		case OpBuildRange:
			// no extra args to print

		default:
			if arg != 0 {
				fmt.Fprintf(sb, " arg=%d", arg)
			}
		}

		fmt.Fprintln(sb)

		// For OpClosure, recursively disassemble the nested chunk.
		if op == OpClosure && int(arg) < len(ch.Constants) && ch.Constants[arg].Kind == ConstChunk {
			inner := ch.Constants[arg].Val.(*Chunk)
			disassembleChunk(sb, inner, indent+1)
		}
	}

	// Print nested chunks embedded as constants (non-closure references).
	for _, c := range ch.Constants {
		if c.Kind == ConstChunk {
			// Only print if it wasn't already printed inline by OpClosure above.
			// We print them all here; duplicates in verbose output are acceptable.
			_ = c
		}
	}
}

// DisassembleInstruction returns a one-line description of the instruction
// at the given offset within ch. Used for runtime error messages.
func DisassembleInstruction(ch *Chunk, offset int) string {
	if offset+InstructionSize > len(ch.Code) {
		return "<out of bounds>"
	}
	op := OpCode(ch.Code[offset])
	return fmt.Sprintf("%04d %s", offset, op.Name())
}
