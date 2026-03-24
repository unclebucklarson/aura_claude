package interpreter

import (
	"math/rand"
	"strings"
	"testing"
)

// ============================================================
// std.regex tests
// ============================================================

func TestStdRegexMatch(t *testing.T) {
	exports := createStdRegexExports()
	matchFn := exports["match"].(*BuiltinFnVal).Fn

	tests := []struct {
		pattern, text string
		expected      bool
	}{
		{`\d+`, "abc123", true},
		{`\d+`, "abcdef", false},
		{`^hello`, "hello world", true},
		{`^hello`, "world hello", false},
		{`[aeiou]`, "rhythm", false},
		{`[aeiou]`, "hello", true},
	}
	for _, tt := range tests {
		result := matchFn([]Value{&StringVal{Val: tt.pattern}, &StringVal{Val: tt.text}})
		b := result.(*BoolVal)
		if b.Val != tt.expected {
			t.Errorf("regex.match(%q, %q) = %v, want %v", tt.pattern, tt.text, b.Val, tt.expected)
		}
	}
}

func TestStdRegexFind(t *testing.T) {
	exports := createStdRegexExports()
	findFn := exports["find"].(*BuiltinFnVal).Fn

	// Found case
	result := findFn([]Value{&StringVal{Val: `\d+`}, &StringVal{Val: "abc123def456"}})
	opt := result.(*OptionVal)
	if !opt.IsSome {
		t.Fatal("expected Some, got None")
	}
	if opt.Val.(*StringVal).Val != "123" {
		t.Errorf("expected '123', got %q", opt.Val.(*StringVal).Val)
	}

	// Not found case
	result = findFn([]Value{&StringVal{Val: `\d+`}, &StringVal{Val: "abcdef"}})
	opt = result.(*OptionVal)
	if opt.IsSome {
		t.Fatal("expected None, got Some")
	}
}

func TestStdRegexFindAll(t *testing.T) {
	exports := createStdRegexExports()
	fn := exports["find_all"].(*BuiltinFnVal).Fn

	result := fn([]Value{&StringVal{Val: `\d+`}, &StringVal{Val: "a1b22c333"}})
	list := result.(*ListVal)
	if len(list.Elements) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(list.Elements))
	}
	expected := []string{"1", "22", "333"}
	for i, e := range expected {
		if list.Elements[i].(*StringVal).Val != e {
			t.Errorf("match[%d] = %q, want %q", i, list.Elements[i].(*StringVal).Val, e)
		}
	}

	// No matches
	result = fn([]Value{&StringVal{Val: `\d+`}, &StringVal{Val: "abc"}})
	list = result.(*ListVal)
	if len(list.Elements) != 0 {
		t.Errorf("expected 0 matches, got %d", len(list.Elements))
	}
}

func TestStdRegexReplace(t *testing.T) {
	exports := createStdRegexExports()
	fn := exports["replace"].(*BuiltinFnVal).Fn

	result := fn([]Value{&StringVal{Val: `\d+`}, &StringVal{Val: "a1b2c3"}, &StringVal{Val: "X"}})
	if result.(*StringVal).Val != "aXbXcX" {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, "aXbXcX")
	}
}

func TestStdRegexSplit(t *testing.T) {
	exports := createStdRegexExports()
	fn := exports["split"].(*BuiltinFnVal).Fn

	result := fn([]Value{&StringVal{Val: `[,;]+`}, &StringVal{Val: "a,b;;c,d"}})
	list := result.(*ListVal)
	if len(list.Elements) != 4 {
		t.Fatalf("expected 4 parts, got %d", len(list.Elements))
	}
	expected := []string{"a", "b", "c", "d"}
	for i, e := range expected {
		if list.Elements[i].(*StringVal).Val != e {
			t.Errorf("part[%d] = %q, want %q", i, list.Elements[i].(*StringVal).Val, e)
		}
	}
}

func TestStdRegexCompile(t *testing.T) {
	exports := createStdRegexExports()
	fn := exports["compile"].(*BuiltinFnVal).Fn

	// Valid pattern
	result := fn([]Value{&StringVal{Val: `\d+`}})
	r := result.(*ResultVal)
	if !r.IsOk {
		t.Fatal("expected Ok for valid pattern")
	}

	// Invalid pattern
	result = fn([]Value{&StringVal{Val: `[invalid`}})
	r = result.(*ResultVal)
	if r.IsOk {
		t.Fatal("expected Err for invalid pattern")
	}
}

func TestStdRegexInvalidPattern(t *testing.T) {
	exports := createStdRegexExports()
	fn := exports["match"].(*BuiltinFnVal).Fn

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for invalid regex")
		}
	}()
	fn([]Value{&StringVal{Val: `[invalid`}, &StringVal{Val: "test"}})
}

// ============================================================
// std.collections tests
// ============================================================

func TestStdCollectionsRange(t *testing.T) {
	exports := createStdCollectionsExports()
	fn := exports["range"].(*BuiltinFnVal).Fn

	// range(5)
	result := fn([]Value{&IntVal{Val: 5}})
	list := result.(*ListVal)
	if len(list.Elements) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(list.Elements))
	}
	for i := 0; i < 5; i++ {
		if list.Elements[i].(*IntVal).Val != int64(i) {
			t.Errorf("range(5)[%d] = %d, want %d", i, list.Elements[i].(*IntVal).Val, i)
		}
	}

	// range(2, 5)
	result = fn([]Value{&IntVal{Val: 2}, &IntVal{Val: 5}})
	list = result.(*ListVal)
	if len(list.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list.Elements))
	}

	// range(0, 10, 3)
	result = fn([]Value{&IntVal{Val: 0}, &IntVal{Val: 10}, &IntVal{Val: 3}})
	list = result.(*ListVal)
	expected := []int64{0, 3, 6, 9}
	if len(list.Elements) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(list.Elements))
	}
	for i, e := range expected {
		if list.Elements[i].(*IntVal).Val != e {
			t.Errorf("[%d] = %d, want %d", i, list.Elements[i].(*IntVal).Val, e)
		}
	}

	// negative step
	result = fn([]Value{&IntVal{Val: 5}, &IntVal{Val: 0}, &IntVal{Val: -1}})
	list = result.(*ListVal)
	if len(list.Elements) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(list.Elements))
	}
}

