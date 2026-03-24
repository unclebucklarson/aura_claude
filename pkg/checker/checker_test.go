package checker

import (
        "encoding/json"
        "strings"
        "testing"

        "github.com/unclebucklarson/aura/pkg/lexer"
        "github.com/unclebucklarson/aura/pkg/parser"
)

// helper to parse and check Aura source code
func checkSource(t *testing.T, src string) []*CheckError {
        t.Helper()
        l := lexer.New(src, "test.aura")
        tokens, lexErrors := l.Tokenize()
        if len(lexErrors) > 0 {
                t.Fatalf("lex errors: %v", lexErrors)
        }
        p := parser.New(tokens, "test.aura")
        module, parseErrors := p.Parse()
        if len(parseErrors) > 0 {
                t.Fatalf("parse errors: %v", parseErrors)
        }
        c := New(module)
        return c.Check()
}

func expectNoErrors(t *testing.T, errs []*CheckError) {
        t.Helper()
        if len(errs) > 0 {
                t.Errorf("expected no errors, got %d:\n%s", len(errs), FormatErrors(errs))
        }
}

func expectErrorCount(t *testing.T, errs []*CheckError, count int) {
        t.Helper()
        if len(errs) != count {
                t.Errorf("expected %d errors, got %d:\n%s", count, len(errs), FormatErrors(errs))
        }
}

func expectErrorCode(t *testing.T, errs []*CheckError, code ErrorCode) {
        t.Helper()
        for _, e := range errs {
                if e.Code == code {
                        return
                }
        }
        t.Errorf("expected error code %s, not found in errors:\n%s", code, FormatErrors(errs))
}

// --- Basic Type Definitions ---

func TestCheckEmptyModule(t *testing.T) {
        errs := checkSource(t, `module test
`)
        expectNoErrors(t, errs)
}

func TestCheckTypeDef(t *testing.T) {
        errs := checkSource(t, `module test

type TaskId = String
type Priority = Int
`)
        expectNoErrors(t, errs)
}

func TestCheckDuplicateType(t *testing.T) {
        errs := checkSource(t, `module test

type TaskId = String
type TaskId = Int
`)
        expectErrorCode(t, errs, ErrRedefinedType)
}

func TestCheckStructDef(t *testing.T) {
        errs := checkSource(t, `module test

struct Task:
    id: String
    title: String
    priority: Int = 3
`)
        expectNoErrors(t, errs)
}

func TestCheckEnumDef(t *testing.T) {
        errs := checkSource(t, `module test

enum TaskError:
    NotFound(String)
    InvalidTitle(String)
`)
        expectNoErrors(t, errs)
}

// --- Function Definitions ---

func TestCheckPureFunction(t *testing.T) {
        errs := checkSource(t, `module test

fn add(a: Int, b: Int) -> Int:
    return a + b
`)
        expectNoErrors(t, errs)
}

func TestCheckFunctionWithEffects(t *testing.T) {
        errs := checkSource(t, `module test

fn get_time() -> String with time:
    return time.now()
`)
        expectNoErrors(t, errs)
}

// --- Variable Definitions ---

func TestCheckLetStatement(t *testing.T) {
        errs := checkSource(t, `module test

fn main() -> Int:
    let x: Int = 42
    return x
`)
        expectNoErrors(t, errs)
}

func TestCheckLetTypeMismatch(t *testing.T) {
        errs := checkSource(t, `module test

fn main() -> Int:
    let x: Int = "hello"
    return x
`)
        expectErrorCode(t, errs, ErrTypeMismatch)
}

func TestCheckUndefinedVariable(t *testing.T) {
        errs := checkSource(t, `module test

fn main() -> Int:
    return x
`)
        expectErrorCode(t, errs, ErrUndefinedName)
}

func TestCheckImmutableAssignment(t *testing.T) {
        errs := checkSource(t, `module test

fn main() -> Int:
    let x: Int = 1
    x = 2
    return x
`)
        expectErrorCode(t, errs, ErrImmutableAssign)
}

