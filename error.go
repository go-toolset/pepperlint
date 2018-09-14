package pepperlint

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
)

// LineNumberError is an interface that wraps an
// error and contains where the error occurred
type LineNumberError interface {
	error
	LineNumber() int
}

// ErrorWrap ...
type ErrorWrap struct {
	pos    token.Position
	prefix string
	msg    string
}

// NewErrorWrap will return a new error and construct a prefix based on the node
// and file set.
// TODO: Wrap all returned errors in visitor.go with line number and file name
func NewErrorWrap(fset *token.FileSet, node ast.Node, msg string) *ErrorWrap {
	pos := fset.Position(node.Pos())
	prefix := pos.String()

	return &ErrorWrap{
		pos:    pos,
		prefix: prefix,
		msg:    msg,
	}
}

func (e *ErrorWrap) Error() string {
	return fmt.Sprintf("%s: %s", e.prefix, e.msg)
}

// LineNumber return the line number to which the error occurred
func (e *ErrorWrap) LineNumber() int {
	return e.pos.Line
}

// BatchError ...
type BatchError struct {
	errors []error
}

// NewBatchError ...
func NewBatchError(errs ...error) *BatchError {
	return &BatchError{
		errors: errs,
	}
}

// Add ...
func (e *BatchError) Add(errs ...error) {
	e.errors = append(e.errors, errs...)
}

// Errors returns the list of errors back
func (e *BatchError) Errors() []error {
	return e.errors
}

// Len ...
func (e *BatchError) Len() int {
	return len(e.errors)
}

// Return ...
func (e *BatchError) Return() error {
	if e.Len() == 0 {
		return nil
	}

	return e
}

func (e BatchError) Error() string {
	buf := bytes.Buffer{}
	for i := 0; i < len(e.errors); i++ {
		buf.WriteString(fmt.Sprintf("%s", e.errors[i].Error()))
		if i+1 < len(e.errors) {
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

// Errors ...
type Errors []error

// Add ...
func (e *Errors) Add(errs ...error) {
	*e = append(*e, errs...)
}

func (e Errors) Error() string {
	buf := bytes.Buffer{}
	for _, err := range e {
		buf.WriteString(err.Error())
		buf.WriteString("\n")
	}

	return buf.String()
}
