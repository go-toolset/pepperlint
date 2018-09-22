package core

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/go-toolset/pepperlint"
)

// DeprecatedFieldRule will check usage of a deprecated field by field name.
type DeprecatedFieldRule struct {
	fset           *token.FileSet
	currentPkgName string
	helper         pepperlint.Helper
}

type fieldInfo struct {
	Name  string
	Spec  *ast.TypeSpec
	Elts  []ast.Expr
	LHS   ast.Expr
	RHS   ast.Expr
	Field *ast.Field
}

type fieldInfos []fieldInfo

// GetByVarName will iterate through all the fieldInfos and find the info associated
// with the given var name. If one cannot be found, false will be returned
func (f fieldInfos) GetByVarName(name string) (fieldInfo, bool) {
	for _, info := range f {
		if info.Name == name {
			return info, true
		}

	}

	return fieldInfo{}, false
}

func (f fieldInfos) GetByField() (fieldInfo, bool) {
	// weird case here where instead of info being a list of infos
	// a single info will be returned due to it being an accessor in
	// a method.
	if len(f) == 1 && f[0].Field != nil {
		return f[0], true
	}

	return fieldInfo{}, false
}

// NewDeprecatedFieldRule will return a new deprecation rule for fields.
// Any field is used that is marked with the deprecated comment will emit
// an error.
func NewDeprecatedFieldRule(fset *token.FileSet) *DeprecatedFieldRule {
	return &DeprecatedFieldRule{
		fset: fset,
	}
}

func (r DeprecatedFieldRule) isDeprecatedField(expr *ast.SelectorExpr) error {
	ident, ok := expr.X.(*ast.Ident)
	if !ok {
		return nil
	}

	infos := r.getFieldInfo(expr.X)
	if len(infos) == 0 {
		return nil
	}

	info, ok := infos.GetByVarName(ident.Name)
	if !ok {
		// Check to see if the info is a field being accessed in object's method
		if info, ok = infos.GetByField(); !ok {
			return nil
		}
	}

	depField := r.helper.GetFieldByName(expr.Sel.Name, info.Spec)
	if depField == nil {
		return nil
	}

	if hasDeprecatedComment(depField.Doc) {
		return pepperlint.NewErrorWrap(r.fset, expr.Sel, fmt.Sprintf("deprecated %q field usage", expr.Sel.Name))
	}
	return nil
}

// checkBinaryExprFields will iterate through the binary expr lhs and rhs to determine
// if deprecated fields are being used
func (r DeprecatedFieldRule) checkBinaryExprFields(bexpr *ast.BinaryExpr) []error {
	errs := []error{}
	exprs := []ast.Expr{bexpr.X, bexpr.Y}

	// check each expression if it is a selector expr. If it is a selector expr
	// it may be using a field that may be deprecated
	for _, expr := range exprs {
		switch exprType := expr.(type) {
		case *ast.SelectorExpr:
			ident, ok := exprType.X.(*ast.Ident)
			if !ok {
				continue
			}

			infos := r.getFieldInfo(exprType.X)
			if len(infos) == 0 {
				continue
			}

			info, ok := infos.GetByVarName(ident.Name)
			if !ok {
				continue
			}

			depField := r.helper.GetFieldByName(exprType.Sel.Name, info.Spec)
			if depField == nil {
				continue
			}

			if hasDeprecatedComment(depField.Doc) {
				errs = append(errs, pepperlint.NewErrorWrap(r.fset, exprType.Sel, fmt.Sprintf("deprecated %q field usage", exprType.Sel.Name)))
			}
		case *ast.BinaryExpr:
			if es := r.checkBinaryExprFields(exprType); len(es) > 0 {
				errs = append(errs, es...)
			}
		case *ast.CallExpr:
			if err := r.ValidateCallExpr(exprType); err != nil {
				berr := err.(*pepperlint.BatchError)
				errs = append(errs, berr.Errors()...)
			}
		default:
			pepperlint.Log("TODO: checkBinaryExprFields %T", exprType)
		}
	}

	return errs
}

// getFieldInfo will inspect the expr and attempt to pull out necessary information
// about the expr to ensure no deprecated field is used.
func (r DeprecatedFieldRule) getFieldInfo(expr ast.Expr) fieldInfos {
	switch t := expr.(type) {
	case *ast.Ident:
		if t.Obj == nil || t.Obj.Decl == nil {
			return nil
		}

		return r.getFieldInfoFromDecl(t.Obj.Decl)
	default:
		pepperlint.Log("TODO: getInternalTypeSpec %T", t)
	}

	return nil
}

