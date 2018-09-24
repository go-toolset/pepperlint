package rules

import (
	"github.com/go-toolset/pepperlint"
)

type registry map[string]pepperlint.CopyRuler

var rulesRegistry = registry{}

func (r *registry) Add(name string, opt pepperlint.CopyRuler) {
	(*r)[name] = opt
}

func (r registry) Get(name string) pepperlint.CopyRuler {
	return r[name]
}

// Add will add a new copy ruler to the rules registry. The reason why this
// take a copy ruler instead of a rule directly is that some rules may need
// to be pointers, and sharing pointers between visitors will be problematic.
func Add(name string, opt pepperlint.CopyRuler) {
	rulesRegistry.Add(name, opt)
}

// Get will return the copy rule based on name
func Get(name string) pepperlint.Option {
	return rulesRegistry.Get(name).CopyRule()
}
