package pepperlint

import (
	"go/ast"
)

// PackageRules is a list of type PackageRule.
type PackageRules []PackageRule

// ValidatePackage will iterate through the list of array types and call
// ValidatePackage. If an error is returned, then that error will be added
// to the batch of errors.
func (rules PackageRules) ValidatePackage(pkg *ast.Package) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidatePackage(pkg); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// PackageRule represents an interface that will allow for validation
// to occur on an ast.Package.
type PackageRule interface {
	ValidatePackage(*ast.Package) error
}
