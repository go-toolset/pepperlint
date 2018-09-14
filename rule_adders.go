package pepperlint

type RulesAdder interface {
	AddRules(*Visitor)
}
