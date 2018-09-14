package pepperlint

import (
	"go/ast"
)

// CallExprRules is a list of type CallExprRule.
type CallExprRules []CallExprRule

// ValidateCallExpr will iterate through the list of array types and call
// ValidateCallExpr. If an error is returned, then that error will be added
// to the batch of errors.
func (rules CallExprRules) ValidateCallExpr(expr *ast.CallExpr) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateCallExpr(expr); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// CallExprRule represents an interface that will allow for validation
// to occur on an ast.CallExpr.
type CallExprRule interface {
	ValidateCallExpr(*ast.CallExpr) error
}
