package pepperlint

import (
	"go/ast"
)

// ArrayTypeRules is a list of type ArrayTypeRule.
type ArrayTypeRules []ArrayTypeRule

// ValidateArrayType will iterate through the list of array types and call
// ValidateArrayType. If an error is returned, then that error will be added
// to the batch of errors.
func (rules ArrayTypeRules) ValidateArrayType(array *ast.ArrayType) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateArrayType(array); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// ArrayTypeRule represents an interface that will allow for validation
// to occur on an ast.ArrayType.
type ArrayTypeRule interface {
	ValidateArrayType(*ast.ArrayType) error
}
