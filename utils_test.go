package pepperlint

import (
	"go/ast"
	"reflect"
	"testing"
)

func TestIsStruct(t *testing.T) {
	cases := []struct {
		name     string
		expr     ast.Expr
		expected bool
	}{
		{
			name:     "struct type",
			expr:     &ast.StructType{},
			expected: true,
		},
		{
			name:     "empty ident node",
			expr:     &ast.Ident{},
			expected: false,
		},
		{
			name: "ident node",
			expr: &ast.Ident{
				Obj: &ast.Object{
					Decl: &ast.TypeSpec{
						Type: &ast.StructType{},
					},
				},
			},
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			helper := Helper{}
			if e, a := c.expected, helper.IsStruct(c.expr); e != a {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}

func TestGetStructType(t *testing.T) {
	cases := []struct {
		name     string
		expr     ast.Expr
		expected *ast.StructType
	}{
		{
			name: "empty struct type",
			expr: &ast.StructType{
				Incomplete: true,
			},
			expected: &ast.StructType{
				Incomplete: true,
			},
		},
		{
			name:     "empty ident node",
			expr:     &ast.Ident{},
			expected: nil,
		},
		{
			name: "ident node",
			expr: &ast.Ident{
				Obj: &ast.Object{
					Decl: &ast.TypeSpec{
						Type: &ast.StructType{},
					},
				},
			},
			expected: &ast.StructType{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			helper := Helper{}
			if e, a := c.expected, helper.GetStructType(c.expr); !reflect.DeepEqual(e, a) {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}

func TestGetImportPathFromFullPath(t *testing.T) {
	cases := []struct {
		name         string
		prefixes     []string
		path         string
		expectedPath string
	}{
		{
			name:         "no prefixes",
			path:         "foobar",
			expectedPath: "foobar",
		},
		{
			name: "simple prefixes",
			prefixes: []string{
				"foo",
				"bar",
			},
			path:         "foobar",
			expectedPath: "bar",
		},
		{
			name: "prefixes not match",
			prefixes: []string{
				"foo",
				"bar",
			},
			path:         "bazbar",
			expectedPath: "bazbar",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if e, a := c.expectedPath, getImportPathFromFullPath(c.prefixes, c.path); e != a {
				t.Errorf("expected %q, but received %q", e, a)
			}
		})
	}
}
