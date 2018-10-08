package deprecated

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/go-toolset/pepperlint"
)

// StructRule will validate that no deprecated struct
// is used.
type StructRule struct {
	fset           *token.FileSet
	currentPkgName string
	helper         pepperlint.Helper

	// need to keep track of which call expr were visited due to
	// assignment statement also calling ValidateCallExpr.
	visitedCallExpr map[*ast.CallExpr]struct{}
}

// NewStructRule return a newly instantiated StructRule
// with the given file set.
func NewStructRule(fset *token.FileSet) *StructRule {
	return &StructRule{
		fset:            fset,
		visitedCallExpr: map[*ast.CallExpr]struct{}{},
	}
}

// isCompositeLitDeprecated will return whether or not a composite literal type has been
// deprecated. The first parameter will be used to get the line number from the fileset if
// an error occurred
func (r StructRule) isCompositeLitDeprecated(node ast.Node, lit *ast.CompositeLit) []error {
	switch t := lit.Type.(type) {
	case *ast.Ident:
		return r.isIdentDeprecated(node, t)
	// This case checks struct use of imported packages
	case *ast.SelectorExpr:
		if err := r.isSelectorExprDeprecated(node, t); err != nil {
			return []error{err}
		}
	}

	return nil
}

// isSelectorExprDeprecated will see if a package.struct has been deprecated by getting the struct
// info that was cached during the ValidateGenDecl call.
func (r StructRule) isSelectorExprDeprecated(node ast.Node, expr *ast.SelectorExpr) error {
	ident, ok := expr.X.(*ast.Ident)
	if !ok {
		return nil
	}

	file, ok := r.helper.PackagesCache.CurrentFile()
	if !ok {
		panic("CurrentFile was not set")
	}

	pkgImportPath := file.Imports[ident.Name]
	pkg, ok := r.helper.PackagesCache.Packages.Get(pkgImportPath)
	if !ok {
		return nil
	}

	info, ok := pkg.Files.GetTypeInfo(expr.Sel.Name)
	if !ok {
		return nil
	}

	if hasDeprecatedComment(info.Doc) {
		return pepperlint.NewErrorWrap(r.fset, node, fmt.Sprintf("deprecated '%s.%s' struct used", ident.Name, expr.Sel.Name))
	}

	return nil
}

// isIdentDeprecated will return an error upon finding a use of a deprecated structure.
// The node parameter is used to determine a line number if an error occurred.
func (r StructRule) isIdentDeprecated(node ast.Node, ident *ast.Ident) []error {
	var spec *ast.TypeSpec
	errs := []error{}

	if ident.Obj == nil ||
		ident.Obj.Decl == nil {
		// not an object meaning we don't care to evaluate whether or not
		// it is a deprecated shape.
		return errs
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
				return errs
			}

			return r.isIdentDeprecated(node, id)
		default:
			return errs
		}

	// occurrs when return parameter is a variable
	case *ast.AssignStmt:
		for _, rhs := range decl.Rhs {
			switch rhsType := rhs.(type) {
			// pointer to variable
			case *ast.UnaryExpr:
				composite, ok := rhsType.X.(*ast.CompositeLit)
				if !ok {
					continue
				}

				if es := r.isCompositeLitDeprecated(ident, composite); len(es) > 0 {
					errs = append(errs, es...)
				}
			case *ast.CompositeLit:
				if es := r.isCompositeLitDeprecated(ident, rhsType); len(es) > 0 {
					errs = append(errs, es...)
				}
			default:
				pepperlint.Log("TODO: isIdentDeprecated %T", rhsType)
			}

		}

		return errs

	case *ast.ValueSpec:
		if es := r.deprecatedStructUsage(ident, decl.Type); len(es) > 0 {
			errs = append(errs, es...)
		}

		return errs
	default:
		pepperlint.Log("TODO: deprecated_struct_rule.isIdentDeprecated %T", decl)
		return errs
	}

	pkg, ok := r.helper.PackagesCache.CurrentPackage()
	if !ok {
		panic("SHOULD NOT BE HERE")
	}

	info, ok := pkg.Files.GetTypeInfo(spec.Name.Name)
	if !ok {
		return nil
	}

	if hasDeprecatedComment(info.Doc) {
		errs = append(errs, pepperlint.NewErrorWrap(r.fset, node, fmt.Sprintf("deprecated %q struct used", spec.Name.Name)))
	}

	return errs
}

