package pepperlint

import (
	"go/ast"
)

type TypeSpecRules []TypeSpecRule

func (rules TypeSpecRules) ValidateTypeSpec(spec *ast.TypeSpec) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateTypeSpec(spec); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type TypeSpecRule interface {
	ValidateTypeSpec(*ast.TypeSpec) error
}
