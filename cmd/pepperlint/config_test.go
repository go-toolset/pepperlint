package main

import (
	"reflect"
	"testing"

	"github.com/go-toolset/pepperlint"
	"github.com/go-toolset/pepperlint/rules"
)

type mockRule struct{}

func (r mockRule) CopyRule() pepperlint.Rule {
	return mockRule{}
}

func TestConfigRules(t *testing.T) {
	cases := []struct {
		configPath      string
		expectedError   bool
		expectedOptions []pepperlint.Option
	}{
		{
			configPath: "testdata/config.yaml",
			expectedOptions: []pepperlint.Option{
				mockRule{},
			},
		},
	}

	for _, c := range cases {
		cfg, err := NewConfig(c.configPath)

		if e, a := c.expectedError, err != nil; e != a {
			t.Errorf("expected error is %t, but received %t error: %v", e, a, err)
		}

		if e, a := c.expectedOptions, cfg.Options(); !reflect.DeepEqual(e, a) {
			t.Errorf("expected %v, but received %v", e, a)
		}
	}
}

func init() {
	rules.Add("mock", mockRule{})
}
