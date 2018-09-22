package pepperlint

import (
	"go/ast"
	"log"
	"os"
	"path/filepath"
	"strings"
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

		pkg, ok := PackagesCache.Packages.Get(ident.Name)
		if !ok {
			return false
		}

		info, ok := pkg.Files.GetTypeInfo(t.Sel.Name)
		if !ok {
			return false
		}
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

		pkg, ok := PackagesCache.Packages.Get(ident.Name)
		if !ok {
			return nil
		}

		info, ok := pkg.Files.GetTypeInfo(t.Sel.Name)
		if !ok || info.Spec == nil {
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

		file, ok := PackagesCache.CurrentFile()
		if !ok {
			return nil
		}

		pkgImportPath := file.Imports[ident.Name]
		pkg, ok := PackagesCache.Packages.Get(pkgImportPath)
		if !ok {
			return nil
		}

		info, ok := pkg.Files.GetTypeInfo(t.Sel.Name)
		if !ok {
			return nil
		}

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

		pkg, ok := PackagesCache.Packages.Get(ident.Name)
		if !ok {
			return ""
		}

		info, ok := pkg.Files.GetTypeInfo(t.Sel.Name)
		if !ok || info.Spec == nil {
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

// GetImportPathFromFullPath will return the import path from the given full path
func GetImportPathFromFullPath(path string) string {
	gopath := os.Getenv("GOPATH")
	// strip of the gopath
	// TODO: Should eventually iterate through multiple gopaths if provided
	// ie, GOPATH=foo/bar:bar/baz
	if strings.HasPrefix(path, gopath) {
		path = path[len(gopath):]
	}

	// strip off src. What should be left is
	// import/path/to/pkg
	if src := "/src/"; strings.HasPrefix(path, src) {
		path = path[len(src):]
	}

	return path
}

// GetPackageNameFromImportPath ...
func GetPackageNameFromImportPath(importPath string) string {
	return filepath.Base(importPath)
}
