package pepperlint

import (
	"go/ast"
)

type FuncTypeRules []FuncTypeRule

func (rules FuncTypeRules) ValidateFuncType(fn *ast.FuncType) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateFuncType(fn); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type FuncTypeRule interface {
	ValidateFuncType(*ast.FuncType) error
}
