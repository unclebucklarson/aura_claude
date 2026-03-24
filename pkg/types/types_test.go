package types

import "testing"

func TestPrimitiveEquality(t *testing.T) {
	if !Equal(BuiltinInt, BuiltinInt) {
		t.Error("Int should equal itself")
	}
	if Equal(BuiltinInt, BuiltinString) {
		t.Error("Int should not equal String")
	}
}

func TestListTypeEquality(t *testing.T) {
	a := NewListType(BuiltinInt)
	b := NewListType(BuiltinInt)
	c := NewListType(BuiltinString)

	if !Equal(a, b) {
		t.Error("[Int] should equal [Int]")
	}
	if Equal(a, c) {
		t.Error("[Int] should not equal [String]")
	}
}

func TestMapTypeEquality(t *testing.T) {
	a := NewMapType(BuiltinString, BuiltinInt)
	b := NewMapType(BuiltinString, BuiltinInt)
	c := NewMapType(BuiltinString, BuiltinBool)

	if !Equal(a, b) {
		t.Error("{String: Int} should equal {String: Int}")
	}
	if Equal(a, c) {
		t.Error("{String: Int} should not equal {String: Bool}")
	}
}

func TestOptionType(t *testing.T) {
	opt := NewOptionType(BuiltinInt)
	if opt.String() != "Int?" {
		t.Errorf("expected Int?, got %s", opt.String())
	}
}

func TestResultType(t *testing.T) {
	res := NewResultType(BuiltinInt, BuiltinString)
	if res.String() != "Result[Int, String]" {
		t.Errorf("expected Result[Int, String], got %s", res.String())
	}
}

func TestUnionType(t *testing.T) {
	u := NewUnionType([]*Type{
		NewStringLitType("pending"),
		NewStringLitType("done"),
	})
	expected := `"pending" | "done"`
	if u.String() != expected {
		t.Errorf("expected %s, got %s", expected, u.String())
	}
}

func TestFunctionType(t *testing.T) {
	fn := NewFunctionType(
		[]*Type{BuiltinInt, BuiltinString},
		BuiltinBool,
		nil,
	)
	expected := "fn(Int, String) -> Bool"
	if fn.String() != expected {
		t.Errorf("expected %s, got %s", expected, fn.String())
	}
}

func TestTupleType(t *testing.T) {
	tup := NewTupleType([]*Type{BuiltinInt, BuiltinString})
	if tup.String() != "(Int, String)" {
		t.Errorf("expected (Int, String), got %s", tup.String())
	}
}

func TestRefinementType(t *testing.T) {
	ref := NewRefinementType(BuiltinInt, "self >= 1 and self <= 5")
	expected := "Int where self >= 1 and self <= 5"
	if ref.String() != expected {
		t.Errorf("expected %s, got %s", expected, ref.String())
	}
}

func TestStringLitType(t *testing.T) {
	slt := NewStringLitType("pending")
	if slt.String() != `"pending"` {
		t.Errorf("expected %q, got %s", "pending", slt.String())
	}
}

// --- Subtyping / Assignability tests ---

func TestNeverAssignableToAnything(t *testing.T) {
	if !IsAssignableTo(BuiltinNever, BuiltinInt) {
		t.Error("Never should be assignable to Int")
	}
	if !IsAssignableTo(BuiltinNever, BuiltinString) {
		t.Error("Never should be assignable to String")
	}
}

func TestAnythingAssignableToAny(t *testing.T) {
	if !IsAssignableTo(BuiltinInt, BuiltinAny) {
		t.Error("Int should be assignable to Any")
	}
}

func TestNoneAssignableToOption(t *testing.T) {
	opt := NewOptionType(BuiltinInt)
	if !IsAssignableTo(BuiltinNone, opt) {
		t.Error("None should be assignable to Int?")
	}
}

func TestValueAssignableToOption(t *testing.T) {
	opt := NewOptionType(BuiltinInt)
	if !IsAssignableTo(BuiltinInt, opt) {
		t.Error("Int should be assignable to Int?")
	}
}

