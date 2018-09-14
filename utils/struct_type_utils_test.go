package utils_test

import (
	"go/ast"
	"reflect"
	"testing"

	"github.com/go-toolset/pepperlint/utils"
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
			if e, a := c.expected, utils.IsStruct(c.expr); e != a {
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
			if e, a := c.expected, utils.GetStructType(c.expr); !reflect.DeepEqual(e, a) {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}
