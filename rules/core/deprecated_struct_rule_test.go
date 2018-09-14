package core_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/go-toolset/pepperlint"
	"github.com/go-toolset/pepperlint/rules/core"
)

func TestDeprecatedStructRule(t *testing.T) {
	cases := []struct {
		name           string
		code           string
		rulesFn        func(fset *token.FileSet) *core.DeprecatedStructRule
		expectedErrors int
	}{
		{
			name: "simple_deprecated_field",
			code: `package foo

// Foo fake docs here
// Deprecated: use Bar instead
type Foo struct {
	Field int32
}

type Bar struct {
	Field int64
}

type Qux Foo

type Baz Qux

func deprecated() int32 {
	f := Foo{
		Field: 123,
	}

	f = Foo{
		123,
	}

	b := Baz{}

	check(Foo{})
	checkPtr(&Foo{})
	return Foo{}
}

func check(foo Foo) Foo {
	return foo
}

func checkPtr(foo *Foo) *Foo {
	return foo
}
			`,
			rulesFn: func(fset *token.FileSet) *core.DeprecatedStructRule {
				return core.NewDeprecatedStructRule(fset)
			},
			expectedErrors: 10,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, c.name, c.code, parser.ParseComments)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if node == nil {
				t.Fatal("unexpected nil expr")
			}

			v := pepperlint.NewVisitor(fset, c.rulesFn(fset))

			ast.Walk(v, node)
			t.Log("\n", "\b\b", v.Errors)

			if e, a := c.expectedErrors, len(v.Errors); e != a {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}
