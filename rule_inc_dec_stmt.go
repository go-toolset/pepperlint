package pepperlint

import (
	"go/ast"
)

// IncDecStmtRules is a list of type IncDecStmtRule.
type IncDecStmtRules []IncDecStmtRule

// ValidateIncDecStmt will iterate through the list of array types and call
// ValidateIncDecStmt. If an error is returned, then that error will be added
// to the batch of errors.
func (rules IncDecStmtRules) ValidateIncDecStmt(fn *ast.IncDecStmt) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateIncDecStmt(fn); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// IncDecStmtRule represents an interface that will allow for validation
// to occur on an ast.IncDecStmt.
type IncDecStmtRule interface {
	ValidateIncDecStmt(*ast.IncDecStmt) error
}
