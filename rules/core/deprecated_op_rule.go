package core

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/go-toolset/pepperlint"
)

// DeprecatedOpRule is used to walk and determine if an operation, whether function
// or method, being used is deprecated. If it is a deprecated operation, the appropriate
// error will be returned.
type DeprecatedOpRule struct {
	fset           *token.FileSet
	currentPkgName string
	helper         pepperlint.Helper
}

// NewDeprecatedOpRule returns a new DeprecatedOpRule with the given file set.
func NewDeprecatedOpRule(fset *token.FileSet) *DeprecatedOpRule {
	return &DeprecatedOpRule{
		fset: fset,
	}
}

// isIdentDeprecated will take a look at the given ident and determine
// whether or not it is an imported operation. With that, it'll get the
// operation definition and this will take a look at the documentation
// to determine if it is a deprecated operation.
func (r *DeprecatedOpRule) isIdentDeprecated(ident *ast.Ident) error {
	var op *ast.FuncDecl
	var ok bool

	if op, ok = r.getExternalPackageOp(ident); ok {
	} else if op, ok = r.getInternalPackageOp(ident); !ok {
		return nil
	}

	if hasDeprecatedComment(op.Doc) {
		return pepperlint.NewErrorWrap(r.fset, ident, fmt.Sprintf("deprecated %q operation used", op.Name.Name))
	}

	return nil
}

func (r *DeprecatedOpRule) getExternalPackageOp(ident *ast.Ident) (*ast.FuncDecl, bool) {
	return nil, false
}

// getInternalPackageOp will walk the nested parameters of the ident attempting to
// grab the obj declaration. If the object declaration is of an *ast.FuncDecl, it
// will the newly found object along with true.
func (r *DeprecatedOpRule) getInternalPackageOp(ident *ast.Ident) (*ast.FuncDecl, bool) {
	if ident.Obj == nil {
		return nil, false
	}

	if ident.Obj.Decl == nil {
		return nil, false
	}

	switch t := ident.Obj.Decl.(type) {
	case *ast.FuncDecl:
		return t, true
	}

	return nil, false
}

func (r *DeprecatedOpRule) getExternalPackageType(expr ast.Expr) ([]pepperlint.TypeInfo, bool) {
	infos := []pepperlint.TypeInfo{}

	switch t := expr.(type) {
	case *ast.Ident:
		// if t.Obj is nil, this implies that this could be a function instead of a method
		if t.Obj == nil {
			infos = append(infos, pepperlint.TypeInfo{
				PkgName: t.Name,
			})
			return infos, true
		}

		if t.Obj.Decl == nil {
			return nil, false
		}

		// attempt to grab the typespecs
		switch decl := t.Obj.Decl.(type) {
		// assign statements are checked because we are investigating
		// a declaration.
		case *ast.AssignStmt:
			for _, rhs := range decl.Rhs {
				info, ok := r.getExternalTypeSpec(rhs)
				if !ok {
					continue
				}

				infos = append(infos, info)
			}
		default:
			pepperlint.Log("TODO: deprecated_op_rule.getInternalPackageType %T", decl)
		}
	}

	return infos, len(infos) > 0
}

func (r *DeprecatedOpRule) getExternalTypeSpec(rhs ast.Expr) (pepperlint.TypeInfo, bool) {
	switch expr := rhs.(type) {
	case *ast.CompositeLit:
		exprType, ok := expr.Type.(*ast.SelectorExpr)
		if !ok {
			return pepperlint.TypeInfo{}, false
		}

		ident, ok := exprType.X.(*ast.Ident)
		pkgName := ident.Name
		typeName := exprType.Sel.Name
		file, ok := r.helper.PackagesCache.CurrentFile()
		if !ok {
			panic("CurrentFile was not set")
		}

		importPath := file.Imports[pkgName]
		pkg, ok := r.helper.PackagesCache.Packages.Get(importPath)
		if !ok {
			// this occurs when an import from the standard library is called or any
			// other import that wasn't included in the config or cli by using -include-pkgs
			return pepperlint.TypeInfo{}, false
		}

		info, ok := pkg.Files.GetTypeInfo(typeName)
		if !ok {
			return pepperlint.TypeInfo{}, false
		}

		return info, ok
	case *ast.UnaryExpr:
		return r.getExternalTypeSpec(expr.X)
	default:
		pepperlint.Log("TODO: deprecated_op_rule.getInternalPackageType %T %v", expr, expr)
	}

	return pepperlint.TypeInfo{}, false
}

// getInternalPackageType will get associated type specs associated with the given
// expression. The second parameter will be true if any type spec was found.
func (r *DeprecatedOpRule) getInternalPackageType(expr ast.Expr) ([]pepperlint.TypeInfo, bool) {
	infos := []pepperlint.TypeInfo{}

	switch t := expr.(type) {
	case *ast.Ident:
		if t.Obj == nil {
			return nil, false
		}

		if t.Obj.Decl == nil {
			return nil, false
		}

		// attempt to grab the typespecs
		switch decl := t.Obj.Decl.(type) {
		// assign statements are checked because we are investigating
		// a declaration.
		case *ast.AssignStmt:
			for _, rhs := range decl.Rhs {
				spec, ok := r.getInternalTypeSpec(rhs)
				if !ok {
					continue
				}

				infos = append(infos, pepperlint.TypeInfo{
					Spec: spec,
				})
			}
		default:
			pepperlint.Log("TODO: deprecated_op_rule.getInternalPackageType %T", decl)
		}
	}

	return infos, len(infos) > 0
}

