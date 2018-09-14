package pepperlint

import (
	"go/ast"
)

type GenDeclRules []GenDeclRule

func (rules GenDeclRules) ValidateGenDecl(decl *ast.GenDecl) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateGenDecl(decl); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type GenDeclRule interface {
	ValidateGenDecl(*ast.GenDecl) error
}
