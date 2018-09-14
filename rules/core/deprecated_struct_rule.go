package core

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/go-toolset/pepperlint"
)

// DeprecatedStructRule will validate that no deprecated struct
// is used.
type DeprecatedStructRule struct {
	fset           *token.FileSet
	currentPkgName string
}

// NewDeprecatedStructRule return a newly instantiated DeprecatedStructRule
// with the given file set.
func NewDeprecatedStructRule(fset *token.FileSet) *DeprecatedStructRule {
	return &DeprecatedStructRule{
		fset: fset,
	}
}

// isCompositeLitDeprecated will return whether or not a composite literal type has been
// deprecated. The first parameter will be used to get the line number from the fileset if
// an error occurred
func (r DeprecatedStructRule) isCompositeLitDeprecated(node ast.Node, lit *ast.CompositeLit) error {
	switch t := lit.Type.(type) {
	case *ast.Ident:
		return r.isIdentDeprecated(node, t)
	// This case checks struct use of imported packages
	case *ast.SelectorExpr:
		if err := r.isSelectorExprDeprecated(node, t); err != nil {
			return err
		}
	}

	return nil
}

// isSelectorExprDeprecated will see if a package.struct has been deprecated by getting the struct
// info that was cached during the ValidateGenDecl call.
func (r DeprecatedStructRule) isSelectorExprDeprecated(node ast.Node, expr *ast.SelectorExpr) error {
	ident, ok := expr.X.(*ast.Ident)
	if !ok {
		return nil
	}

	info := pepperlint.PackagesCache.TypeInfos[ident.Name][expr.Sel.Name]
	if hasDeprecatedComment(info.Doc) {
		return pepperlint.NewErrorWrap(r.fset, node, fmt.Sprintf("deprecated '%s.%s' struct used", ident.Name, expr.Sel.Name))
	}

	return nil
}

// isIdentDeprecated will return an error upon finding a use of a deprecated structure.
// The node parameter is used to determine a line number if an error occurred.
func (r DeprecatedStructRule) isIdentDeprecated(node ast.Node, ident *ast.Ident) error {
	var spec *ast.TypeSpec

	if ident.Obj == nil {
		// not an object meaning we don't care to evaluate whether or not
		// it is a deprecated shape.
		return nil
	}

	switch decl := ident.Obj.Decl.(type) {
	// decl will be an *ast.TypeSpec if it is being passed in as a parameter
	case *ast.TypeSpec:
		spec = decl

	// decl will be an *ast.Field if the object being returned
	// is a struct.
	case *ast.Field:

		switch t := decl.Type.(type) {
		// occurs when deprecated struct is being returned directly
		case *ast.Ident:
			return r.isIdentDeprecated(node, t)

		// occurs when pointer to deprecated struct is being returned
		case *ast.StarExpr:
			id, ok := t.X.(*ast.Ident)
			if !ok {
				// expr is not a structure
				return nil
			}

			return r.isIdentDeprecated(node, id)
		default:
			return nil
		}
	default:
		return nil
	}

	info := pepperlint.PackagesCache.TypeInfos[r.currentPkgName][spec.Name.Name]
	if hasDeprecatedComment(info.Doc) {
		return pepperlint.NewErrorWrap(r.fset, node, fmt.Sprintf("deprecated %q struct used", spec.Name.Name))
	}

	return nil
}

// ValidateAssignStmt will take a look to see if a deprecated structure is
// being assigned to a given variable.
func (r DeprecatedStructRule) ValidateAssignStmt(stmt *ast.AssignStmt) error {
	batchError := pepperlint.NewBatchError()

	// check to see if any struct initialized objects contain any deprecated field
	// being used
	for _, rhs := range stmt.Rhs {
		if errs := r.validateAssignStmt(rhs, rhs); len(errs) > 0 {
			batchError.Add(errs...)
		}
	}

	return batchError.Return()
}

