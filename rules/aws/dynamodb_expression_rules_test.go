package aws

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"testing"

	"github.com/go-toolset/pepperlint"
)

func TestDynamoDBExpressionRule(t *testing.T) {
	cases := []struct {
		name                string
		code                string
		rulesFn             func(fset *token.FileSet) *DynamoDBExpressionRule
		pkgCache            pepperlint.Packages
		expectedLineNumbers []int
	}{
		{
			name: "simple",
			code: `package foo
			import (
				"github.com/aws/aws-sdk-go/service/dynamodb"
				"github.com/aws/aws-sdk-go/service/dynamodb/expression"
			)

			func bar() {
				input := dynamodb.QueryInput{
					FilterExpression: aws.String("some expr"),
				}

				filt := expression.Name("Artist").Equal(expression.Value("No One You Know"))
				proj := expression.NamesList(expression.Name("SongTitle"), expression.Name("AlbumTitle"))
				expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
				if err != nil {
					return
				} 

				input = dynamodb.QueryInput{
					FilterExpression: expr.Filter(),
				}
			}`,
			rulesFn: func(fset *token.FileSet) *DynamoDBExpressionRule {
				r := NewDynamoDBExpressionRule(fset)
				r.hasService = true
				return r
			},
			pkgCache: pepperlint.Packages{
				"": {
					Files: pepperlint.Files{
						{
							Imports: map[string]string{
								"dynamodb":   "github.com/aws/aws-sdk-go/service/dynamodb",
								"expression": "github.com/aws/aws-sdk-go/service/dynamodb/expression",
							},
						},
					},
				},
				"github.com/aws/aws-sdk-go/service/dynamodb": {
					Files: pepperlint.Files{
						{
							TypeInfos: pepperlint.TypeInfos{
								"QueryInput": {
									Spec: &ast.TypeSpec{
										Name: &ast.Ident{
											Name: "QueryInput",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedLineNumbers: []int{
				9,
			},
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
				Packages: c.pkgCache,
			}

			v := pepperlint.NewVisitor(fset, cache, c.rulesFn(fset))

			// populate cache
			ast.Walk(cache, node)

			ast.Walk(v, node)

			errs := []error{}
			for _, err := range v.Errors {
				e := err.(*pepperlint.BatchError)
				es := e.Errors()
				for _, err := range es {
					errs = append(errs, err)
				}
			}

			lines := []int{}
			for _, err := range errs {
				e := err.(*pepperlint.BatchError)
				es := e.Errors()

				for _, err := range es {
					lerr := err.(pepperlint.LineNumber)
					lines = append(lines, lerr.LineNumber())
				}
			}

			t.Log(v.Errors)
			if e, a := c.expectedLineNumbers, lines; !reflect.DeepEqual(e, a) {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}