func TestStdCollectionsZipWith(t *testing.T) {
	exports := createStdCollectionsExports()
	fn := exports["zip_with"].(*BuiltinFnVal).Fn

	addFn := &BuiltinFnVal{Name: "add", Fn: func(args []Value) Value {
		return &IntVal{Val: args[0].(*IntVal).Val + args[1].(*IntVal).Val}
	}}

	l1 := &ListVal{Elements: []Value{&IntVal{Val: 1}, &IntVal{Val: 2}, &IntVal{Val: 3}}}
	l2 := &ListVal{Elements: []Value{&IntVal{Val: 10}, &IntVal{Val: 20}, &IntVal{Val: 30}}}

	result := fn([]Value{addFn, l1, l2})
	list := result.(*ListVal)
	expected := []int64{11, 22, 33}
	for i, e := range expected {
		if list.Elements[i].(*IntVal).Val != e {
			t.Errorf("[%d] = %d, want %d", i, list.Elements[i].(*IntVal).Val, e)
		}
	}
}

func TestStdCollectionsZipWithUnequalLengths(t *testing.T) {
	exports := createStdCollectionsExports()
	fn := exports["zip_with"].(*BuiltinFnVal).Fn

	addFn := &BuiltinFnVal{Name: "add", Fn: func(args []Value) Value {
		return &IntVal{Val: args[0].(*IntVal).Val + args[1].(*IntVal).Val}
	}}
	l1 := &ListVal{Elements: []Value{&IntVal{Val: 1}, &IntVal{Val: 2}}}
	l2 := &ListVal{Elements: []Value{&IntVal{Val: 10}, &IntVal{Val: 20}, &IntVal{Val: 30}}}

	result := fn([]Value{addFn, l1, l2})
	list := result.(*ListVal)
	if len(list.Elements) != 2 {
		t.Errorf("expected 2 elements (min length), got %d", len(list.Elements))
	}
}

func TestStdCollectionsPartition(t *testing.T) {
	exports := createStdCollectionsExports()
	fn := exports["partition"].(*BuiltinFnVal).Fn

	isEven := &BuiltinFnVal{Name: "is_even", Fn: func(args []Value) Value {
		return &BoolVal{Val: args[0].(*IntVal).Val%2 == 0}
	}}
	list := &ListVal{Elements: []Value{
		&IntVal{Val: 1}, &IntVal{Val: 2}, &IntVal{Val: 3}, &IntVal{Val: 4}, &IntVal{Val: 5},
	}}

	result := fn([]Value{isEven, list})
	parts := result.(*ListVal)
	if len(parts.Elements) != 2 {
		t.Fatal("expected 2 partitions")
	}
	evens := parts.Elements[0].(*ListVal)
	odds := parts.Elements[1].(*ListVal)
	if len(evens.Elements) != 2 {
		t.Errorf("expected 2 evens, got %d", len(evens.Elements))
	}
	if len(odds.Elements) != 3 {
		t.Errorf("expected 3 odds, got %d", len(odds.Elements))
	}
}

func TestStdCollectionsGroupBy(t *testing.T) {
	exports := createStdCollectionsExports()
	fn := exports["group_by"].(*BuiltinFnVal).Fn

	modThree := &BuiltinFnVal{Name: "mod3", Fn: func(args []Value) Value {
		return &IntVal{Val: args[0].(*IntVal).Val % 3}
	}}
	list := &ListVal{Elements: []Value{
		&IntVal{Val: 1}, &IntVal{Val: 2}, &IntVal{Val: 3},
		&IntVal{Val: 4}, &IntVal{Val: 5}, &IntVal{Val: 6},
	}}

	result := fn([]Value{modThree, list})
	m := result.(*MapVal)
	if len(m.Keys) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(m.Keys))
	}
}

func TestStdCollectionsChunk(t *testing.T) {
	exports := createStdCollectionsExports()
	fn := exports["chunk"].(*BuiltinFnVal).Fn

	list := &ListVal{Elements: []Value{
		&IntVal{Val: 1}, &IntVal{Val: 2}, &IntVal{Val: 3}, &IntVal{Val: 4}, &IntVal{Val: 5},
	}}

	result := fn([]Value{&IntVal{Val: 2}, list})
	chunks := result.(*ListVal)
	if len(chunks.Elements) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks.Elements))
	}
	// First chunk: [1, 2]
	first := chunks.Elements[0].(*ListVal)
	if len(first.Elements) != 2 {
		t.Errorf("first chunk length = %d, want 2", len(first.Elements))
	}
	// Last chunk: [5]
	last := chunks.Elements[2].(*ListVal)
	if len(last.Elements) != 1 {
		t.Errorf("last chunk length = %d, want 1", len(last.Elements))
	}
}

func TestStdCollectionsTake(t *testing.T) {
	exports := createStdCollectionsExports()
	fn := exports["take"].(*BuiltinFnVal).Fn

	list := &ListVal{Elements: []Value{
		&IntVal{Val: 1}, &IntVal{Val: 2}, &IntVal{Val: 3}, &IntVal{Val: 4}, &IntVal{Val: 5},
	}}

	result := fn([]Value{&IntVal{Val: 3}, list})
	taken := result.(*ListVal)
	if len(taken.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(taken.Elements))
	}

	// Take more than available
	result = fn([]Value{&IntVal{Val: 10}, list})
	taken = result.(*ListVal)
	if len(taken.Elements) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(taken.Elements))
	}
}

func TestStdCollectionsDrop(t *testing.T) {
	exports := createStdCollectionsExports()
	fn := exports["drop"].(*BuiltinFnVal).Fn

	list := &ListVal{Elements: []Value{
		&IntVal{Val: 1}, &IntVal{Val: 2}, &IntVal{Val: 3}, &IntVal{Val: 4}, &IntVal{Val: 5},
	}}

	result := fn([]Value{&IntVal{Val: 2}, list})
	dropped := result.(*ListVal)
	if len(dropped.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(dropped.Elements))
	}
	if dropped.Elements[0].(*IntVal).Val != 3 {
		t.Errorf("first element = %d, want 3", dropped.Elements[0].(*IntVal).Val)
	}
}

