package main

import (
	"fmt"
	"reflect"
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
				21,
				25,
				26,
				28,
				31,
				32,
				34,
				35,
				36,
				38,
				39,
				41,
				43,
				46,
				47,
			},
		},
	}

	for _, c := range cases {
		config := Config{
			Rules: Rules{
				{
					RuleName: "deprecated",
				},
			},
		}
		v, _, err := lint(config, c.includeDirs, c.mainPackage)
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

		lerrs := []pepperlint.LineNumber{}
		for _, err := range errs {
			e := err.(*pepperlint.BatchError)
			es := e.Errors()

			for _, err := range es {
				lerr := err.(pepperlint.LineNumber)
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
		pepperlint.Log("ERRORS %v", v.Errors)
	}
}

type mockError struct {
	line     int
	filename string
}

func (m mockError) LineNumber() int {
	return m.line
}

func (m mockError) Filename() string {
	return m.filename
}

func (m mockError) Error() string {
	return fmt.Sprintf("%s:%d", m.Filename(), m.LineNumber())
}

func TestMainSuppression(t *testing.T) {
	line := 5

	cases := []struct {
		name           string
		errors         []error
		suppressions   Suppressions
		expectedErrors []error
	}{
		{
			name:           "empty case",
			expectedErrors: []error{},
		},
		{
			name: "simple supression by filename",
			suppressions: Suppressions{
				{
					File: &File{
						FilePath: "foo.go",
					},
				},
			},
			errors: []error{
				mockError{
					filename: "foo.go",
				},
				mockError{
					filename: "bar.go",
				},
			},
			expectedErrors: []error{
				mockError{
					filename: "bar.go",
				},
			},
		},
		{
			name: "simple supression by filename and line",
			suppressions: Suppressions{
				{
					File: &File{
						FilePath:   "foo.go",
						LineNumber: &line,
					},
				},
			},
			errors: []error{
				mockError{
					filename: "bar.go",
					line:     1,
				},
				mockError{
					filename: "foo.go",
					line:     5,
				},
				mockError{
					filename: "bar.go",
					line:     10,
				},
			},
			expectedErrors: []error{
				mockError{
					filename: "bar.go",
					line:     1,
				},
				mockError{
					filename: "bar.go",
					line:     10,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := suppress(c.suppressions, c.errors)

			if e, a := c.expectedErrors, errs; !reflect.DeepEqual(e, a) {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}
