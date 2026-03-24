package interpreter

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// createStdJsonExports creates exports for the std.json module.
func createStdJsonExports() map[string]Value {
	exports := make(map[string]Value)

	// parse(str) - Parse JSON string to Aura values
	exports["parse"] = &BuiltinFnVal{
		Name: "json.parse",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "json.parse() requires exactly one argument"})
			}
			s, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "json.parse() argument must be a string"})
			}
			p := &jsonParser{input: s.Val, pos: 0}
			val := p.parseValue()
			p.skipWhitespace()
			if p.pos < len(p.input) {
				panic(&RuntimeError{Message: fmt.Sprintf("json.parse(): unexpected trailing content at position %d", p.pos)})
			}
			return val
		},
	}

	// stringify(value, pretty?) - Convert Aura value to JSON string
	exports["stringify"] = &BuiltinFnVal{
		Name: "json.stringify",
		Fn: func(args []Value) Value {
			if len(args) < 1 || len(args) > 2 {
				panic(&RuntimeError{Message: "json.stringify() requires 1-2 arguments (value, pretty?)"})
			}
			pretty := false
			if len(args) >= 2 {
				if b, ok := args[1].(*BoolVal); ok {
					pretty = b.Val
				}
			}
			result := jsonStringify(args[0], pretty, 0)
			return &StringVal{Val: result}
		},
	}

	return exports
}

// --- JSON Parser ---

type jsonParser struct {
	input string
	pos   int
}

func (p *jsonParser) parseValue() Value {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		panic(&RuntimeError{Message: "json.parse(): unexpected end of input"})
	}

	ch := p.input[p.pos]
	switch {
	case ch == '"':
		return p.parseString()
	case ch == '{':
		return p.parseObject()
	case ch == '[':
		return p.parseArray()
	case ch == 't' || ch == 'f':
		return p.parseBool()
	case ch == 'n':
		return p.parseNull()
	case ch == '-' || (ch >= '0' && ch <= '9'):
		return p.parseNumber()
	default:
		panic(&RuntimeError{Message: fmt.Sprintf("json.parse(): unexpected character '%c' at position %d", ch, p.pos)})
	}
}

func (p *jsonParser) parseString() Value {
	if p.pos >= len(p.input) || p.input[p.pos] != '"' {
		panic(&RuntimeError{Message: "json.parse(): expected '\"'"})
	}
	p.pos++ // skip opening quote

	var sb strings.Builder
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == '"' {
			p.pos++ // skip closing quote
			return &StringVal{Val: sb.String()}
		}
		if ch == '\\' {
			p.pos++
			if p.pos >= len(p.input) {
				panic(&RuntimeError{Message: "json.parse(): unexpected end of input in string escape"})
			}
			esc := p.input[p.pos]
			switch esc {
			case '"':
				sb.WriteByte('"')
			case '\\':
				sb.WriteByte('\\')
			case '/':
				sb.WriteByte('/')
			case 'b':
				sb.WriteByte('\b')
			case 'f':
				sb.WriteByte('\f')
			case 'n':
				sb.WriteByte('\n')
			case 'r':
				sb.WriteByte('\r')
			case 't':
				sb.WriteByte('\t')
			case 'u':
				// Parse 4-digit hex unicode escape
				if p.pos+4 >= len(p.input) {
					panic(&RuntimeError{Message: "json.parse(): invalid unicode escape"})
				}
				hex := p.input[p.pos+1 : p.pos+5]
				code, err := strconv.ParseInt(hex, 16, 32)
				if err != nil {
					panic(&RuntimeError{Message: fmt.Sprintf("json.parse(): invalid unicode escape: \\u%s", hex)})
				}
				sb.WriteRune(rune(code))
				p.pos += 4
			default:
				panic(&RuntimeError{Message: fmt.Sprintf("json.parse(): invalid escape character '\\%c'", esc)})
			}
		} else {
			sb.WriteByte(ch)
		}
		p.pos++
	}
	panic(&RuntimeError{Message: "json.parse(): unterminated string"})
}

func (p *jsonParser) parseNumber() Value {
	start := p.pos
	isFloat := false

	if p.pos < len(p.input) && p.input[p.pos] == '-' {
		p.pos++
	}

	for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
		p.pos++
	}

	if p.pos < len(p.input) && p.input[p.pos] == '.' {
		isFloat = true
		p.pos++
		for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
			p.pos++
		}
	}

	if p.pos < len(p.input) && (p.input[p.pos] == 'e' || p.input[p.pos] == 'E') {
		isFloat = true
		p.pos++
		if p.pos < len(p.input) && (p.input[p.pos] == '+' || p.input[p.pos] == '-') {
			p.pos++
		}
		for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
			p.pos++
		}
	}

	numStr := p.input[start:p.pos]

	if isFloat {
		f, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			panic(&RuntimeError{Message: fmt.Sprintf("json.parse(): invalid number: %s", numStr)})
		}
		return &FloatVal{Val: f}
	}

	i, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		// Fallback to float for very large integers
		f, err2 := strconv.ParseFloat(numStr, 64)
		if err2 != nil {
			panic(&RuntimeError{Message: fmt.Sprintf("json.parse(): invalid number: %s", numStr)})
		}
		return &FloatVal{Val: f}
	}
	return &IntVal{Val: i}
}

func (p *jsonParser) parseBool() Value {
	if strings.HasPrefix(p.input[p.pos:], "true") {
		p.pos += 4
		return &BoolVal{Val: true}
	}
	if strings.HasPrefix(p.input[p.pos:], "false") {
		p.pos += 5
		return &BoolVal{Val: false}
	}
	panic(&RuntimeError{Message: fmt.Sprintf("json.parse(): unexpected token at position %d", p.pos)})
}