func TestStdCollectionsTakeWhile(t *testing.T) {
	exports := createStdCollectionsExports()
	fn := exports["take_while"].(*BuiltinFnVal).Fn

	lessThan4 := &BuiltinFnVal{Name: "lt4", Fn: func(args []Value) Value {
		return &BoolVal{Val: args[0].(*IntVal).Val < 4}
	}}
	list := &ListVal{Elements: []Value{
		&IntVal{Val: 1}, &IntVal{Val: 2}, &IntVal{Val: 3}, &IntVal{Val: 4}, &IntVal{Val: 5},
	}}

	result := fn([]Value{lessThan4, list})
	taken := result.(*ListVal)
	if len(taken.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(taken.Elements))
	}
}

func TestStdCollectionsDropWhile(t *testing.T) {
	exports := createStdCollectionsExports()
	fn := exports["drop_while"].(*BuiltinFnVal).Fn

	lessThan4 := &BuiltinFnVal{Name: "lt4", Fn: func(args []Value) Value {
		return &BoolVal{Val: args[0].(*IntVal).Val < 4}
	}}
	list := &ListVal{Elements: []Value{
		&IntVal{Val: 1}, &IntVal{Val: 2}, &IntVal{Val: 3}, &IntVal{Val: 4}, &IntVal{Val: 5},
	}}

	result := fn([]Value{lessThan4, list})
	remaining := result.(*ListVal)
	if len(remaining.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(remaining.Elements))
	}
	if remaining.Elements[0].(*IntVal).Val != 4 {
		t.Errorf("first element = %d, want 4", remaining.Elements[0].(*IntVal).Val)
	}
}

func TestStdCollectionsEmptyList(t *testing.T) {
	exports := createStdCollectionsExports()
	emptyList := &ListVal{Elements: []Value{}}

	// chunk of empty
	chunkFn := exports["chunk"].(*BuiltinFnVal).Fn
	result := chunkFn([]Value{&IntVal{Val: 3}, emptyList})
	if len(result.(*ListVal).Elements) != 0 {
		t.Error("chunk of empty should be empty")
	}

	// take of empty
	takeFn := exports["take"].(*BuiltinFnVal).Fn
	result = takeFn([]Value{&IntVal{Val: 3}, emptyList})
	if len(result.(*ListVal).Elements) != 0 {
		t.Error("take of empty should be empty")
	}
}

// ============================================================
// std.random tests
// ============================================================

func TestStdRandomInt(t *testing.T) {
	exports := createStdRandomExports()
	fn := exports["int"].(*BuiltinFnVal).Fn

	rand.Seed(42)
	for i := 0; i < 50; i++ {
		result := fn([]Value{&IntVal{Val: 1}, &IntVal{Val: 10}})
		v := result.(*IntVal).Val
		if v < 1 || v > 10 {
			t.Errorf("random.int(1, 10) = %d, out of range", v)
		}
	}

	// Same min and max
	result := fn([]Value{&IntVal{Val: 5}, &IntVal{Val: 5}})
	if result.(*IntVal).Val != 5 {
		t.Errorf("random.int(5, 5) = %d, want 5", result.(*IntVal).Val)
	}
}

func TestStdRandomFloat(t *testing.T) {
	exports := createStdRandomExports()
	fn := exports["float"].(*BuiltinFnVal).Fn

	rand.Seed(42)
	for i := 0; i < 50; i++ {
		result := fn([]Value{})
		v := result.(*FloatVal).Val
		if v < 0.0 || v >= 1.0 {
			t.Errorf("random.float() = %f, out of [0, 1)", v)
		}
	}
}

func TestStdRandomChoice(t *testing.T) {
	exports := createStdRandomExports()
	fn := exports["choice"].(*BuiltinFnVal).Fn

	list := &ListVal{Elements: []Value{&IntVal{Val: 10}, &IntVal{Val: 20}, &IntVal{Val: 30}}}
	rand.Seed(42)
	for i := 0; i < 20; i++ {
		result := fn([]Value{list})
		v := result.(*IntVal).Val
		if v != 10 && v != 20 && v != 30 {
			t.Errorf("random.choice() = %d, not in list", v)
		}
	}
}

func TestStdRandomChoiceEmptyPanics(t *testing.T) {
	exports := createStdRandomExports()
	fn := exports["choice"].(*BuiltinFnVal).Fn

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty list")
		}
	}()
	fn([]Value{&ListVal{Elements: []Value{}}})
}

func TestStdRandomShuffle(t *testing.T) {
	exports := createStdRandomExports()
	fn := exports["shuffle"].(*BuiltinFnVal).Fn

	list := &ListVal{Elements: []Value{
		&IntVal{Val: 1}, &IntVal{Val: 2}, &IntVal{Val: 3}, &IntVal{Val: 4}, &IntVal{Val: 5},
	}}
	result := fn([]Value{list})
	shuffled := result.(*ListVal)
	if len(shuffled.Elements) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(shuffled.Elements))
	}
	// Original should be unchanged
	if list.Elements[0].(*IntVal).Val != 1 {
		t.Error("shuffle modified original list")
	}
}

func TestStdRandomSample(t *testing.T) {
	exports := createStdRandomExports()
	fn := exports["sample"].(*BuiltinFnVal).Fn

	list := &ListVal{Elements: []Value{
		&IntVal{Val: 1}, &IntVal{Val: 2}, &IntVal{Val: 3}, &IntVal{Val: 4}, &IntVal{Val: 5},
	}}

	rand.Seed(42)
	result := fn([]Value{list, &IntVal{Val: 3}})
	sample := result.(*ListVal)
	if len(sample.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(sample.Elements))
	}

	// Sample 0
	result = fn([]Value{list, &IntVal{Val: 0}})
	if len(result.(*ListVal).Elements) != 0 {
		t.Error("sample 0 should return empty list")
	}
}

