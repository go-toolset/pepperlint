package pepperlint

import (
	"go/ast"
)

type CallExprRules []CallExprRule

func (rules CallExprRules) ValidateCallExpr(expr *ast.CallExpr) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateCallExpr(expr); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type CallExprRule interface {
	ValidateCallExpr(*ast.CallExpr) error
}
