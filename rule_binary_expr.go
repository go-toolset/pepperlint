package pepperlint

import (
	"go/ast"
)

// BinaryExprRules is a list of type BinaryExprRule.
type BinaryExprRules []BinaryExprRule

// ValidateBinaryExpr will iterate through the list of array types and call
// ValidateBinaryExpr. If an error is returned, then that error will be added
// to the batch of errors.
func (rules BinaryExprRules) ValidateBinaryExpr(fn *ast.BinaryExpr) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateBinaryExpr(fn); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// BinaryExprRule represents an interface that will allow for validation
// to occur on an ast.BinaryExpr.
type BinaryExprRule interface {
	ValidateBinaryExpr(*ast.BinaryExpr) error
}
