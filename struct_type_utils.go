package pepperlint

import (
	"go/ast"
	"log"
)

// IsStruct will return whether or not an ast.Expr is a
// struct type.
func IsStruct(expr ast.Expr) bool {
	switch t := expr.(type) {
	// Chcek if it is a selector expression, meaning it potentially
	// could be an imported shape
	case *ast.SelectorExpr:
		ident, ok := t.X.(*ast.Ident)
		if !ok {
			return false
		}

		info := PackagesCache.TypeInfos[ident.Name][t.Sel.Name]
		if info.Spec == nil {
			return false
		}

		return IsStruct(info.Spec.Type)
	case *ast.StructType:
		return true
	case *ast.Ident:
		if t.Obj == nil {
			return false
		}

		if t.Obj.Decl == nil {
			return false
		}

		decl := t.Obj.Decl
		switch d := decl.(type) {
		case *ast.TypeSpec:
			return IsStruct(d.Type)
		}
	case *ast.StarExpr:
		return IsStruct(t.X)
	}

	return false
}

// GetStructType will return a struct from the given expr.
func GetStructType(expr ast.Expr) *ast.StructType {
	switch t := expr.(type) {
	// Chcek if it is a selector expression, meaning it potentially
	// could be an imported shape
	case *ast.SelectorExpr:
		ident, ok := t.X.(*ast.Ident)
		if !ok {
			return nil
		}

		info := PackagesCache.TypeInfos[ident.Name][t.Sel.Name]
		if info.Spec == nil {
			return nil
		}

		return GetStructType(info.Spec.Type)
	case *ast.StructType:
		return t
	case *ast.Ident:
		if t.Obj == nil {
			return nil
		}

		if t.Obj.Decl == nil {
			return nil
		}

		decl := t.Obj.Decl
		switch d := decl.(type) {
		case *ast.TypeSpec:
			return GetStructType(d.Type)
		}
	case *ast.StarExpr:
		return GetStructType(t.X)
	}

	return nil
}

// GetTypeSpec will return the given type spec for an expression. nil will
// be returned if one could not be found
func GetTypeSpec(expr ast.Expr) *ast.TypeSpec {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		ident, ok := t.X.(*ast.Ident)
		if !ok {
			return nil
		}

		info := PackagesCache.TypeInfos[ident.Name][t.Sel.Name]
		return info.Spec
	case *ast.Ident:
		if t.Obj == nil {
			return nil
		}

		if t.Obj.Decl == nil {
			return nil
		}

		decl := t.Obj.Decl
		switch d := decl.(type) {
		case *ast.TypeSpec:
			return d
		}
	case *ast.StarExpr:
		return GetTypeSpec(t.X)
	}

	return nil
}

// GetStructName will return a struct from the given expr.
func GetStructName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		ident, ok := t.X.(*ast.Ident)
		if !ok {
			return ""
		}

		info := PackagesCache.TypeInfos[ident.Name][t.Sel.Name]
		if info.Spec == nil {
			return ""
		}

		return info.Spec.Name.Name
	case *ast.Ident:
		if t.Obj == nil {
			return ""
		}

		if t.Obj.Decl == nil {
			return ""
		}

		decl := t.Obj.Decl
		switch d := decl.(type) {
		case *ast.TypeSpec:
			return d.Name.Name
		}
	case *ast.StarExpr:
		return GetStructName(t.X)
	}

	return ""
}

// IsMethod will return whether or not something is a method.
func IsMethod(expr ast.Decl) bool {
	switch t := expr.(type) {
	case *ast.FuncDecl:
		if t.Recv != nil {
			return true
		}
	}

	return false
}

// GetOpFromType ...
func GetOpFromType(spec *ast.TypeSpec, name string) *ast.CallExpr {
	switch t := spec.Type.(type) {
	case *ast.StructType:
		for _, field := range t.Fields.List {
			if field.Type == nil {
				continue
			}

			callExpr, ok := field.Type.(*ast.CallExpr)
			if !ok {
				continue
			}

			for _, n := range field.Names {
				if name == n.Name {
					return callExpr
				}
			}
		}
	default:
		log.Printf("TODO: GetOpFromType %T", t)
	}

	return nil
}

// GetFieldByName will retrieve the ast.Field off of an ast.TypeSpec and
// field name.
func GetFieldByName(fieldName string, spec *ast.TypeSpec) *ast.Field {
	if !IsStruct(spec.Type) {
		return nil
	}

	sType := GetStructType(spec.Type)
	for _, depField := range sType.Fields.List {
		for _, name := range depField.Names {
			if name.Name == fieldName {
				return depField
			}
		}
	}

	return nil
}

// GetFieldByIndex will retrieve the ast.Field off of an ast.TypeSpec and
// field index.
func GetFieldByIndex(index int, spec *ast.TypeSpec) *ast.Field {
	if !IsStruct(spec.Type) {
		return nil
	}

	sType := GetStructType(spec.Type)
	if index < 0 || index >= len(sType.Fields.List) {
		return nil
	}

	return sType.Fields.List[index]
}
