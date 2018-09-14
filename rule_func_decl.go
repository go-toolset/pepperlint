package pepperlint

import (
	"go/ast"
)

// FuncDeclRules is a list of type FuncDeclRule.
type FuncDeclRules []FuncDeclRule

// ValidateFuncDecl will iterate through the list of array types and call
// ValidateFuncDecl. If an error is returned, then that error will be added
// to the batch of errors.
func (rules FuncDeclRules) ValidateFuncDecl(decl *ast.FuncDecl) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateFuncDecl(decl); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// FuncDeclRule represents an interface that will allow for validation
// to occur on an ast.FuncDecl.
type FuncDeclRule interface {
	ValidateFuncDecl(*ast.FuncDecl) error
}
