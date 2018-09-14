package pepperlint

import (
	"go/ast"
)

type AssignStmtRules []AssignStmtRule

func (rules AssignStmtRules) ValidateAssignStmt(stmt *ast.AssignStmt) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateAssignStmt(stmt); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type AssignStmtRule interface {
	ValidateAssignStmt(*ast.AssignStmt) error
}