func (r DeprecatedFieldRule) getFieldInfoFromDecl(decl interface{}) fieldInfos {
	infos := fieldInfos{}

	switch t := decl.(type) {
	case *ast.AssignStmt:

		for i := 0; i < len(t.Lhs) && i < len(t.Rhs); i++ {
			info := fieldInfo{}
			populated := false

			switch lhs := t.Lhs[i].(type) {
			case *ast.SelectorExpr:
				info.LHS = lhs
				populated = true
			case *ast.Ident:
				// Store variable names incase it is needed later to determine which
				// TypeSpec to grab.
				info.Name = lhs.Name
			case *ast.IndexExpr:
				info.LHS = lhs
				populated = true
			default:
				pepperlint.Log("TODO: lhs.getTypeSpecFromDecl %T", lhs)
			}

			switch rhs := t.Rhs[i].(type) {
			case *ast.CompositeLit:
				switch rhsType := rhs.Type.(type) {
				// RHS is a imported package
				case *ast.SelectorExpr:
					spec := r.helper.GetTypeSpec(rhsType)
					info.Spec = spec
					info.Elts = rhs.Elts
					populated = true
				case *ast.Ident:
					if rhsType.Obj == nil || rhsType.Obj.Decl == nil {
						break
					}

					info.Elts = rhs.Elts
					populated = true

					var ok bool
					if info.Spec, ok = rhsType.Obj.Decl.(*ast.TypeSpec); !ok {
						break
					}

				default:
					pepperlint.Log("TODO: rhsType.inner.getTypeSpecFromDecl %T", rhsType)
				}
			case *ast.CallExpr:
				info.RHS = rhs
				populated = true
			case *ast.IndexExpr:
				info.RHS = rhs
				populated = true
			default:
				pepperlint.Log("TODO: rhsType.getTypeSpecFromDecl %T", rhs)
			}

			if populated {
				infos = append(infos, info)
			}
		}
	case *ast.Field:
		infos = append(infos, fieldInfo{
			Spec:  r.helper.GetTypeSpec(t.Type),
			Field: t,
		})
	default:
		pepperlint.Log("TODO: getTypeSpecFromDecl %T", t)
	}

	return infos
}

// ValidateAssignStmt will check to see if a deprecated field is being set or used.
func (r DeprecatedFieldRule) ValidateAssignStmt(stmt *ast.AssignStmt) error {
	batchError := pepperlint.NewBatchError()

	infos := r.getFieldInfoFromDecl(stmt)
	if len(infos) == 0 {
		return nil
	}

	for _, info := range infos {

		// Check LHS field usage during assignment
		if info.LHS != nil {
			switch lhsType := info.LHS.(type) {
			case *ast.SelectorExpr:
				if err := r.isDeprecatedField(lhsType); err != nil {
					batchError.Add(err)
				}
			case *ast.IndexExpr:
				switch t := lhsType.X.(type) {
				case *ast.SelectorExpr:
					if err := r.isDeprecatedField(t); err != nil {
						batchError.Add(err)
					}
				default:
					pepperlint.Log("TODO: deprecated_field_rule.ValidateAssignStmt %T", t)
				}
			}
		}

		if info.RHS != nil {
			switch rhsType := info.RHS.(type) {
			case *ast.CallExpr:
				if err := r.ValidateCallExpr(rhsType); err != nil {
					berr := err.(*pepperlint.BatchError)
					batchError.Add(berr.Errors()...)
				}
			case *ast.SelectorExpr:
				if err := r.isDeprecatedField(rhsType); err != nil {
					batchError.Add(err)
				}
			case *ast.IndexExpr:
				switch t := rhsType.X.(type) {
				case *ast.SelectorExpr:
					if err := r.isDeprecatedField(t); err != nil {
						batchError.Add(err)
					}
				default:
					pepperlint.Log("TODO: deprecated_field_rule.ValidateAssignStmt %T", t)
				}
			}
		}

		// This function will only care about fields being set, which is the Elts field.
		if info.Spec == nil || !r.helper.IsStruct(info.Spec.Type) {
			continue
		}

		// iterate through elts, if there are any, that will be checked to see if
		// any field declared either through KeyValue or Index will return the appropriate
		// error.
		for i, elt := range info.Elts {
			switch t := elt.(type) {
			// Occurs when struct initializing with field names
			case *ast.KeyValueExpr:
				switch keyType := t.Key.(type) {
				case *ast.Ident:
					depField := r.helper.GetFieldByName(keyType.Name, info.Spec)
					if depField == nil {
						continue
					}

					if hasDeprecatedComment(depField.Doc) {
						batchError.Add(pepperlint.NewErrorWrap(r.fset, elt, fmt.Sprintf("deprecated %q field usage", keyType.Name)))
					}
				}

			case *ast.BasicLit, *ast.CompositeLit:
				depField := r.helper.GetFieldByIndex(i, info.Spec)
				if depField == nil {
					continue
				}

				if hasDeprecatedComment(depField.Doc) {
					batchError.Add(pepperlint.NewErrorWrap(r.fset, elt, fmt.Sprintf("deprecated %v field usage", depField.Names)))
				}

			default:
				pepperlint.Log("TODO: ValidateAssignStmt %T", t)
			}
		}
	}

	return batchError.Return()
}