func TestStdRandomSampleTooLargePanics(t *testing.T) {
	exports := createStdRandomExports()
	fn := exports["sample"].(*BuiltinFnVal).Fn

	list := &ListVal{Elements: []Value{&IntVal{Val: 1}, &IntVal{Val: 2}}}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for n > list length")
		}
	}()
	fn([]Value{list, &IntVal{Val: 5}})
}

func TestStdRandomSeed(t *testing.T) {
	exports := createStdRandomExports()
	seedFn := exports["seed"].(*BuiltinFnVal).Fn
	intFn := exports["int"].(*BuiltinFnVal).Fn

	// Set seed and get sequence
	seedFn([]Value{&IntVal{Val: 123}})
	v1 := intFn([]Value{&IntVal{Val: 0}, &IntVal{Val: 100}}).(*IntVal).Val

	// Reset seed and verify same sequence
	seedFn([]Value{&IntVal{Val: 123}})
	v2 := intFn([]Value{&IntVal{Val: 0}, &IntVal{Val: 100}}).(*IntVal).Val

	if v1 != v2 {
		t.Errorf("same seed should produce same result: %d != %d", v1, v2)
	}
}

// ============================================================
// std.format tests
// ============================================================

func TestStdFormatPadLeft(t *testing.T) {
	exports := createStdFormatExports()
	fn := exports["pad_left"].(*BuiltinFnVal).Fn

	result := fn([]Value{&StringVal{Val: "hi"}, &IntVal{Val: 5}})
	if result.(*StringVal).Val != "   hi" {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, "   hi")
	}

	// With custom char
	result = fn([]Value{&StringVal{Val: "hi"}, &IntVal{Val: 5}, &StringVal{Val: "0"}})
	if result.(*StringVal).Val != "000hi" {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, "000hi")
	}

	// Already long enough
	result = fn([]Value{&StringVal{Val: "hello"}, &IntVal{Val: 3}})
	if result.(*StringVal).Val != "hello" {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, "hello")
	}
}

func TestStdFormatPadRight(t *testing.T) {
	exports := createStdFormatExports()
	fn := exports["pad_right"].(*BuiltinFnVal).Fn

	result := fn([]Value{&StringVal{Val: "hi"}, &IntVal{Val: 5}})
	if result.(*StringVal).Val != "hi   " {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, "hi   ")
	}

	result = fn([]Value{&StringVal{Val: "hi"}, &IntVal{Val: 5}, &StringVal{Val: "."}})
	if result.(*StringVal).Val != "hi..." {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, "hi...")
	}
}

func TestStdFormatCenter(t *testing.T) {
	exports := createStdFormatExports()
	fn := exports["center"].(*BuiltinFnVal).Fn

	result := fn([]Value{&StringVal{Val: "hi"}, &IntVal{Val: 6}})
	if result.(*StringVal).Val != "  hi  " {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, "  hi  ")
	}

	// Odd padding
	result = fn([]Value{&StringVal{Val: "hi"}, &IntVal{Val: 7}})
	s := result.(*StringVal).Val
	if len(s) != 7 || !strings.Contains(s, "hi") {
		t.Errorf("got %q, want centered 'hi' in 7 chars", s)
	}

	// With custom char
	result = fn([]Value{&StringVal{Val: "hi"}, &IntVal{Val: 6}, &StringVal{Val: "*"}})
	if result.(*StringVal).Val != "**hi**" {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, "**hi**")
	}
}

func TestStdFormatTruncate(t *testing.T) {
	exports := createStdFormatExports()
	fn := exports["truncate"].(*BuiltinFnVal).Fn

	result := fn([]Value{&StringVal{Val: "hello world"}, &IntVal{Val: 8}})
	if result.(*StringVal).Val != "hello..." {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, "hello...")
	}

	// Custom suffix
	result = fn([]Value{&StringVal{Val: "hello world"}, &IntVal{Val: 8}, &StringVal{Val: "~"}})
	if result.(*StringVal).Val != "hello w~" {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, "hello w~")
	}

	// Short enough
	result = fn([]Value{&StringVal{Val: "hi"}, &IntVal{Val: 10}})
	if result.(*StringVal).Val != "hi" {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, "hi")
	}
}

func TestStdFormatWrap(t *testing.T) {
	exports := createStdFormatExports()
	fn := exports["wrap"].(*BuiltinFnVal).Fn

	result := fn([]Value{&StringVal{Val: "the quick brown fox jumps over the lazy dog"}, &IntVal{Val: 15}})
	lines := strings.Split(result.(*StringVal).Val, "\n")
	for _, line := range lines {
		if len(line) > 15 {
			t.Errorf("line %q exceeds width 15", line)
		}
	}
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines, got %d", len(lines))
	}

	// Empty string
	result = fn([]Value{&StringVal{Val: ""}, &IntVal{Val: 10}})
	if result.(*StringVal).Val != "" {
		t.Errorf("wrap empty should be empty, got %q", result.(*StringVal).Val)
	}
}

func TestStdFormatIndent(t *testing.T) {
	exports := createStdFormatExports()
	fn := exports["indent"].(*BuiltinFnVal).Fn

	result := fn([]Value{&StringVal{Val: "line1\nline2\nline3"}, &IntVal{Val: 4}})
	expected := "    line1\n    line2\n    line3"
	if result.(*StringVal).Val != expected {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, expected)
	}

	// Empty lines should not be indented
	result = fn([]Value{&StringVal{Val: "line1\n\nline3"}, &IntVal{Val: 2}})
	expected = "  line1\n\n  line3"
	if result.(*StringVal).Val != expected {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, expected)
	}
}

func TestStdFormatDedent(t *testing.T) {
	exports := createStdFormatExports()
	fn := exports["dedent"].(*BuiltinFnVal).Fn

	result := fn([]Value{&StringVal{Val: "    line1\n    line2\n    line3"}})
	expected := "line1\nline2\nline3"
	if result.(*StringVal).Val != expected {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, expected)
	}

	// Mixed indentation
	result = fn([]Value{&StringVal{Val: "    line1\n      line2\n    line3"}})
	expected = "line1\n  line2\nline3"
	if result.(*StringVal).Val != expected {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, expected)
	}

	// No common indentation
	result = fn([]Value{&StringVal{Val: "line1\nline2"}})
	if result.(*StringVal).Val != "line1\nline2" {
		t.Errorf("dedent with no indentation should be unchanged")
	}
}