func (p *jsonParser) parseNull() Value {
	if strings.HasPrefix(p.input[p.pos:], "null") {
		p.pos += 4
		return &NoneVal{}
	}
	panic(&RuntimeError{Message: fmt.Sprintf("json.parse(): unexpected token at position %d", p.pos)})
}

func (p *jsonParser) parseArray() Value {
	p.pos++ // skip '['
	p.skipWhitespace()

	elements := make([]Value, 0)

	if p.pos < len(p.input) && p.input[p.pos] == ']' {
		p.pos++
		return &ListVal{Elements: elements}
	}

	for {
		val := p.parseValue()
		elements = append(elements, val)

		p.skipWhitespace()
		if p.pos >= len(p.input) {
			panic(&RuntimeError{Message: "json.parse(): unterminated array"})
		}
		if p.input[p.pos] == ']' {
			p.pos++
			return &ListVal{Elements: elements}
		}
		if p.input[p.pos] != ',' {
			panic(&RuntimeError{Message: fmt.Sprintf("json.parse(): expected ',' or ']' at position %d", p.pos)})
		}
		p.pos++ // skip ','
	}
}

func (p *jsonParser) parseObject() Value {
	p.pos++ // skip '{'
	p.skipWhitespace()

	keys := make([]Value, 0)
	values := make([]Value, 0)

	if p.pos < len(p.input) && p.input[p.pos] == '}' {
		p.pos++
		return &MapVal{Keys: keys, Values: values}
	}

	for {
		p.skipWhitespace()
		// Parse key (must be string)
		key := p.parseString()

		p.skipWhitespace()
		if p.pos >= len(p.input) || p.input[p.pos] != ':' {
			panic(&RuntimeError{Message: fmt.Sprintf("json.parse(): expected ':' at position %d", p.pos)})
		}
		p.pos++ // skip ':'

		val := p.parseValue()

		keys = append(keys, key)
		values = append(values, val)

		p.skipWhitespace()
		if p.pos >= len(p.input) {
			panic(&RuntimeError{Message: "json.parse(): unterminated object"})
		}
		if p.input[p.pos] == '}' {
			p.pos++
			return &MapVal{Keys: keys, Values: values}
		}
		if p.input[p.pos] != ',' {
			panic(&RuntimeError{Message: fmt.Sprintf("json.parse(): expected ',' or '}' at position %d", p.pos)})
		}
		p.pos++ // skip ','
	}
}

func (p *jsonParser) skipWhitespace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}

// --- JSON Stringify ---

func jsonStringify(val Value, pretty bool, indent int) string {
	switch v := val.(type) {
	case *StringVal:
		return strconv.Quote(v.Val)
	case *IntVal:
		return strconv.FormatInt(v.Val, 10)
	case *FloatVal:
		return strconv.FormatFloat(v.Val, 'f', -1, 64)
	case *BoolVal:
		if v.Val {
			return "true"
		}
		return "false"
	case *NoneVal:
		return "null"
	case *OptionVal:
		if !v.IsSome {
			return "null"
		}
		return jsonStringify(v.Val, pretty, indent)
	case *ListVal:
		return jsonStringifyList(v, pretty, indent)
	case *MapVal:
		return jsonStringifyMap(v, pretty, indent)
	case *StructVal:
		return jsonStringifyStruct(v, pretty, indent)
	default:
		return strconv.Quote(val.String())
	}
}

func jsonStringifyList(v *ListVal, pretty bool, indent int) string {
	if len(v.Elements) == 0 {
		return "[]"
	}

	if !pretty {
		parts := make([]string, len(v.Elements))
		for i, el := range v.Elements {
			parts[i] = jsonStringify(el, false, 0)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	}

	var sb strings.Builder
	sb.WriteString("[\n")
	for i, el := range v.Elements {
		sb.WriteString(strings.Repeat("  ", indent+1))
		sb.WriteString(jsonStringify(el, true, indent+1))
		if i < len(v.Elements)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(strings.Repeat("  ", indent))
	sb.WriteString("]")
	return sb.String()
}

func jsonStringifyMap(v *MapVal, pretty bool, indent int) string {
	if len(v.Keys) == 0 {
		return "{}"
	}

	if !pretty {
		parts := make([]string, len(v.Keys))
		for i := range v.Keys {
			key := jsonStringifyKey(v.Keys[i])
			val := jsonStringify(v.Values[i], false, 0)
			parts[i] = key + ": " + val
		}
		return "{" + strings.Join(parts, ", ") + "}"
	}

	var sb strings.Builder
	sb.WriteString("{\n")
	for i := range v.Keys {
		sb.WriteString(strings.Repeat("  ", indent+1))
		key := jsonStringifyKey(v.Keys[i])
		val := jsonStringify(v.Values[i], true, indent+1)
		sb.WriteString(key + ": " + val)
		if i < len(v.Keys)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(strings.Repeat("  ", indent))
	sb.WriteString("}")
	return sb.String()
}

func jsonStringifyStruct(v *StructVal, pretty bool, indent int) string {
	keys := make([]Value, 0, len(v.Fields))
	vals := make([]Value, 0, len(v.Fields))
	for k, val := range v.Fields {
		keys = append(keys, &StringVal{Val: k})
		vals = append(vals, val)
	}
	return jsonStringifyMap(&MapVal{Keys: keys, Values: vals}, pretty, indent)
}

func jsonStringifyKey(v Value) string {
	if s, ok := v.(*StringVal); ok {
		return strconv.Quote(s.Val)
	}
	return strconv.Quote(v.String())
}
