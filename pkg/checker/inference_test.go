package checker

import "testing"

// --- Empty Collection Inference ---

func TestInferenceEmptyListAnnotated(t *testing.T) {
	errs := checkSource(t, `module test

fn main():
    let xs: [Int] = []
`)
	expectNoErrors(t, errs)
}

func TestInferenceEmptyMapAnnotated(t *testing.T) {
	errs := checkSource(t, `module test

fn main():
    let m: {String: Int} = {}
`)
	expectNoErrors(t, errs)
}

func TestInferenceEmptyListUsed(t *testing.T) {
	// Annotated empty list should be usable — variable has correct type
	errs := checkSource(t, `module test

fn use_list(xs: [String]):
    return None

fn main():
    let xs: [String] = []
    use_list(xs)
`)
	expectNoErrors(t, errs)
}

func TestInferenceEmptyListNoAnnotation(t *testing.T) {
	// No annotation — should still compile, just infers [Any]
	errs := checkSource(t, `module test

fn main():
    let xs = []
`)
	expectNoErrors(t, errs)
}

// --- Bidirectional Constructor Inference ---

func TestInferenceSomeMatchesAnnotation(t *testing.T) {
	errs := checkSource(t, `module test

fn main():
    let x: Option[Int] = Some(42)
`)
	expectNoErrors(t, errs)
}

func TestInferenceSomeMismatchAnnotation(t *testing.T) {
	errs := checkSource(t, `module test

fn main():
    let x: Option[String] = Some(42)
`)
	expectErrorCode(t, errs, ErrTypeMismatch)
}

func TestInferenceOkMatchesAnnotation(t *testing.T) {
	errs := checkSource(t, `module test

fn main():
    let x: Result[Int, String] = Ok(1)
`)
	expectNoErrors(t, errs)
}

func TestInferenceOkMismatchAnnotation(t *testing.T) {
	errs := checkSource(t, `module test

fn main():
    let x: Result[Int, String] = Ok("bad")
`)
	expectErrorCode(t, errs, ErrTypeMismatch)
}

func TestInferenceErrMatchesAnnotation(t *testing.T) {
	errs := checkSource(t, `module test

fn main():
    let x: Result[Int, String] = Err("oops")
`)
	expectNoErrors(t, errs)
}

func TestInferenceErrMismatchAnnotation(t *testing.T) {
	errs := checkSource(t, `module test

fn main():
    let x: Result[Int, String] = Err(42)
`)
	expectErrorCode(t, errs, ErrTypeMismatch)
}

// --- Generic Type Aliases ---

func TestInferenceGenericAliasOptionAccepted(t *testing.T) {
	errs := checkSource(t, `module test

type Maybe[T] = Option[T]

fn get_val() -> Maybe[Int]:
    return Some(1)
`)
	expectNoErrors(t, errs)
}

func TestInferenceGenericAliasListAccepted(t *testing.T) {
	// [T] is the Aura syntax for List[T]
	errs := checkSource(t, `module test

type Wrapper[T] = [T]

fn get_list() -> Wrapper[String]:
    return ["hello", "world"]
`)
	expectNoErrors(t, errs)
}

func TestInferenceNonGenericAliasWithGenericBody(t *testing.T) {
	errs := checkSource(t, `module test

type StringList = [String]

fn get_strings() -> StringList:
    return ["a", "b"]
`)
	expectNoErrors(t, errs)
}
