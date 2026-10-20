// Package clierr is the CLI's expected failure: printed as `folkways: <message>` with exit code 1.
package clierr

import "fmt"

// Error is an expected failure, the Go form of the TS CLI's FolkwaysError.
type Error struct{ Msg string }

func (e *Error) Error() string { return e.Msg }

// New returns an expected failure with a formatted message.
func New(format string, args ...any) error {
	return &Error{Msg: fmt.Sprintf(format, args...)}
}
