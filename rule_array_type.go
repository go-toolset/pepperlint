package pepperlint

import (
	"go/ast"
)

type ArrayTypeRules []ArrayTypeRule

func (rules ArrayTypeRules) ValidateArrayType(array *ast.ArrayType) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateArrayType(array); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type ArrayTypeRule interface {
	ValidateArrayType(*ast.ArrayType) error
}