func TestRefinementSubtypeOfBase(t *testing.T) {
	ref := NewRefinementType(BuiltinInt, "self > 0")
	if !IsAssignableTo(ref, BuiltinInt) {
		t.Error("Int where self > 0 should be assignable to Int")
	}
}

func TestStringLitSubtypeOfString(t *testing.T) {
	slt := NewStringLitType("hello")
	if !IsAssignableTo(slt, BuiltinString) {
		t.Error("\"hello\" should be assignable to String")
	}
}

func TestStringLitAssignableToUnion(t *testing.T) {
	u := NewUnionType([]*Type{
		NewStringLitType("pending"),
		NewStringLitType("done"),
	})
	pending := NewStringLitType("pending")
	other := NewStringLitType("cancelled")

	if !IsAssignableTo(pending, u) {
		t.Error("\"pending\" should be assignable to union")
	}
	if IsAssignableTo(other, u) {
		t.Error("\"cancelled\" should not be assignable to union")
	}
}

func TestIntToFloatWidening(t *testing.T) {
	if !IsAssignableTo(BuiltinInt, BuiltinFloat) {
		t.Error("Int should be assignable to Float")
	}
	if IsAssignableTo(BuiltinFloat, BuiltinInt) {
		t.Error("Float should not be assignable to Int")
	}
}

func TestStructWidthSubtyping(t *testing.T) {
	point2D := NewStructType("Point2D", []*Field{
		{Name: "x", Type: BuiltinFloat},
		{Name: "y", Type: BuiltinFloat},
	}, nil)
	point3D := NewStructType("Point3D", []*Field{
		{Name: "x", Type: BuiltinFloat},
		{Name: "y", Type: BuiltinFloat},
		{Name: "z", Type: BuiltinFloat},
	}, nil)

	if !IsAssignableTo(point3D, point2D) {
		t.Error("Point3D should be assignable to Point2D (width subtyping)")
	}
	if IsAssignableTo(point2D, point3D) {
		t.Error("Point2D should not be assignable to Point3D")
	}
}

func TestAliasUnwrapping(t *testing.T) {
	alias := NewAliasType("TaskId", BuiltinString)
	if !IsAssignableTo(alias, BuiltinString) {
		t.Error("TaskId alias should be assignable to String")
	}
	if !Equal(alias, BuiltinString) {
		t.Error("TaskId alias should equal String")
	}
}

func TestUnionSubtyping(t *testing.T) {
	// Union of two string lits should be assignable to String
	u := NewUnionType([]*Type{
		NewStringLitType("a"),
		NewStringLitType("b"),
	})
	if !IsAssignableTo(u, BuiltinString) {
		t.Error("union of string lits should be assignable to String")
	}
}

// --- Registry tests ---

func TestRegistryBuiltins(t *testing.T) {
	r := NewRegistry()
	intT, ok := r.Lookup("Int")
	if !ok {
		t.Fatal("expected to find Int in registry")
	}
	if intT != BuiltinInt {
		t.Error("Int should be the builtin singleton")
	}
}

func TestRegistryRegisterAndLookup(t *testing.T) {
	r := NewRegistry()
	taskType := NewStructType("Task", nil, nil)
	if err := r.Register("Task", taskType); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, ok := r.Lookup("Task")
	if !ok {
		t.Fatal("expected to find Task")
	}
	if found != taskType {
		t.Error("should return same type object")
	}
}

func TestRegistryDuplicateRegister(t *testing.T) {
	r := NewRegistry()
	r.Register("Task", NewStructType("Task", nil, nil))
	if err := r.Register("Task", NewStructType("Task", nil, nil)); err == nil {
		t.Error("expected error for duplicate registration")
	}
}

func TestRegistryCannotRedefineBuiltin(t *testing.T) {
	r := NewRegistry()
	if err := r.Register("Int", NewStructType("Int", nil, nil)); err == nil {
		t.Error("expected error for redefining builtin")
	}
}

func TestTypeKindString(t *testing.T) {
	if KindPrimitive.String() != "primitive" {
		t.Errorf("expected 'primitive', got %q", KindPrimitive.String())
	}
}
