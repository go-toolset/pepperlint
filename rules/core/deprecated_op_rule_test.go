package core_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/go-toolset/pepperlint"
	"github.com/go-toolset/pepperlint/rules/core"
)

func TestDeprecateOpRule(t *testing.T) {
	cases := []struct {
		name           string
		code           string
		rulesFn        func(fset *token.FileSet) *core.DeprecatedOpRule
		expectedErrors int
	}{
		{
			name: "simple_deprecated_op",
			code: `package foo
			type Foo struct {}

			// DeprecatedOp op
			//
			// Deprecated: Use Foo instead
			func (f Foo) DeprecatedOp() {
			}

			// DeprecatedPtrOp op
			//
			// Deprecated: Use Foo instead
			func (f *Foo) DeprecatedPtrOp() {
			}

			// DeprecatedFunction op
			//
			// Deprecated: Use Foo instead
			func DeprecatedFunction() int {
				return 1
			}

			func deprecated() {
				f := Foo{}

				f.DeprecatedOp()
				f.DeprecatedPtrOp()

				DeprecatedFunction()
				v := DeprecatedFunction()
			}
			`,
			rulesFn: func(fset *token.FileSet) *core.DeprecatedOpRule {
				return core.NewDeprecatedOpRule(fset)
			},
			expectedErrors: 4,
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

			cache := &pepperlint.Cache{
				Packages: pepperlint.Packages{},
			}
			cache.Packages[""] = &pepperlint.Package{}

			v := pepperlint.NewVisitor(fset, cache, c.rulesFn(fset))

			// populate cache
			ast.Walk(cache, node)

			ast.Walk(v, node)
			pepperlint.Log("%v", v.Errors)

			if e, a := c.expectedErrors, len(v.Errors); e != a {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}