// ValidateAssignStmt will take a look to see if a deprecated structure is
// being assigned to a given variable.
func (r StructRule) ValidateAssignStmt(stmt *ast.AssignStmt) error {
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

func (r StructRule) validateAssignStmt(expr ast.Expr, rhs ast.Expr) []error {
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

	// occurs when referencing a type during assignment
	case *ast.UnaryExpr:
		return r.validateAssignStmt(expr, t.X)
	default:
		pepperlint.Log("TODO: RHS StructRule.ValidateAssignStmt %T %v\n", t, t)
	}

	return nil
}

// deprecatedStructUsage will determine if a struct is being used in an assignment
// statement. The RHS parameter is used to grab the line number if an error has
// occurred.
func (r StructRule) deprecatedStructUsage(rhs ast.Expr, expr ast.Expr) []error {
	errs := []error{}
	switch tstruct := expr.(type) {
	case *ast.Ident:
		if tstruct.Obj == nil {
			return errs
		}

		if tstruct.Obj.Decl == nil {
			return errs
		}

		switch decl := tstruct.Obj.Decl.(type) {
		case *ast.ValueSpec:
			errs = append(errs, r.deprecatedStructUsage(rhs, decl.Type)...)

		case *ast.TypeSpec:

			decl, ok := tstruct.Obj.Decl.(*ast.TypeSpec)
			if !ok {
				return errs
			}

			pkg, ok := r.helper.PackagesCache.CurrentPackage()
			if !ok {
				panic("CurrentPackage was not set")
			}

			info, ok := pkg.Files.GetTypeInfo(decl.Name.Name)
			if !ok {
				return nil
			}

			if hasDeprecatedComment(info.Doc) {
				errs = append(errs, pepperlint.NewErrorWrap(r.fset, rhs, fmt.Sprintf("deprecated %q struct used", decl.Name.Name)))
			} else if es := r.checkTypeAliases(rhs, decl.Type); len(es) > 0 {
				errs = append(errs, es...)
			}
		}

	// this case will happen when a imported package structure is being used.
	// The AST will not contain a type spec and only *ast.Ident with no object
	// tied to it.
	case *ast.SelectorExpr:
		if err := r.isSelectorExprDeprecated(rhs, tstruct); err != nil {
			errs = append(errs, err)
		}
	case *ast.ArrayType:
		if es := r.deprecatedStructUsage(rhs, tstruct.Elt); len(es) > 0 {
			errs = append(errs, es...)
		}
	case *ast.StarExpr:
		if es := r.deprecatedStructUsage(rhs, tstruct.X); len(es) > 0 {
			errs = append(errs, es...)
		}
	default:
		pepperlint.Log("TODO: deprecated_struct_rule.deprecatedStructUsage %T", tstruct)
	}

	return errs
}

