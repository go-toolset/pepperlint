package pepperlint

import (
	"go/ast"
)

// ChanTypeRules is a list of type ChanTypeRule.
type ChanTypeRules []ChanTypeRule

// ValidateChanType will iterate through the list of array types and call
// ValidateChanType. If an error is returned, then that error will be added
// to the batch of errors.
func (rules ChanTypeRules) ValidateChanType(ch *ast.ChanType) error {
	batchError := NewBatchError()
	for _, rule := range rules {
		if err := rule.ValidateChanType(ch); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// ChanTypeRule represents an interface that will allow for validation
// to occur on an ast.ChanType.
type ChanTypeRule interface {
	ValidateChanType(*ast.ChanType) error
}
