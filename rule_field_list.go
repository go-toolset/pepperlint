package pepperlint

import (
	"go/ast"
)

// FieldListRules is a list of type FieldListRule.
type FieldListRules []FieldListRule

// ValidateFieldList will iterate through the list of array types and call
// ValidateFieldList. If an error is returned, then that error will be added
// to the batch of errors.
func (rules FieldListRules) ValidateFieldList(fields *ast.FieldList) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateFieldList(fields); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// FieldListRule represents an interface that will allow for validation
// to occur on an ast.FieldList.
type FieldListRule interface {
	ValidateFieldList(*ast.FieldList) error
}
