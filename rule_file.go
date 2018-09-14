package pepperlint

import (
	"go/ast"
)

type FileRules []FileRule

func (rules FileRules) ValidateFile(stmt *ast.File) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateFile(stmt); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type FileRule interface {
	ValidateFile(*ast.File) error
}
