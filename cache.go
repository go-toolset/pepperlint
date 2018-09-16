package pepperlint

import (
	"go/ast"
)

// PackagesCache contains all type information in a given cache.
// This allows packages to inspect other packages types
var PackagesCache = &Cache{
	TypeInfos: TypeInfos{},
	OpInfos:   OpInfos{},
}

// Cache defintion that contain type information per package.
type Cache struct {
	TypeInfos TypeInfos
	OpInfos   OpInfos

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
	Doc     *ast.CommentGroup
	Spec    *ast.TypeSpec
	PkgName string
}

// OpInfos maintains package operation information
type OpInfos map[string]PackageOpInfos

// PackageOpInfos is a map of key operation name and OpInfo containing
// relevant per package operation declarations.
type PackageOpInfos map[string]OpInfo

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
		if _, ok := c.OpInfos[c.currentPkgName]; !ok {
			c.OpInfos[c.currentPkgName] = PackageOpInfos{}
		}

		method := false
		opInfo := OpInfo{
			Decl:    t,
			PkgName: c.currentPkgName,
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

		c.OpInfos[c.currentPkgName][t.Name.Name] = opInfo
	case *ast.Package:
		c.currentPkgName = t.Name
	case *ast.GenDecl:
		for _, spec := range t.Specs {
			switch spec := spec.(type) {
			case *ast.TypeSpec:
				if _, ok := c.TypeInfos[c.currentPkgName]; !ok {
					c.TypeInfos[c.currentPkgName] = PackageTypeInfos{}
				}

				c.TypeInfos[c.currentPkgName][spec.Name.Name] = TypeInfo{
					t.Doc,
					spec,
					c.currentPkgName,
				}
			}
		}
	}

	return c
}
