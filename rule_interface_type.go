package pepperlint

import (
	"go/ast"
)

type InterfaceTypeRules []InterfaceTypeRule

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

type InterfaceTypeRule interface {
	ValidateInterfaceType(*ast.InterfaceType) error
}
