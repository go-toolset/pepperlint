package pepperlint

import (
	"go/ast"
)

// PackagesCache contains all type information in a given cache.
// This allows packages to inspect other packages types
var PackagesCache = &Cache{
	TypeInfos: TypeInfos{},
}

// Cache defintion that contain type information per package.
type Cache struct {
	TypeInfos TypeInfos

	currentPkgName string
}

// TypeInfos represents a map of TypeInfos
type TypeInfos map[string]PackageTypeInfos

// Get will return the TypeInfos of a given package
func (infos TypeInfos) Get(packagename string) (PackageTypeInfos, bool) {
	v, ok := infos[packagename]
	return v, ok
}

// PackageTypeInfos groups TypeInfo by package
type PackageTypeInfos map[string]TypeInfo

// Get will return the type info based on the struct name.
func (p PackageTypeInfos) Get(structname string) (TypeInfo, bool) {
	v, ok := p[structname]
	return v, ok
}

// TypeInfo contains the type definition and documentation tied to that definition.
type TypeInfo struct {
	Doc  *ast.CommentGroup
	Spec *ast.TypeSpec
}

// Visit will cache all specifications and docs
func (c *Cache) Visit(node ast.Node) ast.Visitor {
	decl, ok := node.(*ast.GenDecl)
	if !ok {
		if pkg, ok := node.(*ast.Package); ok {
			c.currentPkgName = pkg.Name
		}

		return c
	}

	for _, spec := range decl.Specs {
		switch t := spec.(type) {
		case *ast.TypeSpec:
			if _, ok := c.TypeInfos[c.currentPkgName]; !ok {
				c.TypeInfos[c.currentPkgName] = PackageTypeInfos{}
			}

			c.TypeInfos[c.currentPkgName][t.Name.Name] = TypeInfo{
				decl.Doc,
				t,
			}
		}
	}

	return c
}