// ============================================================
// std.result tests
// ============================================================

func TestStdResultAllOk(t *testing.T) {
	exports := createStdResultExports()
	fn := exports["all_ok"].(*BuiltinFnVal).Fn

	allOk := &ListVal{Elements: []Value{
		&ResultVal{IsOk: true, Val: &IntVal{Val: 1}},
		&ResultVal{IsOk: true, Val: &IntVal{Val: 2}},
	}}
	if !fn([]Value{allOk}).(*BoolVal).Val {
		t.Error("expected true for all Ok")
	}

	mixed := &ListVal{Elements: []Value{
		&ResultVal{IsOk: true, Val: &IntVal{Val: 1}},
		&ResultVal{IsOk: false, Val: &StringVal{Val: "err"}},
	}}
	if fn([]Value{mixed}).(*BoolVal).Val {
		t.Error("expected false for mixed")
	}

	// Empty list
	empty := &ListVal{Elements: []Value{}}
	if !fn([]Value{empty}).(*BoolVal).Val {
		t.Error("expected true for empty list")
	}
}

func TestStdResultAnyOk(t *testing.T) {
	exports := createStdResultExports()
	fn := exports["any_ok"].(*BuiltinFnVal).Fn

	allErr := &ListVal{Elements: []Value{
		&ResultVal{IsOk: false, Val: &StringVal{Val: "e1"}},
		&ResultVal{IsOk: false, Val: &StringVal{Val: "e2"}},
	}}
	if fn([]Value{allErr}).(*BoolVal).Val {
		t.Error("expected false for all Err")
	}

	mixed := &ListVal{Elements: []Value{
		&ResultVal{IsOk: false, Val: &StringVal{Val: "e1"}},
		&ResultVal{IsOk: true, Val: &IntVal{Val: 42}},
	}}
	if !fn([]Value{mixed}).(*BoolVal).Val {
		t.Error("expected true for mixed")
	}
}

func TestStdResultCollect(t *testing.T) {
	exports := createStdResultExports()
	fn := exports["collect"].(*BuiltinFnVal).Fn

	// All ok -> Ok([1, 2, 3])
	allOk := &ListVal{Elements: []Value{
		&ResultVal{IsOk: true, Val: &IntVal{Val: 1}},
		&ResultVal{IsOk: true, Val: &IntVal{Val: 2}},
		&ResultVal{IsOk: true, Val: &IntVal{Val: 3}},
	}}
	result := fn([]Value{allOk}).(*ResultVal)
	if !result.IsOk {
		t.Fatal("expected Ok")
	}
	collected := result.Val.(*ListVal)
	if len(collected.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(collected.Elements))
	}

	// With error -> first Err
	withErr := &ListVal{Elements: []Value{
		&ResultVal{IsOk: true, Val: &IntVal{Val: 1}},
		&ResultVal{IsOk: false, Val: &StringVal{Val: "fail"}},
		&ResultVal{IsOk: true, Val: &IntVal{Val: 3}},
	}}
	result = fn([]Value{withErr}).(*ResultVal)
	if result.IsOk {
		t.Fatal("expected Err")
	}
	if result.Val.(*StringVal).Val != "fail" {
		t.Errorf("expected 'fail', got %q", result.Val.(*StringVal).Val)
	}
}

func TestStdResultPartitionResults(t *testing.T) {
	exports := createStdResultExports()
	fn := exports["partition_results"].(*BuiltinFnVal).Fn

	list := &ListVal{Elements: []Value{
		&ResultVal{IsOk: true, Val: &IntVal{Val: 1}},
		&ResultVal{IsOk: false, Val: &StringVal{Val: "e1"}},
		&ResultVal{IsOk: true, Val: &IntVal{Val: 2}},
		&ResultVal{IsOk: false, Val: &StringVal{Val: "e2"}},
	}}
	result := fn([]Value{list}).(*ListVal)
	oks := result.Elements[0].(*ListVal)
	errs := result.Elements[1].(*ListVal)
	if len(oks.Elements) != 2 {
		t.Errorf("expected 2 oks, got %d", len(oks.Elements))
	}
	if len(errs.Elements) != 2 {
		t.Errorf("expected 2 errs, got %d", len(errs.Elements))
	}
}

func TestStdResultFromOption(t *testing.T) {
	exports := createStdResultExports()
	fn := exports["from_option"].(*BuiltinFnVal).Fn

	// Some -> Ok
	some := &OptionVal{IsSome: true, Val: &IntVal{Val: 42}}
	result := fn([]Value{some, &StringVal{Val: "missing"}}).(*ResultVal)
	if !result.IsOk || result.Val.(*IntVal).Val != 42 {
		t.Error("Some should convert to Ok(42)")
	}

	// None -> Err
	none := &OptionVal{IsSome: false}
	result = fn([]Value{none, &StringVal{Val: "missing"}}).(*ResultVal)
	if result.IsOk {
		t.Error("None should convert to Err")
	}
	if result.Val.(*StringVal).Val != "missing" {
		t.Errorf("expected 'missing', got %q", result.Val.(*StringVal).Val)
	}
}

// ============================================================
// std.option tests
// ============================================================

func TestStdOptionAllSome(t *testing.T) {
	exports := createStdOptionExports()
	fn := exports["all_some"].(*BuiltinFnVal).Fn

	allSome := &ListVal{Elements: []Value{
		&OptionVal{IsSome: true, Val: &IntVal{Val: 1}},
		&OptionVal{IsSome: true, Val: &IntVal{Val: 2}},
	}}
	if !fn([]Value{allSome}).(*BoolVal).Val {
		t.Error("expected true for all Some")
	}

	mixed := &ListVal{Elements: []Value{
		&OptionVal{IsSome: true, Val: &IntVal{Val: 1}},
		&OptionVal{IsSome: false},
	}}
	if fn([]Value{mixed}).(*BoolVal).Val {
		t.Error("expected false for mixed")
	}

	empty := &ListVal{Elements: []Value{}}
	if !fn([]Value{empty}).(*BoolVal).Val {
		t.Error("expected true for empty")
	}
}

