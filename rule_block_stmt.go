package pepperlint

import (
	"go/ast"
)

type BlockStmtRules []BlockStmtRule

func (rules BlockStmtRules) ValidateBlockStmt(stmt *ast.BlockStmt) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateBlockStmt(stmt); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type BlockStmtRule interface {
	ValidateBlockStmt(*ast.BlockStmt) error
}