func (r DeprecatedStructRule) validateAssignStmt(expr ast.Expr, rhs ast.Expr) []error {
	switch t := rhs.(type) {
	case *ast.CompositeLit:
		if errs := r.deprecatedStructUsage(expr, t.Type); len(errs) > 0 {
			return errs
		}

	// occurs when function has a return value and is being used in an
	// assignment statement
	case *ast.CallExpr:
		if err := r.ValidateCallExpr(t); err != nil {
			batchError := err.(*pepperlint.BatchError)
			return batchError.Errors()
		}
	default:
		pepperlint.Log("TODO: RHS DeprecatedStructRule.ValidateAssignStmt %T %v\n", t, t)
	}

	return nil
}

// deprecatedStructUsage will determine if a struct is being used in an assignment
// statement. The RHS parameter is used to grab the line number if an error has
// occurred.
func (r DeprecatedStructRule) deprecatedStructUsage(rhs ast.Expr, expr ast.Expr) []error {
	errs := []error{}
	switch tstruct := expr.(type) {
	case *ast.Ident:
		if tstruct.Obj == nil {
			return errs
		}

		if tstruct.Obj.Decl == nil {
			return errs
		}

		decl, ok := tstruct.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return errs
		}

		info := pepperlint.PackagesCache.TypeInfos[r.currentPkgName][decl.Name.Name]
		if hasDeprecatedComment(info.Doc) {
			errs = append(errs, pepperlint.NewErrorWrap(r.fset, rhs, fmt.Sprintf("deprecated %q struct used", decl.Name.Name)))
		} else if es := r.checkTypeAliases(rhs, decl.Type); len(es) > 0 {
			errs = append(errs, es...)
		}

	// this case will happen when a imported package structure is being used.
	// The AST will not contain a type spec and only *ast.Ident with no object
	// tied to it.
	case *ast.SelectorExpr:
		if err := r.isSelectorExprDeprecated(rhs, tstruct); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// checkTypeAliases will ensure that a type alias is not type aliasing
// a deprecated structure.
func (r DeprecatedStructRule) checkTypeAliases(rhs ast.Node, expr ast.Expr) []error {
	errs := []error{}

	switch t := expr.(type) {
	case *ast.Ident:
		spec, ok := t.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return errs
		}

		info := pepperlint.PackagesCache.TypeInfos[r.currentPkgName][spec.Name.Name]
		if hasDeprecatedComment(info.Doc) {
			errs = append(errs, pepperlint.NewErrorWrap(r.fset, rhs, fmt.Sprintf("deprecated %q struct used", spec.Name.Name)))
		} else {
			errs = append(errs, r.checkTypeAliases(rhs, spec.Type)...)
		}
	// This case occurs when using an imported structure
	case *ast.SelectorExpr:
		if err := r.isSelectorExprDeprecated(rhs, t); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// ValidateCallExpr ...
func (r DeprecatedStructRule) ValidateCallExpr(expr *ast.CallExpr) error {
	batchError := pepperlint.NewBatchError()

	for _, arg := range expr.Args {
		if errs := r.validateCallExpr(expr, arg); len(errs) > 0 {
			batchError.Add(errs...)
		}
	}

	return batchError.Return()
}

func (r DeprecatedStructRule) validateCallExpr(node ast.Node, arg ast.Expr) []error {
	errs := []error{}

	switch t := arg.(type) {
	case *ast.CompositeLit:
		if err := r.isCompositeLitDeprecated(arg, t); err != nil {
			errs = append(errs, err)
		}
	case *ast.UnaryExpr:
		if es := r.validateCallExpr(arg, t.X); len(es) > 0 {
			errs = append(errs, es...)
		}
	case *ast.Ident:
		if t.Obj == nil {
			return errs
		}

		stmt, ok := t.Obj.Decl.(*ast.AssignStmt)
		if !ok {
			return errs
		}

		for _, rhs := range stmt.Rhs {
			if es := r.validateAssignStmt(arg, rhs); len(es) > 0 {
				errs = append(errs, es...)
			}
		}
	default:
		pepperlint.Log("TODO: dep_struct_rule.ValidateCallExpr: %T\n", t)
	}

	return errs
}

// ValidateReturnStmt will validate that return items are not deprecated structures.
func (r DeprecatedStructRule) ValidateReturnStmt(stmt *ast.ReturnStmt) error {
	batchError := pepperlint.NewBatchError()

	for _, result := range stmt.Results {
		switch t := result.(type) {
		case *ast.CompositeLit:
			if err := r.isCompositeLitDeprecated(result, t); err != nil {
				batchError.Add(err)
			}
		case *ast.Ident:
			if err := r.isIdentDeprecated(result, t); err != nil {
				batchError.Add(err)
			}

		default:
			pepperlint.Log("DeprecatedStructRule.ValidateReturnStmt: %T %v\n", t, t)
		}
	}

	return batchError.Return()
}

// ValidateFuncDecl will validate function declaractions and ensure no
// deprecated structure is being used
func (r DeprecatedStructRule) ValidateFuncDecl(decl *ast.FuncDecl) error {
	batchError := pepperlint.NewBatchError()

	for _, param := range decl.Type.Params.List {
		switch t := param.Type.(type) {
		case *ast.Ident:
			if err := r.isIdentDeprecated(param, t); err != nil {
				batchError.Add(err)
			}

		case *ast.StarExpr:
			id, ok := t.X.(*ast.Ident)
			if !ok {
				// expr is not a structure
				continue
			}

			if err := r.isIdentDeprecated(param, id); err != nil {
				batchError.Add(err)
			}

		// This cases is to see if an imported deprecated struct
		// is being used
		case *ast.SelectorExpr:
			if err := r.isSelectorExprDeprecated(param, t); err != nil {
				batchError.Add(err)
			}
		default:
			pepperlint.Log("TODO: ValidateFuncDecl.Param %T\n", t)
		}
	}

	// If there are no results, that means there is no
	// result being returned. Which doesn't allow for
	// the possibility of a deprecated struct to be
	// returned.
	if decl.Type.Results == nil {
		return batchError.Return()
	}

	for _, result := range decl.Type.Results.List {
		switch t := result.Type.(type) {
		case *ast.Ident:
			if err := r.isIdentDeprecated(result, t); err != nil {
				batchError.Add(err)
			}

		// Pointer case
		case *ast.StarExpr:
			id, ok := t.X.(*ast.Ident)
			if !ok {
				// expr is not a structure
				continue
			}

			if err := r.isIdentDeprecated(result, id); err != nil {
				batchError.Add(err)
			}

		case *ast.SelectorExpr:
			if err := r.isSelectorExprDeprecated(result, t); err != nil {
				batchError.Add(err)
			}
		default:
			pepperlint.Log("TODO: ValidateFuncDecl.Result %T\n", t)
		}
	}

	return batchError.Return()
}

// ValidatePackage will set the current package name to the package that is currently
// being visited.
func (r *DeprecatedStructRule) ValidatePackage(pkg *ast.Package) error {
	r.currentPkgName = pkg.Name
	return nil
}

// ValidateTypeSpec ...
func (r DeprecatedStructRule) ValidateTypeSpec(spec *ast.TypeSpec) error {
	batchError := pepperlint.NewBatchError()

	switch t := spec.Type.(type) {
	case *ast.SelectorExpr:
		if err := r.isSelectorExprDeprecated(spec, t); err != nil {
			batchError.Add(err)
		}
	default:
		pepperlint.Log("TODO: DeprecatedStructRule.ValidateTypeSpec %T", t)
	}
	return batchError.Return()
}

// AddRules will add the DeprecatedStructRule to the given visitor
func (r *DeprecatedStructRule) AddRules(v *pepperlint.Visitor) {
	rules := pepperlint.Rules{
		AssignStmtRules: pepperlint.AssignStmtRules{r},
		CallExprRules:   pepperlint.CallExprRules{r},
		ReturnStmtRules: pepperlint.ReturnStmtRules{r},
		FuncDeclRules:   pepperlint.FuncDeclRules{r},
		PackageRules:    pepperlint.PackageRules{r},
		TypeSpecRules:   pepperlint.TypeSpecRules{r},
	}

	v.Rules.Merge(rules)
}