func TestCheckMutableAssignment(t *testing.T) {
        errs := checkSource(t, `module test

fn main() -> Int:
    let mut x: Int = 1
    x = 2
    return x
`)
        expectNoErrors(t, errs)
}

func TestCheckDuplicateVariable(t *testing.T) {
        errs := checkSource(t, `module test

fn main() -> Int:
    let x: Int = 1
    let x: Int = 2
    return x
`)
        expectErrorCode(t, errs, ErrRedefinedName)
}

// --- Control Flow ---

func TestCheckIfStatement(t *testing.T) {
        errs := checkSource(t, `module test

fn check(x: Int) -> Int:
    if x > 0:
        return x
    else:
        return 0
`)
        expectNoErrors(t, errs)
}

func TestCheckIfConditionMustBeBool(t *testing.T) {
        errs := checkSource(t, `module test

fn check(x: Int) -> Int:
    if 42:
        return x
    return 0
`)
        expectErrorCode(t, errs, ErrTypeMismatch)
}

func TestCheckWhileConditionMustBeBool(t *testing.T) {
        errs := checkSource(t, `module test

fn loop() -> Int:
    while "yes":
        break
    return 0
`)
        expectErrorCode(t, errs, ErrTypeMismatch)
}

func TestCheckBreakOutsideLoop(t *testing.T) {
        errs := checkSource(t, `module test

fn bad() -> Int:
    break
    return 0
`)
        expectErrorCode(t, errs, ErrBreakOutside)
}

func TestCheckContinueOutsideLoop(t *testing.T) {
        errs := checkSource(t, `module test

fn bad() -> Int:
    continue
    return 0
`)
        expectErrorCode(t, errs, ErrContinueOutside)
}

func TestCheckBreakInsideLoop(t *testing.T) {
        errs := checkSource(t, `module test

fn good() -> Int:
    let mut i: Int = 0
    while true:
        if i > 10:
            break
        i = i + 1
    return i
`)
        expectNoErrors(t, errs)
}

func TestCheckForLoop(t *testing.T) {
        errs := checkSource(t, `module test

fn sum(items: [Int]) -> Int:
    let mut total: Int = 0
    for item in items:
        total = total + item
    return total
`)
        expectNoErrors(t, errs)
}

// --- Match/Pattern ---

func TestCheckMatchStatement(t *testing.T) {
        errs := checkSource(t, `module test

enum Color:
    Red
    Green
    Blue

fn name(c: Color) -> String:
    match c:
        case Color.Red:
            return "red"
        case Color.Green:
            return "green"
        case Color.Blue:
            return "blue"
`)
        expectNoErrors(t, errs)
}

func TestCheckMatchNonExhaustive(t *testing.T) {
        errs := checkSource(t, `module test

enum Color:
    Red
    Green
    Blue

fn name(c: Color) -> String:
    match c:
        case Color.Red:
            return "red"
`)
        expectErrorCode(t, errs, ErrNonExhaustive)
}

func TestCheckMatchWithWildcard(t *testing.T) {
        errs := checkSource(t, `module test

enum Color:
    Red
    Green
    Blue

fn name(c: Color) -> String:
    match c:
        case Color.Red:
            return "red"
        case _:
            return "other"
`)
        expectNoErrors(t, errs)
}

// --- Effect Tracking ---

func TestCheckEffectCapabilityUsage(t *testing.T) {
        errs := checkSource(t, `module test

fn bad() -> String:
    return db.query("tasks")
`)
        expectErrorCode(t, errs, ErrMissingEffect)
}

func TestCheckEffectCapabilityDeclared(t *testing.T) {
        errs := checkSource(t, `module test

fn good() -> String with db:
    return db.query("tasks")
`)
        expectNoErrors(t, errs)
}

func TestCheckEffectPropagation(t *testing.T) {
        errs := checkSource(t, `module test

fn helper() -> String with db:
    return db.query("tasks")

fn caller() -> String:
    return helper()
`)
        expectErrorCode(t, errs, ErrMissingEffect)
}

