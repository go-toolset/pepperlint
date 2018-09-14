package pepperlint

// RulesAdder will add rules to the visitor
type RulesAdder interface {
	AddRules(*Visitor)
}
