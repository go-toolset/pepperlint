package pepperlint

import (
	"go/ast"
)

// ReturnStmtRules is a list of type ReturnStmtRule.
type ReturnStmtRules []ReturnStmtRule

// ValidateReturnStmt will iterate through the list of array types and call
// ValidateReturnStmt. If an error is returned, then that error will be added
// to the batch of errors.
func (rules ReturnStmtRules) ValidateReturnStmt(stmt *ast.ReturnStmt) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateReturnStmt(stmt); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// ReturnStmtRule represents an interface that will allow for validation
// to occur on an ast.ReturnStmt.
type ReturnStmtRule interface {
	ValidateReturnStmt(*ast.ReturnStmt) error
}
