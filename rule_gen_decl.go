package pepperlint

import (
	"go/ast"
)

// GenDeclRules is a list of type GenDeclRule.
type GenDeclRules []GenDeclRule

// ValidateGenDecl will iterate through the list of array types and call
// ValidateGenDecl. If an error is returned, then that error will be added
// to the batch of errors.
func (rules GenDeclRules) ValidateGenDecl(decl *ast.GenDecl) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateGenDecl(decl); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// GenDeclRule represents an interface that will allow for validation
// to occur on an ast.GenDecl.
type GenDeclRule interface {
	ValidateGenDecl(*ast.GenDecl) error
}