func TestStdOptionAnySome(t *testing.T) {
	exports := createStdOptionExports()
	fn := exports["any_some"].(*BuiltinFnVal).Fn

	allNone := &ListVal{Elements: []Value{
		&OptionVal{IsSome: false},
		&OptionVal{IsSome: false},
	}}
	if fn([]Value{allNone}).(*BoolVal).Val {
		t.Error("expected false for all None")
	}

	mixed := &ListVal{Elements: []Value{
		&OptionVal{IsSome: false},
		&OptionVal{IsSome: true, Val: &IntVal{Val: 42}},
	}}
	if !fn([]Value{mixed}).(*BoolVal).Val {
		t.Error("expected true for mixed")
	}
}

func TestStdOptionCollect(t *testing.T) {
	exports := createStdOptionExports()
	fn := exports["collect"].(*BuiltinFnVal).Fn

	// All Some -> Some([1, 2, 3])
	allSome := &ListVal{Elements: []Value{
		&OptionVal{IsSome: true, Val: &IntVal{Val: 1}},
		&OptionVal{IsSome: true, Val: &IntVal{Val: 2}},
		&OptionVal{IsSome: true, Val: &IntVal{Val: 3}},
	}}
	result := fn([]Value{allSome}).(*OptionVal)
	if !result.IsSome {
		t.Fatal("expected Some")
	}
	collected := result.Val.(*ListVal)
	if len(collected.Elements) != 3 {
		t.Fatalf("expected 3, got %d", len(collected.Elements))
	}

	// With None -> None
	withNone := &ListVal{Elements: []Value{
		&OptionVal{IsSome: true, Val: &IntVal{Val: 1}},
		&OptionVal{IsSome: false},
		&OptionVal{IsSome: true, Val: &IntVal{Val: 3}},
	}}
	result = fn([]Value{withNone}).(*OptionVal)
	if result.IsSome {
		t.Fatal("expected None")
	}
}

func TestStdOptionFirstSome(t *testing.T) {
	exports := createStdOptionExports()
	fn := exports["first_some"].(*BuiltinFnVal).Fn

	list := &ListVal{Elements: []Value{
		&OptionVal{IsSome: false},
		&OptionVal{IsSome: true, Val: &IntVal{Val: 42}},
		&OptionVal{IsSome: true, Val: &IntVal{Val: 99}},
	}}
	result := fn([]Value{list}).(*OptionVal)
	if !result.IsSome || result.Val.(*IntVal).Val != 42 {
		t.Error("expected Some(42)")
	}

	// All None
	allNone := &ListVal{Elements: []Value{
		&OptionVal{IsSome: false},
		&OptionVal{IsSome: false},
	}}
	result = fn([]Value{allNone}).(*OptionVal)
	if result.IsSome {
		t.Error("expected None")
	}
}

func TestStdOptionFromResult(t *testing.T) {
	exports := createStdOptionExports()
	fn := exports["from_result"].(*BuiltinFnVal).Fn

	// Ok -> Some
	ok := &ResultVal{IsOk: true, Val: &IntVal{Val: 42}}
	result := fn([]Value{ok}).(*OptionVal)
	if !result.IsSome || result.Val.(*IntVal).Val != 42 {
		t.Error("Ok should convert to Some(42)")
	}

	// Err -> None
	errV := &ResultVal{IsOk: false, Val: &StringVal{Val: "err"}}
	result = fn([]Value{errV}).(*OptionVal)
	if result.IsSome {
		t.Error("Err should convert to None")
	}
}

// ============================================================
// std.iter tests
// ============================================================

func TestStdIterCycle(t *testing.T) {
	exports := createStdIterExports()
	fn := exports["cycle"].(*BuiltinFnVal).Fn

	list := &ListVal{Elements: []Value{&IntVal{Val: 1}, &IntVal{Val: 2}}}
	result := fn([]Value{list, &IntVal{Val: 3}})
	cycled := result.(*ListVal)
	if len(cycled.Elements) != 6 {
		t.Fatalf("expected 6 elements, got %d", len(cycled.Elements))
	}
	expected := []int64{1, 2, 1, 2, 1, 2}
	for i, e := range expected {
		if cycled.Elements[i].(*IntVal).Val != e {
			t.Errorf("[%d] = %d, want %d", i, cycled.Elements[i].(*IntVal).Val, e)
		}
	}

	// Cycle 0 times
	result = fn([]Value{list, &IntVal{Val: 0}})
	if len(result.(*ListVal).Elements) != 0 {
		t.Error("cycle 0 should be empty")
	}
}

func TestStdIterRepeat(t *testing.T) {
	exports := createStdIterExports()
	fn := exports["repeat"].(*BuiltinFnVal).Fn

	result := fn([]Value{&StringVal{Val: "x"}, &IntVal{Val: 4}})
	list := result.(*ListVal)
	if len(list.Elements) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(list.Elements))
	}
	for _, elem := range list.Elements {
		if elem.(*StringVal).Val != "x" {
			t.Errorf("expected 'x', got %q", elem.(*StringVal).Val)
		}
	}

	// Repeat 0
	result = fn([]Value{&IntVal{Val: 42}, &IntVal{Val: 0}})
	if len(result.(*ListVal).Elements) != 0 {
		t.Error("repeat 0 should be empty")
	}
}

func TestStdIterChain(t *testing.T) {
	exports := createStdIterExports()
	fn := exports["chain"].(*BuiltinFnVal).Fn

	lists := &ListVal{Elements: []Value{
		&ListVal{Elements: []Value{&IntVal{Val: 1}, &IntVal{Val: 2}}},
		&ListVal{Elements: []Value{&IntVal{Val: 3}}},
		&ListVal{Elements: []Value{&IntVal{Val: 4}, &IntVal{Val: 5}}},
	}}
	result := fn([]Value{lists})
	chained := result.(*ListVal)
	if len(chained.Elements) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(chained.Elements))
	}
	for i, expected := range []int64{1, 2, 3, 4, 5} {
		if chained.Elements[i].(*IntVal).Val != expected {
			t.Errorf("[%d] = %d, want %d", i, chained.Elements[i].(*IntVal).Val, expected)
		}
	}

	// Empty list of lists
	result = fn([]Value{&ListVal{Elements: []Value{}}})
	if len(result.(*ListVal).Elements) != 0 {
		t.Error("chain empty should be empty")
	}
}

