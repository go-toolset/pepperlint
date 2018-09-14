package pepperlint

import (
	"go/ast"
)

// AssignStmtRules is a list of type AssignStmtRule.
type AssignStmtRules []AssignStmtRule

// ValidateAssignStmt will iterate through the list of array types and call
// ValidateAssignStmt. If an error is returned, then that error will be added
// to the batch of errors.
func (rules AssignStmtRules) ValidateAssignStmt(stmt *ast.AssignStmt) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateAssignStmt(stmt); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// AssignStmtRule represents an interface that will allow for validation
// to occur on an ast.AssignStmt.
type AssignStmtRule interface {
	ValidateAssignStmt(*ast.AssignStmt) error
}
