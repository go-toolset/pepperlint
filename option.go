package pepperlint

import (
	"go/token"
)

// Option will allow rules to be prepped with defined visitor
// Rules or caching can be initialized during the 'With' call of
// the option
type Option interface{}

// Rule is an empty interface but will allow for modifications later
// if this needs to be a more specific type.
type Rule interface{}

// RulesAdder will add rules to sets of rules
type RulesAdder interface {
	AddRules(*Rules)
}

// CacheOption will allow rules to perform action based on what is
// in the cache.
type CacheOption interface {
	WithCache(*Cache)
}

// FileSetOption will set the token.FileSet to a given
// rule.
type FileSetOption interface {
	WithFileSet(*token.FileSet)
}

// CopyRuler is used to copy a pre-existing rule allowing for a new
// rule to be acted upon indenpendently from the rule it is copied from.
type CopyRuler interface {
	CopyRule() Rule
}
