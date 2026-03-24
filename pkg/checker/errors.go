// Package checker implements the type checker and semantic analysis for Aura.
package checker

import (
	"encoding/json"
	"fmt"

	"github.com/unclebucklarson/aura/pkg/token"
)

// ErrorCode categorizes checker errors for AI consumption.
type ErrorCode string

const (
	// Type errors
	ErrTypeMismatch     ErrorCode = "TYPE_MISMATCH"
	ErrUndefinedName    ErrorCode = "UNDEFINED_NAME"
	ErrRedefinedName    ErrorCode = "REDEFINED_NAME"
	ErrUndefinedType    ErrorCode = "UNDEFINED_TYPE"
	ErrRedefinedType    ErrorCode = "REDEFINED_TYPE"
	ErrNotCallable      ErrorCode = "NOT_CALLABLE"
	ErrArgCount         ErrorCode = "ARGUMENT_COUNT"
	ErrFieldNotFound    ErrorCode = "FIELD_NOT_FOUND"
	ErrNotIndexable     ErrorCode = "NOT_INDEXABLE"
	ErrNotIterable      ErrorCode = "NOT_ITERABLE"
	ErrNotAssignable    ErrorCode = "NOT_ASSIGNABLE"
	ErrImmutableAssign  ErrorCode = "IMMUTABLE_ASSIGN"
	ErrReturnOutside    ErrorCode = "RETURN_OUTSIDE_FUNCTION"
	ErrBreakOutside     ErrorCode = "BREAK_OUTSIDE_LOOP"
	ErrContinueOutside  ErrorCode = "CONTINUE_OUTSIDE_LOOP"
	ErrMissingReturn    ErrorCode = "MISSING_RETURN"
	ErrUnreachable      ErrorCode = "UNREACHABLE_CODE"

	// Effect errors
	ErrMissingEffect    ErrorCode = "MISSING_EFFECT"
	ErrEffectMismatch   ErrorCode = "EFFECT_MISMATCH"

	// Spec errors
	ErrSpecNotFound     ErrorCode = "SPEC_NOT_FOUND"
	ErrSpecInputMismatch ErrorCode = "SPEC_INPUT_MISMATCH"
	ErrSpecEffectMismatch ErrorCode = "SPEC_EFFECT_MISMATCH"
	ErrSpecDuplicate    ErrorCode = "SPEC_DUPLICATE"

	// Pattern errors
	ErrNonExhaustive    ErrorCode = "NON_EXHAUSTIVE_MATCH"
	ErrDuplicateCase    ErrorCode = "DUPLICATE_CASE"

	// Generic errors
	ErrTypeParamCount   ErrorCode = "TYPE_PARAM_COUNT"
	ErrInvalidOperation ErrorCode = "INVALID_OPERATION"
)

// Severity indicates the severity of a diagnostic.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// CheckError is a structured, AI-parseable error from the type checker.
type CheckError struct {
	Code     ErrorCode  `json:"code"`
	Severity Severity   `json:"severity"`
	Message  string     `json:"message"`
	Span     token.Span `json:"span"`
	File     string     `json:"file"`
	Line     int        `json:"line"`
	Column   int        `json:"column"`
	EndLine  int        `json:"end_line,omitempty"`
	EndCol   int        `json:"end_column,omitempty"`
	Context  string     `json:"context,omitempty"`  // surrounding code context
	Fix      string     `json:"fix,omitempty"`      // suggested fix
	Expected string     `json:"expected,omitempty"` // expected type/value
	Got      string     `json:"got,omitempty"`      // actual type/value
}

// Error implements the error interface.
func (e *CheckError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s: %s", e.File, e.Line, e.Column, e.Code, e.Message)
}

// JSON returns the error as a JSON string for AI consumption.
func (e *CheckError) JSON() string {
	data, _ := json.Marshal(e)
	return string(data)
}

// newError creates a new CheckError from a span.
func newError(code ErrorCode, span token.Span, msg string) *CheckError {
	return &CheckError{
		Code:     code,
		Severity: SeverityError,
		Message:  msg,
		Span:     span,
		File:     span.File,
		Line:     span.Start.Line,
		Column:   span.Start.Column,
		EndLine:  span.End.Line,
		EndCol:   span.End.Column,
	}
}

// withExpectedGot adds expected/got info to an error.
func (e *CheckError) withExpectedGot(expected, got string) *CheckError {
	e.Expected = expected
	e.Got = got
	return e
}

// withFix adds a suggested fix to an error.
func (e *CheckError) withFix(fix string) *CheckError {
	e.Fix = fix
	return e
}

// withContext adds surrounding context to an error.
func (e *CheckError) withContext(ctx string) *CheckError {
	e.Context = ctx
	return e
}

// asWarning changes the severity to warning.
func (e *CheckError) asWarning() *CheckError {
	e.Severity = SeverityWarning
	return e
}
