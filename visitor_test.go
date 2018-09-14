package pepperlint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestVisitor(t *testing.T) {
	cases := []struct {
		name           string
		visitor        Visitor
		expectedErrors Errors
	}{
		{
			name: "exclude struct rule",
			visitor: Visitor{
				Rules: Rules{
					TypeSpecRules: TypeSpecRules{
						testExcludeNameTypeSpecRule{
							Name: "Foo",
						},
					},
					StructTypeRules: StructTypeRules{
						testExcludeField{
							Name: "Boo",
						},
					},
					FuncDeclRules: FuncDeclRules{
						testExcludeMethod{
							StructName: "Foo",
							Name:       "Baz",
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			filenames, err := filepath.Glob("testdata/*")
			if err != nil {
				t.Fatal(err)
			}

			for _, filepath := range filenames {
				if strings.Contains(filepath, ".expected.go") {
					continue
				}

				fset := token.NewFileSet()
				node, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
				if err != nil {
					t.Errorf("unexpected error %v", err)
				}

				ast.Walk(&c.visitor, node)

				expectedFile := strings.TrimRight(filepath, ".go") + ".expected.go"
				b, err := ioutil.ReadFile(expectedFile)
				if err != nil {
					t.Errorf("unexpected error %v", err)
				}

				if e, a := string(b), c.visitor.Errors.Error(); e != a {
					t.Errorf("expected %v, but received %v", e, a)
				}
			}
		})
	}
}
