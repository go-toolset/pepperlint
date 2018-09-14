package core

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/go-toolset/pepperlint"
	"github.com/go-toolset/pepperlint/utils"
)

// DeprecatedFieldRule will check usage of a deprecated field by field name.
type DeprecatedFieldRule struct {
	fset           *token.FileSet
	currentPkgName string
}

// NewDeprecatedFieldRule will return a new deprecation rule for fields.
// Any field is used that is marked with the deprecated comment will emit
// an error.
func NewDeprecatedFieldRule(fset *token.FileSet) *DeprecatedFieldRule {
	return &DeprecatedFieldRule{
		fset: fset,
	}
}

// containsDeprecatedFields will look inside a composite literal and check
// if any fields that contain a deprecated comment will return the number of
// deprecated fields in use
func (r DeprecatedFieldRule) containsDeprecatedFields(node ast.Node, cl *ast.CompositeLit) []error {
	errs := []error{}
	var s *ast.StructType

	switch t := cl.Type.(type) {
	case *ast.Ident:
		s = utils.GetStructType(t)

	// using imported package field
	case *ast.SelectorExpr:
		pkgName := ""
		structName := ""

		// attempt to get the package name from an *ast.Ident
		switch ident := t.X.(type) {
		case *ast.Ident:
			pkgName = ident.Name
		default:
			return nil
		}

		structName = t.Sel.Name

		info, ok := pepperlint.PackagesCache.TypeInfos[pkgName][structName]
		if !ok {
			// TODO: Should probably return regular error like could not find pkg.foo
			return nil
		}

		if info.Spec == nil {
			return nil
		}

		switch st := info.Spec.Type.(type) {
		case *ast.StructType:
			s = st
		default:
			return nil
		}
	default:
		pepperlint.Log("TODO: deprecated_field_rule.containsDeprecatedField %T", t)
	}

	if s == nil {
		return nil
	}

	depFields := getDeprecatedFields(s.Fields)
	for i, elt := range cl.Elts {
		if isDeprecatedField(depFields, i, elt) {
			names := s.Fields.List[i].Names

			structNames := make([]string, len(names))
			for i := 0; i < len(names); i++ {
				structNames[i] = names[i].Name
			}

			lineNumber := ast.Node(elt)
			if node != nil {
				lineNumber = node
			}

			errs = append(errs, pepperlint.NewErrorWrap(r.fset, lineNumber, fmt.Sprintf("deprecated %v field used in struct initialization", structNames)))
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

// isDeprecatedField evaluates whether or not a field is deprecated by looking at
// either the deprecated field being set during struct initialization or by setting
// the index of that deprecated field.
func isDeprecatedField(depFields deprecatedCache, index int, expr ast.Expr) bool {
	switch t := expr.(type) {
	// this case happens when setting the index of the field, ie
	//
	// f := Foo {
	//   123,
	// }
	case *ast.BasicLit:
		// the LHS of this expr should never occur, but may occur due to some bug.
		if index < len(depFields.IndexLookup) && depFields.IndexLookup[index] {
			return true
		}

	// this case occurrs when struct initialization with setting fields by their
	// name, ie
	//
	// f := Foo {
	//   Field: 123,
	// }
	case *ast.KeyValueExpr:
		switch keyType := t.Key.(type) {
		case *ast.Ident:
			if _, ok := depFields.KeyLookup[keyType.Name]; ok {
				return true
			}
		}
	default:
		pepperlint.Log("TODO: isDeprecatedField %T\n", t)
	}

	return false
}

func (r DeprecatedFieldRule) isSelectorExprDeprecated(node ast.Node, lhsName string, e ast.Expr) []error {
	errs := []error{}
	t, decl, es := r.getDeclFromSelectorExpr(e)

	if len(es) > 0 {
		errs = append(errs, es...)
	}

	if decl == nil {
		return errs
	}

	lhsIndex := -1
	// Grab the proper lhs to rhs. This is done by
	// find the lhs name and map that to an index
	for i, lhs := range decl.Lhs {
		lhsIdent, ok := lhs.(*ast.Ident)
		if !ok {
			continue
		}

		if lhsIdent.Name == lhsName {
			lhsIndex = i
			break
		}
	}

	if lhsIndex == -1 {
		return errs
	}

	// this occurs when there are more lhs then rhs.
	//	* k, v := range m
	//	* v, ok := m[key]
	//	* v, ok := iface.(objType)
	if lhsIndex >= len(decl.Rhs) {
		lhsIndex = len(decl.Rhs) - 1
	}

	switch rhs := decl.Rhs[lhsIndex].(type) {
	case *ast.CompositeLit:
		if es := r.containsDeprecatedFields(node, rhs); len(es) > 0 {
			errs = append(errs, es...)
		}
	default:
		pepperlint.Log("TODO: LHS DeprecatedFieldRule.ValidateAssignStmt %T\n", t)
	}

	return errs
}
func (r DeprecatedFieldRule) checkDeprecatedFieldUse(t *ast.SelectorExpr, field *ast.Field) []error {
	errs := []error{}

	s := utils.GetStructType(field.Type)
	if s == nil {
		return nil
	}

	depFields := getDeprecatedFields(s.Fields)
	// Check the selector name, which is struct.field = somevalue, if it is
	// in the deprecated field container.
	if _, ok := depFields.KeyLookup[t.Sel.Name]; ok {
		errs = append(errs, pepperlint.NewErrorWrap(r.fset, t.Sel, fmt.Sprintf("deprecated %q field usage", t.Sel.Name)))
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

func (r DeprecatedFieldRule) getDeclFromSelectorExpr(e ast.Expr) (*ast.SelectorExpr, *ast.AssignStmt, []error) {
	t, ok := e.(*ast.SelectorExpr)
	if !ok {
		return nil, nil, nil
	}

	expr, ok := t.X.(*ast.Ident)
	if !ok ||
		expr.Obj == nil ||
		expr.Obj.Decl == nil {

		return t, nil, nil
	}

	// This occurrs when field assignment occurrs on an object
	if field, found := expr.Obj.Decl.(*ast.Field); found {
		return t, nil, r.checkDeprecatedFieldUse(t, field)
	}

	// Get object declaration to get struct definition. This will contain
	// the struct information such as which fields are deprecated.
	decl, ok := expr.Obj.Decl.(*ast.AssignStmt)
	if !ok {
		return t, nil, nil
	}

	return t, decl, nil
}

func (r DeprecatedFieldRule) isReturnSelectorExprDeprecated(fset *token.FileSet, e ast.Expr) []error {
	errs := []error{}

	switch t := e.(type) {
	case *ast.SelectorExpr:
		switch x := t.X.(type) {
		case *ast.Ident:
			if x.Obj == nil || x.Obj.Decl == nil {
				return errs
			}

			switch decl := x.Obj.Decl.(type) {
			case *ast.AssignStmt:
				if es := r.isAssignStmtDeprecated(t.Sel, t.Sel.Name, decl); len(es) > 0 {
					errs = append(errs, es...)
				}

			// Occurs when using imported package structure field
			case *ast.Field:
				field, ok := x.Obj.Decl.(*ast.Field)
				if !ok {
					return errs
				}

				// gets the package.structname
				selExpr, ok := field.Type.(*ast.SelectorExpr)
				if !ok {
					return errs
				}

				// package name identifier
				pkgIdent, ok := selExpr.X.(*ast.Ident)
				if !ok {
					return errs
				}

				// get struct info from the cache
				info := pepperlint.PackagesCache.TypeInfos[pkgIdent.Name][selExpr.Sel.Name]

				depField := utils.GetFieldByName(t.Sel.Name, info.Spec)
				if depField == nil {
					return errs
				}

				if hasDeprecatedComment(depField.Doc) {
					errs = append(errs, pepperlint.NewErrorWrap(r.fset, t.Sel, fmt.Sprintf("deprecated %q field usage", t.Sel.Name)))
				}
			}
		}
	}

	return errs
}

func (r DeprecatedFieldRule) isAssignStmtDeprecated(node ast.Node, fieldName string, stmt *ast.AssignStmt) []error {
	errs := []error{}

	for _, stmtRHS := range stmt.Rhs {
		switch rhs := stmtRHS.(type) {
		case *ast.CompositeLit:
			if !utils.IsStruct(rhs.Type) {
				continue
			}

			s := utils.GetStructType(rhs.Type)
			if s == nil {
				continue
			}

			depFields := getDeprecatedFields(s.Fields)
			// Check the selector name, which is struct.field = somevalue, if it is
			// in the deprecated field container.
			if _, ok := depFields.KeyLookup[fieldName]; ok {
				errs = append(errs, pepperlint.NewErrorWrap(r.fset, node, fmt.Sprintf("deprecated %q field usage", fieldName)))
			}
		default:
			pepperlint.Log("TODO: LHS DeprecatedFieldRule.ValidateAssignStmt %T\n", rhs)
		}
	}

	return errs
}

// ValidateAssignStmt will check to see if a deprecated field is being set or used.
func (r DeprecatedFieldRule) ValidateAssignStmt(stmt *ast.AssignStmt) error {
	batchError := pepperlint.NewBatchError()

	// checks LHS for any selector expressions. If the struct field that
	// is being used is deprecated, the error will be added to the batch
	// error.
	for _, lhs := range stmt.Lhs {
		lhsType, ok := lhs.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		ident, ok := lhsType.X.(*ast.Ident)
		if !ok {
			continue
		}

		if errs := r.isSelectorExprDeprecated(ident, ident.Name, lhs); len(errs) > 0 {
			batchError.Add(errs...)
		}
	}

	// check to see if any struct initialized objects contain any deprecated field
	// being used
	for _, rhs := range stmt.Rhs {
		switch t := rhs.(type) {
		case *ast.CompositeLit:
			if errs := r.containsDeprecatedFields(nil, t); len(errs) > 0 {
				batchError.Add(errs...)
			}
		default:
			pepperlint.Log("TODO: RHS DeprecatedFieldRule.ValidateAssignStmt %T %v\n", t, t)
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
			ident, ok := t.X.(*ast.Ident)
			if !ok {
				continue
			}

			if errs := r.isSelectorExprDeprecated(t.Sel, ident.Name, t); len(errs) > 0 {
				batchError.Add(errs...)
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
		if errs := r.isReturnSelectorExprDeprecated(r.fset, result); len(errs) > 0 {
			batchError.Add(errs...)
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

// AddRules will add the DeprecatedFieldRule to the given visitor
func (r *DeprecatedFieldRule) AddRules(v *pepperlint.Visitor) {
	rules := pepperlint.Rules{
		AssignStmtRules: pepperlint.AssignStmtRules{r},
		CallExprRules:   pepperlint.CallExprRules{r},
		PackageRules:    pepperlint.PackageRules{r},
		ReturnStmtRules: pepperlint.ReturnStmtRules{r},
	}

	v.Rules.Merge(rules)
}