func (r *DeprecatedOpRule) getInternalTypeSpec(rhs ast.Expr) (*ast.TypeSpec, bool) {
	switch expr := rhs.(type) {
	case *ast.CompositeLit:
		exprType, ok := expr.Type.(*ast.Ident)
		if !ok {
			return nil, false
		}

		if exprType.Obj == nil {
			return nil, false
		}

		if exprType.Obj.Decl == nil {
			return nil, false
		}

		spec, ok := exprType.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return nil, false
		}

		return spec, true

	case *ast.UnaryExpr:
		return r.getInternalTypeSpec(expr.X)
	default:
		pepperlint.Log("TODO: deprecated_op_rule.getInternalPackageType %T %v", expr, expr)
	}

	return nil, false
}

// isSelectorExprDeprecated will take a look at a selector expression and grab the
// type spec off of the selector expression's X field. The type spec will be used
// to determine whether or not that it is apart of the operation found via method name
// and current package name.
func (r *DeprecatedOpRule) isSelectorExprDeprecated(sel *ast.SelectorExpr) []error {
	methodName := sel.Sel.Name

	var infos []pepperlint.TypeInfo
	var ok bool
	errs := []error{}

	externalPkg := false
	if infos, ok = r.getExternalPackageType(sel.X); ok {
		externalPkg = true
	} else if infos, ok = r.getInternalPackageType(sel.X); !ok {
		return nil
	}

	for _, info := range infos {
		spec := info.Spec
		var pkg *pepperlint.Package
		found := false

		if externalPkg {
			file, ok := r.helper.PackagesCache.CurrentFile()
			if !ok {
				panic("CurrentFile was not properly set")
			}

			pkgImportPath := file.Imports[info.PkgName]
			pkg, found = r.helper.PackagesCache.Packages.Get(pkgImportPath)
		} else {
			pkg, found = r.helper.PackagesCache.CurrentPackage()
		}

		// Potentially received a package we didn't crawl for the cache
		if !found {
			continue
		}

		opInfo, ok := pkg.Files.GetOpInfo(methodName)
		if !ok {
			continue
		}

		// Check to see if the type has the associated method name
		if spec != nil && !opInfo.HasReceiverType(spec) {
			continue
		}

		if hasDeprecatedComment(opInfo.Decl.Doc) {
			errs = append(errs, pepperlint.NewErrorWrap(
				r.fset,
				sel,
				fmt.Sprintf("deprecated '%s.%s' struct used", r.currentPkgName, opInfo.Decl.Name),
			))
		}
	}

	return errs
}

// ValidateAssignStmt will determine whether or not the RHS of an assignment expression
// contains any deprecated operation.
func (r *DeprecatedOpRule) ValidateAssignStmt(stmt *ast.AssignStmt) error {
	batchError := pepperlint.NewBatchError()

	for _, rhs := range stmt.Rhs {
		switch t := rhs.(type) {
		case *ast.CallExpr:
			if err := r.ValidateCallExpr(t); err != nil {
				if be, ok := err.(*pepperlint.BatchError); ok {
					batchError.Add(be.Errors()...)
				} else {
					batchError.Add(err)
				}
			}
		}
	}
	return batchError.Return()
}

// ValidateCallExpr will determine if the operation in the CallExpr is deprecated.
func (r *DeprecatedOpRule) ValidateCallExpr(expr *ast.CallExpr) error {
	batchError := pepperlint.NewBatchError()

	switch fun := expr.Fun.(type) {
	case *ast.Ident:
		if err := r.isIdentDeprecated(fun); err != nil {
			batchError.Add(err)
		}
	case *ast.SelectorExpr:
		if errs := r.isSelectorExprDeprecated(fun); len(errs) > 0 {
			batchError.Add(errs...)
		}
	default:
		pepperlint.Log("TODO: deprecated_op_rule.ValidateCallExpr %T", fun)
	}

	return batchError.Return()
}

// ValidatePackage is used to keep track of the current package scope that is
// being traversed.
func (r *DeprecatedOpRule) ValidatePackage(pkg *ast.Package) error {
	r.currentPkgName = pkg.Name
	return nil
}

// AddRules will add the DeprecatedFieldRule to the given visitor
func (r *DeprecatedOpRule) AddRules(visitorRules *pepperlint.Rules) {
	rules := pepperlint.Rules{
		AssignStmtRules: pepperlint.AssignStmtRules{r},
		CallExprRules:   pepperlint.CallExprRules{r},
		PackageRules:    pepperlint.PackageRules{r},
	}

	visitorRules.Merge(rules)
}

// WithCache will create a new helper with the given cache. This is used
// to determine infomation about a specific ast.Node.
func (r *DeprecatedOpRule) WithCache(cache *pepperlint.Cache) {
	r.helper = pepperlint.NewHelper(cache)
}

// WithFileSet will set the token.FileSet to the rule allowing for more
// in depth errors.
func (r *DeprecatedOpRule) WithFileSet(fset *token.FileSet) {
	r.fset = fset
}
