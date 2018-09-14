package pepperlint

import (
	"go/ast"
)

type FieldListRules []FieldListRule

func (rules FieldListRules) ValidateFieldList(fields *ast.FieldList) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateFieldList(fields); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type FieldListRule interface {
	ValidateFieldList(*ast.FieldList) error
}
