package pepperlint

import (
	"go/ast"
)

// TypeSpecRules is a list of type TypeSpecRule.
type TypeSpecRules []TypeSpecRule

// ValidateTypeSpec will iterate through the list of array types and call
// ValidateTypeSpec. If an error is returned, then that error will be added
// to the batch of errors.
func (rules TypeSpecRules) ValidateTypeSpec(spec *ast.TypeSpec) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateTypeSpec(spec); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// TypeSpecRule represents an interface that will allow for validation
// to occur on an ast.TypeSpec.
type TypeSpecRule interface {
	ValidateTypeSpec(*ast.TypeSpec) error
}