func TestCheckEffectPropagationDeclared(t *testing.T) {
        errs := checkSource(t, `module test

fn helper() -> String with db:
    return db.query("tasks")

fn caller() -> String with db:
    return helper()
`)
        expectNoErrors(t, errs)
}

// --- Spec Validation ---

func TestCheckSpecSatisfies(t *testing.T) {
        errs := checkSource(t, `module test

spec CreateTask:
    doc: "Creates a task"
    inputs:
        title: String
    effects: db, time

fn create_task(title: String) -> String with db, time satisfies CreateTask:
    return "ok"
`)
        expectNoErrors(t, errs)
}

func TestCheckSpecNotFound(t *testing.T) {
        errs := checkSource(t, `module test

fn create_task(title: String) -> String satisfies NonExistent:
    return "ok"
`)
        expectErrorCode(t, errs, ErrSpecNotFound)
}

func TestCheckSpecInputMismatch(t *testing.T) {
        errs := checkSource(t, `module test

spec CreateTask:
    doc: "Creates a task"
    inputs:
        title: String

fn create_task(name: String) -> String satisfies CreateTask:
    return "ok"
`)
        expectErrorCode(t, errs, ErrSpecInputMismatch)
}

func TestCheckSpecEffectMismatch(t *testing.T) {
        errs := checkSource(t, `module test

spec CreateTask:
    doc: "Creates a task"
    inputs:
        title: String
    effects: db, time

fn create_task(title: String) -> String with db satisfies CreateTask:
    return "ok"
`)
        expectErrorCode(t, errs, ErrSpecEffectMismatch)
}

func TestCheckDuplicateSpec(t *testing.T) {
        errs := checkSource(t, `module test

spec MySpec:
    doc: "first"

spec MySpec:
    doc: "second"
`)
        expectErrorCode(t, errs, ErrSpecDuplicate)
}

// --- Struct Construction ---

func TestCheckStructConstruction(t *testing.T) {
        errs := checkSource(t, `module test

struct Point:
    x: Float
    y: Float

fn make() -> Point:
    return Point(x: 1.0, y: 2.0)
`)
        expectNoErrors(t, errs)
}

func TestCheckStructMissingField(t *testing.T) {
        errs := checkSource(t, `module test

struct Point:
    x: Float
    y: Float

fn make() -> Point:
    return Point(x: 1.0)
`)
        expectErrorCode(t, errs, ErrFieldNotFound)
}

func TestCheckStructUnknownField(t *testing.T) {
        errs := checkSource(t, `module test

struct Point:
    x: Float
    y: Float

fn make() -> Point:
    return Point(x: 1.0, y: 2.0, z: 3.0)
`)
        expectErrorCode(t, errs, ErrFieldNotFound)
}

// --- Expressions ---

func TestCheckIfExpression(t *testing.T) {
        errs := checkSource(t, `module test

fn label(p: Int) -> String:
    let result = if p > 3 then "high" else "low"
    return result
`)
        expectNoErrors(t, errs)
}

func TestCheckListComprehension(t *testing.T) {
        errs := checkSource(t, `module test

fn filter(items: [Int]) -> [Int]:
    let result = [x for x in items if x > 0]
    return result
`)
        expectNoErrors(t, errs)
}

// --- Type Alias with Refinement ---

func TestCheckRefinementType(t *testing.T) {
        errs := checkSource(t, `module test

type Priority = Int where self >= 1 and self <= 5
`)
        expectNoErrors(t, errs)
}

func TestCheckUnionStringLitType(t *testing.T) {
        errs := checkSource(t, `module test

type TaskStatus = "pending" | "done"
`)
        expectNoErrors(t, errs)
}

// --- Constants ---

func TestCheckConstant(t *testing.T) {
        errs := checkSource(t, `module test

pub let max_val: Int = 200
`)
        expectNoErrors(t, errs)
}

