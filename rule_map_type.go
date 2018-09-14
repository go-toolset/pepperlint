package pepperlint

import (
	"go/ast"
)

// MapTypeRules is a list of type MapTypeRule.
type MapTypeRules []MapTypeRule

// ValidateMapType will iterate through the list of array types and call
// ValidateMapType. If an error is returned, then that error will be added
// to the batch of errors.
func (rules MapTypeRules) ValidateMapType(m *ast.MapType) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateMapType(m); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// MapTypeRule represents an interface that will allow for validation
// to occur on an ast.MapType.
type MapTypeRule interface {
	ValidateMapType(*ast.MapType) error
}