// ValidateCallExpr will ensure that the deprecated field is not being passed
// as a parameter to a function or method.
func (r DeprecatedFieldRule) ValidateCallExpr(expr *ast.CallExpr) error {
	batchError := pepperlint.NewBatchError()

	for _, arg := range expr.Args {

		switch t := arg.(type) {
		case *ast.SelectorExpr:
			if err := r.isDeprecatedField(t); err != nil {
				batchError.Add(err)
			}
		default:
			pepperlint.Log("TODO: dep_field_rule.ValidateCallExpr: %T %v", t, t)
		}
	}

	return batchError.Return()
}

// ValidateReturnStmt will ensure that the deprecated field is not being returned
// by any method or function.
func (r DeprecatedFieldRule) ValidateReturnStmt(stmt *ast.ReturnStmt) error {
	batchError := pepperlint.NewBatchError()

	for _, result := range stmt.Results {
		switch t := result.(type) {
		case *ast.SelectorExpr:
			if err := r.isDeprecatedField(t); err != nil {
				batchError.Add(err)
			}
		default:
			pepperlint.Log("TODO: deprecated_field_rule.ValidateReturnStmt %T", t)
		}
	}

	return batchError.Return()
}

// ValidatePackage will set the current package name to the package that is currently
// being visited.
func (r *DeprecatedFieldRule) ValidatePackage(pkg *ast.Package) error {
	r.currentPkgName = pkg.Name
	return nil
}

// ValidateIncDecStmt will ensure that deprecated fields that utilize ++ or -- will
// return an error.
func (r DeprecatedFieldRule) ValidateIncDecStmt(stmt *ast.IncDecStmt) error {
	batchError := pepperlint.NewBatchError()

	switch t := stmt.X.(type) {
	case *ast.SelectorExpr:
		if err := r.isDeprecatedField(t); err != nil {
			batchError.Add(err)
		}
	case *ast.IndexExpr:
		switch expr := t.X.(type) {
		case *ast.SelectorExpr:
			if err := r.isDeprecatedField(expr); err != nil {
				batchError.Add(err)
			}
		}
	default:
		pepperlint.Log("TODO: ValidateIncDecExpr %T", t)
	}

	return batchError.Return()
}

// ValidateBinaryExpr will ensure that neither the LHS or RHS of the expr uses a
// deprecated field
func (r DeprecatedFieldRule) ValidateBinaryExpr(expr *ast.BinaryExpr) error {
	batchError := pepperlint.NewBatchError()

	if errs := r.checkBinaryExprFields(expr); len(errs) > 0 {
		batchError.Add(errs...)
	}

	return batchError.Return()
}

// ValidateRangeStmt will ensure that deprecated fields are not used within a
// range statement.
func (r DeprecatedFieldRule) ValidateRangeStmt(expr *ast.RangeStmt) error {
	batchError := pepperlint.NewBatchError()

	// TODO Also check if key or value is being assigned to a deprecated field
	switch t := expr.X.(type) {
	case *ast.SelectorExpr:
		if err := r.isDeprecatedField(t); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Return()
}

// AddRules will add the DeprecatedFieldRule to the given visitor
func (r *DeprecatedFieldRule) AddRules(visitorRules *pepperlint.Rules) {
	rules := pepperlint.Rules{
		AssignStmtRules: pepperlint.AssignStmtRules{r},
		BinaryExprRules: pepperlint.BinaryExprRules{r},
		CallExprRules:   pepperlint.CallExprRules{r},
		PackageRules:    pepperlint.PackageRules{r},
		IncDecStmtRules: pepperlint.IncDecStmtRules{r},
		ReturnStmtRules: pepperlint.ReturnStmtRules{r},
		RangeStmtRules:  pepperlint.RangeStmtRules{r},
	}

	visitorRules.Merge(rules)
}

// WithCache .
func (r *DeprecatedFieldRule) WithCache(cache *pepperlint.Cache) {
	r.helper = pepperlint.NewHelper(cache)
}
