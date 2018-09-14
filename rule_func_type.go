package pepperlint

import (
	"go/ast"
)

// FuncTypeRules is a list of type FuncTypeRule.
type FuncTypeRules []FuncTypeRule

// ValidateFuncType will iterate through the list of array types and call
// ValidateFuncType. If an error is returned, then that error will be added
// to the batch of errors.
func (rules FuncTypeRules) ValidateFuncType(fn *ast.FuncType) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateFuncType(fn); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// FuncTypeRule represents an interface that will allow for validation
// to occur on an ast.FuncType.
type FuncTypeRule interface {
	ValidateFuncType(*ast.FuncType) error
}
