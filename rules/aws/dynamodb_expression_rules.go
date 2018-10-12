package aws

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/go-toolset/pepperlint"
	"github.com/go-toolset/pepperlint/rules"
)

// dynamodbImport path represents the AWS SDK for Go's import path
// to dynamodb. It is quoted since this is visiting the ImportPathSpec,
// which includes quotes.
const dynamodbImport = `"github.com/aws/aws-sdk-go/service/dynamodb"`
const dynamodbExpressionImport = "github.com/aws/aws-sdk-go/service/dynamodb/expression"
const dynamodbName = "dynamodb"

var expressionFields = map[string]object{
	"QueryInput": object{
		fields: map[string]object{
			"FilterExpression": object{
				expression: true,
			},
			"ExpressionAttributeNames": object{
				expression: true,
			},
			"ExpressionAttributeValues": object{
				expression: true,
			},
			"KeyConditionExpression": object{
				expression: true,
			},
			"ProjectionExpression": object{
				expression: true,
			},
		},
	},

	"ScanInput": object{
		fields: map[string]object{
			"FilterExpression": object{
				expression: true,
			},
			"ExpressionAttributeNames": object{
				expression: true,
			},
			"ExpressionAttributeValues": object{
				expression: true,
			},
			"KeyConditionExpression": object{
				expression: true,
			},
			"ProjectionExpression": object{
				expression: true,
			},
		},
	},
}

type object struct {
	fields     map[string]object
	expression bool
}

// DynamoDBExpressionRule is a rule that validates whether or not
// an expressions package should have been used. Currently this will
// use the expressionFields cache to check whether or not it is usable
// by the expression package.
type DynamoDBExpressionRule struct {
	hasService bool
	pkgName    string

	fset   *token.FileSet
	helper pepperlint.Helper
}

// NewDynamoDBExpressionRule returns a new rule with the given token.FileSet
func NewDynamoDBExpressionRule(fset *token.FileSet) *DynamoDBExpressionRule {
	return &DynamoDBExpressionRule{
		fset: fset,
	}
}

// ValidateFile will iterate through the import specs to determine whether or not
// dynamodb package is being used. This will also peel off the pkg name incase of
// package name overriding.
func (r *DynamoDBExpressionRule) ValidateFile(file *ast.File) error {
	r.hasService = false
	r.pkgName = dynamodbName

	for _, spec := range file.Imports {
		if spec.Path.Value == dynamodbImport {
			r.hasService = true

			if spec.Name == nil {
				return nil
			}

			r.pkgName = spec.Name.Name
			return nil
		}
	}

	return nil
}

// ValidateAssignStmt will check if an assignment statement is using an expression
// builder compatible structure and field. If a field is being used, it will see
// if the expression builder is being used. If it is not, then an error will be
// returned
func (r DynamoDBExpressionRule) ValidateAssignStmt(stmt *ast.AssignStmt) error {
	if !r.hasService {
		return nil
	}

	exprs := []ast.Expr{}
	for _, rhs := range stmt.Rhs {
		switch t := rhs.(type) {
		case *ast.CompositeLit:
			spec := r.helper.GetTypeSpec(t.Type)
			field, ok := expressionFields[spec.Name.Name]
			if !ok {
				continue
			}

			for _, elt := range t.Elts {
				if es := r.IsUsingExpression(field, elt); len(es) > 0 {
					exprs = append(exprs, es...)
				}
			}
		default:
			pepperlint.Log("TODO: dynamodb_expression_rule ValidateAssignStmt: %T", t)
		}
	}

	batchErr := pepperlint.NewBatchError()
	for _, expr := range exprs {
		if !r.UsingExpressionsPackage(expr) {
			batchErr.Add(pepperlint.NewErrorWrap(
				r.fset,
				expr,
				fmt.Sprintf("expressions package can be used here"),
			))
		}
	}

	return batchErr.Return()
}

// IsUsingExpression will return a list of expressions that are an dynamodb.Expression.
//
// TODO: Figure out how to get field index
// TODO: This does not check whether or not typed aliased expression builders or
// if an expression string was returned by a function or method.
func (r DynamoDBExpressionRule) IsUsingExpression(obj object, elt ast.Expr) []ast.Expr {
	asts := []ast.Expr{}
	if obj.expression {
		return []ast.Expr{elt}
	}

	if elt == nil {
		return nil
	}

	switch eltType := elt.(type) {
	case *ast.BasicLit:
		return nil
	case *ast.CompositeLit:
		spec := r.helper.GetTypeSpec(eltType.Type)
		field, ok := expressionFields[spec.Name.Name]
		if !ok {
			return nil
		}

		for _, elt := range eltType.Elts {
			if exprs := r.IsUsingExpression(field, elt); len(exprs) > 0 {
				asts = append(asts, exprs...)
			}
		}

		return asts
	case *ast.KeyValueExpr:
		ident, ok := eltType.Key.(*ast.Ident)
		if !ok {
			return nil
		}

		v, ok := obj.fields[ident.Name]
		if !ok {
			return nil
		}

		return r.IsUsingExpression(v, eltType.Value)
	default:
		pepperlint.Log("TODO: dynamodb_expression_rules.IsUsingExpression %T", eltType)
	}

	return nil
}

