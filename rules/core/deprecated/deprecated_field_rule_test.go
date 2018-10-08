package deprecated_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/go-toolset/pepperlint"
	"github.com/go-toolset/pepperlint/rules/core/deprecated"
)

func TestDeprecateFieldRule(t *testing.T) {
	cases := []struct {
		name           string
		code           string
		rulesFn        func(fset *token.FileSet) *deprecated.FieldRule
		expectedErrors int
	}{
		{
			name: "simple_deprecated_field",
			code: `package foo
			type Foo struct {
				// Deprecated: foo use Field
				DeprecatedField int32
			}

			func deprecated() interface{} {
				f := Foo{
					DeprecatedField: 123,
				}

				f = Foo{
					123,
				}
			}`,
			rulesFn: func(fset *token.FileSet) *deprecated.FieldRule {
				return deprecated.NewFieldRule(fset)
			},
			expectedErrors: 2,
		},
		{
			name: "deprecated_field_math",
			code: `package foo
			type Foo struct {
				// Deprecated: foo use Field
				DeprecatedField int32
			}

			func deprecated() interface{} {
				f := Foo{}
				f.DeprecatedField = 0	
				f.DeprecatedField++	
				f.DeprecatedField	+= 1
				f.DeprecatedField	*= 1
				f.DeprecatedField	/= 1
				f.DeprecatedField	%= 2

				f.DeprecatedField	= 2 + 1
				n := f.DeprecatedField + 1
			}`,
			rulesFn: func(fset *token.FileSet) *deprecated.FieldRule {
				return deprecated.NewFieldRule(fset)
			},
			expectedErrors: 8,
		},
		{
			name: "deprecated_field_array",
			code: `package foo
			type Foo struct {
				// Deprecated: foo use Field
				DeprecatedField []int
			}

			func deprecated() interface{} {
				f := Foo{
					DeprecatedField: []int{0,1,2},
				}

				f.DeprecatedField = []int{4,5,6}
				f.DeprecatedField = append(f.DeprecatedField, 7)
			}`,
			rulesFn: func(fset *token.FileSet) *deprecated.FieldRule {
				return deprecated.NewFieldRule(fset)
			},
			expectedErrors: 4,
		},
		{
			name: "deprecated_field_operations",
			code: `package foo
			type Foo struct {
				// Deprecated: foo use Field
				DeprecatedField int
			}

			func (f *Foo) SetDeprecatedField(v int) {
				f.DeprecatedField =  v
			}

			func (f *Foo) DeprecatedRet() int {
				return f.DeprecatedField
			}

			func deprecated() interface{} {
				f := Foo{}
				f.SetDeprecatedField(1)
				depOp(f.DeprecatedField)
			}
			
			func depOp(v int) {
			}`,
			rulesFn: func(fset *token.FileSet) *deprecated.FieldRule {
				return deprecated.NewFieldRule(fset)
			},
			expectedErrors: 3,
		},
		{
			name: "deprecated_field_stmts",
			code: `package foo
			type Foo struct {
				// Deprecated: deprecated
				DeprecatedField int

				// Deprecated: deprecated
				DeprecatedArray []int
			}

			func deprecated() interface{} {
				f := Foo{}

				if f.DeprecatedField == 0 {
					return f.DeprecatedField
				}

				if len(f.DeprecatedArray) == 0 {
					return f.DeprecatedArray
				}

				for i := 0; i < len(f.DeprecatedArray); i++ {
					n := f.DeprecatedArray[i] // TODO
					n += 1
					f.DeprecatedArray[i] = n
				}

				return nil
			}
			`,
			rulesFn: func(fset *token.FileSet) *deprecated.FieldRule {
				return deprecated.NewFieldRule(fset)
			},
			expectedErrors: 7,
		},
		{
			name: "complex_deprecated_field",
			code: `package foo
			type Foo struct {
				// Deprecated: foo use Field
				DeprecatedField int32

				// Deprecated: is deprecated
				DeprecatedArrayField []int32
				Field int64
			}

			func (fife *Foo) SetDeprecatedField(v int32) {
				fife.DeprecatedField = v
			}

			func (fife *Foo) SetDeprecatedArrayField(arr []int32) {
				fife.DeprecatedArrayField = arr
			}

			func deprecated() interface{} {
				f := Foo{
					DeprecatedField: 123,
					DeprecatedArrayField: []int32{1,2,3},
				}

				f = Foo{
					123,
					[]int32{4,5},
					456,
				}

				f.DeprecatedField++
				f.DeprecatedField += 1
				f.DeprecatedField = 0xFFFF
				f.SetDeprecatedField(-10)

				f.DeprecatedArrayField = []int32{6}
				f.SetDeprecatedArrayField([]int32{2,4,6})
				for _, elem := range f.DeprecatedArrayField { // TODO
					check(elem)
				}
				for i := 0; i < len(f.DeprecatedArrayField); i++ {
					f.DeprecatedArrayField[i]++
					f.DeprecatedArrayField[i] += 2
					f.DeprecatedArrayField[i] = 3
				}

				check(1337)
				check(f.DeprecatedField)
				check(f.DeprecatedField + 1)
				check(f.DeprecatedArrayField)

				if f.DeprecatedField > 0 {
					return f.DeprecatedField
				}

				if len(f.DeprecatedArrayField) > 0 {
					return f.DeprecatedArrayField
				}

				return 0
			}

			func check(v int32) {
			}

			func checkArr(v []int32) {
			}
			`,
			rulesFn: func(fset *token.FileSet) *deprecated.FieldRule {
				return deprecated.NewFieldRule(fset)
			},
			// TODO: Is this number correct?
			expectedErrors: 22,
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

			if e, a := c.expectedErrors, v.Errors.Count(); e != a {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}
