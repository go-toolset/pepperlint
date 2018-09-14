package pepperlint

import (
	"go/ast"
)

// InterfaceTypeRules is a list of type InterfaceTypeRule.
type InterfaceTypeRules []InterfaceTypeRule

// ValidateInterfaceType will iterate through the list of array types and call
// ValidateInterfaceType. If an error is returned, then that error will be added
// to the batch of errors.
func (rules InterfaceTypeRules) ValidateInterfaceType(iface *ast.InterfaceType) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateInterfaceType(iface); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

// InterfaceTypeRule represents an interface that will allow for validation
// to occur on an ast.InterfaceType.
type InterfaceTypeRule interface {
	ValidateInterfaceType(*ast.InterfaceType) error
}
