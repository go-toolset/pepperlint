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
			name: "simple_deprecated_struct",
			code: `package foo

// Foo fake docs here
// Deprecated: use Bar instead
type Foo struct {
	Field int32
}

func deprecated() interface{} {
	f := Foo{
		Field: 123,
	}
}
			`,
			rulesFn: func(fset *token.FileSet) *core.DeprecatedStructRule {
				return core.NewDeprecatedStructRule(fset)
			},
			expectedErrors: 1,
		},
		{
			name: "simple_deprecated_struct_pointers",
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

func deprecated() interface{} {
	fPtr := &Foo{
		Field: 123,
	}

	var moo *Foo

	if moo == nil {
	}
}`,
			rulesFn: func(fset *token.FileSet) *core.DeprecatedStructRule {
				return core.NewDeprecatedStructRule(fset)
			},
			expectedErrors: 3,
		},
		{
			name: "simple_deprecated_struct_array",
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

func deprecated() interface{} {
	fArray := []Foo{
		{},
	}

	fPtrArray := []*Foo{
		{},
	}

	bazArray := []Baz{
		{},
	}

	bazPtrArray := []*Baz{
		{},
	}
}`,
			rulesFn: func(fset *token.FileSet) *core.DeprecatedStructRule {
				return core.NewDeprecatedStructRule(fset)
			},
			expectedErrors: 4,
		},
		{
			name: "deprecated_struct_return_stmts",
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

var x int

func deprecated() interface{} {
	foo := Foo{}
	fooPtr := &Foo{}
	var moo *Foo

	if x > 0 {
		return foo
	} else if fooPtr != nil {
		return fooPtr
	}

	return moo
}`,
			rulesFn: func(fset *token.FileSet) *core.DeprecatedStructRule {
				return core.NewDeprecatedStructRule(fset)
			},
			expectedErrors: 6,
		},
		{
			name: "deprecated_struct_function_param",
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

var x int

func deprecated() interface{} {
	foo := Foo{}
	fooPtr := &Foo{}

	check(foo)
	checkPtr(fooPtr)
	check(Foo{})
	checkPtr(&Foo{})
}

func check(v interface{}) {
}`,
			rulesFn: func(fset *token.FileSet) *core.DeprecatedStructRule {
				return core.NewDeprecatedStructRule(fset)
			},
			expectedErrors: 6,
		},
		{
			name: "deprecated_struct_function_definitions",
			code: `package foo

// Foo fake docs here
// Deprecated: use Bar instead
type Foo struct {
	Field int32
}

func Deprecated(f Foo) {
}

func DeprecatedReturn() Foo {
	return Foo{}
}
`,
			rulesFn: func(fset *token.FileSet) *core.DeprecatedStructRule {
				return core.NewDeprecatedStructRule(fset)
			},
			expectedErrors: 3,
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

			pepperlint.PackagesCache.Packages[""] = &pepperlint.Package{}
			// populate cache
			ast.Walk(pepperlint.PackagesCache, node)

			ast.Walk(v, node)
			t.Log("\n", "\b\b", v.Errors)

			if e, a := c.expectedErrors, v.Errors.Count(); e != a {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}