// AddRules will add the DeprecatedFieldRule to the given visitor
func (r *DynamoDBExpressionRule) AddRules(visitorRules *pepperlint.Rules) {
	rules := pepperlint.Rules{
		AssignStmtRules: pepperlint.AssignStmtRules{r},
		FileRules:       pepperlint.FileRules{r},
	}

	visitorRules.Merge(rules)
}

// WithCache will create a new helper with the given cache. This is used
// to determine infomation about a specific ast.Node.
func (r *DynamoDBExpressionRule) WithCache(cache *pepperlint.Cache) {
	r.helper = pepperlint.NewHelper(cache)
}

// UsingExpressionsPackage checks to see if the expression package is being used.
func (r DynamoDBExpressionRule) UsingExpressionsPackage(expr ast.Expr) bool {
	f, ok := r.helper.PackagesCache.CurrentFile()
	if !ok {
		return false
	}

	pkgName, hasImport := getExpressionsPkgName(f)

	if !hasImport {
		return false
	}

	// TODO: now that the pkg name has been found, we need to walk the expr to ensure the package
	// is being used
	switch t := expr.(type) {
	case *ast.CallExpr:
		switch fun := t.Fun.(type) {
		case *ast.SelectorExpr:
			ident, ok := fun.X.(*ast.Ident)
			if !ok {
				return false
			}

			// This assumes that the ident expression is a imported function
			if ident.Obj == nil {
				// This may not be correct. For instance, the expression package could be
				// used as a return item in a function. This would not capture that.
				if ident.Name != pkgName {
					return false
				}

				return true
			}

			// Does not have a declaration?
			if ident.Obj.Decl == nil {
				return false
			}

			switch decl := ident.Obj.Decl.(type) {
			case *ast.AssignStmt:
				fieldIndex := -1
				for i, lhs := range decl.Lhs {

					switch lhsType := lhs.(type) {
					case *ast.Ident:
						if lhsType.Name == ident.Name {
							fieldIndex = i
							break
						}
					default:
						return false
					}
				}

				if fieldIndex == -1 {
					return false
				}

				rhs := decl.Rhs[fieldIndex]
				if hasExpressionBuilder(pkgName, rhs) {
					return true
				}
			}

			return false
		default:
			pepperlint.Log("TODO: dynamodb_expression_rules.UsingExpressionsPackage %T", fun)
		}
	default:
		pepperlint.Log("TODO: dynamodb_expression_rules.UsingExpressionsPackage %T", t)
	}

	return false
}

// CopyRule returns a new copy of the rule
func (r DynamoDBExpressionRule) CopyRule() pepperlint.Rule {
	return &DynamoDBExpressionRule{}
}

// WithFileSet sets the rule's file set
func (r *DynamoDBExpressionRule) WithFileSet(fset *token.FileSet) {
	r.fset = fset
}

// getExpressoinsPkgName will attempt to get the package name of the dynamodb.expressions
// package. This needs to be done due to local overrides of package names. If a dynamodb import
// path could not be found in the pepperlint.File, then false will be returned.
func getExpressionsPkgName(f *pepperlint.File) (string, bool) {
	foundImport := false
	pkgName := ""
	for imp, path := range f.Imports {
		if path == dynamodbExpressionImport {
			foundImport = true
			pkgName = imp
			break
		}
	}

	return pkgName, foundImport
}

// hasExpressionBuilder will iterate through the expr AST and check to see if
// an expression builder is within it.
func hasExpressionBuilder(pkgName string, expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.CallExpr:
		return hasExpressionBuilder(pkgName, t.Fun)
	case *ast.SelectorExpr:
		ident, ok := t.X.(*ast.Ident)
		if !ok {
			return hasExpressionBuilder(pkgName, t.X)
		}

		if ident.Name != pkgName {
			return false
		}

		if t.Sel.Name != "NewBuilder" {
			return false
		}

		// This means that an expression.NewBuilder was called meaning
		// that is being used to generate expressions.
		return true
	default:
		pepperlint.Log("TODO: dynamodb_expression_rules.hasExpressionBuilder %T", t)
	}

	return false
}

func init() {
	rules.Add("aws/dynamodb", &DynamoDBExpressionRule{})
}
