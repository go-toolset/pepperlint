package core_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/go-toolset/pepperlint"
	"github.com/go-toolset/pepperlint/rules/core"
)

func TestDeprecateFieldRule(t *testing.T) {
	cases := []struct {
		name           string
		code           string
		rulesFn        func(fset *token.FileSet) *core.DeprecatedFieldRule
		expectedErrors int
	}{
		{
			name: "simple_deprecated_field",
			code: `package foo
			type Foo struct {
				// Deprecated: foo use Field
				DeprecatedField int32
				Field int64
			}

			func (fife *Foo) SetDeprecatedField(v int32) {
				fife.DeprecatedField = v
			}

			func deprecated() int32 {
				f := Foo{
					DeprecatedField: 123,
				}

				f = Foo{
					123,
					456,
				}

				f.DeprecatedField = 0xFFFF
				check(f.DeprecatedField)
				check(1337)
				f.SetDeprecatedField(-10)

				return f.DeprecatedField
			}

			func check(v int32) {
			}
			`,
			rulesFn: func(fset *token.FileSet) *core.DeprecatedFieldRule {
				return core.NewDeprecatedFieldRule(fset)
			},
			expectedErrors: 6,
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
