package domain

import "fmt"

type ErrorKind string

const (
	ErrKindNotFound     ErrorKind = "not_found"
	ErrKindConflict     ErrorKind = "conflict"
	ErrKindValidation   ErrorKind = "validation"
	ErrKindUnauthorized ErrorKind = "unauthorized"
	ErrKindForbidden    ErrorKind = "forbidden"
	ErrKindInternal     ErrorKind = "internal"
)

type Error struct {
	Kind    ErrorKind
	Message string
	Cause   error
}

func (e Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func (e Error) Unwrap() error { return e.Cause }

func NotFound(msg string) error   { return Error{Kind: ErrKindNotFound, Message: msg} }
func Conflict(msg string) error   { return Error{Kind: ErrKindConflict, Message: msg} }
func Validation(msg string) error { return Error{Kind: ErrKindValidation, Message: msg} }
func Internal(msg string, cause error) error {
	return Error{Kind: ErrKindInternal, Message: msg, Cause: cause}
}
