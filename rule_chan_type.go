package pepperlint

import (
	"go/ast"
)

type ChanTypeRules []ChanTypeRule

func (rules ChanTypeRules) ValidateChanType(ch *ast.ChanType) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateChanType(ch); err != nil {
			batchError.Add(err)
		}
	}

	if batchError.Len() == 0 {
		return nil
	}

	return batchError
}

type ChanTypeRule interface {
	ValidateChanType(*ast.ChanType) error
}
