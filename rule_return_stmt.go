package pepperlint

import (
	"go/ast"
)

type ReturnStmtRules []ReturnStmtRule

func (rules ReturnStmtRules) ValidateReturnStmt(stmt *ast.ReturnStmt) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateReturnStmt(stmt); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type ReturnStmtRule interface {
	ValidateReturnStmt(*ast.ReturnStmt) error
}
