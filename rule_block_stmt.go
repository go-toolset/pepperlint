package pepperlint

import (
	"go/ast"
)

// BlockStmtRules is a list of type BlockStmtRule.
type BlockStmtRules []BlockStmtRule

// ValidateBlockStmt will iterate through the list of array types and call
// ValidateBlockStmt. If an error is returned, then that error will be added
// to the batch of errors.
func (rules BlockStmtRules) ValidateBlockStmt(stmt *ast.BlockStmt) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateBlockStmt(stmt); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// BlockStmtRule represents an interface that will allow for validation
// to occur on an ast.BlockStmt.
type BlockStmtRule interface {
	ValidateBlockStmt(*ast.BlockStmt) error
}
