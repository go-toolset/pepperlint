package pepperlint

import (
	"go/ast"
)

type StructTypeRules []StructTypeRule

func (rules StructTypeRules) ValidateStructType(s *ast.StructType) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateStructType(s); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type StructTypeRule interface {
	ValidateStructType(*ast.StructType) error
}
