package main

import (
	"io/ioutil"

	"github.com/go-toolset/pepperlint"
	"github.com/go-toolset/pepperlint/rules"

	"github.com/go-yaml/yaml"
)

// Config is used to determine which rules will be run along with which errors
// will be suppressed.
type Config struct {
	Rules        Rules        `yaml:"rules"`
	Suppressions Suppressions `yaml:"suppressions"`
}

// NewConfig returns a new config at a given path.
func NewConfig(path string) (Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	config := Config{}
	if err := yaml.Unmarshal(b, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

// Options will return a list of options from a given rule name
func (cfg Config) Options() []pepperlint.Option {
	opts := []pepperlint.Option{}

	for _, rule := range cfg.Rules {
		r := rules.Get(rule.RuleName)
		opts = append(opts, r)
	}

	return opts
}

// Rules represents a list of rules
type Rules []Rule

// Rule is a shape definition of what a rule object will look like
// in the yaml configuration.
type Rule struct {
	RuleName string `yaml:"rule_name"`
}

// Suppressions represents a list of suppressions
type Suppressions []Suppression

// Suppression is a shape definition of what a suppression object will look like
// in the yaml configuration.
type Suppression struct {
	File *File `yaml:"file"`
}

// File represents a File object that will be used for file based suppressions
type File struct {
	FilePath string `yaml:"file_path"`
	// LineNumber is an optional field, if it isn't set, then the
	// whole file will be suppressed.
	LineNumber *int `yaml:"line"`
}

func buildConfig(configPath string) Config {
	if len(configPath) > 0 {
		cfg, err := NewConfig(configPath)
		if err != nil {
			panic(err)
		}

		return cfg
	}

	// TODO get rules and suppressions from flags instead

	return Config{}
}
