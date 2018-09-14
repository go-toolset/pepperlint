package pepperlint

import (
	"go/ast"
)

type PackageRules []PackageRule

func (rules PackageRules) ValidatePackage(pkg *ast.Package) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidatePackage(pkg); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type PackageRule interface {
	ValidatePackage(*ast.Package) error
}
