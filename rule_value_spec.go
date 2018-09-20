package pepperlint

import (
	"go/ast"
)

// ValueSpecRules is a list of type ValueSpecRule.
type ValueSpecRules []ValueSpecRule

// ValidateValueSpec will iterate through the list of array types and call
// ValidateValueSpec. If an error is returned, then that error will be added
// to the batch of errors.
func (rules ValueSpecRules) ValidateValueSpec(spec *ast.ValueSpec) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateValueSpec(spec); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// ValueSpecRule represents an interface that will allow for validation
// to occur on an ast.ValueSpec.
type ValueSpecRule interface {
	ValidateValueSpec(*ast.ValueSpec) error
}
