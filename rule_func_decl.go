package pepperlint

import (
	"go/ast"
)

type FuncDeclRules []FuncDeclRule

func (rules FuncDeclRules) ValidateFuncDecl(decl *ast.FuncDecl) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateFuncDecl(decl); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type FuncDeclRule interface {
	ValidateFuncDecl(*ast.FuncDecl) error
}