func TestStdIterInterleave(t *testing.T) {
	exports := createStdIterExports()
	fn := exports["interleave"].(*BuiltinFnVal).Fn

	l1 := &ListVal{Elements: []Value{&IntVal{Val: 1}, &IntVal{Val: 3}, &IntVal{Val: 5}}}
	l2 := &ListVal{Elements: []Value{&IntVal{Val: 2}, &IntVal{Val: 4}, &IntVal{Val: 6}}}
	result := fn([]Value{l1, l2})
	interleaved := result.(*ListVal)
	expected := []int64{1, 2, 3, 4, 5, 6}
	if len(interleaved.Elements) != 6 {
		t.Fatalf("expected 6 elements, got %d", len(interleaved.Elements))
	}
	for i, e := range expected {
		if interleaved.Elements[i].(*IntVal).Val != e {
			t.Errorf("[%d] = %d, want %d", i, interleaved.Elements[i].(*IntVal).Val, e)
		}
	}

	// Unequal lengths
	l1 = &ListVal{Elements: []Value{&IntVal{Val: 1}}}
	l2 = &ListVal{Elements: []Value{&IntVal{Val: 2}, &IntVal{Val: 3}, &IntVal{Val: 4}}}
	result = fn([]Value{l1, l2})
	interleaved = result.(*ListVal)
	if len(interleaved.Elements) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(interleaved.Elements))
	}
}

func TestStdIterPairwise(t *testing.T) {
	exports := createStdIterExports()
	fn := exports["pairwise"].(*BuiltinFnVal).Fn

	list := &ListVal{Elements: []Value{
		&IntVal{Val: 1}, &IntVal{Val: 2}, &IntVal{Val: 3}, &IntVal{Val: 4},
	}}
	result := fn([]Value{list})
	pairs := result.(*ListVal)
	if len(pairs.Elements) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(pairs.Elements))
	}
	// Check first pair: [1, 2]
	pair := pairs.Elements[0].(*ListVal)
	if pair.Elements[0].(*IntVal).Val != 1 || pair.Elements[1].(*IntVal).Val != 2 {
		t.Error("first pair should be [1, 2]")
	}
	// Check last pair: [3, 4]
	pair = pairs.Elements[2].(*ListVal)
	if pair.Elements[0].(*IntVal).Val != 3 || pair.Elements[1].(*IntVal).Val != 4 {
		t.Error("last pair should be [3, 4]")
	}

	// Single element
	result = fn([]Value{&ListVal{Elements: []Value{&IntVal{Val: 1}}}})
	if len(result.(*ListVal).Elements) != 0 {
		t.Error("pairwise of single element should be empty")
	}

	// Empty
	result = fn([]Value{&ListVal{Elements: []Value{}}})
	if len(result.(*ListVal).Elements) != 0 {
		t.Error("pairwise of empty should be empty")
	}
}

// ============================================================
// Integration tests - modules via import
// ============================================================

