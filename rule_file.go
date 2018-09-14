package pepperlint

import (
	"go/ast"
)

// FileRules is a list of type FileRule.
type FileRules []FileRule

// ValidateFile will iterate through the list of array types and call
// ValidateFile. If an error is returned, then that error will be added
// to the batch of errors.
func (rules FileRules) ValidateFile(stmt *ast.File) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateFile(stmt); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// FileRule represents an interface that will allow for validation
// to occur on an ast.File.
type FileRule interface {
	ValidateFile(*ast.File) error
}
