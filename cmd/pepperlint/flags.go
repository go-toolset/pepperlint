package main

import (
	"flag"
	"strings"
)

// Flags represent flags passed in via command line. These fields
// have the highest priority and will be overwrite any config values
// if the field is set.
type flags struct {
	ConfigPath  string
	IncludePkgs []string

	RuleNames    []string
	Suppressions []string
}

func newFlags() flags {
	f := flags{}

	includePkgs := ""
	flag.StringVar(
		&includePkgs,
		"include-pkgs",
		"",
		"comma separated list of directories to be included during linting",
	)

	ruleNames := ""
	flag.StringVar(
		&ruleNames,
		"rules",
		"",
		"comma separated list of rule names to run",
	)

	flag.StringVar(
		&f.ConfigPath,
		"config-path",
		"",
		"path to yaml config",
	)

	flag.Parse()

	if len(ruleNames) > 0 {
		f.RuleNames = strings.Split(ruleNames, ",")
	}

	if len(includePkgs) > 0 {
		f.IncludePkgs = strings.Split(includePkgs, ",")
	}

	return f
}

// Merge will merge flag based configuration in the config structure
func (f flags) Merge(config Config) Config {
	if len(f.IncludePkgs) > 0 {
		config.IncludePkgs = f.IncludePkgs
	}

	if len(f.RuleNames) > 0 {
		for _, ruleName := range f.RuleNames {
			config.Rules = append(config.Rules, Rule{
				RuleName: ruleName,
			})
		}
	}

	return config
}
