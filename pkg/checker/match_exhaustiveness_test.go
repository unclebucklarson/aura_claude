package checker

import "testing"

// =============================================================================
// Phase 3.2 Chunk 4: Exhaustiveness Checking Tests
// =============================================================================

// ---------------------------------------------------------------------------
// MatchExpr — enum exhaustiveness
// ---------------------------------------------------------------------------

func TestCheckMatchExprExhaustiveEnum(t *testing.T) {
	src := `module test
enum Color:
    Red
    Green
    Blue

fn describe(c: Color) -> String:
    return match c:
        Color.Red   -> "red"
        Color.Green -> "green"
        Color.Blue  -> "blue"
`
	expectNoErrors(t, checkSource(t, src))
}

func TestCheckMatchExprNonExhaustiveEnum(t *testing.T) {
	src := `module test
enum Color:
    Red
    Green
    Blue

fn describe(c: Color) -> String:
    return match c:
        Color.Red   -> "red"
        Color.Green -> "green"
`
	errs := checkSource(t, src)
	expectErrorCode(t, errs, ErrNonExhaustive)
}

func TestCheckMatchExprWildcardCoversAll(t *testing.T) {
	src := `module test
enum Color:
    Red
    Green
    Blue

fn describe(c: Color) -> String:
    return match c:
        Color.Red -> "red"
        _ -> "other"
`
	expectNoErrors(t, checkSource(t, src))
}

func TestCheckMatchExprOrPatternExhaustive(t *testing.T) {
	src := `module test
enum Color:
    Red
    Green
    Blue

fn describe(c: Color) -> String:
    return match c:
        Color.Red | Color.Green | Color.Blue -> "a color"
`
	expectNoErrors(t, checkSource(t, src))
}

func TestCheckMatchExprOrPatternPartial(t *testing.T) {
	src := `module test
enum Color:
    Red
    Green
    Blue

fn describe(c: Color) -> String:
    return match c:
        Color.Red | Color.Green -> "warm"
`
	errs := checkSource(t, src)
	expectErrorCode(t, errs, ErrNonExhaustive)
}

// ---------------------------------------------------------------------------
// MatchStmt — or-pattern and as-pattern exhaustiveness
// ---------------------------------------------------------------------------

func TestCheckMatchStmtOrPatternExhaustive(t *testing.T) {
	src := `module test
enum Dir:
    North
    South
    East
    West

fn handle(d: Dir) -> String:
    let mut result = ""
    match d:
        case Dir.North | Dir.South:
            result = "vertical"
        case Dir.East | Dir.West:
            result = "horizontal"
    return result
`
	expectNoErrors(t, checkSource(t, src))
}

func TestCheckMatchStmtAsPatternIsWildcard(t *testing.T) {
	src := `module test
enum Color:
    Red
    Green
    Blue

fn handle(c: Color):
    match c:
        case _ as x:
            let n = 1
`
	expectNoErrors(t, checkSource(t, src))
}

// ---------------------------------------------------------------------------
// Bool exhaustiveness
// ---------------------------------------------------------------------------

func TestCheckMatchExprBoolBothCovered(t *testing.T) {
	src := `module test
fn label(b: Bool) -> String:
    return match b:
        true  -> "yes"
        false -> "no"
`
	expectNoErrors(t, checkSource(t, src))
}

func TestCheckMatchExprBoolMissingFalse(t *testing.T) {
	src := `module test
fn label(b: Bool) -> String:
    return match b:
        true -> "yes"
`
	errs := checkSource(t, src)
	expectErrorCode(t, errs, ErrNonExhaustive)
}

func TestCheckMatchStmtBoolBothCovered(t *testing.T) {
	src := `module test
fn handle(b: Bool):
    match b:
        case true:
            let _ = 1
        case false:
            let _ = 2
`
	expectNoErrors(t, checkSource(t, src))
}

func TestCheckMatchStmtBoolMissingTrue(t *testing.T) {
	src := `module test
fn handle(b: Bool):
    match b:
        case false:
            let _ = 2
`
	errs := checkSource(t, src)
	expectErrorCode(t, errs, ErrNonExhaustive)
}

// ---------------------------------------------------------------------------
// Guard type checking
// ---------------------------------------------------------------------------

func TestCheckMatchExprGuardType(t *testing.T) {
	src := `module test
fn check(n: Int) -> String:
    return match n:
        x if 42 -> "bad"
        _ -> "ok"
`
	errs := checkSource(t, src)
	expectErrorCode(t, errs, ErrTypeMismatch)
}
