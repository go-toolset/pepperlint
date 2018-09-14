package pepperlint

import (
	"go/ast"
)

type MapTypeRules []MapTypeRule

func (rules MapTypeRules) ValidateMapType(m *ast.MapType) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateMapType(m); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type MapTypeRule interface {
	ValidateMapType(*ast.MapType) error
}