func TestCheckConstantTypeMismatch(t *testing.T) {
        errs := checkSource(t, `module test

pub let max_val: Int = "hello"
`)
        expectErrorCode(t, errs, ErrTypeMismatch)
}

// --- Test Blocks ---

func TestCheckTestBlock(t *testing.T) {
        errs := checkSource(t, `module test

fn add(a: Int, b: Int) -> Int:
    return a + b

test "add works":
    let result = add(1, 2)
    assert result == 3
`)
        expectNoErrors(t, errs)
}

// --- Error Message Quality (AI-parseable) ---

func TestErrorsAreJSONSerializable(t *testing.T) {
        errs := checkSource(t, `module test

fn bad() -> Int:
    return "hello"
`)
        if len(errs) == 0 {
                t.Fatal("expected at least one error")
        }

        jsonStr := FormatErrorsJSON(errs)
        var parsed []map[string]interface{}
        if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
                t.Fatalf("errors not valid JSON: %v\nJSON: %s", err, jsonStr)
        }

        // Check required fields exist
        for _, e := range parsed {
                if _, ok := e["code"]; !ok {
                        t.Error("error missing 'code' field")
                }
                if _, ok := e["message"]; !ok {
                        t.Error("error missing 'message' field")
                }
                if _, ok := e["line"]; !ok {
                        t.Error("error missing 'line' field")
                }
        }
}

func TestErrorMessageContainsContext(t *testing.T) {
        errs := checkSource(t, `module test

fn bad() -> Int:
    let x: Int = "hello"
    return x
`)

        found := false
        for _, e := range errs {
                if e.Code == ErrTypeMismatch {
                        found = true
                        if e.Expected == "" || e.Got == "" {
                                t.Error("type mismatch error should have expected/got")
                        }
                }
        }
        if !found {
                t.Error("expected TYPE_MISMATCH error")
        }
}

func TestErrorHasSuggestedFix(t *testing.T) {
        errs := checkSource(t, `module test

fn bad() -> Int:
    let x: Int = 1
    x = 2
    return x
`)

        for _, e := range errs {
                if e.Code == ErrImmutableAssign {
                        if e.Fix == "" {
                                t.Error("immutable assign error should have suggested fix")
                        }
                        if !strings.Contains(e.Fix, "let mut") {
                                t.Errorf("fix should suggest 'let mut', got: %s", e.Fix)
                        }
                        return
                }
        }
        t.Error("expected IMMUTABLE_ASSIGN error with fix")
}

// --- Complex Integration Test ---

func TestCheckCompleteModule(t *testing.T) {
        src := `module auratask.models

type TaskId = String
type Priority = Int where self >= 1 and self <= 5
type TaskStatus = "pending" | "done"

struct Task:
    id: TaskId
    title: String
    priority: Priority = 3
    status: TaskStatus = "pending"

enum TaskError:
    NotFound(TaskId)
    InvalidTitle(String)

spec CreateTask:
    doc: "Creates a new task"
    inputs:
        title: String
        priority: Priority
    effects: db, time

fn add(a: Int, b: Int) -> Int:
    return a + b

pub fn create_task(title: String, priority: Priority) -> String with db, time satisfies CreateTask:
    let id = "t-001"
    return id

test "add works":
    let result = add(1, 2)
    assert result == 3
`
        errs := checkSource(t, src)
        expectNoErrors(t, errs)
}

func TestCheckReturnTypeMismatch(t *testing.T) {
        errs := checkSource(t, `module test

fn get_num() -> Int:
    return "not a number"
`)
        expectErrorCode(t, errs, ErrTypeMismatch)
}

func TestCheckOptionType(t *testing.T) {
        errs := checkSource(t, `module test

fn find() -> Int?:
    return none
`)
        expectNoErrors(t, errs)
}

func TestCheckResultType(t *testing.T) {
        errs := checkSource(t, `module test

enum MyError:
    Bad(String)

fn try_it() -> Result[Int, MyError]:
    return Ok(42)
`)
        expectNoErrors(t, errs)
}
