package pepperlint

import (
	"go/ast"
)

// StructTypeRules is a list of type StructTypeRule
type StructTypeRules []StructTypeRule

// ValidateStructType will iterate through all provided StructTypeRules and call
// the ValidateStructType. Errors will be added in the BatchError type
func (rules StructTypeRules) ValidateStructType(s *ast.StructType) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateStructType(s); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// StructTypeRule allows for rules to be handled on ast.StructType definitions
type StructTypeRule interface {
	ValidateStructType(*ast.StructType) error
}
