package pepperlint

import (
	"go/ast"
)

// RangeStmtRules is a list of type RangeStmtRule.
type RangeStmtRules []RangeStmtRule

// ValidateRangeStmt will iterate through the list of array types and call
// ValidateRangeStmt. If an error is returned, then that error will be added
// to the batch of errors.
func (rules RangeStmtRules) ValidateRangeStmt(stmt *ast.RangeStmt) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateRangeStmt(stmt); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// RangeStmtRule represents an interface that will allow for validation
// to occur on an ast.RangeStmt.
type RangeStmtRule interface {
	ValidateRangeStmt(*ast.RangeStmt) error
}