// checkTypeAliases will ensure that a type alias is not type aliasing
// a deprecated structure.
func (r StructRule) checkTypeAliases(rhs ast.Node, expr ast.Expr) []error {
	errs := []error{}

	switch t := expr.(type) {
	case *ast.Ident:
		if t.Obj == nil {
			return nil
		}

		if t.Obj.Decl == nil {
			return nil
		}

		spec, ok := t.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return nil
		}

		pkg, ok := r.helper.PackagesCache.CurrentPackage()
		if !ok {
			panic("CurrentPackage was not set")
		}

		info, ok := pkg.Files.GetTypeInfo(spec.Name.Name)
		if !ok {
			return nil
		}

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

// ValidateCallExpr will ensure a deprecated struct is not being passed as a parameter
func (r StructRule) ValidateCallExpr(expr *ast.CallExpr) error {
	batchError := pepperlint.NewBatchError()
	if _, ok := r.visitedCallExpr[expr]; ok {
		return nil
	}

	defer func() {
		r.visitedCallExpr[expr] = struct{}{}
	}()

	for _, arg := range expr.Args {
		if errs := r.validateCallExpr(expr, arg); len(errs) > 0 {
			batchError.Add(errs...)
		}
	}

	return batchError.Return()
}

func (r StructRule) validateCallExpr(node ast.Node, arg ast.Expr) []error {
	errs := []error{}

	switch t := arg.(type) {
	case *ast.CompositeLit:
		if es := r.isCompositeLitDeprecated(arg, t); len(es) > 0 {
			errs = append(errs, es...)
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
func (r StructRule) ValidateReturnStmt(stmt *ast.ReturnStmt) error {
	batchError := pepperlint.NewBatchError()

	for _, result := range stmt.Results {
		switch t := result.(type) {
		case *ast.CompositeLit:
			if es := r.isCompositeLitDeprecated(result, t); len(es) > 0 {
				batchError.Add(es...)
			}
		case *ast.Ident:
			if es := r.isIdentDeprecated(result, t); len(es) > 0 {
				batchError.Add(es...)
			}

		default:
			pepperlint.Log("TODO: StructRule.ValidateReturnStmt: %T", t)
		}
	}

	return batchError.Return()
}

// ValidateFuncDecl will validate function declaractions and ensure no
// deprecated structure is being used
func (r StructRule) ValidateFuncDecl(decl *ast.FuncDecl) error {
	batchError := pepperlint.NewBatchError()

	for _, param := range decl.Type.Params.List {
		switch t := param.Type.(type) {
		case *ast.Ident:
			if es := r.isIdentDeprecated(param, t); len(es) > 0 {
				batchError.Add(es...)
			}

		case *ast.StarExpr:
			id, ok := t.X.(*ast.Ident)
			if !ok {
				// expr is not a structure
				continue
			}

			if es := r.isIdentDeprecated(param, id); len(es) > 0 {
				batchError.Add(es...)
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
			if es := r.isIdentDeprecated(result, t); len(es) > 0 {
				batchError.Add(es...)
			}

		// Pointer case
		case *ast.StarExpr:
			id, ok := t.X.(*ast.Ident)
			if !ok {
				// expr is not a structure
				continue
			}

			if es := r.isIdentDeprecated(result, id); len(es) > 0 {
				batchError.Add(es...)
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
func (r *StructRule) ValidatePackage(pkg *ast.Package) error {
	r.currentPkgName = pkg.Name
	return nil
}

// ValidateTypeSpec will ensure that the type spec's type isn't a deprecated structure.
func (r StructRule) ValidateTypeSpec(spec *ast.TypeSpec) error {
	batchError := pepperlint.NewBatchError()

	switch t := spec.Type.(type) {
	case *ast.SelectorExpr:
		if err := r.isSelectorExprDeprecated(spec, t); err != nil {
			batchError.Add(err)
		}
	default:
		pepperlint.Log("TODO: StructRule.ValidateTypeSpec %T", t)
	}

	return batchError.Return()
}

// ValidateValueSpec will check structures used as values are not deprecated structs.
func (r StructRule) ValidateValueSpec(spec *ast.ValueSpec) error {
	batchError := pepperlint.NewBatchError()

	switch specType := spec.Type.(type) {
	case *ast.Ident:
		if es := r.isIdentDeprecated(spec, specType); len(es) > 0 {
			batchError.Add(es...)
		}
	case *ast.StarExpr:
		id, ok := specType.X.(*ast.Ident)
		if !ok {
			break
		}

		if es := r.isIdentDeprecated(spec, id); len(es) > 0 {
			batchError.Add(es...)
		}
	}

	return batchError.Return()
}

// ValidateBinaryExpr will ensure no deprecated struct is being used on either the LHS
// or RHS of the expression.
func (r StructRule) ValidateBinaryExpr(expr *ast.BinaryExpr) error {
	batchError := pepperlint.NewBatchError()

	if es := r.deprecatedStructUsage(expr.X, expr.X); len(es) > 0 {
		batchError.Add(es...)
	}
	if es := r.deprecatedStructUsage(expr.Y, expr.Y); len(es) > 0 {
		batchError.Add(es...)
	}

	return batchError.Return()
}

// AddRules will add the StructRule to the given visitor
func (r *StructRule) AddRules(visitorRules *pepperlint.Rules) {
	rules := pepperlint.Rules{
		AssignStmtRules: pepperlint.AssignStmtRules{r},
		BinaryExprRules: pepperlint.BinaryExprRules{r},
		CallExprRules:   pepperlint.CallExprRules{r},
		ReturnStmtRules: pepperlint.ReturnStmtRules{r},
		FuncDeclRules:   pepperlint.FuncDeclRules{r},
		PackageRules:    pepperlint.PackageRules{r},
		TypeSpecRules:   pepperlint.TypeSpecRules{r},
		ValueSpecRules:  pepperlint.ValueSpecRules{r},
	}

	visitorRules.Merge(rules)
}

// WithCache will create a new helper with the given cache. This is used
// to determine infomation about a specific ast.Node.
func (r *StructRule) WithCache(cache *pepperlint.Cache) {
	r.helper = pepperlint.NewHelper(cache)
}

// WithFileSet will set the token.FileSet to the rule allowing for more
// in depth errors.
func (r *StructRule) WithFileSet(fset *token.FileSet) {
	r.fset = fset
}
