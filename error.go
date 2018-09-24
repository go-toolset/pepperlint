package pepperlint

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
)

// FileError is an error interface that will represent an error and where it
// occurred.
type FileError interface {
	error

	LineNumber
	Filename
}

// LineNumber is used to return a line number of where an error occurred.
type LineNumber interface {
	LineNumber() int
}

// Filename is the filename of where the error occurred.
type Filename interface {
	Filename() string
}

// ErrorWrap will extract the line number from the ast.Node provided.
// The prefix usually represents the file name that the lint error was
// found in.
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

// Filename will return the filename of where the error occurred
func (e *ErrorWrap) Filename() string {
	return e.pos.Filename
}

// BatchError groups a set of errors together usually to organize
// them by Validator but is not limited to.
type BatchError struct {
	errors []error
}

// NewBatchError returns a new BatchError
func NewBatchError(errs ...error) *BatchError {
	return &BatchError{
		errors: errs,
	}
}

// Add will add a new error to the BatchError
func (e *BatchError) Add(errs ...error) {
	e.errors = append(e.errors, errs...)
}

// Errors returns the list of errors back
func (e *BatchError) Errors() []error {
	return e.errors
}

// Len returns the length of the errors contained in the BatchError
func (e *BatchError) Len() int {
	return len(e.errors)
}

// Return will return BatchError if there is at least 1 error in the container.
// If not, nil will be returned
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

// Errors is a list of errors. This type is mostly used for pretty printing the
// error message.
type Errors []error

// Add will add the series of errors to the list.
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

// Count will return the number of errors in the list. If there are any batch error, this will
// make recursive calls until a non-batch error is found.
func (e Errors) Count() int {
	errAmount := 0
	for _, err := range e {
		errAmount += count(err)
	}
	return errAmount
}

func count(e error) int {
	berr, ok := e.(*BatchError)
	if !ok {
		return 1
	}

	errAmount := 0
	for _, err := range berr.Errors() {
		errAmount += count(err)
	}

	return errAmount
}
