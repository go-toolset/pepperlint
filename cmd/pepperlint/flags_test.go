package main

import (
	"reflect"
	"testing"
)

func TestFlagsMerge(t *testing.T) {
	cases := []struct {
		name           string
		flagsConfig    flags
		config         Config
		expectedConfig Config
	}{
		{
			name:           "simple case",
			flagsConfig:    flags{},
			config:         Config{},
			expectedConfig: Config{},
		},
		{
			name: "include packages case",
			flagsConfig: flags{
				IncludePkgs: []string{
					"foo",
					"bar",
				},
			},
			config: Config{},
			expectedConfig: Config{
				IncludePkgs: []string{
					"foo",
					"bar",
				},
			},
		},
		{
			name: "rules case",
			flagsConfig: flags{
				RuleNames: []string{
					"foo",
					"bar",
				},
			},
			config: Config{},
			expectedConfig: Config{
				Rules: []Rule{
					{
						"foo",
					},
					{
						"bar",
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			config := c.flagsConfig.Merge(c.config)

			if e, a := c.expectedConfig, config; !reflect.DeepEqual(e, a) {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}
