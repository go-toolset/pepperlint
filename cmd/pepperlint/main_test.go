package main

import (
	"fmt"
	"testing"

	"github.com/go-toolset/pepperlint"
)

func TestMain(t *testing.T) {
	cases := []struct {
		includeDirs         []string
		mainPackage         string
		expectedLineNumbers []int
	}{
		{
			mainPackage: "github.com/go-toolset/pepperlint/cmd/pepperlint/testdata/core",
			includeDirs: []string{
				"github.com/go-toolset/pepperlint/cmd/pepperlint/testdata/deprecated",
			},
			expectedLineNumbers: []int{
				10,
				12,
				12,
				13,
				14,
				17,
				18,
				23,
				24,
				25,
				28,
				29,
			},
		},
	}

	for _, c := range cases {
		v, _, err := lint(c.includeDirs, c.mainPackage)
		if err != nil {
			t.Fatal(err)
		}

		errs := []error{}
		for _, err := range v.Errors {
			e := err.(*pepperlint.BatchError)
			es := e.Errors()
			for _, err := range es {
				errs = append(errs, err)
			}
		}

		lerrs := []pepperlint.LineNumberError{}
		for _, err := range errs {
			e := err.(*pepperlint.BatchError)
			es := e.Errors()

			for _, err := range es {
				lerr := err.(pepperlint.LineNumberError)
				lerrs = append(lerrs, lerr)
			}
		}

		if e, a := len(c.expectedLineNumbers), len(lerrs); e != a {
			numbers := []int{}
			for _, lerr := range lerrs {
				numbers = append(numbers, lerr.LineNumber())
			}
			t.Fatal(fmt.Sprintf("expected %v, but received %v: %v", e, a, numbers))
		}

		for i, e := range lerrs {
			if e, a := c.expectedLineNumbers[i], e.LineNumber(); e != a {
				t.Errorf("expected %v, but received %v", e, a)
			}
		}
		t.Log("ERRORS", v.Errors)
	}
}
