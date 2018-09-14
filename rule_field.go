package pepperlint

import (
	"go/ast"
)

// FieldRules is a list of type FieldRule.
type FieldRules []FieldRule

// ValidateField will iterate through the list of array types and call
// ValidateField. If an error is returned, then that error will be added
// to the batch of errors.
func (rules FieldRules) ValidateField(field *ast.Field) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateField(field); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// FieldRule represents an interface that will allow for validation
// to occur on an ast.Field.
type FieldRule interface {
	ValidateField(*ast.Field) error
}
