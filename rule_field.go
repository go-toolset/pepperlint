package pepperlint

import (
	"go/ast"
)

type FieldRules []FieldRule

func (rules FieldRules) ValidateField(field *ast.Field) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateField(field); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type FieldRule interface {
	ValidateField(*ast.Field) error
}
