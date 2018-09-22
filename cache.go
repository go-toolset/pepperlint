package pepperlint

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"strconv"
)

// PackagesCache contains all type information in a given cache.
// This allows packages to inspect other packages types
var PackagesCache = &Cache{
	Packages: Packages{},
}

// Cache defintion that contain type information per package.
type Cache struct {
	Packages             Packages
	currentPkgImportPath string
	currentFile          *ast.File
}

// CurrentPackage will attempt to return the current package. If currentPkgImportPath
// was not found in the map, then false will be returned.
func (c Cache) CurrentPackage() (*Package, bool) {
	v, ok := c.Packages[c.currentPkgImportPath]
	return v, ok
}

// CurrentFile will attempt to return the current file that is being visited. If the file
// could not be found, then false will be returned.
func (c Cache) CurrentFile() (*File, bool) {
	pkg, ok := c.CurrentPackage()
	if !ok {
		return nil, false
	}

	for _, f := range pkg.Files {
		if f.ASTFile == c.currentFile {
			return f, true
		}
	}

	return nil, false
}

// Packages is a map of Packages that keyed off of the import path.
type Packages map[string]*Package

// Get attempts to grab the package by name and return it if it exists.
func (p Packages) Get(pkg string) (*Package, bool) {
	v, ok := p[pkg]
	return v, ok
}

// Package is a container for a list of files
type Package struct {
	Name  string
	Files Files
}

// Files is a list of Files
type Files []*File

// GetTypeInfo will iterate through all the files in an attempt to grab the
// specified type info by type name provided.
func (fs Files) GetTypeInfo(typeName string) (TypeInfo, bool) {
	for _, f := range fs {
		info, ok := f.TypeInfos[typeName]
		if !ok {
			continue
		}

		return info, true
	}

	return TypeInfo{}, false
}

// GetOpInfo will iterate through all the files in an atttempt to grab the specified
// op info by the op name provided.
func (fs Files) GetOpInfo(opName string) (OpInfo, bool) {
	for _, f := range fs {
		info, ok := f.OpInfos[opName]
		if !ok {
			continue
		}

		return info, true
	}

	return OpInfo{}, false
}

// File contains the file scope of types and operation infos
type File struct {
	ASTFile   *ast.File
	TypeInfos TypeInfos
	OpInfos   OpInfos

	Imports map[string]string
}

// NewFile will return a new file with any instantiation that needs to be
// done.
func NewFile(t *ast.File) *File {
	return &File{
		ASTFile:   t,
		TypeInfos: TypeInfos{},
		OpInfos:   OpInfos{},
		Imports:   map[string]string{},
	}
}

// TypeInfos represents a map of TypeInfos
type TypeInfos map[string]TypeInfo

// TypeInfo contains the type definition and documentation tied to that definition.
type TypeInfo struct {
	Doc     *ast.CommentGroup
	Spec    *ast.TypeSpec
	PkgName string
}

// OpInfos is a map of key operation name and OpInfo containing
// relevant per package operation declarations.
type OpInfos map[string]OpInfo

// OpInfo signifies an operation which is a method or function.
type OpInfo struct {
	IsMethod  bool
	TypeSpecs []*ast.TypeSpec
	Decl      *ast.FuncDecl
	PkgName   string
}

// HasReceiverType will return true if the method has a receiver of type
// spec.
func (oi OpInfo) HasReceiverType(spec *ast.TypeSpec) bool {
	if !oi.IsMethod {
		return false
	}

	for _, s := range oi.TypeSpecs {
		if s == spec {
			return true
		}
	}

	return false
}

func (c *Cache) getTypeFromField(expr ast.Expr) (*ast.TypeSpec, bool) {
	switch fieldType := expr.(type) {
	case *ast.Ident:
		if fieldType.Obj == nil {
			return nil, false
		}

		if fieldType.Obj.Decl == nil {
			return nil, false
		}

		spec, ok := fieldType.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return nil, false
		}

		return spec, true
	case *ast.StarExpr:
		return c.getTypeFromField(fieldType.X)
	default:
		Log("OPPOPOPOO %T %v", fieldType, fieldType)
	}

	return nil, false
}

// Visit will cache all specifications and docs
func (c *Cache) Visit(node ast.Node) ast.Visitor {
	switch t := node.(type) {
	case *ast.FuncDecl:
		pkg, ok := c.Packages.Get(c.currentPkgImportPath)

		if !ok {
			pkg = &Package{}
			c.Packages[c.currentPkgImportPath] = pkg
		}

		// Get most recent file
		f := pkg.Files[len(pkg.Files)-1]

		method := false
		opInfo := OpInfo{
			Decl:    t,
			PkgName: pkg.Name,
		}

		// If receiver is nil, this implies that this is a function and
		// not a method.
		if t.Recv != nil {

			// methods can have multiple receivers it looks like.
			for _, field := range t.Recv.List {
				method = true
				spec, ok := c.getTypeFromField(field.Type)
				if !ok {
					continue
				}

				opInfo.TypeSpecs = append(opInfo.TypeSpecs, spec)
			}

			opInfo.IsMethod = method
		}

		f.OpInfos[t.Name.Name] = opInfo

	case *ast.ImportSpec:
		pkg, ok := c.CurrentPackage()
		if !ok {
			panic("Current package could not be found")
		}

		importPath, err := strconv.Unquote(t.Path.Value)
		if err != nil {
			panic(err)
		}

		f := pkg.Files[len(pkg.Files)-1]
		if t.Name != nil {
			name, err := strconv.Unquote(t.Name.Name)
			if err != nil {
				panic(err)
			}

			f.Imports[name] = importPath
		} else {
			f.Imports[GetPackageNameFromImportPath(importPath)] = importPath
		}

	case *ast.Package:
		// iterate through files to get the full path of the file
		for k := range t.Files {
			c.currentPkgImportPath = GetImportPathFromFullPath(filepath.Dir(k))
			break
		}

		c.Packages[c.currentPkgImportPath] = &Package{
			Name: GetPackageNameFromImportPath(c.currentPkgImportPath),
		}
	case *ast.File:
		pkg, ok := c.Packages.Get(c.currentPkgImportPath)
		if !ok {
			panic(fmt.Errorf("package import path not found: %q", c.currentPkgImportPath))
		}

		pkg.Files = append(pkg.Files, NewFile(t))
	case *ast.GenDecl:
		pkg, ok := c.Packages.Get(c.currentPkgImportPath)
		if !ok {
			panic(fmt.Errorf("package import path not found: %q", c.currentPkgImportPath))
		}

		f := pkg.Files[len(pkg.Files)-1]
		pkgName := GetPackageNameFromImportPath(c.currentPkgImportPath)

		for _, spec := range t.Specs {
			switch spec := spec.(type) {
			case *ast.TypeSpec:
				f.TypeInfos[spec.Name.Name] = TypeInfo{
					t.Doc,
					spec,
					pkgName,
				}
			}
		}
	}

	return c
}