func TestStdRegexViaImport(t *testing.T) {
	exports := createStdRegexExports()
	matchFn := exports["match"].(*BuiltinFnVal).Fn

	// Email-like pattern
	result := matchFn([]Value{
		&StringVal{Val: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`},
		&StringVal{Val: "user@example.com"},
	})
	if !result.(*BoolVal).Val {
		t.Error("should match valid email-like pattern")
	}
}

func TestStdCollectionsChainedOps(t *testing.T) {
	exports := createStdCollectionsExports()
	rangeFn := exports["range"].(*BuiltinFnVal).Fn
	takeFn := exports["take"].(*BuiltinFnVal).Fn
	dropFn := exports["drop"].(*BuiltinFnVal).Fn

	// range(10) |> drop(3) |> take(4)
	rangeResult := rangeFn([]Value{&IntVal{Val: 10}})
	dropResult := dropFn([]Value{&IntVal{Val: 3}, rangeResult})
	takeResult := takeFn([]Value{&IntVal{Val: 4}, dropResult})
	list := takeResult.(*ListVal)
	expected := []int64{3, 4, 5, 6}
	if len(list.Elements) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(list.Elements))
	}
	for i, e := range expected {
		if list.Elements[i].(*IntVal).Val != e {
			t.Errorf("[%d] = %d, want %d", i, list.Elements[i].(*IntVal).Val, e)
		}
	}
}

func TestStdResultOptionRoundTrip(t *testing.T) {
	resultExports := createStdResultExports()
	optionExports := createStdOptionExports()

	fromOptionFn := resultExports["from_option"].(*BuiltinFnVal).Fn
	fromResultFn := optionExports["from_result"].(*BuiltinFnVal).Fn

	// Some(42) -> Ok(42) -> Some(42)
	original := &OptionVal{IsSome: true, Val: &IntVal{Val: 42}}
	asResult := fromOptionFn([]Value{original, &StringVal{Val: "err"}}).(*ResultVal)
	if !asResult.IsOk || asResult.Val.(*IntVal).Val != 42 {
		t.Fatal("conversion to Result failed")
	}
	backToOption := fromResultFn([]Value{asResult}).(*OptionVal)
	if !backToOption.IsSome || backToOption.Val.(*IntVal).Val != 42 {
		t.Fatal("round-trip conversion failed")
	}

	// None -> Err("err") -> None
	noneVal := &OptionVal{IsSome: false}
	asResult = fromOptionFn([]Value{noneVal, &StringVal{Val: "err"}}).(*ResultVal)
	if asResult.IsOk {
		t.Fatal("None should convert to Err")
	}
	backToOption = fromResultFn([]Value{asResult}).(*OptionVal)
	if backToOption.IsSome {
		t.Fatal("Err should convert to None")
	}
}

func TestStdIterChainWithCollections(t *testing.T) {
	iterExports := createStdIterExports()
	collExports := createStdCollectionsExports()

	// Chain + chunk
	chainFn := iterExports["chain"].(*BuiltinFnVal).Fn
	chunkFn := collExports["chunk"].(*BuiltinFnVal).Fn

	lists := &ListVal{Elements: []Value{
		&ListVal{Elements: []Value{&IntVal{Val: 1}, &IntVal{Val: 2}, &IntVal{Val: 3}}},
		&ListVal{Elements: []Value{&IntVal{Val: 4}, &IntVal{Val: 5}, &IntVal{Val: 6}}},
	}}
	chained := chainFn([]Value{lists})
	chunks := chunkFn([]Value{&IntVal{Val: 2}, chained})
	chunked := chunks.(*ListVal)
	if len(chunked.Elements) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunked.Elements))
	}
}

func TestStdFormatIndentDedentRoundTrip(t *testing.T) {
	exports := createStdFormatExports()
	indentFn := exports["indent"].(*BuiltinFnVal).Fn
	dedentFn := exports["dedent"].(*BuiltinFnVal).Fn

	original := &StringVal{Val: "line1\nline2\nline3"}
	indented := indentFn([]Value{original, &IntVal{Val: 4}})
	dedented := dedentFn([]Value{indented})

	if dedented.(*StringVal).Val != "line1\nline2\nline3" {
		t.Errorf("round-trip failed: got %q", dedented.(*StringVal).Val)
	}
}

// ============================================================
// Edge case / error tests
// ============================================================

func TestStdCollectionsRangeZeroStep(t *testing.T) {
	exports := createStdCollectionsExports()
	fn := exports["range"].(*BuiltinFnVal).Fn

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero step")
		}
	}()
	fn([]Value{&IntVal{Val: 0}, &IntVal{Val: 10}, &IntVal{Val: 0}})
}

func TestStdIterCycleNegativePanics(t *testing.T) {
	exports := createStdIterExports()
	fn := exports["cycle"].(*BuiltinFnVal).Fn

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for negative n")
		}
	}()
	fn([]Value{&ListVal{Elements: []Value{}}, &IntVal{Val: -1}})
}

func TestStdIterRepeatNegativePanics(t *testing.T) {
	exports := createStdIterExports()
	fn := exports["repeat"].(*BuiltinFnVal).Fn

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for negative n")
		}
	}()
	fn([]Value{&IntVal{Val: 1}, &IntVal{Val: -1}})
}

func TestStdRandomIntMinGreaterThanMaxPanics(t *testing.T) {
	exports := createStdRandomExports()
	fn := exports["int"].(*BuiltinFnVal).Fn

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for min > max")
		}
	}()
	fn([]Value{&IntVal{Val: 10}, &IntVal{Val: 5}})
}

func TestStdCollectionsChunkZeroPanics(t *testing.T) {
	exports := createStdCollectionsExports()
	fn := exports["chunk"].(*BuiltinFnVal).Fn

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero chunk size")
		}
	}()
	fn([]Value{&IntVal{Val: 0}, &ListVal{Elements: []Value{&IntVal{Val: 1}}}})
}

func TestStdRegexFindNoMatch(t *testing.T) {
	exports := createStdRegexExports()
	fn := exports["find"].(*BuiltinFnVal).Fn

	result := fn([]Value{&StringVal{Val: `xyz`}, &StringVal{Val: "abc"}})
	opt := result.(*OptionVal)
	if opt.IsSome {
		t.Error("expected None for no match")
	}
}

func TestStdResultCollectEmpty(t *testing.T) {
	exports := createStdResultExports()
	fn := exports["collect"].(*BuiltinFnVal).Fn

	empty := &ListVal{Elements: []Value{}}
	result := fn([]Value{empty}).(*ResultVal)
	if !result.IsOk {
		t.Error("empty collect should be Ok")
	}
	if len(result.Val.(*ListVal).Elements) != 0 {
		t.Error("empty collect should have empty list")
	}
}

func TestStdOptionCollectEmpty(t *testing.T) {
	exports := createStdOptionExports()
	fn := exports["collect"].(*BuiltinFnVal).Fn

	empty := &ListVal{Elements: []Value{}}
	result := fn([]Value{empty}).(*OptionVal)
	if !result.IsSome {
		t.Error("empty collect should be Some")
	}
	if len(result.Val.(*ListVal).Elements) != 0 {
		t.Error("empty collect should have empty list")
	}
}

func TestStdFormatTruncateShortSuffix(t *testing.T) {
	exports := createStdFormatExports()
	fn := exports["truncate"].(*BuiltinFnVal).Fn

	// Max length less than suffix - should just cut
	result := fn([]Value{&StringVal{Val: "hello world"}, &IntVal{Val: 2}})
	if result.(*StringVal).Val != "he" {
		t.Errorf("got %q, want %q", result.(*StringVal).Val, "he")
	}
}

// Module registration test
func TestStdModuleRegistration(t *testing.T) {
	modules := []string{
		"std.regex", "std.collections", "std.random",
		"std.format", "std.result", "std.option", "std.iter",
	}
	// Create a dummy interpreter to test createStdModule
	interp := &Interpreter{env: NewEnvironment()}
	interp.registerBuiltins()

	for _, mod := range modules {
		modVal := interp.createStdModule(mod)
		if modVal == nil {
			t.Errorf("createStdModule(%q) returned nil", mod)
			continue
		}
		if len(modVal.Exports) == 0 {
			t.Errorf("module %q has no exports", mod)
		}
	}
}

// Test that all existing modules still work
func TestStdExistingModulesStillWork(t *testing.T) {
	modules := []string{
		"std.math", "std.string", "std.io", "std.testing", "std.json",
	}
	interp := &Interpreter{env: NewEnvironment()}
	interp.registerBuiltins()

	for _, mod := range modules {
		modVal := interp.createStdModule(mod)
		if modVal == nil {
			t.Errorf("existing module %q returned nil", mod)
		}
	}
}

func TestStdUnknownModuleReturnsNil(t *testing.T) {
	interp := &Interpreter{env: NewEnvironment()}
	interp.registerBuiltins()

	if modVal := interp.createStdModule("std.nonexistent"); modVal != nil {
		t.Error("unknown module should return nil")
	}
}
