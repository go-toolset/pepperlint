package pepperlint

// Option  will allow rules to be prepped with defined visitor
// Rules or caching can be initialized during the 'With' call of
// the option
type Option interface{}

// RulesAdder will add rules to sets of rules
type RulesAdder interface {
	AddRules(*Rules)
}

// CacheOption will allow rules to perform action based on what is
// in the cache.
type CacheOption interface {
	WithCache(*Cache)
}
